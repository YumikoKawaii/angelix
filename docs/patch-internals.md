# Angelix Binary Patch Internals

How angelix injects OTEL rate-limit telemetry into the Claude Code Bun SFE binary.

---

## Overview

Claude Code ships as a **Bun Single-File Executable (SFE)** — a Mach-O binary with a `__bun` section that contains a self-contained JS bundle. Angelix patches this binary at startup to wrap the API success handler, so every Claude API response emits a minimal OTLP trace span carrying `five_hour_utilization` and `seven_day_utilization` to the angelix server.

The patch is applied once per binary version, cached at `~/.angelix/patched/<sha256>/claude`, and re-codesigned for macOS.

---

## Bun SFE Binary Layout

```
<Mach-O headers and load commands>
  LC_SEGMENT_64 __BUN
    filesize = N   (total bytes allocated in file, including null padding)

__bun section (at file offset secOff, size = secSize):
  [secOff+0  : secOff+8]          uint64 LE bundle_length = secSize - 8
  [secOff+8  : secOff+secSize-52] JS bundle data  ← jsStart = secOff+8
  [secOff+secSize-52 : secOff+secSize]  52-byte Bun footer

Null padding (secSize → __BUN filesize):
  Available for the section to grow into without moving other data.
  v2.1.138 has ~9541 bytes of padding.
```

### JS Bundle Structure (offsets relative to jsStart)

```
[0          : ~2.5MB]       Module path strings + null bytes
[111955744  : 126242505)    cli.js module code (14,286,761 bytes of minified JS)
[126242505  : 130848838)    Other JS modules + native .node blobs
[130848838  : 130849410)    Module table (572 bytes)
[130849410  : 130849462)    52-byte Bun footer
```

### 52-byte Bun Footer Layout

```
[+0  : +4]   padding / version
[+4  : +12]  uint64 field A — internal pointer within module table area
[+12 : +16]  uint32 field B — moduleTableOff (relative to jsStart)
[+16 : +24]  uint64 field C — module table size in bytes  ← DO NOT UPDATE
[+24 : +32]  uint64 field D — moduleTableEnd (relative to jsStart)
[+32 : +36]  padding
[+36 : +52]  "\n---- Bun! ----\n" magic
```

---

## Module Table Format

The module table starts at `jsStart + moduleTableOff`. The first entry is always cli.js:

```
[0  : 4]   uint32 path_offset  — offset of path string, relative to jsStart
[4  : 8]   uint32 path_length  — byte length of path string
[8  : 12]  uint32 code_offset  — start of JS source, relative to jsStart
[12 : 16]  uint32 code_size    — byte length of JS source
```

For v2.1.138:
- path = `/$bunfs/root/src/entrypoints/cli.js`
- code_offset = 111955744
- code_size   = 14286761
- code end    = 126242505  ← injection point

Subsequent entries (52-byte native entries flagged `01 01 02 00`) cover `image-processor.js`, `audio-capture.js`, `url-handler.js`, etc.

**Critical**: Bun only executes code within byte ranges registered in the module table. Code written outside those ranges is silently never run.

---

## Injection Strategy

### What we inject

An IIFE appended to the end of cli.js code, inside its CJS wrapper:

```js
(function(exports, require, module, __filename, __dirname) {
  // ... all of cli.js — L48, K0H, Bo_, etc. declared here ...
  So3();
  ;(function(){ /* ANGELIX IIFE injected here */ }());   ← NEW
})\n
```

The entire cli.js module is wrapped in `(function(exports, require, module, __filename, __dirname) { ... })\n`. We inject our IIFE **before** the closing `})\n` so that module-scoped variables (`L48`, `K0H`) are accessible.

### Extracted names (re-derived per binary from stable anchors)

| Name | Role | Anchor |
|------|------|--------|
| `K0H` | Rate-limit claims object (holds `five_hour`, `seven_day`) | Identifier immediately before `.five_hour?.utilization` |
| `Bo_` | 429-error handler function | Nearest `function NAME(` scanning 8KB backward from anchor |
| `L48` | API success handler (updates claims) | Nearest `function NAME(` scanning 1KB before `Bo_` |

