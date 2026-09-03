package tree

import (
	"path"
	"sort"
	"strings"
)

// PathItem is a leaf placed in the namespace tree.
type PathItem struct {
	ID    string
	Path  string
	Label string
}

// Entry is an immediate child of the current folder.
type Entry struct {
	ID       string
	Name     string
	Path     string
	Label    string
	IsFolder bool
}

// NormalizeBranch removes leading technical namespaces. A remaining branch
// with no hierarchy is grouped under misc so it cannot crowd the root.
func NormalizeBranch(branch string, stripPrefixes []string) string {
	parts := split(branch)
	for len(parts) > 1 && isPrefix(parts[0], stripPrefixes) {
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return "misc/unnamed"
	}
	if len(parts) == 1 {
		return "misc/" + parts[0]
	}
	return strings.Join(parts, "/")
}

func isPrefix(candidate string, prefixes []string) bool {
	for _, prefix := range prefixes {
		prefix = strings.Trim(strings.TrimSpace(prefix), "/")
		if prefix != "" && strings.EqualFold(candidate, prefix) {
			return true
		}
	}
	return false
}

func split(value string) []string {
	raw := strings.Split(strings.Trim(strings.TrimSpace(value), "/"), "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		if part != "" && part != "." {
			parts = append(parts, part)
		}
	}
	return parts
}

// Entries returns the immediate folders and leaves visible at current. Query
// keeps a folder when any descendant path or label matches.
func Entries(items []PathItem, current, query string) []Entry {
	current = cleanFolder(current)
	query = strings.ToLower(strings.TrimSpace(query))
	folders := make(map[string]Entry)
	leaves := make([]Entry, 0)
	prefix := current
	if prefix != "" {
		prefix += "/"
	}

	for _, item := range items {
		itemPath := strings.Join(split(item.Path), "/")
		if itemPath == "" || (prefix != "" && !strings.HasPrefix(itemPath, prefix)) {
			continue
		}
		remainder := strings.TrimPrefix(itemPath, prefix)
		if remainder == itemPath && prefix != "" {
			continue
		}
		segments := split(remainder)
		if len(segments) == 0 {
			continue
		}
		matches := query == "" || strings.Contains(strings.ToLower(itemPath), query) ||
			strings.Contains(strings.ToLower(item.Label), query)
		if !matches {
			continue
		}
		if len(segments) > 1 {
			folderPath := joinFolder(current, segments[0])
			folders[segments[0]] = Entry{
				ID:       "folder:" + folderPath,
				Name:     segments[0],
				Path:     folderPath,
				Label:    segments[0] + "/",
				IsFolder: true,
			}
			continue
		}
		leaves = append(leaves, Entry{
			ID:    item.ID,
			Name:  segments[0],
			Path:  itemPath,
			Label: item.Label,
		})
	}

	entries := make([]Entry, 0, len(folders)+len(leaves))
	for _, folder := range folders {
		entries = append(entries, folder)
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	sort.Slice(leaves, func(i, j int) bool {
		left := strings.ToLower(leaves[i].Label)
		right := strings.ToLower(leaves[j].Label)
		if left == right {
			return leaves[i].ID < leaves[j].ID
		}
		return left < right
	})
	return append(entries, leaves...)
}

// ResolveFolder returns saved when it exists, otherwise its nearest existing
// ancestor. A completely stale path resolves to the root.
func ResolveFolder(items []PathItem, saved string) string {
	folder := cleanFolder(saved)
	for folder != "" {
		if FolderExists(items, folder) {
			return folder
		}
		folder = parent(folder)
	}
	return ""
}

func FolderExists(items []PathItem, folder string) bool {
	folder = cleanFolder(folder)
	if folder == "" {
		return true
	}
	prefix := folder + "/"
	for _, item := range items {
		itemPath := strings.Join(split(item.Path), "/")
		if strings.HasPrefix(itemPath, prefix) {
			return true
		}
	}
	return false
}

func Parent(folder string) string { return parent(cleanFolder(folder)) }

func cleanFolder(folder string) string {
	folder = strings.Join(split(folder), "/")
	if folder == "." {
		return ""
	}
	return folder
}

func parent(folder string) string {
	if folder == "" {
		return ""
	}
	parentFolder := path.Dir(folder)
	if parentFolder == "." {
		return ""
	}
	return parentFolder
}

func joinFolder(folder, child string) string {
	if folder == "" {
		return child
	}
	return folder + "/" + child
}
