---
name: stremcli
description: Stream a movie or TV episode through Real-Debrid and open it in IINA, VLC or mpv using the stremcli command-line tool. Use when the user wants to watch, play or stream a movie or show, get a direct streaming link for a title, pick a torrent by quality, or see what they have watched recently.
---

# stremcli

`stremcli` finds a movie or TV episode by name, gets a torrent that is cached
on Real-Debrid, turns it into a direct streaming link and optionally opens it
in a media player. Source: https://github.com/mebn/stremcli

## 1. Make sure it's installed

Always run this first. It installs stremcli if needed, then prints the binary's
absolute path:

```sh
sh "<this skill's directory>/scripts/ensure-installed.sh"
```

Use the printed path to run stremcli (it may not be on `PATH`). Below,
`stremcli` means that path.

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

## History

Every play is recorded. Show the most recent first with:

```sh
stremcli history        # last 20
stremcli history -n 0   # everything
```

Use this to answer questions like "what did I watch last?" or to continue a
show with the next episode. Raw data is in `~/.stremcli/history.jsonl`.
