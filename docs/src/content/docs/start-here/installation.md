---
title: Installation
description: All install options, prerequisites, update, and uninstall.
---

## macOS / Linux

```sh
curl -fsSL https://raw.githubusercontent.com/eugenelo/no-mistakes/main/docs/install.sh | sh
```

The installer downloads release artifacts from `eugenelo/no-mistakes` by default. Override that with `NO_MISTAKES_RELEASE_REPO=owner/repo` when testing another release source. If this fork does not have release artifacts yet, use the source or existing-wrapper install below.

The installer keeps the real binary in `~/.no-mistakes/bin` and exposes `no-mistakes` through a symlink in `~/.local/bin` or `/usr/local/bin`. That keeps future `no-mistakes update` runs in a user-owned location instead of rewriting a system binary in place.

It also installs or refreshes the background daemon for you by running `no-mistakes daemon restart`, preferring a managed service (launchd on macOS, systemd user service on Linux) and falling back to a detached daemon if that path is unavailable. If the restart fails, the install command fails.

Official release binaries installed this way include the default self-hosted telemetry host and website ID. Disable telemetry with `NO_MISTAKES_TELEMETRY=0`, or override the host and website ID with `NO_MISTAKES_UMAMI_HOST` and `NO_MISTAKES_UMAMI_WEBSITE_ID`.

## Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/eugenelo/no-mistakes/main/docs/install.ps1 | iex
```

Installs the binary and restarts the background daemon automatically with `no-mistakes.exe daemon restart`, preferring a managed Task Scheduler task and falling back to a detached daemon if needed. If the restart fails, the install command fails.

Official release binaries installed this way include the default self-hosted telemetry host and website ID. Disable telemetry with `NO_MISTAKES_TELEMETRY=0`, or override the host and website ID with `NO_MISTAKES_UMAMI_HOST` and `NO_MISTAKES_UMAMI_WEBSITE_ID`.

## Go install

```sh
go install github.com/kunchenguid/no-mistakes/cmd/no-mistakes@latest
```

`go install` builds the CLI without an embedded telemetry website ID, so telemetry stays off by default unless you later set `NO_MISTAKES_UMAMI_WEBSITE_ID` at runtime. This uses the upstream module path; use the source install below for fork-specific patches.

## From source

```sh
git clone git@github.com:eugenelo/no-mistakes.git
cd no-mistakes
make install VERSION=dev-fork
```

`make build` embeds the telemetry host from `NO_MISTAKES_UMAMI_HOST` in a repo-local `.env` first, then `UMAMI_HOST` from the shell, then the default self-hosted host. It embeds the telemetry website ID from `NO_MISTAKES_UMAMI_WEBSITE_ID` in `.env` first, then `UMAMI_WEBSITE_ID` from the shell, then the default website ID.

### Existing wrapper install

If your existing install has `~/.local/bin/no-mistakes` pointing at a wrapper script in `~/.no-mistakes/bin/no-mistakes`, with the actual binary at `~/.no-mistakes/bin/no-mistakes.real`, replace only the real binary after active runs finish:

```sh
git clone git@github.com:eugenelo/no-mistakes.git
cd no-mistakes
make build VERSION=dev-fork
no-mistakes daemon stop
install -m 755 bin/no-mistakes ~/.no-mistakes/bin/no-mistakes.real
no-mistakes daemon start
```

Building with a development version such as `dev-fork` disables `no-mistakes update`, so the upstream updater cannot overwrite the forked binary.

### Updating an existing fork install

After you change the fork, rebuild and replace the real binary. Do this only when there are no `pending` or `running` runs, because restarting the daemon can interrupt active pipelines. If this checkout has a machine-local `LOCAL_INSTALL.md`, read it first for host-specific paths; that file is gitignored and should not be committed.

```sh
# Use the local clone of your no-mistakes fork.
FORK_DIR="${NO_MISTAKES_FORK_DIR:-$HOME/repos/no-mistakes}"

# Use the Go toolchain on PATH, or prepend a locally installed Go if needed.
# Example: PATH="$HOME/.local/share/go1.25.0/bin:$PATH"
cd "$FORK_DIR"
git pull --ff-only
make build VERSION=dev-fork

