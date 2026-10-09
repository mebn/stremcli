# stremcli

Stream movies and TV episodes through [Real-Debrid](https://real-debrid.com)
into IINA, VLC or mpv. It's built to be driven by an AI agent: you say "play
Dune in 1080p" or "continue watching", and the agent runs stremcli.

stremcli looks up the title with Stremio's Cinemeta catalog, finds torrents with
[Torrentio](https://torrentio.strem.fun), and asks Real-Debrid for a direct link
to the first one that's cached.

## Install

**Claude Code:** install the plugin, which adds the stremcli skill:

```
/plugin marketplace add mebn/stremcli
/plugin install stremcli@stremcli
```

The skill downloads the CLI the first time it's needed. Update the skill with
`/plugin marketplace update stremcli`.

**Codex:** copy the skill into your skills folder:

```sh
mkdir -p ~/.agents/skills/stremcli
curl -fsSL https://raw.githubusercontent.com/mebn/stremcli/main/skills/stremcli/SKILL.md \
  -o ~/.agents/skills/stremcli/SKILL.md
```

**CLI only:** `go install github.com/mebn/stremcli@latest`, or download a
binary from the [latest release](https://github.com/mebn/stremcli/releases/latest).
Update with `stremcli update`.

Then save your Real-Debrid API key from https://real-debrid.com/apitoken with
`stremcli config -token <KEY>` (or just `stremcli config` to be prompted).

### Instructions for AI agents

If a user asks you to install this repository:

1. In Claude Code, run
   `claude plugin marketplace add mebn/stremcli && claude plugin install stremcli@stremcli`.
   In Codex, copy the skill as shown above.
2. Install the CLI as described under "Setup" in
   [the skill](skills/stremcli/SKILL.md).
3. Ask for the user's Real-Debrid API key and save it with
   `stremcli config -token <KEY>`. Never run `stremcli config` without flags;
   it waits for input.
4. Tell the user to restart their session so the skill is loaded.

## Usage

```sh
stremcli -p iina "The Matrix"
stremcli -t tv -s 1 -e 3 -p vlc "Breaking Bad"
stremcli -t tv -s 1 -e 3 -start 41:12 -p iina "Breaking Bad"
stremcli -q 1080p -l "Dune"              # list numbered results
stremcli -q 1080p -pick 3 -p iina "Dune" # play result 3
stremcli history                         # what you've watched, and how far
stremcli config -player iina -quality 1080p
```

Run `stremcli help` for every flag. The link is printed to stdout, so
`stremcli "Dune" | pbcopy` works too.

Everything is stored in `~/.stremcli/`: `config.json` for settings and
`history.jsonl` for what you've watched. With IINA and mpv, each history entry
also records where playback stopped, which the agent uses to resume or move on
to the next episode.
