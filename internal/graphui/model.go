package graphui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Hans-Einar/gh-tree/internal/graph"
)

type Backend interface {
	LoadGraph(context.Context, string, int, int) (graph.Snapshot, error)
}

type loadMsg struct { snapshot graph.Snapshot; append bool; err error }

type Model struct {
	repo string
	backend Backend
	commits []graph.Commit
	rows []graph.Row
	cursor int
	detailScroll int
	focusDetails bool
	width int
	height int
	loading bool
	hasMore bool
	freshness string
	status string
}

func New(repo string, backend Backend) Model {
	return Model{repo:repo,backend:backend,loading:true,status:"Loading Git graph…"}
}

func (m Model) Init() tea.Cmd { return m.loadCmd(100,0,false) }
func (m Model) loadCmd(limit,skip int,appendMode bool) tea.Cmd {
	return func() tea.Msg { s,err:=m.backend.LoadGraph(context.Background(),m.repo,limit,skip);return loadMsg{snapshot:s,append:appendMode,err:err} }
}
func (m Model) Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch x:=msg.(type){
	case tea.WindowSizeMsg:m.width,m.height=x.Width,x.Height;return m,nil
	case loadMsg:
		m.loading=false
		if x.err!=nil{m.status="Graph load failed: "+x.err.Error();return m,nil}
		if x.append{m.commits=append(m.commits,x.snapshot.Commits...)}else{m.commits=x.snapshot.Commits;m.cursor=0;m.detailScroll=0}
		m.rows=graph.Rows(m.commits);m.hasMore=x.snapshot.HasMore;m.freshness=x.snapshot.RemoteFreshness;m.status=fmt.Sprintf("Loaded %d commits",len(m.commits));return m,nil
	case tea.KeyMsg:
		if x.String()=="ctrl+c"||x.String()=="q"{return m,tea.Quit}
		if m.loading{return m,nil}
		if x.String()=="tab"||x.String()=="shift+tab"{m.focusDetails=!m.focusDetails;return m,nil}
		if x.String()=="r"{m.loading=true;m.status="Refreshing graph from local refs…";return m,m.loadCmd(100,0,false)}
		if x.String()=="L"&&m.hasMore{m.loading=true;m.status="Loading more history…";return m,m.loadCmd(100,len(m.commits),true)}
		if m.focusDetails{
			switch x.String(){case "up","k":if m.detailScroll>0{m.detailScroll--};case "down","j":m.detailScroll++;case "pgup":m.detailScroll=max(0,m.detailScroll-10);case "pgdown":m.detailScroll+=10}
			return m,nil
		}
		switch x.String(){case "up","k":if m.cursor>0{m.cursor--;m.detailScroll=0};case "down","j":if m.cursor+1<len(m.rows){m.cursor++;m.detailScroll=0};case "home","g":m.cursor=0;m.detailScroll=0;case "end","G":if len(m.rows)>0{m.cursor=len(m.rows)-1;m.detailScroll=0}}
	}
	return m,nil
}

var (
	header=lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	selected=lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	dim=lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	panel=lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("238")).Padding(0,1)
	focused=lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63")).Padding(0,1)
)

func (m Model) View() string {
	w:=m.width;if w<70{w=100};h:=m.height;if h<18{h=30}
	var b strings.Builder
	b.WriteString(header.Render("gh tree · Git graph"));b.WriteString("  ");b.WriteString(m.repo);b.WriteString("\n")
	bodyH:=max(10,h-5)
	if w>=92{
		lw:=w*3/5;rw:=w-lw
		left:=m.renderPanel("Git DAG",!m.focusDetails,m.renderRows(lw-4,bodyH-3),lw-1,bodyH)
		right:=m.renderPanel("Commit details",m.focusDetails,m.renderDetails(rw-4,bodyH-3),rw-1,bodyH)
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,left,right))
	}else{
		listH:=bodyH/2
		b.WriteString(m.renderPanel("Git DAG",!m.focusDetails,m.renderRows(w-4,listH-3),w-1,listH));b.WriteString("\n")
		b.WriteString(m.renderPanel("Commit details",m.focusDetails,m.renderDetails(w-4,bodyH-listH-3),w-1,max(7,bodyH-listH)))
	}
	b.WriteString("\n");b.WriteString(dim.Render(m.status+" · "+m.freshness));b.WriteString("\n")
	if m.focusDetails{b.WriteString(dim.Render("[Tab] graph  [↑/↓ PgUp/PgDn] scroll details  [r] reload  [L] load more  [q] quit"))}else{b.WriteString(dim.Render("[Tab] details  [↑/↓] select  [r] reload  [L] load more  [q] quit"))}
	return b.String()
}

