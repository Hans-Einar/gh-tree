import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

BASE = Path('C:/Users/hanse/.codex/tmp/cr21-helper-review-001')
REPO = Path('C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets')
GO = 'C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe'
HEAD = 'd110a068db8ea68cb2d848774d79b7b9d71b0557'
def call(args,env=None):
    return subprocess.run(args,cwd=REPO,env=env,encoding='utf-8',errors='strict',stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
assert call(['git','rev-parse','HEAD']).stdout.strip()==HEAD
env = os.environ.copy()
env.update({'GOTOOLCHAIN':'local','GOWORK':'off','CGO_ENABLED':'0','GOFLAGS':'-mod=readonly'})
env.pop('GOROOT',None)
env.pop('GOARCH',None)
env.pop('GOOS',None)
def snapshot():
    return {str(p.relative_to(REPO)).replace('\\','/'):{'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'mtimeNs':p.stat().st_mtime_ns,'length':p.stat().st_size} for p in (REPO/'internal/runtime').rglob('*') if p.is_file()}
before=snapshot()
started=time.time()
args=[GO,'run','./internal/runtime/cmd/helpergen','-check']
r=call(args,env)
after=snapshot()
record={'source':HEAD,'command':args,'exit':r.returncode,'seconds':round(time.time()-started,2),'output':r.stdout,'noRewrite':before==after,'runtimeFiles':len(before)}
(BASE/'positive-check.json').write_text(json.dumps(record,indent=2)+'\n',encoding='utf-8')
print(json.dumps(record),flush=True)
args=[GO,'test','-overlay',str(BASE/'overlay.json'),'./internal/runtime/cmd/helpergen','-run','TestReviewer','-count=1','-v']
started=time.time()
r=call(args,env)
(BASE/'reviewer-controls.log').write_text(r.stdout,encoding='utf-8')
print(json.dumps({'command':args,'exit':r.returncode,'seconds':round(time.time()-started,2),'output':r.stdout}),flush=True)
ci=call(['gh','run','view','34068851101','--json','headSha,attempt,status,conclusion,jobs,url'])
assert ci.returncode==0
j=json.loads(ci.stdout)
j['jobs']=[{key:v[key] for key in ['name','conclusion','databaseId','url']} for v in j['jobs']]
(BASE/'ci-attempt2.json').write_text(json.dumps(j,indent=2)+'\n',encoding='utf-8')
print(json.dumps({'ci':j['url'],'attempt':j['attempt'],'conclusion':j['conclusion'],'counts':{state:sum(v['conclusion']==state for v in j['jobs']) for state in ['success','failure','skipped']}}),flush=True)
for job in ['101583102130','101583039279']:
    log=call(['gh','api','repos/Hans-Einar/gh-tree/actions/jobs/'+job+'/logs'])
    assert log.returncode==0
    lines=log.stdout.splitlines()
    indexes=set()
    for i,line in enumerate(lines):
        if any(word in line for word in ['helpergen:','TestActualSelectedDependencyClosure','closure_test.go:24','module lookup disabled']):
            indexes.update(range(max(0,i-2),min(len(lines),i+5)))
    snippet='Job https://github.com/Hans-Einar/gh-tree/actions/runs/34068851101/job/'+job+'; attempt 2\nSelected diagnostic context, not full job log:\n'+'\n'.join(lines[i] for i in sorted(indexes))+'\n'
    (BASE/('ci-job-'+job+'.txt')).write_text(snippet,encoding='utf-8')