Names are extracted fresh from each binary by `extractClaimsVar`, `extractRlFunc`, `extractSuccessHandler` in `internal/patch/patches.go`. They change with every minified Claude Code release but the anchors remain stable.

### The IIFE

```js
;(function(){
  var _ax_ep   = process.env.OTEL_EXPORTER_OTLP_ENDPOINT || '';
  var _ax_raw  = process.env.OTEL_EXPORTER_OTLP_HEADERS  || '';
  var _ax_m    = _ax_raw.match(/Authorization=(.+)/i);
  var _ax_auth = _ax_m ? _ax_m[1] : '';

  function _ax_hex(n) { /* n random hex bytes */ }

  function _ax_emit(cv) {
    if (!_ax_ep) return;
    var t  = Date.now();
    var ns = String(t) + '000000';          // nanoseconds (avoid float64 precision loss)
    var fh = (cv && cv.five_hour  && cv.five_hour.utilization)  || 0;
    var sd = (cv && cv.seven_day  && cv.seven_day.utilization)  || 0;
    fetch(_ax_ep + '/v1/traces', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': _ax_auth },
      body: JSON.stringify({ resourceSpans: [{ scopeSpans: [{ spans: [{
        traceId: _ax_hex(16), spanId: _ax_hex(8),
        name: 'angelix.rate_limit',
        startTimeUnixNano: ns, endTimeUnixNano: ns,
        attributes: [
          { key: 'five_hour_utilization', value: { doubleValue: fh } },
          { key: 'seven_day_utilization', value: { doubleValue: sd } },
        ]
      }] }] }] })
    }).catch(function(){});
  }

  try {
    var _ax_orig = L48;
    L48 = function(H) { var r = _ax_orig(H); _ax_emit(K0H); return r; };
  } catch (_ax_e) {
    process.stderr.write('[angelix] wrap err: ' + _ax_e + '\n');
  }
}());
```

---

## Patch Application Algorithm (`applyPatches`)

Let `n` = bytes added (size of injected IIFE).

### 1. Data movement

```
Before:
  [cliCode][gap][otherModules][moduleTable][footer][nullPadding]
                ↑ injAbs = jsStart + cliCodeOff + cliCodeSize

After:
  [cliCode+IIFE][otherModules shifted +n][moduleTable shifted +n][footer shifted +n][nullPadding shrunk]
```

Operations in order:
1. `copy(data[footerStart+n : footerEnd+n], data[footerStart : footerEnd])` — move footer forward
2. `tmp = data[injAbs : footerStart]` → `copy(data[injAbs+n : ...], tmp)` — shift gap + other modules + module table
3. `copy(data[cliCodeAbs:], newCliCode)` — write patched cli.js (original + IIFE, no trailing gap)

### 2. Module table updates

The module table is now at `jsStart + moduleTableOff + n`. Within it:
- `mt[12:16]` (cli.js `code_size`) += n
- Every `uint32` value in range `[injPoint, moduleTableOff)` += n — these are path/code offsets for modules after cli.js

Size fields (small numbers like 35, 572) are naturally below `injPoint` (~126M) so they are not accidentally updated.

### 3. Footer updates (at new position `footerStart + n`)

- field A (uint64) += n
- field B (uint32) += n  (moduleTableOff)
- field C (uint64) — **do not update** (module table size is unchanged)
- field D (uint64) += n  (moduleTableEnd)

### 4. Header updates

- `bundle_length` (uint64 at `secOff`) += n
- Section64 `size` field in Mach-O header += n

The `__BUN` segment `filesize` is NOT updated — the null padding absorbs the growth.

---

## Cache and Codesigning

```
EnsurePatch(claudePath):
  hash = sha256(claudePath)
  dest = ~/.angelix/patched/<hash>/claude
  if exists(dest): return dest         // cache hit
  data = readFile(claudePath)
  applyPatches(data)
  writeFile(dest, data, 0755)
  if macOS: codesign --remove-signature dest; codesign -s - dest
  return dest
```

