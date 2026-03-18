# flywheel

> Agents learn daily. Vault notes go stale. Flywheel closes the loop.

A Go CLI that syncs tagged learnings from agent daily logs into your Obsidian vault.

## How It Works

Agents mark learnings in their daily logs:

```markdown
[LEARNING] topic | the actual learning
[UPDATE] topic | something that changed
[STALE] topic | this note is outdated
```

Flywheel reads those, matches them to vault notes, and creates/updates accordingly.

## Install

```bash
go install github.com/cyperx84/flywheel/cmd/flywheel@latest
# or
brew install cyperx84/tap/flywheel
```

## Commands

### `flywheel sync`

```bash
flywheel sync                          # today's logs, all agents
flywheel sync --since 2026-03-15       # date range
flywheel sync --agent builder          # single agent
flywheel sync --dry-run                # preview only
flywheel sync --json                   # machine-readable output
```

### `flywheel freshness`

```bash
flywheel freshness                     # notes stale 30+ days
flywheel freshness --days 60           # custom threshold
flywheel freshness --json
```

### `flywheel verify`

```bash
flywheel verify "ollama-setup"         # mark as current
flywheel verify --all                  # verify everything
```

## Config

`~/.config/flywheel/config.json` (optional — has sane defaults):

```json
{
  "agents": ["builder", "researcher", "ops"],
  "workspace": "~/.openclaw/agents",
  "vault": "~/path/to/obsidian/vault",
  "freshness_days": 30
}
```

## Dependencies

- `obsidian-cli` for vault writes
- That's it. No database, no LLM calls, no runtime deps.

## License

MIT
