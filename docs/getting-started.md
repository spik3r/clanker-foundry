# Getting started

The skill pack can be used on its own. Project Memory, the model-driven example
and the orchestration runtime are optional.

## Install the pack

### Quick start (recommended)

Use Git, Bash and GNU Stow to create every personal skill symlink and wire the shared global agent
instructions. Keep this repository in a permanent location; do not run Stow from a
temporary clone.

1. Install GNU Stow with your system package manager. On macOS with Homebrew:

   ```bash
   brew install stow
   ```

2. Clone this repository, then enter it:

   ```bash
   git clone https://github.com/spik3r/clanker-foundry.git ~/code/clanker-foundry
   cd ~/code/clanker-foundry
   ```

3. Review [scripts/stow.sh](../scripts/stow.sh), then run it:

   ```bash
   scripts/stow.sh
   ```

4. Start a new agent session. For both `~/.agents/` and `~/.claude/`, the script has
   installed `skills/`, `agents/`, `workflows/`, `templates/`, and `checklists/` as
   symlinks to this repository. It has also stowed the tracked global instructions to
   `${XDG_CONFIG_HOME:-$HOME/.config}/agents/AGENTS.md` and wired Codex, Claude Code, opencode, and Gemini CLI
   to that canonical link.

   The subagent profiles use portable model tiers. Resolve their current client-specific
   model through `scripts/agent-model.sh <claude|codex> <agent-name>`; see the
   [subagent model mapping](../agents/README.md#model-selection).

The script creates missing target directories and uses `stow --restow`, so it is safe
to run again after pulling updates. Repository-only `docs/` and `scripts/` are not
installed. The script is not transactional and may stop after some links are created if an
existing non-Stow file conflicts with a target; inspect and resolve the conflict rather than using Stow's `--adopt` option blindly.

Older installs may have a copied `${XDG_CONFIG_HOME:-$HOME/.config}/agents/AGENTS.md`. If it matches the tracked
file, the script moves it to `AGENTS.md.pre-stow` before creating the link. If it differs,
the script stops so you can merge those local changes into `global-agents/AGENTS.md`.
A differing legacy copy or an already occupied backup path for a matching copy is
rejected before any Stow call. The script rechecks before the late migration; this
does not make installation transactional or protect against concurrent changes.
Commit and pull changes to the tracked file to share them across machines.

### Project-scoped use

For a team or a single project, use the [pinned shared-base setup](project-shared-base.md).
It keeps a reviewed `.agent-baseline` submodule, a short local instruction overlay
and selected relative skill links under project version control. It makes no
home-directory changes and requires no Stow installation. Workflows, templates and
checklists can be referenced through the submodule; placing them in a client-named
directory does not itself guarantee discovery.

### One-off manual link

For a temporary or single-skill installation, run these commands from this
repository. The checks refuse an existing target, including a dangling symlink;
inspect it rather than replacing it. Keep the checkout in place while linked.

```bash
mkdir -p "$HOME/.agents/skills" "$HOME/.claude/agents"

# Add one skill to Codex, only if the destination is absent.
target="$HOME/.agents/skills/engineer"
test ! -e "$target" && test ! -L "$target" && ln -s "$PWD/skills/engineer" "$target"

# Add one Claude Code subagent profile, only if the destination is absent.
target="$HOME/.claude/agents/explorer-researcher.md"
test ! -e "$target" && test ! -L "$target" && ln -s "$PWD/agents/explorer-researcher.md" "$target"
```

Later, after confirming these are the links you created, remove them with:

```bash
unlink "$HOME/.agents/skills/engineer"
unlink "$HOME/.claude/agents/explorer-researcher.md"
```

Do not manually replace a link managed by Stow; use `scripts/stow.sh` after changing
the pack instead. `unlink` will not remove a directory.

### Global instruction conflicts

The canonical path is `${XDG_CONFIG_HOME:-$HOME/.config}/agents/AGENTS.md`;
Codex's user path respects `CODEX_HOME`. Existing regular instruction files or
symlinks pointing elsewhere for Codex, opencode or Gemini stop wiring. A
non-directory in a required canonical or client directory path stops preflight
before Stow runs. Symlinks to real directories remain supported. An existing
regular Claude `CLAUDE.md` is preserved and receives an import, but a Claude
symlink or directory is refused. Review [the wiring script](../global-agents/wire-global-agents.sh)
and the [global agent guide](../global-agents/README.md) before resolving conflicts.
A successful link or import alone does not prove a client used the context; see
[context portability](agent-context-portability.md).

## Validation

Run the validator after any change:

```bash
scripts/validate.sh
```

It checks that the folders follow the Agent Skills open format:

- directory name matches the `name`;
- names use lowercase letters and hyphens;
- each `SKILL.md` contains `name` and `description` frontmatter;
- all skills remain well below the recommended 500-line limit;
- each skill ships a `README.md`, model mappings are nonempty, shell scripts parse, and
  every relative Markdown link in tracked or unignored untracked Markdown resolves on disk.
  One literal outer `<...>` destination pair is removed before URL handling.
  External `http://`, `https://` and `mailto:` prefixes ignore ASCII letter case.
  Link paths decode valid `%HH` escapes once, after removing the raw query and fragment.
  Encoded question marks stay in filenames. Plus signs and malformed escapes stay
  literal; encoded NUL bytes in paths are refused. Queries are not decoded.
  Inline links are checked; anchors, reference-style links, and paths in inline code are
  outside this check.

CI runs the same script on pull requests, pushes to `main`, all tag pushes and
manual runs. Other branch pushes wait for a pull request. When the `skills-ref`
validator is installed, both the script and CI additionally run it per skill:

```bash
skills-ref validate ./skills/engineer
```

Installer regression tests run with:

```bash
bash tests/test-install.sh
```
