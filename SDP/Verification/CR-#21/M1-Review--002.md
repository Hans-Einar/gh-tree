# M1 corrected-source independent re-review

Disposition: CHANGES_REQUIRED. Authority: #57 under #21; I-03 / M1.
Reviewed source: b1f533e062c8ab94b81e9309aaffadee096a5c7d.
Prior review: [M1-Review--001](M1-Review--001.md), source6685b4f.
Independent reviewer: m1_composition_review, read-only and separate from author.

## Actual disposition

M157-M01 (physical selected-root normalization), M157-M02 (original nested
callbacks) and M157-M05 (trimpath/toolchain metadata) pass their required
regressions at this source. Valid private Git plumbing works, Git-root native/
callback wrapper escapes reject, and pure private values/API methods remain usable.
These local corrections do not complete M1 or resolve baseline product findings.

| ID / severity | Remaining or new actual evidence | Required correction |
|---|---|---|
| M157-M03 / MEDIUM remains | check.go:484 traverses same-role names only. Runtime returning broker.Result with *os.File or []func(), and Application returning usecases/native.Result with *os.File, all compile and receive checker PASS0. These private decompositions have distinct role tags within an allowed owner graph. | Recursively inspect boundary-reachable module values across the accepted private decomposition, preserving private native plumbing and pure/control-value positives. |
| M157-M04 / MEDIUM remains | clean.go:46 omits legacy crlf attribute. Under core.autocrlf=true and -crlf, actual clean blob f9c880865942fcf2d20e179bf941f34751b84174 differs from accepted4f8be20c2beff7db3f180a102fe146310c96a3aa, but checker passes61 allowances. +text passes; -text now correctly rejects. | Handle or explicitly refuse the actual legacy crlf profile before granting an exception; no implicit filter execution. |
| M157-M06 / MEDIUM new | check.go:289-295 compares selected-source paths lexically. A child-directory Windows junction under internal/git/native points outside the fixture and still passes; a junction used solely as the selected root correctly passes. | Bind/check physical selected-source parents/objects relative to the acquired root; preserve selected-root aliases, reject outside child source, and add actual Windows junction proof without requiring symlink privilege. |

The reviewer used an exact-source standalone checker, SHA256
25B1FF80B3F3074CA15419B7AB89089CA7A390D27B4019372BC5732167AECA32,
an owned frozen snapshot and independent compiled fixtures. Existing39 fixture
scenarios, standalone regressions, vet and formatting passed. Master read actual
corrected code plus all three result captures and both probe sources; binary and
archived-file hashes are independently checked. [Archive manifest](Evidence/M1-Review--002/ArchiveManifest.json)
preserves the exact native bytes. These are review evidence, not product code.

## Exact CI and native evidence

[CI34030099746](https://github.com/Hans-Einar/gh-tree/actions/runs/34030099746)
at b1f533e completed18 SUCCESS and one deliberate M3 helper SKIP: twelve build/
architecture jobs, four native test/vet/build jobs, Linux race and prerequisite
inventory all passed. This does not override independent CHANGES_REQUIRED.

Master inspected actual job101477921712: Windows ARM64 image win11-arm64
20260830.155.1, hostArm64, Go1.25.0 windows/arm64, Git2.55.0.windows.5,
gh2.98.0; test/vet/build actually ran. macOS job101477921735 also passed.
TestPhysicalRepositoryRoot skips symlink privilege only on Windows; Linux/macOS
errors are fatal, so their passing complete suites exercise that regression.
Native Runtime/helper/ABI and full vertical product proof remain future gates.

## Next gate

Worker is authorized to correct remaining M03/M04 and new M06, add actual
regressions, retain the passing controls and freeze a new exact source. Reviewer
reopens the complete correction and Master requires fresh applicable branch CI
before integration. No integration, completed Slice, baseline finding closure,
product PR, tag or publication follows from this green but rejected source.
