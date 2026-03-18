package extractor

import "testing"

func TestExtract(t *testing.T) {
	text := `# 2026-03-18

[LEARNING] ollama-setup | M4 Mac Mini handles 8 models with 16GB RAM
[UPDATE] clawforge | v2.0.0 re-architected from tmux util to fleet manager
[STALE] openclaw-setup | Install guide references v0.8, we're on v2026.3

Some regular text here.
[LEARNING] ghostty-config
`

	entries := Extract(text, "builder", "memory/2026-03-18.md")

	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// First entry
	if entries[0].Tag != TagLearning {
		t.Errorf("expected tag %s, got %s", TagLearning, entries[0].Tag)
	}
	if entries[0].Topic != "ollama-setup" {
		t.Errorf("expected topic ollama-setup, got %s", entries[0].Topic)
	}
	if entries[0].Content != "M4 Mac Mini handles 8 models with 16GB RAM" {
		t.Errorf("unexpected content: %s", entries[0].Content)
	}

	// No pipe separator
	if entries[3].Topic != "ghostty-config" {
		t.Errorf("expected topic ghostty-config, got %s", entries[3].Topic)
	}
}

func TestEmpty(t *testing.T) {
	entries := Extract("no tags here", "test", "test.md")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
