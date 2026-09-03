package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	activeStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	focusedTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Padding(0, 1)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dialogStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("238")).Padding(0, 1)
	focusPanelStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
)

func (m Model) View() string {
	width := m.width
	if width < 40 { width = 100 }
	height := m.height
	if height < 18 { height = 30 }

	var out strings.Builder
	out.WriteString(m.renderHeader(width))
	out.WriteString("\n")

	upperHeight := max(8, (height-7)*2/3)
	lowerHeight := max(6, height-upperHeight-6)
	if m.mode == modeCommits {
		out.WriteString(m.renderCommitCockpit(width, upperHeight+lowerHeight+1))
	} else {
		out.WriteString(m.renderCockpit(width, upperHeight, lowerHeight))
	}
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine())
	out.WriteString("\n")
	out.WriteString(dimStyle.Render(m.renderFooter()))

	if m.dialog != dialogNone || m.deploying || m.busy {
		if m.dialog != dialogNone || m.deploying {
			out.WriteString("\n\n")
			out.WriteString(m.renderDialog())
		}
	}
	return out.String()
}

func (m Model) renderHeader(width int) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("gh tree"))
	b.WriteString("  ")
	b.WriteString(m.repo)
	b.WriteString("  ")
	switch m.mode {
	case modePullRequests:
		b.WriteString(activeStyle.Render("[PRs]")); b.WriteString(dimStyle.Render("  branches  commits"))
	case modeBranches:
		b.WriteString(dimStyle.Render("PRs  ")); b.WriteString(activeStyle.Render("[branches]")); b.WriteString(dimStyle.Render("  commits"))
	case modeCommits:
		b.WriteString(dimStyle.Render("PRs  branches  ")); b.WriteString(activeStyle.Render("[commits: "+m.commitSource+"]"))
	}
	if m.mode != modeCommits {
		breadcrumb := "/"
		if m.folder != "" { breadcrumb += m.folder + "/" }
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(breadcrumb))
		if m.query != "" || m.searching { b.WriteString("  "); b.WriteString(activeStyle.Render("filter: "+m.query+cursorMarker(m.searching))) }
	}
	return truncate(b.String(), max(20,width))
}

func (m Model) renderCockpit(width, upperHeight, lowerHeight int) string {
	if width >= 88 {
		leftWidth := width/2
		rightWidth := width-leftWidth
		left := m.panel("Navigator", m.focus==paneNavigator, m.renderEntries(leftWidth-4, upperHeight-3), leftWidth-1, upperHeight)
		right := m.panel("Local worktrees", m.focus==paneWorktrees, m.renderWorktreePane(rightWidth-4, upperHeight-3), rightWidth-1, upperHeight)
		upper := lipgloss.JoinHorizontal(lipgloss.Top,left,right)
		lower := m.panel("Active worktree", m.focus==paneDetails, m.renderWorktreeStatus(width-4, lowerHeight-3), width-1, lowerHeight)
		return upper+"\n"+lower
	}
	left := m.panel("Navigator", m.focus==paneNavigator, m.renderEntries(width-4, max(6,upperHeight/2-2)), width-1, max(8,upperHeight/2))
	right := m.panel("Local worktrees", m.focus==paneWorktrees, m.renderWorktreePane(width-4, max(5,upperHeight/2-3)), width-1, max(7,upperHeight/2))
	lower := m.panel("Active worktree", m.focus==paneDetails, m.renderWorktreeStatus(width-4,lowerHeight-3), width-1,lowerHeight)
	return left+"\n"+right+"\n"+lower
}

func (m Model) renderCommitCockpit(width, height int) string {
	if width >= 88 {
		leftWidth:=width/2; rightWidth:=width-leftWidth
		left:=m.panel("Commits",m.focus!=paneDetails,m.renderCommitList(leftWidth-4,height-3),leftWidth-1,height)
		right:=m.panel("Commit details",m.focus==paneDetails,m.renderCommitDetails(rightWidth-4,height-3),rightWidth-1,height)
		return lipgloss.JoinHorizontal(lipgloss.Top,left,right)
	}
	listHeight:=max(8,height/2)
	left:=m.panel("Commits",m.focus!=paneDetails,m.renderCommitList(width-4,listHeight-3),width-1,listHeight)
	right:=m.panel("Commit details",m.focus==paneDetails,m.renderCommitDetails(width-4,height-listHeight-3),width-1,max(7,height-listHeight))
	return left+"\n"+right
}

