"""Independent #72 exact-source checker and native binding execution observer."""
import gzip
import hashlib
import json
import pathlib
import platform
import struct
import subprocess
import time

ROOT = pathlib.Path(__file__).resolve().parents[4]
OUT = pathlib.Path(__file__).resolve().parent
GO = r"C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe"
PREFIX = "M3-HelperBinding-Review--001-"

def digest(data):
    return hashlib.sha256(data).hexdigest()

def invoke(args):
    return subprocess.run(args, cwd=ROOT, encoding="utf-8", errors="strict", capture_output=True)

def snapshot():
    files = list((ROOT / "internal/runtime").rglob("*")) + [ROOT / "go.mod", ROOT / "go.sum"]
    return {p.relative_to(ROOT).as_posix(): [len(p.read_bytes()), digest(p.read_bytes()), p.stat().st_mtime_ns]
            for p in files if p.is_file()}

result = {"head": invoke(["git", "rev-parse", "HEAD"]).stdout.strip(), "platform": platform.platform(),
          "toolchain": invoke([GO, "version"]).stdout.strip()}
before = snapshot()
started = time.monotonic()
check = invoke([GO, "run", "./internal/runtime/cmd/helpergen", "-check"])
result["checker"] = {"exit": check.returncode, "seconds": round(time.monotonic() - started, 3)}
(OUT / (PREFIX + "checker.log")).write_text(check.stdout + check.stderr, encoding="utf-8", newline="\n")
after = snapshot()
result["noRewrite"] = {"count": len(before), "sameFilesBytesLengthsMtimes": before == after}
assert check.returncode == 0 and before == after, result

manifest_path = ROOT / "internal/runtime/brokerassets/manifest.json"
manifest_bytes = manifest_path.read_bytes()
manifest = json.loads(manifest_bytes)
result["manifest"] = {"sha256": digest(manifest_bytes), "sourceDigest": manifest["sourceDigest"],
                      "sourceCount": len(manifest["sources"]), "targets": manifest["targets"]}
repo_entries = 0
recipe_entries = 0
for source in manifest["sources"]:
    prefix, _, name = source["path"].partition("/")
    if prefix in ("repo", "recipe"):
        data = (ROOT / name).read_bytes().replace(b"\r\n", b"\n")
        assert len(data) == source["length"] and digest(data) == source["sha256"], source["path"]
        repo_entries += prefix == "repo"
        recipe_entries += prefix == "recipe"
result["repositorySourceEntriesVerified"] = repo_entries
result["recipeSourceEntriesVerified"] = recipe_entries
for target in manifest["targets"]:
    compressed = (ROOT / "internal/runtime/brokerassets" / ("broker-" + target["arch"] + ".gz")).read_bytes()
    image = gzip.decompress(compressed)
    pe_offset = struct.unpack_from("<I", image, 0x3C)[0]
    machine = struct.unpack_from("<H", image, pe_offset + 4)[0]
    assert image[pe_offset:pe_offset + 4] == b"PE\0\0" and machine == target["machine"]
    assert len(image) == target["length"] and digest(image) == target["sha256"]
    assert len(compressed) == target["compressedLength"] and digest(compressed) == target["compressedSHA256"]

started = time.monotonic()
native = invoke([GO, "test", "-race", "./internal/runtime", "-run", "^TestWindowsCommittedHelperBinding$", "-count=1", "-timeout=180s", "-v"])
result["nativeBinding"] = {"exit": native.returncode, "seconds": round(time.monotonic() - started, 3)}
(OUT / (PREFIX + "native.log")).write_text(native.stdout + native.stderr, encoding="utf-8", newline="\n")
result["sourceStillUnchanged"] = snapshot() == before
assert native.returncode == 0 and result["sourceStillUnchanged"], result
(OUT / (PREFIX + "result.json")).write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8", newline="\n")
print(json.dumps(result, indent=2))
