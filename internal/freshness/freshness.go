package freshness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Run scans vault notes and reports those not modified within the threshold.
func Run(vaultPath string, days int, folder string, jsonOut bool) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	var stale []staleNote

	searchDir := vaultPath
	if folder != "" {
		searchDir = filepath.Join(vaultPath, folder)
	}

	files, err := filepath.Glob(filepath.Join(searchDir, "*.md"))
	if err != nil {
		return fmt.Errorf("scanning vault: %w", err)
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		modified := parseModified(string(data))
		if modified.IsZero() {
			// Fall back to file mtime
			info, _ := os.Stat(f)
			if info != nil {
				modified = info.ModTime()
			}
		}

		if modified.Before(cutoff) {
			daysOld := int(time.Since(modified).Hours() / 24)
			name := strings.TrimSuffix(filepath.Base(f), ".md")
			stale = append(stale, staleNote{
				Name:     name,
				DaysOld:  daysOld,
				Modified: modified.Format("2006-01-02"),
			})
		}
	}

	if jsonOut {
		printJSON(stale)
	} else {
		printReport(stale, days)
	}
	return nil
}

type staleNote struct {
	Name     string `json:"name"`
	DaysOld  int    `json:"days_old"`
	Modified string `json:"modified"`
}

func printReport(notes []staleNote, threshold int) {
	fmt.Printf("📋 Vault freshness — as of %s\n", time.Now().Format("2006-01-02"))
	fmt.Printf("   Threshold: %d days\n\n", threshold)

	var warning []staleNote
	var critical []staleNote

	for _, n := range notes {
		if n.DaysOld >= 60 {
			critical = append(critical, n)
		} else {
			warning = append(warning, n)
		}
	}

	if len(warning) > 0 {
		fmt.Printf("⚠️  %d-%d days stale (%d notes)\n", threshold, 59, len(warning))
		for _, n := range warning {
			fmt.Printf("    %s (%d days, modified %s)\n", n.Name, n.DaysOld, n.Modified)
		}
		fmt.Println()
	}

	if len(critical) > 0 {
		fmt.Printf("🔴 60+ days stale (%d notes)\n", len(critical))
		for _, n := range critical {
			fmt.Printf("    %s (%d days, modified %s)\n", n.Name, n.DaysOld, n.Modified)
		}
		fmt.Println()
	}

	if len(notes) == 0 {
		fmt.Println("✅ All notes are fresh!")
	} else {
		fmt.Println("Run: flywheel verify <note> to mark as current")
	}
}

func printJSON(notes []staleNote) {
	fmt.Print("[")
	for i, n := range notes {
		if i > 0 {
			fmt.Print(",")
		}
		fmt.Printf(`{"name":"%s","days_old":%d,"modified":"%s"}`, n.Name, n.DaysOld, n.Modified)
	}
	fmt.Println("]")
}

func parseModified(content string) time.Time {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "modified:") {
			val := strings.TrimSpace(trimmed[len("modified:"):])
			for _, fmt := range []string{
				"2006-01-02 15:04",
				"2006-01-02T15:04:05",
				"2006-01-02",
			} {
				if t, err := time.Parse(fmt, val); err == nil {
					return t
				}
			}
		}
	}
	return time.Time{}
}