func (m Model) panel(title string, focused bool, content string, width,height int) string {
	titleRender:=headerStyle.Render(title)
	style:=panelStyle
	if focused { titleRender=focusedTitle.Render(title); style=focusPanelStyle }
	body:=titleRender+"\n"+content
	return style.Width(max(10,width-2)).Height(max(3,height-2)).Render(body)
}

func (m Model) renderEntries(width,height int) string {
	if m.loading && len(m.entries)==0 { return "Loading…" }
	if len(m.entries)==0 { return dimStyle.Render("No matching items") }
	visible:=max(3,height)
	start:=0; if m.cursor>=visible { start=m.cursor-visible+1 }
	end:=min(len(m.entries),start+visible)
	lines:=make([]string,0,end-start)
	for i:=start;i<end;i++ {
		e:=m.entries[i]; marker:="  "; if i==m.cursor {marker="> "}
		label:=e.Label; if e.IsFolder {label=e.Name+"/"}
		line:=truncate(marker+label,max(8,width))
		if i==m.cursor {line=selectedStyle.Render(line)}
		lines=append(lines,line)
	}
	return strings.Join(lines,"\n")
}

func (m Model) renderWorktreePane(width,height int) string {
	if !m.snapshot.WorktreesEnabled { return dimStyle.Render("Unavailable outside the selected local repo") }
	if len(m.snapshot.Worktrees)==0 { return dimStyle.Render("No worktrees yet\n\n[c] create one from the selected PR/branch") }
	visible:=max(3,height-2); start:=0; if m.worktreeCursor>=visible {start=m.worktreeCursor-visible+1}; end:=min(len(m.snapshot.Worktrees),start+visible)
	lines:=make([]string,0,end-start+2)
	for i:=start;i<end;i++ {
		wt:=m.snapshot.Worktrees[i]; marker:="  "; if i==m.worktreeCursor {marker="> "}
		branch:=wt.Branch; if wt.Detached {branch="DETACHED"}
		flag:=""; if sameDisplayPath(wt.Path,m.activeWorktree) {flag=" *ACTIVE"} else if wt.Primary {flag=" primary"}
		line:=truncate(fmt.Sprintf("%s%s  %s @ %s%s",marker,filepath.Base(wt.Path),branch,shortSHA(wt.Head),flag),width)
		if i==m.worktreeCursor {line=selectedStyle.Render(line)}
		lines=append(lines,line)
	}
	if wt,ok:=m.selectedWorktree();ok { lines=append(lines,"",dimStyle.Render(truncate(wt.Path,width))) }
	return strings.Join(lines,"\n")
}

func (m Model) renderWorktreeStatus(width,height int) string {
	if m.activeWorktree=="" { return dimStyle.Render("No active worktree. Select one above and press Enter, or press c to create one.") }
	if !m.haveWTStatus { return dimStyle.Render("Active: "+wrapText(m.activeWorktree,width)+"\nStatus loading…") }
	s:=m.worktreeStatus
	branch:=s.Info.Branch; if s.Info.Detached {branch="DETACHED HEAD"}
	clean:="CLEAN"; if !s.Clean { clean=fmt.Sprintf("DIRTY · staged %d · modified %d · untracked %d · conflicts %d",s.Staged,s.Modified,s.Untracked,s.Conflicted) }
	remote:="no upstream"; if s.Upstream!="" {remote=fmt.Sprintf("%s · ahead %d / behind %d",s.Upstream,s.Ahead,s.Behind)}
	prLine:="no open PR linked to this branch"
	if pr,ok:=m.prForBranch(s.Info.Branch);ok { state:="OPEN"; if pr.IsDraft {state="DRAFT"}; match:=""; if strings.EqualFold(pr.HeadSHA,s.Info.Head){match=" · HEAD matches PR"}else{match=" · local HEAD differs"}; prLine=fmt.Sprintf("PR #%d %s · %s → %s%s",pr.Number,state,pr.HeadBranch,pr.BaseBranch,match) }
	flags:=[]string{}; if s.Info.Primary {flags=append(flags,"PRIMARY")}; if s.Info.Current {flags=append(flags,"START DIRECTORY")}
	lines:=[]string{
		labeledValue("path:    ",s.Info.Path,width),
		labeledValue("branch:  ",branch,width),
		labeledValue("HEAD:    ",s.Info.Head,width),
		labeledValue("working: ",clean,width),
		labeledValue("remote:  ",remote,width),
		labeledValue("PR:      ",prLine,width),
	}
	if len(flags)>0 {lines=append(lines,"flags:   "+strings.Join(flags,", "))}
	if len(lines)>height && height>0 {lines=lines[:height]}
	return strings.Join(lines,"\n")
}