# Safety check: from any repo initialized with no-mistakes, list recent runs
# and wait until none are pending/running.
no-mistakes runs --limit 10

no-mistakes daemon stop
install -m 755 "$FORK_DIR/bin/no-mistakes" "$HOME/.no-mistakes/bin/no-mistakes.real"
no-mistakes daemon start
no-mistakes --version
```

If you are not inside a repo initialized with `no-mistakes`, use the state database to check for active runs before restarting:

```sh
python3 - <<'PY'
import os, sqlite3
path = os.path.expanduser('~/.no-mistakes/state.sqlite')
conn = sqlite3.connect(path)
for run_id, branch, head_sha, status in conn.execute(
    "select id, branch, substr(head_sha,1,8), status "
    "from runs where status in ('pending','running') "
    "order by created_at desc"
):
    print(status, branch, run_id, head_sha)
PY
```

The command path stays unchanged:

```text
~/.local/bin/no-mistakes -> ~/.no-mistakes/bin/no-mistakes -> ~/.no-mistakes/bin/no-mistakes.real
```

## Prerequisites

- **git** - required
- **One supported agent binary** - `claude`, `codex`, `acli` (Rovo Dev), `opencode`, or `pi`, or a separately installed `acpx` binary for `agent: acp:<target>`
- **Optional, for PRs and CI:**
  - `gh` CLI (GitHub)
  - `glab` CLI (GitLab)
  - `NO_MISTAKES_BITBUCKET_EMAIL` and `NO_MISTAKES_BITBUCKET_API_TOKEN` (Bitbucket Cloud)

Run `no-mistakes doctor` to check native agents and provider tools.
For ACP agents, verify `acpx` or `acpx_path` separately because `doctor` does not validate ACP targets.

See [Provider Integration](/no-mistakes/guides/provider-integration/) for PR and CI setup per host.

## Update

```sh
no-mistakes update
no-mistakes update --beta
no-mistakes update -y
```

This downloads the latest release from `eugenelo/no-mistakes`, verifies the SHA-256 checksum, atomically replaces the binary, and resets the daemon so it picks up the new executable. It prefers the managed service path and falls back to a detached daemon if service startup is unavailable or fails.

`no-mistakes update` installs the latest stable release from this fork.
Use `no-mistakes update --beta` to opt into prereleases and install the latest beta when one is newer than the current stable release.
Use `no-mistakes update -y` to answer yes to update safety prompts.

Because `update` installs the latest fork release binary, it installs a binary with the default self-hosted telemetry host and website ID. Disable telemetry with `NO_MISTAKES_TELEMETRY=0`, or override the host and website ID with `NO_MISTAKES_UMAMI_HOST` and `NO_MISTAKES_UMAMI_WEBSITE_ID`.

If pending or running pipeline runs exist, the update warns that restarting the daemon can cause those runs to fail, prints each active run's ID, status, branch, and short head SHA, and prompts before continuing.
If the running daemon was started from a different binary, the update prompts before replacing it.
Pass `-y` or `--yes` to continue through these prompts while still printing warnings.
If the daemon executable path cannot be determined, the update aborts before replacing the binary.
If the daemon does not come back cleanly after a successful replacement, the new binary stays installed but the command reports the daemon reset failure.

Background update checks run automatically on each CLI invocation (except `update` itself). Suppress with `NO_MISTAKES_NO_UPDATE_CHECK=1`.

## Remove from a repo

```sh
no-mistakes eject
```

Removes the `no-mistakes` remote, deletes the bare repo, cleans up worktrees, and removes the database record.
It does not remove repo-local agent skill files created by `no-mistakes init`.

## Uninstall

Stop the daemon, delete the binary, and clear state:

```sh
no-mistakes daemon stop
rm -f ~/.local/bin/no-mistakes /usr/local/bin/no-mistakes
rm -rf ~/.no-mistakes
```

On macOS, also remove `~/Library/LaunchAgents/com.kunchenguid.no-mistakes.daemon.*.plist`. On Linux, also remove `~/.config/systemd/user/no-mistakes-daemon-*.service`. On Windows, remove the `no-mistakes-daemon-*` Task Scheduler task.
