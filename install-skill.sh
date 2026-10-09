#!/bin/sh
# Installs the stremcli skill for Claude Code (~/.claude/skills) and Codex
# (~/.agents/skills), then installs the stremcli binary.
#
#   curl -fsSL https://raw.githubusercontent.com/mebn/stremcli/main/install-skill.sh | sh
set -eu

REPO="mebn/stremcli"
RAW="https://raw.githubusercontent.com/$REPO/main/skills/stremcli"
FILES="SKILL.md scripts/ensure-installed.sh"

install_to() {
	dest="$1/stremcli"
	mkdir -p "$dest/scripts"
	for f in $FILES; do
		curl -fsSL "$RAW/$f" -o "$dest/$f"
	done
	chmod +x "$dest/scripts/ensure-installed.sh"
	echo "Installed skill to $dest"
}

install_to "$HOME/.claude/skills"
install_to "$HOME/.agents/skills"

bin="$(sh "$HOME/.claude/skills/stremcli/scripts/ensure-installed.sh")"
bin_dir="$(dirname "$bin")"
case ":$PATH:" in
	*":$bin_dir:"*) ;;
	*)
		echo
		echo "Note: $bin_dir is not on your PATH. Add it to use stremcli directly:"
		echo "  echo 'export PATH=\"$bin_dir:\$PATH\"' >> ~/.zshrc"
		;;
esac

echo
echo "Done. Save your Real-Debrid token (https://real-debrid.com/apitoken) with:"
echo "  stremcli config -token <KEY>"
echo "or just ask Claude/Codex to play something; they'll ask you for it."
