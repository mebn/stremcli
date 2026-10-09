# stremcli

Stream movies and TV episodes through [Real-Debrid](https://real-debrid.com)
from the terminal, straight into IINA, VLC or mpv.

stremcli looks up the title with Stremio's Cinemeta catalog, finds torrents with
[Torrentio](https://torrentio.strem.fun), and asks Real-Debrid for a direct link
to the first one that's cached.

## Install

```sh
go install github.com/mebn/stremcli/cmd/stremcli@latest
```

or download a binary from the [latest release](https://github.com/mebn/stremcli/releases/latest).

Then save your Real-Debrid API token (from https://real-debrid.com/apitoken) once:

```sh
stremcli config -token <KEY>
stremcli config -player iina   # optional default player
```

## Usage

```sh
stremcli -p iina "The Matrix"
stremcli -t tv -s 1 -e 3 -p vlc "Breaking Bad"
stremcli -q 1080p -l "Dune"              # list numbered results
stremcli -q 1080p -pick 3 -p iina "Dune" # play result 3
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

Settings live in `~/.stremcli/config.json` and watch history in
`~/.stremcli/history.jsonl`.

## Claude Code / Codex skill

Let Claude Code or Codex play things for you ("play the next episode of The
Office in IINA"):

```sh
curl -fsSL https://raw.githubusercontent.com/mebn/stremcli/main/install-skill.sh | sh
```

This installs the skill into `~/.claude/skills/stremcli` and
`~/.agents/skills/stremcli`, and installs the `stremcli` binary. The skill
also installs the binary on first use if it's missing.
