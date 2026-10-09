# stremcli

Stream movies and TV episodes through [Real-Debrid](https://real-debrid.com)
from the terminal, straight into IINA, VLC or mpv.

stremcli looks up the title with Stremio's Cinemeta catalog, finds torrents with
[Torrentio](https://torrentio.strem.fun), and asks Real-Debrid for a direct link
to the first one that's cached.

## Install

### With Claude Code or Codex (easiest)

Tell your agent:

> install https://github.com/mebn/stremcli

It installs both the `stremcli` CLI and the stremcli skill, so afterwards you
can just say things like "play Dune in 1080p in IINA" or "play the next
episode of The Office".

### Manually

```sh
curl -fsSL https://raw.githubusercontent.com/mebn/stremcli/main/install-skill.sh | sh
```

This installs the skill into `~/.claude/skills/stremcli` (Claude Code) and
`~/.agents/skills/stremcli` (Codex), and the CLI via `go install` or, without
Go, a prebuilt binary in `~/.local/bin`.

Only want the CLI?

```sh
go install github.com/mebn/stremcli@latest
```

or download a binary from the [latest release](https://github.com/mebn/stremcli/releases/latest).

### Set your Real-Debrid key

Once, with your API key from https://real-debrid.com/apitoken:

```sh
stremcli config -token <KEY>
stremcli config -player iina   # optional default player
stremcli config -quality 1080p # optional preferred quality
```

Running `stremcli config` with no flags asks for the key interactively.

### Instructions for AI agents

If a user asks you to install this repository:

1. Run `curl -fsSL https://raw.githubusercontent.com/mebn/stremcli/main/install-skill.sh | sh`.
   It installs the skill for both Claude Code and Codex and installs the
   `stremcli` binary. Report any PATH note it prints to the user.
2. Ask the user for their Real-Debrid API key (from
   https://real-debrid.com/apitoken) and save it with
   `stremcli config -token <KEY>`. Never run `stremcli config` without
   `-token`, because it waits for input on stdin.
3. Tell the user to restart their session so the new skill is loaded.

## Usage

```sh
stremcli -p iina "The Matrix"
stremcli -t tv -s 1 -e 3 -p vlc "Breaking Bad"
stremcli -q 1080p -l "Dune"              # list numbered results
stremcli -q 1080p -pick 3 -p iina "Dune" # play result 3
stremcli continue                        # next episode of the last show
stremcli continue -p iina mobland        # next episode of MobLand
stremcli history                         # what you've watched
```

| Flag | Meaning |
|------|---------|
| `-type`, `-t` | `movie` (default) or `tv` |
| `-season`, `-s` / `-episode`, `-e` | required for `tv` |
| `-player`, `-p` | `iina`, `vlc` or `mpv`; without it the link is only printed |
| `-quality`, `-q` | `4k`, `1080p`, `720p` or `480p` |
| `-list`, `-l` | print numbered results and exit |
| `-pick N` | use only result N from `-list` |

The link is printed to stdout, so `stremcli "Dune" | pbcopy` works too.
`$REAL_DEBRID_TOKEN` overrides the saved token.

With IINA or mpv, stremcli remembers how far you got: playing the same movie
or episode again resumes there, and `stremcli continue` picks up a
half-watched episode before moving on to the next one. (VLC can't report its
position, so episodes played in VLC count as watched.)

Settings live in `~/.stremcli/config.json`, watch history in
`~/.stremcli/history.jsonl` and playback positions in
`~/.stremcli/progress.json`.
