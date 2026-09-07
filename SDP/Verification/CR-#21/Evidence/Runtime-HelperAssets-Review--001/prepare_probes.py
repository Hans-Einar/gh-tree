import base64
import hashlib
import json
import os
from pathlib import Path
import subprocess

BASE = Path('C:/Users/hanse/.codex/tmp/cr21-helper-review-001')
REPO = Path('C:/Users/hanse/GIT/gh-tree-wt/runtime-helper-assets')
GO = 'C:/Users/hanse/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.0.windows-amd64/bin/go.exe'
def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(value if isinstance(value, bytes) else value.encode('utf-8'))
def h1(files):
    rows = ''.join(hashlib.sha256(data).hexdigest()+'  '+name+'\n' for name,data in sorted(files.items()))
    return 'h1:'+base64.b64encode(hashlib.sha256(rows.encode('utf-8')).digest()).decode('ascii')

mod = 'example.com/helperfixture'
version = 'v1.0.0'
modfile = 'module '+mod+'\n\ngo 1.25.0\n'
original = 'package helperfixture\nconst Value = "REVIEW_ORIGINAL_MODULE_BYTES_111111"\n'
changed = 'package helperfixture\nconst Value = "REVIEW_CHANGED_MODULE_BYTES_222222"\n'
for case in ['pinned', 'forged']:
    root = BASE/case/'source'
    cache = BASE/case/'modcache'
    prefix = mod+'@'+version
    body = original if case == 'pinned' else changed
    expected = h1({prefix+'/go.mod':modfile.encode(),prefix+'/value.go':original.encode()})
    observed = h1({prefix+'/go.mod':modfile.encode(),prefix+'/value.go':body.encode()})
    write(cache/prefix/'go.mod',modfile)
    write(cache/prefix/'value.go',body)
    download = cache/'cache/download'/mod/'@v'
    write(download/(version+'.mod'),modfile)
    write(download/(version+'.info'),json.dumps({'Version':version,'Time':'2026-01-01T00:00:00Z'}))
    write(download/(version+'.ziphash'),observed)
    write(download/'list',version+'\n')
    write(root/'go.mod','module github.com/Hans-Einar/gh-tree\n\ngo 1.25.0\n\nrequire '+mod+' '+version+'\n')
    write(root/'go.sum',mod+' '+version+' '+expected+'\n'+mod+' '+version+'/go.mod '+h1({'go.mod':modfile.encode()})+'\n')
    write(root/'internal/runtime/broker/protocol.go','package broker\nconst ProtocolVersion uint16 = 1\n')
    write(root/'internal/runtime/broker/cmd/main_windows.go','package main\nimport ("fmt"; "example.com/helperfixture"; "github.com/Hans-Einar/gh-tree/internal/runtime/broker")\nfunc main(){fmt.Print(helperfixture.Value, broker.ProtocolVersion)}\n')
    recipe = root/'internal/runtime/cmd/helpergen'
    for file in (REPO/'internal/runtime/cmd/helpergen').glob('*.go'):
        if not file.name.endswith('_test.go'):
            write(recipe/file.name,file.read_bytes())
    print(json.dumps({'case':case,'expectedGoSum':expected,'actualCacheHash':observed}))

write(BASE/'overlay.json',json.dumps({'Replace':{str(REPO/'internal/runtime/cmd/helpergen/reviewer_test.go'):str(BASE/'reviewer_test.go.txt')}}))
empty = BASE/'emptycache'
empty.mkdir(exist_ok=True)
env = os.environ.copy()
env.update({'GOTOOLCHAIN':'local','CGO_ENABLED':'0','GOFLAGS':'-mod=readonly','GOMODCACHE':str(empty),'GOPROXY':'off','GOSUMDB':'off','PYTHONIOENCODING':'utf-8'})
before = {str(p.relative_to(REPO)):(hashlib.sha256(p.read_bytes()).hexdigest(),p.stat().st_mtime_ns) for p in (REPO/'internal/runtime').rglob('*') if p.is_file()}
result = subprocess.run([GO,'run','./internal/runtime/cmd/helpergen','-check'],cwd=REPO,env=env,encoding='utf-8',errors='strict',stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
after = {str(p.relative_to(REPO)):(hashlib.sha256(p.read_bytes()).hexdigest(),p.stat().st_mtime_ns) for p in (REPO/'internal/runtime').rglob('*') if p.is_file()}
write(BASE/'emptycache.log',result.stdout)
print(json.dumps({'control':'exact-check-empty-module-cache','exit':result.returncode,'noRewrite':before==after,'files':len(before),'output':result.stdout}))
