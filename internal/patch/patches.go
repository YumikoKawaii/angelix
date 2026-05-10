package patch

import (
	"bytes"
	"fmt"
)

// rateLimitAnchor is a stable property access in the rate-limit claims object.
const rateLimitAnchor = `.five_hour?.utilization`

// patchJS applies all JS instrumentation patches to the cli.js code region and
// returns the modified code bytes with the injection inserted inside the CJS wrapper.
//
// The cli.js module is wrapped in a CJS function:
//   (function(exports, require, module, __filename, __dirname) { ... })\n
// We inject just before the closing '})\n' so the IIFE runs in the same scope
// as L48 and K0H, and the module still ends with the expected wrapper pattern.
func patchJS(cliCode []byte) ([]byte, error) {
	claimsVar, err := extractClaimsVar(cliCode)
	if err != nil {
		return nil, err
	}

	rlFunc, err := extractRlFunc(cliCode)
	if err != nil {
		return nil, err
	}

	successHandler, err := extractSuccessHandler(cliCode, rlFunc)
	if err != nil {
		return nil, err
	}

	injection := buildInjection(claimsVar, successHandler)

	// The module code ends with '}\n)\n' or '})\n' — the closing of the CJS wrapper.
	// Inject just before the final '}' so we remain inside the wrapper's scope.
	const suffix = "})\n"
	if len(cliCode) < len(suffix) || string(cliCode[len(cliCode)-len(suffix):]) != suffix {
		return nil, fmt.Errorf("cli.js does not end with expected %q suffix (got %q)", suffix, cliCode[max(0, len(cliCode)-10):])
	}

	var buf bytes.Buffer
	buf.Write(cliCode[:len(cliCode)-len(suffix)])
	buf.WriteString(injection)
	buf.Write(cliCode[len(cliCode)-len(suffix):])
	return buf.Bytes(), nil
}

// extractClaimsVar returns the identifier used as the rate-limit claims object.
// It finds the identifier immediately before `.five_hour?.utilization`.
func extractClaimsVar(js []byte) (string, error) {
	anchorPos := bytes.Index(js, []byte(rateLimitAnchor))
	if anchorPos < 0 {
		return "", fmt.Errorf("rate-limit anchor (.five_hour?.utilization) not found — unsupported claude version")
	}
	if anchorPos == 0 {
		return "", fmt.Errorf("rate-limit anchor at start of bundle — unexpected")
	}

	name := identBefore(js, anchorPos)
	if name == "" {
		return "", fmt.Errorf("could not extract claims variable name from rate-limit anchor (byte before: 0x%02x)", js[anchorPos-1])
	}
	return name, nil
}

// extractRlFunc returns the name of the 429-error handler (e.g. Bo_).
// It scans backward from `.five_hour?.utilization` to find the nearest enclosing
// named function declaration.
func extractRlFunc(js []byte) (string, error) {
	anchorPos := bytes.Index(js, []byte(rateLimitAnchor))
	if anchorPos < 0 {
		return "", fmt.Errorf("rate-limit anchor not found")
	}

	window := js[max(0, anchorPos-8192):anchorPos]
	funcKw := []byte("function ")
	off := len(window)
	for off > 0 {
		prevOff := bytes.LastIndex(window[:off], funcKw)
		if prevOff < 0 {
			break
		}
		afterFunc := window[prevOff+len(funcKw):]
		name := identStart(afterFunc)
		if name != "" {
			return name, nil
		}
		off = prevOff
	}

	return "", fmt.Errorf("could not find enclosing function name for rate-limit anchor")
}

