import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

BASE=Path('C:/Users/hanse/.codex/tmp/cr21-helper-review-002')
REPO=Path('C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets')
SDK=Path('C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64')
HEAD='c5f36edfcfd2e418f28e32fc0613143ed31649c5'
def run(args,env=None):
    return subprocess.run(args,cwd=REPO,env=env,encoding='utf-8',errors='strict',stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
assert run(['git','rev-parse','HEAD']).stdout.strip()==HEAD
alias=BASE/'canonical-go-alias'
assert not os.path.lexists(alias)
create=run(['powershell.exe','-NoProfile','-NonInteractive','-Command',f"$null=New-Item -ItemType Junction -Path '{alias}' -Target '{SDK}'"])
assert create.returncode==0,create.stdout
cache=BASE/'fresh-exact-modcache'
assert not cache.exists()
cache.mkdir()
env=os.environ.copy()
env.update({'GOTOOLCHAIN':'local','CGO_ENABLED':'0','GOFLAGS':'-mod=readonly','GOWORK':'off','GOROOT':str(alias),'GOMODCACHE':str(cache),'GOPROXY':'https://proxy.golang.org,direct','GOSUMDB':'sum.golang.org'})
env.pop('GOARCH',None)
env.pop('GOOS',None)
def snapshot():
    files=[p for p in (REPO/'internal/runtime').rglob('*') if p.is_file()]+[REPO/'go.mod',REPO/'go.sum']
    return {str(p.relative_to(REPO)).replace('\\','/'):{'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'length':p.stat().st_size,'mtimeNs':p.stat().st_mtime_ns} for p in files}
before=snapshot()
started=time.time()
args=[str(alias/'bin/go.exe'),'run','./internal/runtime/cmd/helpergen','-check']
r=run(args,env)
after=snapshot()
record={'head':HEAD,'command':args,'exit':r.returncode,'seconds':round(time.time()-started,2),'emptyCacheAtStart':True,'GOROOTIsOwnedJunction':True,'fileSetHashesLengthsMtimesUnchanged':before==after,'files':len(before),'output':r.stdout,'retainedOwnedCache':str(cache),'retainedOwnedJunction':str(alias)}
(BASE/'fresh-junction-no-rewrite.json').write_bytes((json.dumps(record,indent=2)+'\n').encode('utf-8'))
print(json.dumps(record),flush=True)