func (m Model) renderPanel(title string,active bool,content string,w,h int) string { s:=panel;if active{s=focused};return s.Width(max(10,w-2)).Height(max(4,h-2)).Render(header.Render(title)+"\n"+content) }
func (m Model) renderRows(w,h int) string {
	if m.loading&&len(m.rows)==0{return "Loading…"};if len(m.rows)==0{return dim.Render("No commits")}
	visible:=max(3,h);start:=0;if m.cursor>=visible{start=m.cursor-visible+1};end:=min(len(m.rows),start+visible);lines:=make([]string,0,end-start)
	for i:=start;i<end;i++{r:=m.rows[i];deco:=decorations(r.Commit);line:=fmt.Sprintf("%s %-8s %s",r.Prefix,short(r.Commit.SHA),r.Commit.Subject);if deco!=""{line+="  "+deco};line=truncate(line,w);if i==m.cursor{line=selected.Render(line)};lines=append(lines,line)}
	return strings.Join(lines,"\n")
}
func (m Model) renderDetails(w,h int) string {
	if m.cursor<0||m.cursor>=len(m.rows){return dim.Render("No commit selected")};c:=m.rows[m.cursor].Commit;parents:=strings.Join(c.Parents," ");if parents==""{parents="(root)"}
	var lines []string
	lines=append(lines,"commit:  "+c.SHA,"parents: "+parents,"author:  "+c.Author,"date:    "+c.Authored.Local().Format(time.RFC1123Z),"subject: "+c.Subject)
	for _,d:=range c.Decorations{lines=append(lines,fmt.Sprintf("ref:     %s %s",d.Kind,d.Name))}
	for _,pr:=range c.PRs{state:="OPEN";if pr.Draft{state="DRAFT"};lines=append(lines,fmt.Sprintf("PR:      #%d %s · %s → %s",pr.Number,state,pr.Head,pr.Base))}
	for _,wt:=range c.Worktrees{state:=wt.Branch;if wt.Detached{state="DETACHED"};tag:="";if wt.Primary{tag=" PRIMARY"};lines=append(lines,fmt.Sprintf("WT:      %s · %s%s",filepath.Base(wt.Path),state,tag))}
	wrapped:=wrapLines(lines,w);maxScroll:=max(0,len(wrapped)-max(1,h));scroll:=m.detailScroll;if scroll>maxScroll{scroll=maxScroll};end:=min(len(wrapped),scroll+max(1,h));return strings.Join(wrapped[scroll:end],"\n")
}
func decorations(c graph.Commit) string {var parts []string;for _,d:=range c.Decorations{switch d.Kind{case graph.RefHEAD:parts=append(parts,"HEAD");case graph.RefLocal:parts=append(parts,"[L:"+d.Name+"]");case graph.RefRemote:parts=append(parts,"[R:"+d.Name+"]");case graph.RefTag:parts=append(parts,"[T:"+d.Name+"]")}};for _,p:=range c.PRs{parts=append(parts,fmt.Sprintf("[PR#%d]",p.Number))};for _,wt:=range c.Worktrees{parts=append(parts,"[WT:"+filepath.Base(wt.Path)+"]")};return strings.Join(parts," ")}
func wrapLines(lines []string,w int)[]string{var out []string;for _,line:=range lines{if len(line)<=w{out=append(out,line);continue};r:=[]rune(line);for len(r)>0{n:=min(len(r),max(1,w));out=append(out,string(r[:n]));r=r[n:]}};return out}
func short(s string)string{if len(s)<=8{return s};return s[:8]}
func truncate(s string,w int)string{r:=[]rune(s);if len(r)<=w{return s};if w<2{return s};return string(r[:w-1])+"…"}
func min(a,b int)int{if a<b{return a};return b}
func max(a,b int)int{if a>b{return a};return b}