func (m Model) renderCommitList(width,height int) string {
	if m.busy&&len(m.commits)==0{return "Loading history…"}
	if len(m.commits)==0{return dimStyle.Render("No commits")}
	visible:=max(3,height);start:=0;if m.commitCursor>=visible{start=m.commitCursor-visible+1};end:=min(len(m.commits),start+visible)
	lines:=make([]string,0,end-start)
	for i:=start;i<end;i++{c:=m.commits[i];marker:="  ";if i==m.commitCursor{marker="> "};when:=c.Authored.Local().Format("2006-01-02");line:=truncate(fmt.Sprintf("%s%s  %s  %s",marker,shortSHA(c.SHA),when,c.Subject),width);if i==m.commitCursor{line=selectedStyle.Render(line)};lines=append(lines,line)}
	return strings.Join(lines,"\n")
}

func (m Model) renderCommitDetails(width,height int) string {
	c,ok:=m.currentCommit();if !ok{return dimStyle.Render("No commit selected")}
	parents:=strings.Join(c.Parents," ");if parents==""{parents="(root)"}
	text:=strings.Join([]string{
		labeledValue("commit:  ",c.SHA,width),
		labeledValue("parents: ",parents,width),
		labeledValue("author:  ",fmt.Sprintf("%s <%s>",c.Author,c.Email),width),
		labeledValue("date:    ",c.Authored.Local().Format(time.RFC1123Z),width),
		"",
		wrapText(c.Message,width),
	},"\n")
	lines:=strings.Split(text,"\n");maxScroll:=max(0,len(lines)-max(1,height));scroll:=m.commitScroll;if scroll>maxScroll{scroll=maxScroll};end:=min(len(lines),scroll+max(1,height));return strings.Join(lines[scroll:end],"\n")
}

func (m Model) renderStatusLine() string {
	if strings.HasPrefix(m.status,"✓") {return successStyle.Render(m.status)}
	low:=strings.ToLower(m.status);if strings.Contains(low,"fail")||strings.Contains(low,"error")||strings.Contains(low,"refus")||strings.Contains(low,"cannot") {return errorStyle.Render(m.status)}
	return dimStyle.Render(m.status)
}

func (m Model) renderFooter() string {
	if m.mode==modeCommits {
		if m.focus==paneDetails{return "[Tab] commits  [↑/↓ PgUp/PgDn] scroll message  [Esc] back  [p] PRs  [b] branches  [q] quit"}
		return "[Tab] details  [↑/↓] select  [L] load more  [c] worktree at commit  [n] branch at commit  [Esc] back  [q] quit"
	}
	switch m.focus {
	case paneWorktrees:
		return "[Tab] focus  [Enter/a] activate  [c] create  [x] checkout selected PR/branch  [f] fetch  [p] pull  [m] commit  [P] push  [n] new branch  [o] create PR  [h] history  [q] quit"
	case paneDetails:
		return "[Tab] focus  [Enter] refresh status  [h] history  [p] PRs  [b] branches  [r] refresh  [q] quit"
	default:
		return "[Tab] focus  [Enter] open/select  [Backspace] parent  [p] PRs  [b] branches  [h] history  [w] deploy  [/] search  [r] refresh  [q] quit"
	}
}