// extractSuccessHandler returns the name of the function that is called on every
// successful API response and updates the rate-limit claims object.
// It is the named function declared immediately before the 429-error handler.
func extractSuccessHandler(js []byte, rlFunc string) (string, error) {
	rlFuncKw := []byte("function " + rlFunc + "(")
	rlPos := bytes.Index(js, rlFuncKw)
	if rlPos < 0 {
		return "", fmt.Errorf("could not locate function %s in bundle", rlFunc)
	}

	window := js[max(0, rlPos-1024):rlPos]
	funcKw := []byte("function ")
	off := len(window)
	for off > 0 {
		prevOff := bytes.LastIndex(window[:off], funcKw)
		if prevOff < 0 {
			break
		}
		afterFunc := window[prevOff+len(funcKw):]
		name := identStart(afterFunc)
		if name != "" && name != rlFunc {
			return name, nil
		}
		off = prevOff
	}

	return "", fmt.Errorf("could not find success handler before %s", rlFunc)
}

// buildInjection constructs the JS IIFE injected inside the cli.js CJS wrapper.
// On every successful API response it POSTs a minimal OTEL trace span carrying
// five_hour_utilization and seven_day_utilization to the angelix server.
func buildInjection(claimsVar, successHandler string) string {
	// _ax_ep  — OTLP endpoint base URL (e.g. http://host/otel)
	// _ax_hdr — raw "Authorization=Bearer <token>" header string
	// _ax_hex — generate n random hex bytes
	// _ax_emit — POST a rate_limit span; fires-and-forgets, never throws
	return `;(function(){` +
		`var _ax_ep=process.env.OTEL_EXPORTER_OTLP_ENDPOINT||'';` +
		`var _ax_raw=process.env.OTEL_EXPORTER_OTLP_HEADERS||'';` +
		`var _ax_m=_ax_raw.match(/Authorization=(.+)/i);` +
		`var _ax_auth=_ax_m?_ax_m[1]:'';` +
		`function _ax_hex(n){var h='';for(var i=0;i<n;i++)h+=Math.floor(Math.random()*256).toString(16).padStart(2,'0');return h;}` +
		`function _ax_emit(cv){` +
		`if(!_ax_ep)return;` +
		`var t=Date.now();` +
		`var ns=String(t)+'000000';` +
		`var fh=(cv&&cv.five_hour&&cv.five_hour.utilization)||0;` +
		`var sd=(cv&&cv.seven_day&&cv.seven_day.utilization)||0;` +
		`fetch(_ax_ep+'/v1/traces',{` +
		`method:'POST',` +
		`headers:{'Content-Type':'application/json','Authorization':_ax_auth},` +
		`body:JSON.stringify({resourceSpans:[{scopeSpans:[{spans:[{` +
		`traceId:_ax_hex(16),spanId:_ax_hex(8),` +
		`name:'angelix.rate_limit',` +
		`startTimeUnixNano:ns,endTimeUnixNano:ns,` +
		`attributes:[` +
		`{key:'five_hour_utilization',value:{doubleValue:fh}},` +
		`{key:'seven_day_utilization',value:{doubleValue:sd}}` +
		`]}]}]}]})` +
		`}).catch(function(){});` +
		`}` +
		`try{` +
		`var _ax_orig=` + successHandler + `;` +
		successHandler + `=function(H){var r=_ax_orig(H);_ax_emit(` + claimsVar + `);return r;};` +
		`}catch(_ax_e){process.stderr.write('[angelix] wrap err: '+_ax_e+'\n');}` +
		`}());`
}

// identStart returns the leading identifier from bs, or "" if none.
func identStart(bs []byte) string {
	if len(bs) == 0 || !isIdentStart(bs[0]) {
		return ""
	}
	end := 1
	for end < len(bs) && isIdentChar(bs[end]) {
		end++
	}
	return string(bs[:end])
}

// identBefore returns the identifier that ends at position pos in js.
func identBefore(js []byte, pos int) string {
	end := pos
	start := end - 1
	for start >= 0 && isIdentChar(js[start]) {
		start--
	}
	start++
	if start >= end {
		return ""
	}
	return string(js[start:end])
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '$'
}

func isIdentChar(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

