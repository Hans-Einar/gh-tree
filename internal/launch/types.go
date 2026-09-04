package launch

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Candidate struct {
	Provider string
	ID       string
	Path     []string
	Script   string
	Targets  []string
	Command  string
}

type Invocation struct {
	Provider string
	Name     string
	Command  string
	Args     []string
	Dir      string
}

type Provider interface {
	Name() string
	Detect(root string) bool
	Discover(root string) ([]Candidate, error)
	Build(root string, candidate Candidate) (Invocation, error)
}

type Registry struct { providers []Provider }

func DefaultRegistry() Registry { return Registry{providers: []Provider{NPMProvider{}, MakeProvider{}}} }
func NewRegistry(providers ...Provider) Registry { return Registry{providers: append([]Provider(nil), providers...)} }

func (r Registry) Discover(root string) ([]Candidate,error) {
	var all []Candidate
	for _,p:=range r.providers {
		if !p.Detect(root){continue}
		items,err:=p.Discover(root);if err!=nil{return nil,fmt.Errorf("discover %s launch points: %w",p.Name(),err)}
		all=append(all,items...)
	}
	sort.Slice(all,func(i,j int)bool{if all[i].Provider!=all[j].Provider{return all[i].Provider<all[j].Provider};return strings.Join(all[i].Path,"/")<strings.Join(all[j].Path,"/")})
	return all,nil
}

func (r Registry) Build(root string,c Candidate)(Invocation,error){
	for _,p:=range r.providers{if p.Name()==c.Provider{return p.Build(root,c)}}
	return Invocation{},fmt.Errorf("unknown launch provider %q",c.Provider)
}

func cleanRoot(root string)(string,error){if strings.TrimSpace(root)==""{return "",fmt.Errorf("launch root is empty")};abs,err:=filepath.Abs(root);if err!=nil{return "",err};return filepath.Clean(abs),nil}
