#!/bin/sh
# Run only inside the disposable FreeBSD guest, as the ordinary ghci user.
set -eu
umask 077
cd /home/ghci/source
results=/home/ghci/results
mkdir -p "$results"
test "$(id -u)" -ne 0
test "$(uname -s)" = FreeBSD
test "$(uname -m)" = amd64
test "$(go env GOHOSTOS)/$(go env GOHOSTARCH)" = freebsd/amd64
test "$(go env GOOS)/$(go env GOARCH)" = freebsd/amd64
test "$(go env GOVERSION)" = go1.25.0
test "$(git rev-parse HEAD)" = "$(cat /home/ghci/expected-source)"
test -z "$(git status --porcelain --untracked-files=all)"

run_tests() {
    output=$1
    shift
    status=0
    go test -count=1 -json -timeout=10m "$@" > "$output" 2>&1 || status=$?
    # Preserve complete output as an artifact, with failures and skips visible in CI.
    awk '/"Action":"(pass|fail|skip)"/ || !/^\{/' "$output"
    printf 'go_test_exit=%s\n' "$status"
    return "$status"
}

if [ "${1-}" = --failure-probe ]; then
    probe=$(mktemp -d "$TMPDIR/failure-probe.XXXXXX")
    printf 'module failureprobe\n\ngo 1.25.0\n' > "$probe/go.mod"
    printf 'package failureprobe\nimport "testing"\nfunc TestExpectedFailure(t *testing.T) { t.Fatal("CI_EXPECTED_FAILURE_CONTROL") }\n' > "$probe/probe_test.go"
    cd "$probe"
    run_tests "$results/failure-control.jsonl" ./...
    exit 0
fi
test "$#" -eq 0

printf 'source=%s\n' "$(git rev-parse HEAD)"
printf 'source_tree=%s\n' "$(git rev-parse HEAD^{tree})"
uname -a
freebsd-version -kru
id
go version
go env GOHOSTOS GOHOSTARCH GOOS GOARCH GOVERSION CGO_ENABLED GOTOOLCHAIN
git --version
pkg query '%n %v' git-lite ca_root_nss
go list -m all
for directory in /home/ghci/source "$TMPDIR"; do
    df -T "$directory"
    fs=$(df -T "$directory" | awk 'NR == 2 { print $2 }')
    case "$fs" in ufs|zfs) ;; *) echo "Unsupported native test filesystem: $fs"; exit 1 ;; esac
    stat -f 'device=%d inode=%i mode=%Sp owner=%Su group=%Sg' "$directory"
done

# Diagnostics are evidence about this ordinary-user profile, not adapter acceptance.
profile=$(mktemp "$TMPDIR/metadata-profile.XXXXXX")
for namespace in user system; do
    status=0
    lsextattr "$namespace" "$profile" > "$results/xattr-$namespace.txt" 2>&1 || status=$?
    cat "$results/xattr-$namespace.txt"
    printf 'lsextattr_namespace=%s exit=%s\n' "$namespace" "$status"
done
status=0
getfacl "$profile" > "$results/acl.txt" 2>&1 || status=$?
cat "$results/acl.txt"
printf 'getfacl_exit=%s (diagnostic; refused profiles remain unverified)\n' "$status"

: > "$results/packages.txt"
for directory in internal/domain internal/application/api internal/application/ports internal/tuistate/viewmodel internal/composition/architecture; do
    go list "./$directory/..." >> "$results/packages.txt"
done
for directory in internal/git internal/github/adapter internal/persistence internal/launchdiscovery internal/runtime; do
    if [ -d "$directory" ]; then
        go list "./$directory/..." > "$results/current-packages.txt"
        test -s "$results/current-packages.txt"
        cat "$results/current-packages.txt" >> "$results/packages.txt"
    else
        printf 'NOT RUN: %s absent at this source; no adapter proof\n' "$directory"
    fi
done
sort -u "$results/packages.txt" > "$results/sorted-packages.txt"
mv "$results/sorted-packages.txt" "$results/packages.txt"
cat "$results/packages.txt"
echo 'Scope: named new-stack packages only; legacy suite and full V-PER/V-RUN gates are not claimed.'
# Go import paths cannot contain shell whitespace; every path comes from go list.
set -- $(cat "$results/packages.txt")
test "$#" -gt 0
run_tests "$results/tests.jsonl" "$@"
test -z "$(git status --porcelain --untracked-files=all)"
echo 'native_new_stack_tests=PASS; inspect tests.jsonl for individual skips'
