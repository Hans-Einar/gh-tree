# M3 committed helper-image binding — Issue #72

Status: partial generated-asset milestone; native committed-image binding tests and independent review remain required. Worker owns only the three generated assets, exact Runtime root helper_binding_windows_test.go exception and this report. No product host cutover or Runtime/Slice acceptance.

Authority: #72 under #65/#69/#70/#71/#21, Sprint-004-v04/I-03/M3; Master e43297acbbff385315062bbaec7b5a58c79678a8 / ledger94. Clean preparation 7e19d617ae430cd2c7ad414b9ea4622eb1e00617 combines independently accepted Windows review9b06f8bd (technical bd78deafd4dd36e22d5b106eb7ef9c4edcd2e832) and helper review6f385a9c (technical b6f161f5189d70b66b95129237b51f9984d58e35). Broker tree is 04afda252be325beaa6bf1f22c154a094b8daed9. No reviewed native, generator, loader, module, workflow or frozen source was changed.

Native Windows amd64 Go1.25.0 at C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe executed `go run ./internal/runtime/cmd/helpergen`, then the exact `go run ./internal/runtime/cmd/helpergen -check`. Both succeeded, with two independent builds per target each. Checker exit0 in28.264s. An external Python3.14 UTF-8 observer verified all75 Runtime/go.mod/go.sum files retained their complete set, bytes, lengths and nanosecond mtimes across -check, and independently matched each normalized repository source entry to the manifest.

The unchanged accepted generator captures the complete selected source/assembly/include/embed/module/toolchain recipe into guarded isolated snapshots, with continuous one-shot native name invalidation through actual compiler exit and joined cancellation. CGO0, explicit microarchitecture, trimpath/buildvcsfalse/PGOoff/emptybuildid/internal-link and deterministic gzip remain intact. Ordinary build/run/install uses committed inputs and performs no helper compilation or download; there are three generated outputs and twelve public assets.

Source closure:955 inputs; digest23bf82b051123cd1aa31c5a2368d1cc732f4b09cbc33ea2c9abf4f08f0cfdde5. Manifest SHA25603c9db1f329aa29a4d5b4c279dcc2fc17ac0cd42967d7b4527c2bef5c482a3e8. No self-referential commit SHA is embedded.

| Target | Image length / SHA256 | Gzip length / SHA256 |
|---|---|---|
| amd64 0x8664 |3612160 / bfdb8eb2ec496222b8033bbeca2331c319fd1a0cfcacb6bb7adf3e79c138781c|2109197 / b62f559568d69b369ea5cf2ecef93d56f65ae0c5f3d03ab4241f7b64a447f44b|
| ARM64 0xAA64 |3422208 / 617b7b8a06b5333e146af86a63d5c93f89548f98ab9ab240e9a6600aa34dc5bf|1968006 / 54ecab0a18d9ca2bc762cb75baf8e9d6e07190c1370efdd86df41d6206b032c7|

Next permitted action: publish this coherent partial milestone, then add real committed-image native parent/extraction/client binding tests in the explicitly owned root file. Actual amd64/386 and ARM64/emulated execution, typed failure/retained receipt compatibility, clean owned teardown, exact CI/race/vet/architecture/twelve builds and separate independent binding review remain open. Earlier native ABI/loader/TLS/ACL and generator adversarial reviews are reused at their exact unchanged source; their standalone acceptance is not this binding proof. Policy-rejected residues and unrelated review/access controls remain untouched.
