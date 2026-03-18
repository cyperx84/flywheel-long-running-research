package verify

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Run marks a vault note as verified by setting last-verified in frontmatter.
func Run(vaultPath, note string, all bool) error {
	if all {
		return verifyAll(vaultPath)
	}
	return verifyOne(note)
}

func verifyOne(noteID string) error {
	now := time.Now().Format("2006-01-02")
	cmd := exec.Command("obsidian-cli", "frontmatter", noteID, "--edit", "--key", "last-verified", "--value", now)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verify %s: %w\n%s", noteID, err, string(out))
	}
	fmt.Printf("✅ Verified: %s (last-verified: %s)\n", noteID, now)
	return nil
}

func verifyAll(vaultPath string) error {
	files, err := filepath.Glob(filepath.Join(vaultPath, "*.md"))
	if err != nil {
		return fmt.Errorf("scanning vault: %w", err)
	}

	now := time.Now().Format("2006-01-02")
	verified := 0
	errors := 0

	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".md")
		cmd := exec.Command("obsidian-cli", "frontmatter", name, "--edit", "--key", "last-verified", "--value", now)
		if err := cmd.Run(); err != nil {
			errors++
			fmt.Printf("❌ %s: %v\n", name, err)
		} else {
			verified++
			fmt.Printf("✅ %s\n", name)
		}
	}

	fmt.Printf("\nVerified: %d, Errors: %d\n", verified, errors)
	return nil
}
