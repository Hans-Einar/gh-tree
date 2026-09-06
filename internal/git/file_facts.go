package git

import (
	"crypto/sha256"
	"io"
	"os"
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func (s *readSession) fileState(scope api.WorktreeScope, path api.GitPath) (api.FileState, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, err
	}
	d := scope.Data()
	root, err := acquireDirectory(directoryObservation{path: d.RootLocator, identity: d.RootIdentity})
	if err != nil {
		return nil, err
	}
	parents := []*nativeDirectory{root}
	defer func() {
		for i := len(parents) - 1; i >= 0; i-- {
			parents[i].close()
		}
	}()
	components := strings.Split(path.String(), "/")
	parent := root
	for _, component := range components[:len(components)-1] {
		child, e := parent.openChild(component)
		if e != nil {
			if os.IsNotExist(e) {
				return api.NewAbsentFile(api.AbsentFileData{Path: path})
			}
			return nil, e
		}
		parents = append(parents, child)
		parent = child
	}
	name := components[len(components)-1]
	f, err := parent.openRegular(name)
	if os.IsNotExist(err) {
		return api.NewAbsentFile(api.AbsentFileData{Path: path})
	}
	var identity string
	var mode uint32
	kind := api.RegularFile
	var target api.Optional[string]
	var digest []byte
	if err != nil {
		link, id, m, e := parent.linkObservation(name)
		if e != nil {
			return nil, e
		}
		identity = id
		mode = m
		kind = api.SymlinkFile
		target = api.Some(link)
		sum := sha256.Sum256([]byte(link))
		digest = sum[:]
	} else {
		defer f.Close()
		identity, mode, err = regularIdentity(f)
		if err != nil {
			return nil, err
		}
		hash := sha256.New()
		buffer := make([]byte, 64<<10)
		var total int64
		for {
			if err := s.ctx.Err(); err != nil {
				return nil, err
			}
			n, e := f.Read(buffer)
			if n > 0 {
				total += int64(n)
				if total > 64<<20 {
					return nil, diagnostic(api.Unavailable, "FileObservationLimit", "The file exceeds the bounded full-content observation profile.")
				}
				hash.Write(buffer[:n])
			}
			if e == io.EOF {
				break
			}
			if e != nil {
				return nil, e
			}
		}
		digest = hash.Sum(nil)
		after, m, e := regularIdentity(f)
		if e != nil {
			return nil, e
		}
		if after != identity || m != mode {
			return nil, diagnostic(api.StaleObservation, "FileChangedDuringRead", "The file object or mode changed during content observation.")
		}
	}
	subject := queryBinding(d.ID.Repository().Token(), d.ID.AdministrativeKey(), path.String())
	return api.NewPresentFile(api.PresentFileData{Path: path, ObjectIdentity: sourceVersion("file-object", subject, s.a.lifetime, []byte(identity)), Kind: kind, Mode: mode, Content: sourceVersion("file-content", subject, s.a.lifetime, digest), LinkTarget: target, ParentIdentity: sourceVersion("file-parent", subject, s.a.lifetime, []byte(directoryKey(parent.expected)))})
}
