#!/bin/sh
# Root setup is confined to this disposable guest; all Go execution uses ghci.
set -eu
umask 077
test "$(id -u)" -eq 0
test "$(uname -s)" = FreeBSD
test "$(uname -m)" = amd64
test "$(freebsd-version -u | cut -d- -f1)" = 15.0
pw useradd ghci -m -s /bin/sh
cp -R "$GITHUB_WORKSPACE" /home/ghci/source
printf '%s\n' "$GITHUB_SHA" > /home/ghci/expected-source
mkdir /home/ghci/results /home/ghci/tmp
chown -R ghci:ghci /home/ghci
chmod 700 /home/ghci /home/ghci/tmp /home/ghci/results

# Clear the inherited action/SSH environment before running any source or tests.
command='exec env -i HOME=/home/ghci USER=ghci LOGNAME=ghci PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin LANG=C.UTF-8 TMPDIR=/home/ghci/tmp GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOWORK=off CGO_ENABLED=0 /bin/sh /home/ghci/source/.github/scripts/freebsd-tests.sh'
status=0
su -m ghci -c "$command --failure-probe" > /home/ghci/results/failure-control.log 2>&1 || status=$?
cat /home/ghci/results/failure-control.log
test "$status" -eq 1
grep -q CI_EXPECTED_FAILURE_CONTROL /home/ghci/results/failure-control.jsonl
echo 'failure_control=PASS (intentional Go failure propagated through helper and su as exit 1)'

status=0
su -m ghci -c "$command" > /home/ghci/results/native.log 2>&1 || status=$?
cat /home/ghci/results/native.log
printf 'native_suite_exit=%s\n' "$status"
exit "$status"
