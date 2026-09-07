"""Bounded independent Windows review: additive test overlays, owned fixtures only."""
from pathlib import Path
import argparse
import hashlib
import json
import os
import subprocess
import tempfile

parser = argparse.ArgumentParser()
parser.add_argument("--go", required=True)
parser.add_argument("--arch", choices=["amd64", "386"], default="amd64")
parser.add_argument("--wow64-loader", action="store_true")
parser.add_argument("--race", action="store_true")
args = parser.parse_args()
evidence = Path(__file__).resolve().parent
repo = evidence.parents[3]
scratch = Path(tempfile.mkdtemp(prefix="m69-review-overlay-"))
overlay = {str(repo / "internal/runtime/broker/reviewer_controls_windows_test.go"):
           str(evidence / "M3-Windows-Review--001-controls.go.txt")}
selector = "^TestReviewWindows"
if args.wow64_loader:
    # Reuse the owned real C fixture compiler/runner while selecting x86 DLL/exe.
    # Add a new test file; no original source or test file is replaced.
    loader = (repo / "internal/runtime/broker/loader_windows_test.go").read_text(encoding="utf-8")
    loader = loader.replace("TestWindowsStaticDLLTLSAndDebugHeap", "TestReviewWindowsWOW64StaticDLLTLS")
    loader = loader.replace('architecture := map[uint16]string{machine386: "x86", machineAMD64: "x64", machineARM64: "arm64"}[machine]',
                            'architecture := "x86"')
    loader = loader.replace("image.Machine != machine", "image.Machine != machine386")
    loader = loader.replace("image.Machine, machine)", "image.Machine, machine386)")
    loader = loader.replace("machine, emulated, err := MachineRoute()", "_, emulated, err := MachineRoute()")
    derived = scratch / "reviewer_wow64_loader_windows_test.go"
    derived.write_text(loader, encoding="utf-8", newline="\n")
    overlay[str(repo / "internal/runtime/broker/reviewer_wow64_loader_windows_test.go")] = str(derived)
    selector = "^TestReviewWindowsWOW64StaticDLLTLS$"
overlay_path = scratch / "overlay.json"
overlay_path.write_text(json.dumps({"Replace": overlay}, indent=2), encoding="utf-8", newline="\n")
env = os.environ.copy()
env.update(GOARCH=args.arch, CGO_ENABLED="0")
command = [args.go, "test", "-overlay", str(overlay_path), "./internal/runtime/broker", "-run", selector, "-count=1", "-timeout=120s", "-v"]
if args.race:
    env["CGO_ENABLED"] = "1"
    command.insert(2, "-race")
result = subprocess.run(command, cwd=repo, env=env, text=True, encoding="utf-8", errors="strict", stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=180)
suffix = "wow64-loader" if args.wow64_loader else "controls-" + args.arch + ("-race" if args.race else "")
log = evidence / ("M3-Windows-Review--002-" + suffix + ".log")
source = evidence / "M3-Windows-Review--001-controls.go.txt"
metadata = {"sourceCommit": "6decc16a952dad45a07e7e35ea01a11e5df32c00", "command": command,
            "GOARCH": args.arch, "CGO_ENABLED": env["CGO_ENABLED"], "exitCode": result.returncode,
            "testSHA256": hashlib.sha256(source.read_bytes()).hexdigest()}
# Preserve evidence content while normalizing line endings/trailing whitespace.
normalized = "\n".join(line.rstrip() for line in result.stdout.splitlines()) + "\n"
log.write_text(json.dumps(metadata, indent=2) + "\n" + normalized, encoding="utf-8", newline="\n")
print(result.stdout, end="")
print("exitCode=" + str(result.returncode) + "; log=" + str(log))
raise SystemExit(result.returncode)