The patched binary is exec'd in place of the original. The `OTEL_EXPORTER_OTLP_ENDPOINT` and `OTEL_EXPORTER_OTLP_HEADERS` env vars are set by `internal/otel/env.go` before the exec.

---

## Known Fragility Points

1. **CJS suffix check**: `patchJS` requires cli.js to end with `})\n`. If Bun changes the module wrapper format, this fails with a clear error.
2. **Module table layout**: First 16 bytes assumed to be the cli.js entry. If Bun adds a header, module table parsing breaks.
3. **Minified name anchors**: `.five_hour?.utilization` must survive minification. If the property access is renamed or removed, `extractClaimsVar` returns an error.
4. **Scope accessibility**: `L48` and `K0H` must be top-level within the CJS wrapper. If Bun moves them into inner closures, the wrap fails at runtime — the IIFE catch block logs the error to stderr instead of crashing.
5. **Null padding**: Injection must fit within the segment's null padding. v2.1.138 has ~9541 bytes; the IIFE is ~294 bytes. If a future release has less padding, `applyPatches` returns an error.

---

## Alternative Considered: Claude's Built-in OTEL Tracer

Investigated using Claude's own OTEL SDK to emit rate-limit spans instead of the custom `fetch()` (v2.1.138). Rejected — current approach is better for this use case.

### How Claude's OTEL is structured

- `X3` = `@opentelemetry/api` module, assigned once inside a lazy init block: `X3=m(j4(),1)`
- `oN()` = the getTracer wrapper: `return X3.trace.getTracer("com.anthropic.claude_code.tracing","1.0.0")`
- A real OTLP provider is registered at startup via two paths:
  1. `BETA_TRACING_ENDPOINT` env var → Anthropic internal, irrelevant to us
  2. Enhanced telemetry path: activates only if **both** `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=true` **and** `OTEL_TRACES_EXPORTER=otlp`. Reads `OTEL_EXPORTER_OTLP_ENDPOINT` and calls `setGlobalTracerProvider(...)`.

### Why timing would be fine (proxy pattern)

`oN()` returns a `ProxyTracer` backed by `ProxyTracerProvider`. OTEL's proxy handles late provider registration — by the time `L48` fires on the first real API request, `setGlobalTracerProvider` has already run and the proxy delegates to the real OTLP exporter.

### Why we rejected this approach

1. **All of Claude's own spans flood our server.** Enabling `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=true` causes Claude to export every interaction, tool_use, and internal span to `OTEL_EXPORTER_OTLP_ENDPOINT` — not just our rate-limit spans. Significant extra volume to filter or store.
2. **Another unstable minified name.** `oN` changes every release. The anchor would be the string `"com.anthropic.claude_code.tracing"` — doable, but adds another fragility point.
3. **Current `fetch()` approach already works** — self-contained, emits only rate-limit data, no dependency on Claude's internal telemetry flag.

### When to reconsider

If parent span correlation is ever needed (linking rate-limit spans to their parent interaction span), this approach becomes worthwhile. Path to implement:
1. Set `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=true` and `OTEL_TRACES_EXPORTER=otlp` in `internal/otel/env.go`
2. Extract the getTracer wrapper name via anchor string `"com.anthropic.claude_code.tracing"`
3. Replace `_ax_emit`'s `fetch()` with a call to `oN().startSpan("angelix.rate_limit", ...)`
4. Handle the extra Claude spans in `server/otlp/parser.go` (filter or store separately)

---

## Key Files

| File | Purpose |
|------|---------|
| `internal/patch/patch.go` | `EnsurePatch`, `applyPatches`, `findBunSection`, Mach-O parsing |
| `internal/patch/patches.go` | `patchJS`, name extraction (`extractClaimsVar`, `extractRlFunc`, `extractSuccessHandler`), `buildInjection` |
| `internal/patch/patch_test.go` | Integration tests; `TestPatchedBinaryRuns` is the end-to-end validator |
| `internal/exec/claude.go` | Calls `EnsurePatch` before exec-ing claude |
| `internal/otel/env.go` | Sets OTLP env vars before exec |
| `server/otlp/parser.go` | Reads `five_hour_utilization` / `seven_day_utilization` from incoming spans |
