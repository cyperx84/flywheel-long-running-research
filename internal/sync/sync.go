package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyperx84/flywheel/internal/extractor"
	"github.com/cyperx84/flywheel/internal/vault"
)

type Options struct {
	Since  string
	Agent  string
	DryRun bool
	JSON   bool
	Agents []string
	Dir    string
	Vault  string
}

type Result struct {
	Updated int      `json:"updated"`
	Created int      `json:"created"`
	Stale   int      `json:"stale"`
	Skipped int      `json:"skipped"`
	Errors  int      `json:"errors"`
	Details []Detail `json:"details,omitempty"`
}

type Detail struct {
	Action string `json:"action"` // UPDATED, CREATED, STALE, SKIPPED
	Topic  string `json:"topic"`
	Note   string `json:"note,omitempty"`
	Source string `json:"source"`
}

func Run(opts Options) error {
	agents := opts.Agents
	if opts.Agent != "" {
		agents = []string{opts.Agent}
	}

	var allEntries []extractor.Entry
	for _, agent := range agents {
		entries := scanAgent(opts.Dir, agent, opts.Since)
		allEntries = append(allEntries, entries...)
	}

	if len(allEntries) == 0 {
		if opts.JSON {
			out, _ := json.Marshal(Result{})
			fmt.Println(string(out))
		} else {
			fmt.Printf("No learnings found since %s\n", opts.Since)
		}
		return nil
	}

	result := Result{}
	now := time.Now().Format("2006-01-02 15:04")

	for _, entry := range allEntries {
		detail := Detail{Topic: entry.Topic, Source: entry.Source}

		switch entry.Tag {
		case extractor.TagStale:
			result.Stale++
			detail.Action = "STALE"
			fmt.Printf("⚠️  STALE   %s\n   %s\n   Via: %s\n", entry.Topic, entry.Content, entry.Source)

		case extractor.TagLearning:
			if opts.DryRun {
				result.Created++
				detail.Action = "CREATED (dry-run)"
				fmt.Printf("📄 CREATE  %s (dry-run)\n   %s\n", entry.Topic, entry.Content)
			} else {
				noteID := topicToID(entry.Topic)
				err := vault.Create(vault.Note{
					ID:       noteID,
					Title:    entry.Topic,
					Content:  entry.Content,
					Tags:     []string{},
					Aliases:  []string{entry.Topic},
					Modified: now,
				})
				if err != nil {
					result.Errors++
					detail.Action = "ERROR"
					fmt.Printf("❌ ERROR   %s: %v\n", entry.Topic, err)
				} else {
					result.Created++
					detail.Action = "CREATED"
					detail.Note = noteID
					fmt.Printf("✅ CREATED %s\n   %s\n   Via: %s\n", entry.Topic, entry.Content, entry.Source)
				}
			}

		case extractor.TagUpdate:
			if opts.DryRun {
				result.Updated++
				detail.Action = "UPDATED (dry-run)"
				fmt.Printf("✏️  UPDATE  %s (dry-run)\n   %s\n", entry.Topic, entry.Content)
			} else {
				noteID := topicToID(entry.Topic)
				err := vault.UpdateFrontmatter(noteID, "modified", now)
				if err != nil {
					// Note might not exist — try creating
					err2 := vault.Create(vault.Note{
						ID:       noteID,
						Title:    entry.Topic,
						Content:  entry.Content,
						Tags:     []string{},
						Aliases:  []string{entry.Topic},
						Modified: now,
					})
					if err2 != nil {
						result.Errors++
						detail.Action = "ERROR"
						fmt.Printf("❌ ERROR   %s: %v\n", entry.Topic, err2)
					} else {
						result.Created++
						detail.Action = "CREATED"
						detail.Note = noteID
						fmt.Printf("✅ CREATED %s\n   %s\n   Via: %s\n", entry.Topic, entry.Content, entry.Source)
					}
				} else {
					result.Updated++
					detail.Action = "UPDATED"
					detail.Note = noteID
					fmt.Printf("✅ UPDATED %s\n   Via: %s\n", entry.Topic, entry.Source)
				}
			}

		default:
			result.Skipped++
		}

		result.Details = append(result.Details, detail)
		fmt.Println()
	}

	fmt.Printf("Summary: %d updated, %d created, %d stale, %d errors\n",
		result.Updated, result.Created, result.Stale, result.Errors)

	if opts.JSON {
		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	}

	return nil
}

func scanAgent(workDir, agent, since string) []extractor.Entry {
	memoryDir := filepath.Join(workDir, agent, "memory")
	var entries []extractor.Entry

	files, err := filepath.Glob(filepath.Join(memoryDir, "*.md"))
	if err != nil {
		return entries
	}

	for _, f := range files {
		base := filepath.Base(f)
		// Simple date check: filename is YYYY-MM-DD.md
		if len(base) >= 10 && base[:10] < since {
			continue
		}

		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		extracted := extractor.Extract(string(data), agent, base)
		entries = append(entries, extracted...)
	}

	return entries
}

func topicToID(topic string) string {
	id := strings.ToLower(topic)
	id = strings.ReplaceAll(id, " ", "-")
	// Trim to reasonable length
	if len(id) > 60 {
		id = id[:60]
	}
	return id
}
