import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
import sys

BASE=Path('C:/Users/hanse/.codex/tmp/cr21-helper-review-003')
REPO=Path('C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets')
GO='C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe'
HEAD='4d59b27fb15d5f7885c92ac98dce365063c60a5e'
def run(args,env=None):
    return subprocess.run(args,cwd=REPO,env=env,encoding='utf-8',errors='strict',stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
assert run(['git','rev-parse','HEAD']).stdout.strip()==HEAD
env=os.environ.copy()
env.update({'GOTOOLCHAIN':'local','CGO_ENABLED':'0','GOFLAGS':'-mod=readonly','GOWORK':'off'})
for key in ['GOARCH','GOOS','GOROOT']:
    env.pop(key,None)
mode=sys.argv[1]
if mode=='controls':
    overlay={'Replace':{str(REPO/'internal/runtime/cmd/helpergen/reviewer_acceptance_test.go'):str(BASE/'reviewer_acceptance_test.go.txt')}}
    (BASE/'overlay.json').write_bytes(json.dumps(overlay).encode('utf-8'))
    selectors='Test(ReviewerInputSetAcceptanceBoundary|ReviewerCompletedNameChangesCannotCancelCleanly|DirectoryChangeGuardReleaseOverflowAndRefusal|PartialDirectoryWatchAcquisitionUnwinds|DirectoryWatchReadProbe|PostSelectionInputSetChangesRefuseActualBuild|BuildConsumesCapturedModuleAndToolchain)$'
    args=[GO,'test','-overlay',str(BASE/'overlay.json'),'./internal/runtime/cmd/helpergen','-run',selectors,'-count=1','-v']
    start=time.time()
    r=run(args,env)
    (BASE/'final-controls.log').write_bytes(r.stdout.encode('utf-8'))
    result={'head':HEAD,'command':args,'exit':r.returncode,'seconds':round(time.time()-start,2),'log':'final-controls.log'}
else:
    def snapshot():
        files=[p for p in (REPO/'internal/runtime').rglob('*') if p.is_file()]+[REPO/'go.mod',REPO/'go.sum']
        return {str(p.relative_to(REPO)).replace('\\','/'):{'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'length':p.stat().st_size,'mtimeNs':p.stat().st_mtime_ns} for p in files}
    before=snapshot()
    start=time.time()
    args=[GO,'run','./internal/runtime/cmd/helpergen','-check']
    r=run(args,env)
    after=snapshot()
    result={'head':HEAD,'command':args,'exit':r.returncode,'seconds':round(time.time()-start,2),'fileSetHashesLengthsMtimesUnchanged':before==after,'files':len(before),'output':r.stdout}
(BASE/('final-'+mode+'.json')).write_bytes((json.dumps(result,indent=2)+'\n').encode('utf-8'))
print(json.dumps(result),flush=True)
if mode=='controls':
    print(r.stdout,flush=True)
