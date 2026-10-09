---
name: stremcli
description: Stream a movie or TV episode through Real-Debrid and open it in IINA, VLC or mpv using the stremcli command-line tool. Use when the user wants to watch, play or stream a movie or show, get a direct streaming link for a title, pick a torrent by quality, continue watching a show (play the next episode), or see what they have watched recently.
---

# stremcli

`stremcli` finds a movie or TV episode by name, gets a torrent that is cached
on Real-Debrid, turns it into a direct streaming link and optionally opens it
in a media player. Source: https://github.com/mebn/stremcli

## 1. Make sure it's installed

Always check first:

```sh
command -v stremcli || ls ~/.local/bin/stremcli
```

If neither finds it, download the latest release binary from GitHub into
`~/.local/bin`. Releases exist for macOS and Linux on amd64 and arm64:

```sh
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; esac
mkdir -p ~/.local/bin
curl -fsSL "https://github.com/mebn/stremcli/releases/latest/download/stremcli-$os-$arch" -o ~/.local/bin/stremcli
chmod +x ~/.local/bin/stremcli
```

If `~/.local/bin` isn't on `PATH`, run it as `~/.local/bin/stremcli` and tell
the user they can add `export PATH="$HOME/.local/bin:$PATH"` to their shell
profile. Below, `stremcli` means whichever path works.

## 2. Make sure a Real-Debrid token is set

If a command fails with `no Real-Debrid token configured` or
`real-debrid rejected the API token`, ask the user for their API token
(found at https://real-debrid.com/apitoken) and save it:

```sh
stremcli config -token <TOKEN>
```

Never run `stremcli config` without `-token`, because it then waits for input
on stdin. Don't echo the token back to the user or write it anywhere else.

## 3. Find and play

| Flag | Meaning |
|------|---------|
| `-type`, `-t` | `movie` (default) or `tv` |
| `-season`, `-s` / `-episode`, `-e` | required for `tv` |
| `-player`, `-p` | `iina`, `vlc` or `mpv`; without it the link is only printed |
| `-quality`, `-q` | `4k`, `1080p`, `720p` or `480p` |
| `-list`, `-l` | print numbered results and exit (no token needed) |
| `-pick N` | use only result N from `-list` |

```sh
stremcli -p iina "The Matrix"
stremcli -t tv -s 1 -e 3 -p vlc "Breaking Bad"
stremcli -q 1080p -l "Dune"              # show the options
stremcli -q 1080p -pick 3 -p iina "Dune" # play option 3
```

Details:
- The direct link is printed to stdout. Progress goes to stderr.
- Without `-pick`, it tries the best results in order and uses the first one
  cached on Real-Debrid. With `-pick`, it tries only that result and fails if
  it isn't cached; if so, suggest another number.
- The first output line, `Found <Title> (<year>) [<imdb id>]`, shows which
  title matched. If it's the wrong one (a remake, say), add the year or more
  words to the query.
- If the user doesn't name a player, check the saved default with
  `cat ~/.stremcli/config.json`. If none is set, ask which player they want,
  or just give them the link.
- Set a default player with `stremcli config -player iina`.
- If the user asks to update stremcli, run `stremcli update`. It replaces the
  binary with the latest release. In Claude Code, this skill is updated
  through the plugin: `claude plugin marketplace update stremcli`, then
  restart the session.
- Set a preferred quality with `stremcli config -quality 1080p`. Results in
  that quality are tried first, then the rest. `-quality` on a play command
  overrides it and uses only that quality.

## History

Every play is recorded. Show the most recent first with:

```sh
stremcli history        # last 20
stremcli history -n 0   # everything
```

Use this to answer questions like "what did I watch last?". Raw data is in
`~/.stremcli/history.jsonl`.

## Continue watching

When the user says "continue watching" or asks for the next episode, run:

```sh
stremcli continue            # next episode of the most recently watched show
stremcli continue mobland    # next episode of a specific show in history
```

It takes the same `-player`, `-quality`, `-list` and `-pick` flags as a play
command, rolls over to the next season, and says so if the user is caught up.

With IINA or mpv, stremcli records how far the user got. `continue` resumes a
half-watched episode instead of skipping it, and replaying a movie or episode
resumes where it stopped. `stremcli history` shows `[12:34 / 45:00]` or
`[finished]` next to tracked entries.