func (m Model) renderDialog() string {
	if m.deploying{return dialogStyle.Render(fmt.Sprintf("Deploying PR #%d\n%s\n\nPlease wait…",m.pendingPR.Number,m.pendingPR.HeadSHA))}
	switch m.dialog {
	case dialogTargetPicker:
		lines:=[]string{headerStyle.Render(fmt.Sprintf("Deploy PR #%d to which legacy target?",m.pendingPR.Number))};for i,t:=range m.targets{marker:="  ";if i==m.targetCursor{marker="> "};lines=append(lines,fmt.Sprintf("%s%s → %s (%s)",marker,t.Name,t.Path,t.Branch))};lines=append(lines,"",dimStyle.Render("[Enter] choose  [Esc] cancel"));return dialogStyle.Render(strings.Join(lines,"\n"))
	case dialogConfirm:
		return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Confirm local test deployment"),fmt.Sprintf("PR:     #%d",m.pendingPR.Number),"SHA:    "+m.pendingPR.HeadSHA,"Target: "+m.pendingTarget.Name,"Path:   "+m.pendingTarget.Path,"Branch: "+m.pendingTarget.Branch,"",activeStyle.Render("Deploy? [y/N]")},"\n"))
	case dialogCreateWorktree:
		source:="active commit";if m.pendingPR.Number>0{source=fmt.Sprintf("PR #%d · %s",m.pendingPR.Number,m.pendingPR.HeadBranch)}else if m.pendingBranch.Name!=""{source="branch "+m.pendingBranch.Name}else if m.pendingRevision!=""{source="commit "+shortSHA(m.pendingRevision)}
		return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Create worktree"),"source: "+source,"",inputLine("Path",m.inputA,m.inputField==0),inputLine("Branch (blank = detached)",m.inputB,m.inputField==1),"",dimStyle.Render("[Tab] field  [Enter] next/create  [Esc] cancel")},"\n"))
	case dialogCheckoutConfirm:
		target:="";if m.pendingPR.Number>0{target=fmt.Sprintf("PR #%d · %s @ %s",m.pendingPR.Number,m.pendingPR.HeadBranch,m.pendingPR.HeadSHA)}else{target="branch "+m.pendingBranch.Name}
		return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Retarget active worktree"),labeledValue("path:   ",m.activeWorktree,90),"target: "+target,"","Dirty worktrees are refused; the primary worktree is protected.",activeStyle.Render("Checkout? [y/N]")},"\n"))
	case dialogNewBranch:
		return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Create branch in active worktree"),inputLine("Branch",m.inputA,true),"start: "+coalesce(m.pendingRevision,"HEAD"),"",dimStyle.Render("[Enter] create  [Esc] cancel")},"\n"))
	case dialogCommit:
		return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Commit active worktree"),"All current changes will be staged explicitly.",inputLine("Message",m.inputA,true),"",dimStyle.Render("[Enter] stage + commit  [Esc] cancel")},"\n"))
	case dialogPushConfirm:
		dest:=m.worktreeStatus.Upstream;if dest==""{dest="origin/"+m.worktreeStatus.Info.Branch+" (set upstream)"};return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Push active branch"),"branch: "+m.worktreeStatus.Info.Branch,"to:     "+dest,"","No force push is used.",activeStyle.Render("Push? [y/N]")},"\n"))
	case dialogCreatePR:
		return dialogStyle.Render(strings.Join([]string{headerStyle.Render("Create draft PR"),"head: "+m.worktreeStatus.Info.Branch,inputLine("Base",m.inputA,m.inputField==0),inputLine("Title",m.inputB,m.inputField==1),inputLine("Body",m.inputC,m.inputField==2),"",dimStyle.Render("[Tab] field  [Enter] next/create  [Esc] cancel")},"\n"))
	default:return ""
	}
}

func inputLine(label,value string,active bool) string {cursor:="";if active{cursor="▏"};prefix:=label+": ";if active{return activeStyle.Render(prefix)+value+cursor};return prefix+value}
func labeledValue(label,value string,width int) string {if lipgloss.Width(label+value)<=width{return label+value};return label+"\n"+wrapText(value,width)}
func wrapText(value string,width int) string {if width<1||lipgloss.Width(value)<=width{return value};var lines []string;runes:=[]rune(value);for len(runes)>0{cut:=len(runes);for cut>1&&lipgloss.Width(string(runes[:cut]))>width{cut--};lines=append(lines,string(runes[:cut]));runes=runes[cut:]};return strings.Join(lines,"\n")}
func truncate(value string,width int) string {if width<2||lipgloss.Width(value)<=width{return value};runes:=[]rune(value);for len(runes)>0&&lipgloss.Width(string(runes))+1>width{runes=runes[:len(runes)-1]};return string(runes)+"…"}
func shortSHA(sha string) string {if len(sha)<=8{return sha};return sha[:8]}
func cursorMarker(active bool) string {if active{return "▏"};return ""}
func max(left,right int) int {if left>right{return left};return right}
