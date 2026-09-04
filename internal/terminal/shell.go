package terminal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Shell struct {
	Name string
	Path string
	Args []string
}

func DetectShell() (Shell, error) {
	if override := strings.TrimSpace(os.Getenv("GH_TREE_SHELL")); override != "" {
		return shellFromCandidate(override)
	}
	for _, candidate := range ancestryShellCandidates() {
		if sh, ok := knownShell(candidate); ok {
			return sh, nil
		}
	}
	if runtime.GOOS == "windows" {
		if comspec := strings.TrimSpace(os.Getenv("COMSPEC")); comspec != "" {
			return shellFromCandidate(comspec)
		}
		return shellFromCandidate("cmd.exe")
	}
	if shell := strings.TrimSpace(os.Getenv("SHELL")); shell != "" {
		return shellFromCandidate(shell)
	}
	return shellFromCandidate("sh")
}

func ancestryShellCandidates() []string {
	if runtime.GOOS == "windows" {
		return windowsAncestry()
	}
	return unixAncestry()
}

func unixAncestry() []string {
	pid := os.Getppid()
	var out []string
	for i := 0; i < 10 && pid > 1; i++ {
		name, err := psField(pid, "comm")
		if err == nil && name != "" {
			out = append(out, name)
		}
		parentText, err := psField(pid, "ppid")
		if err != nil {
			break
		}
		parent, err := strconv.Atoi(strings.TrimSpace(parentText))
		if err != nil || parent <= 0 || parent == pid {
			break
		}
		pid = parent
	}
	return out
}

func psField(pid int, field string) (string, error) {
	cmd := exec.Command("ps", "-o", field+"=", "-p", strconv.Itoa(pid))
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func windowsAncestry() []string {
	// gh.exe is normally the immediate parent. Query the process chain through
	// CIM so we can recover the shell which launched gh itself.
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {
		return nil
	}
	script := `$id=` + strconv.Itoa(os.Getppid()) + `; for($i=0;$i -lt 10 -and $id -gt 0;$i++){ $p=Get-CimInstance Win32_Process -Filter ("ProcessId="+$id); if(-not $p){break}; [Console]::WriteLine($p.Name); $next=[int]$p.ParentProcessId; if($next -eq $id){break}; $id=$next }`
	cmd := exec.Command(ps, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script)
	b, err := cmd.Output()
	if err != nil {
		return nil
	}
	var out []string
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		if v := strings.TrimSpace(s.Text()); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func knownShell(candidate string) (Shell, bool) {
	base := strings.ToLower(candidate)
	base = strings.ReplaceAll(base, "\\", "/")
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	switch base {
	case "cmd", "cmd.exe":
		sh, err := shellFromCandidate("cmd.exe")
		return sh, err == nil
	case "powershell", "powershell.exe":
		sh, err := shellFromCandidate("powershell.exe")
		if err == nil {
			sh.Name = "Windows PowerShell"
			sh.Args = []string{"-NoLogo"}
		}
		return sh, err == nil
	case "pwsh", "pwsh.exe":
		sh, err := shellFromCandidate("pwsh")
		if err == nil {
			sh.Name = "PowerShell"
			sh.Args = []string{"-NoLogo"}
		}
		return sh, err == nil
	case "bash", "bash.exe", "zsh", "zsh.exe", "fish", "fish.exe", "sh", "sh.exe":
		sh, err := shellFromCandidate(candidate)
		return sh, err == nil
	default:
		return Shell{}, false
	}
}

func shellFromCandidate(candidate string) (Shell, error) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return Shell{}, fmt.Errorf("shell candidate is empty")
	}
	path := candidate
	if resolved, err := exec.LookPath(candidate); err == nil {
		path = resolved
	} else if _, statErr := os.Stat(candidate); statErr != nil {
		return Shell{}, fmt.Errorf("shell %q not found", candidate)
	}
	base := strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	name := strings.TrimSuffix(base, ".exe")
	if name == "pwsh" {
		name = "PowerShell"
	}
	return Shell{Name: name, Path: path}, nil
}
