---
name: stremcli
description: Stream a movie or TV episode through Real-Debrid and open it in IINA, VLC or mpv using the stremcli command-line tool. Use when the user wants to watch, play or stream a movie or show, continue watching something, play the next episode, get a direct streaming link, pick a torrent by quality, or see what they have watched.
---

# stremcli

`stremcli` finds a movie or episode by name, gets a torrent cached on
Real-Debrid, turns it into a direct streaming link and opens it in a player.
It does what it's told; deciding *what* to play (resume, next episode, which
quality) is your job, using the history it keeps.

## Setup

Check that it's installed with `command -v stremcli || ls ~/.local/bin/stremcli`.
If not, download the release binary (macOS and Linux, amd64 and arm64):

```sh
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; esac
mkdir -p ~/.local/bin
curl -fsSL "https://github.com/mebn/stremcli/releases/latest/download/stremcli-$os-$arch" -o ~/.local/bin/stremcli
chmod +x ~/.local/bin/stremcli
```

If `~/.local/bin` isn't on `PATH`, run it by full path and tell the user they
can add `export PATH="$HOME/.local/bin:$PATH"` to their shell profile.

If a command fails with `no Real-Debrid token configured` or
`real-debrid rejected the API token`, ask the user for their API token (from
https://real-debrid.com/apitoken) and run `stremcli config -token <TOKEN>`.
Never run `stremcli config` without flags: it waits for input. Don't echo the
token back or write it anywhere else.

## Playing

```sh
stremcli -p iina "The Matrix"
stremcli -t tv -s 1 -e 3 -p iina "Breaking Bad"
stremcli -t tv -s 1 -e 3 -start 41:12 -p iina "Breaking Bad"  # resume
stremcli -q 1080p -l "Dune"                                     # list results
stremcli -q 1080p -pick 3 -p iina "Dune"                        # play result 3
```

| Flag | Meaning |
|------|---------|
| `-type`, `-t` | `movie` (default) or `tv` |
| `-season`, `-s` / `-episode`, `-e` | required for `tv` |
| `-player`, `-p` | `iina`, `vlc` or `mpv`; without one the link is only printed |
| `-quality`, `-q` | only `4k`, `1080p`, `720p` or `480p` results |
| `-start` | start position, `h:mm:ss` or seconds (IINA and mpv only) |
| `-list`, `-l` | print numbered results and exit |
| `-pick N` | try only result N from `-list` |

- The link goes to stdout, progress to stderr. Without `-pick`, the best
  results are tried in order until one is cached; with `-pick`, only that
  one, so suggest another number if it isn't cached.
- `Found <Title> (<year>) [<imdb id>]` shows which title matched. If it's the
  wrong one, add the year or more words to the query.
- No player named? Use the saved default (`cat ~/.stremcli/config.json`); if
  there is none, ask, or just give the link.
- Saved defaults: `stremcli config -player iina` and
  `stremcli config -quality 1080p` (tried first, other qualities as fallback;
  `-q` overrides it).

## History and continue watching

```sh
stremcli history         # last 20, newest first
stremcli history -n 0    # everything
```

```
2026-10-09 21:14  MobLand (2025-) S01E04  [0:41:12 / 0:58:20]
2026-10-08 22:30  Dune (2021)  [2:28:10 / 2:35:00]
2026-10-07 20:02  Severance (2022) S02E01
```

The `[position / duration]` is where playback stopped, recorded for IINA and
mpv. Entries without it were played in VLC or only as a link, so assume they
were watched. Raw data is in `~/.stremcli/history.jsonl`.

When the user asks to continue watching, or for "the next episode":

1. Pick the entry: the most recent one, or the most recent matching the title
   they named.
2. Decide whether it's finished. Treat it as unfinished if a meaningful part
   is left, i.e. they stopped before the end credits. A few minutes left in a
   movie can still be the ending, so lean towards resuming. If it's really
   unclear, ask.
3. Unfinished: replay the same movie or episode with `-start` a few seconds
   before the saved position.
4. Finished episode: play the next one (`-e` + 1). If that finds no torrents,
   try episode 1 of the next season. If that fails too, the user is probably
   caught up; say so.
5. Finished movie: tell the user and ask what they'd like to watch.

Say what you're doing, e.g. "Resuming MobLand S01E04 at 41:12".

## Updating

`stremcli update` replaces the binary with the latest release
(`stremcli update -check` only reports). This skill updates with the Claude
Code plugin: `claude plugin marketplace update stremcli`, then restart the
session.
