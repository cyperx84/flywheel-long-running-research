package vault

import (
	"fmt"
	"os/exec"
	"strings"
)

// Note represents a vault note to create or update.
type Note struct {
	ID       string
	Title    string
	Content  string
	Tags     []string
	Aliases  []string
	Folder   string // optional subfolder (e.g., "claw", "ideas")
	Modified string // ISO date
}

// Create writes a new note to the vault using obsidian-cli.
func Create(note Note) error {
	tags := append([]string{"reference"}, note.Tags...)
	tagLines := make([]string, len(tags))
	for i, t := range tags {
		tagLines[i] = fmt.Sprintf("  - %s", t)
	}

	aliasLines := make([]string, len(note.Aliases))
	for i, a := range note.Aliases {
		aliasLines[i] = fmt.Sprintf("  - %s", a)
	}

	content := fmt.Sprintf(`---
id: %s
title: %s
created: %s
modified: %s
tags:
%s
refs: []
aliases:
%s
---

# %s

%s
`,
		note.ID,
		note.Title,
		note.Modified,
		note.Modified,
		strings.Join(tagLines, "\n"),
		strings.Join(aliasLines, "\n"),
		note.Title,
		note.Content,
	)

	args := []string{"create", note.ID, "--overwrite", "--content", content}
	if note.Folder != "" {
		args = append(args, "--folder", note.Folder)
	}

	cmd := exec.Command("obsidian-cli", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("obsidian-cli create: %w\n%s", err, string(out))
	}
	return nil
}

// UpdateFrontmatter sets a frontmatter field on an existing note.
func UpdateFrontmatter(noteID, key, value string) error {
	cmd := exec.Command("obsidian-cli", "frontmatter", noteID, "--edit", "--key", key, "--value", value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("obsidian-cli frontmatter: %w\n%s", err, string(out))
	}
	return nil
}

// Search finds notes matching a query.
func Search(query string) (string, error) {
	cmd := exec.Command("obsidian-cli", "search-content", query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("obsidian-cli search: %w\n%s", err, string(out))
	}
	return string(out), nil
}
