#!/usr/bin/env python3
"""Downloads the claude binary from npm into ~/.angelix/bin/claude.

Usage:
    python3 scripts/download-claude.py [version]

If version is omitted, the latest published version is used.
"""
import io
import json
import os
import platform
import shutil
import ssl
import sys
import tarfile
import urllib.request

REGISTRY = "https://registry.npmjs.org"


def _ssl_context() -> ssl.SSLContext:
    """Return an SSL context with trusted CA certificates.

    On macOS the standalone Python installer ships without the Apple root CAs,
    so urllib fails with CERTIFICATE_VERIFY_FAILED.  We try certifi first, then
    the default context, then (last resort) the macOS system keychain bundle.
    """
    try:
        import certifi
        return ssl.create_default_context(cafile=certifi.where())
    except ImportError:
        pass

    ctx = ssl.create_default_context()
    if sys.platform == "darwin":
        # macOS system root CA bundle — works when certifi is absent.
        system_bundle = "/etc/ssl/cert.pem"
        if os.path.exists(system_bundle):
            ctx.load_verify_locations(system_bundle)
    return ctx


def _urlopen(url: str):
    req = urllib.request.Request(url, headers={"User-Agent": "angelix-setup/1.0"})
    return urllib.request.urlopen(req, context=_ssl_context())


def platform_suffix() -> str:
    os_ = sys.platform
    arch = platform.machine().lower()
    if os_ == "darwin":
        if arch == "arm64":
            return "darwin-arm64"
        if arch in ("x86_64", "amd64"):
            return "darwin-x64"
    elif os_.startswith("linux"):
        musl = _is_musl()
        suffix = "-musl" if musl else ""
        if arch in ("aarch64", "arm64"):
            return f"linux-arm64{suffix}"
        if arch in ("x86_64", "amd64"):
            return f"linux-x64{suffix}"
    raise RuntimeError(f"Unsupported platform: {os_}/{arch}")


def _is_musl() -> bool:
    try:
        with open("/proc/version") as f:
            return "musl" in f.read().lower()
    except OSError:
        return False


def latest_version() -> str:
    url = f"{REGISTRY}/@anthropic-ai/claude-code/latest"
    with _urlopen(url) as resp:
        meta = json.load(resp)
    version = meta.get("version", "")
    if not version:
        raise RuntimeError("Empty version in npm metadata")
    return version


def download_claude(version: str, suffix: str) -> bytes:
    pkg = f"@anthropic-ai/claude-code-{suffix}"
    filename = f"claude-code-{suffix}-{version}.tgz"
    url = f"{REGISTRY}/{pkg}/-/{filename}"

    print(f"Downloading claude {version} for {suffix}...")
    with _urlopen(url) as resp:
        data = resp.read()

    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as tf:
        member = tf.getmember("package/claude")
        f = tf.extractfile(member)
        if f is None:
            raise RuntimeError("claude entry in tarball is not a regular file")
        return f.read()


def install(binary: bytes, dest: str) -> None:
    os.makedirs(os.path.dirname(dest), mode=0o700, exist_ok=True)
    tmp = dest + ".tmp"
    try:
        with open(tmp, "wb") as f:
            f.write(binary)
        os.chmod(tmp, 0o755)
        shutil.move(tmp, dest)
    except Exception:
        try:
            os.remove(tmp)
        except OSError:
            pass
        raise


def main() -> None:
    version = sys.argv[1] if len(sys.argv) > 1 else ""

    suffix = platform_suffix()

    if not version:
        print("Fetching latest claude-code version...")
        version = latest_version()

    binary = download_claude(version, suffix)

    home = os.path.expanduser("~")
    dest = os.path.join(home, ".angelix", "bin", "claude")
    install(binary, dest)

    print(f"claude {version} installed to {dest}")


if __name__ == "__main__":
    main()
