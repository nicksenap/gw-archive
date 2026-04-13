# gw-archive

Grove plugin that archives workspaces before deletion and revives them later. Zero-copy — uses git stash commits stored as custom refs, with metadata in a single JSONL file.

## Install

```bash
gw plugin install nicksenap/gw-archive
```

## Setup

Add the pre-delete hook to your Grove config:

```bash
gw archive hook install
```

Or manually add to `~/.grove/config.toml`:

```toml
[hooks]
pre_delete = "gw archive save {name} {path} {branch}"
```

Now every `gw delete` automatically archives the workspace first.

## Usage

```bash
# List archived workspaces
gw archive list

# Show archive details
gw archive show <id>

# Revive a workspace (recreates it with uncommitted changes restored)
gw archive revive <id>

# Manually archive a workspace without deleting it
gw archive save <name> <path> <branch>

# Remove an archive
gw archive remove <id>

# Clean up old archives
gw archive prune --older-than 90d
```

## How it works

### Save

When `gw delete` fires the `pre_delete` hook (before any cleanup):

1. For each repo worktree, `git stash create --include-untracked` captures all uncommitted changes as a commit object
2. The stash commit is stored as a custom ref: `refs/grove-archive/<workspace>/<repo>`
3. Workspace metadata is appended to `~/.grove/archives.jsonl`

No files are copied. The stash commit lives in the source repo's object store, and the custom ref keeps it safe from garbage collection.

### Revive

1. Runs `gw create` with the archived workspace name, branch, and repos
2. Applies `git stash apply <ref>` in each worktree to restore uncommitted changes
3. Claude Code sessions reconnect automatically — the worktree path is the same, so `~/.claude/projects/` picks up where it left off

### Custom refs

Archive refs live under `refs/grove-archive/` — invisible to `git branch`, `git tag`, and `git stash list`, but safe from `git gc`. Same pattern GitHub uses for `refs/pull/*/head`.

```bash
# See all archive refs in a repo
git for-each-ref refs/grove-archive/
```

## Storage

The entire archive store is:

- **`~/.grove/archives.jsonl`** — one JSON line per archived workspace (~200 bytes each)
- **Git refs** — one `refs/grove-archive/<ws>/<repo>` per repo with uncommitted changes

Archiving 100 workspaces adds essentially zero disk overhead.
