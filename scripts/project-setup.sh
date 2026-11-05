#!/usr/bin/env bash
# Wire a pinned Foundry submodule into the current project, never the user's home.
set -euo pipefail

fail() { printf 'FAIL  %s\n' "$*" >&2; exit 1; }
check=false
case "${1:-}" in
  --help|-h)
    printf '%s\n' 'Usage: bash .agent-baseline/scripts/project-setup.sh [--check] [skill ...]' \
      'Run at the consumer repository root after adding and staging .agent-baseline.' \
      'First setup records selected skills; later runs read .agents/foundry-skills.'
    exit 0 ;;
  --check) check=true; shift ;;
esac

# Read-only Git queries must not refresh the index, fetch objects or run FSMonitor.
export GIT_OPTIONAL_LOCKS=0 GIT_NO_LAZY_FETCH=1 GIT_NO_REPLACE_OBJECTS=1
git_local() { git -c protocol.allow=never -c core.fsmonitor=false "$@"; }
root="$(pwd -P)"
top="$(git_local rev-parse --show-toplevel 2>/dev/null)" || fail 'Run inside a Git repository.'
[ "$root" = "$top" ] || fail 'Run at the consumer repository root.'
baseline="$root/.agent-baseline"
[ -d "$baseline" ] && [ ! -L "$baseline" ] || fail 'Initialize the real .agent-baseline submodule first.'
baseline_top="$(git_local -C "$baseline" rev-parse --show-toplevel 2>/dev/null)" || fail 'Baseline is not initialized.'
[ "$baseline_top" = "$baseline" ] || fail 'Baseline must be its own Git checkout.'
pin="$(git_local ls-files --stage -- .agent-baseline)"
head="$(git_local -C "$baseline" rev-parse --verify HEAD)"
[ "$pin" = "$(printf '160000 %s 0\t.agent-baseline' "$head")" ] \
  || fail 'Baseline HEAD must match the staged gitlink. Review the pin, then git add .agent-baseline.'
[ -f .gitmodules ] && [ ! -L .gitmodules ] || fail 'Missing regular .gitmodules.'
modules_blob="$(git_local hash-object --no-filters .gitmodules)"
modules_entry="$(git_local ls-files --stage -- .gitmodules)"
case "$modules_entry" in
  "$(printf '100644 %s 0\t.gitmodules' "$modules_blob")"|"$(printf '100755 %s 0\t.gitmodules' "$modules_blob")") ;;
  *) fail '.gitmodules must match its staged regular file. Review and stage it first.' ;;
esac
module_key=''
while IFS=' ' read -r key value; do
  if [ "$value" = .agent-baseline ]; then
    [ -z "$module_key" ] || fail 'Duplicate .agent-baseline submodule mappings.'
    module_key="${key%.path}"
  fi
done < <(git_local config --no-includes --file .gitmodules --get-regexp '^submodule\..*\.path$' || true)
[ -n "$module_key" ] || fail 'Missing .agent-baseline path in .gitmodules.'
[ "$(git_local config --no-includes --file .gitmodules --get-all "$module_key.path")" = .agent-baseline ] \
  || fail 'Expected exactly one path for the baseline submodule.'
module_url="$(git_local config --no-includes --file .gitmodules --get-all "$module_key.url")" \
  || fail 'Missing baseline URL in .gitmodules.'
[ -n "$module_url" ] && [[ "$module_url" != *$'\n'* ]] || fail 'Expected one nonempty baseline URL.'
[ -z "$(git_local -C "$baseline" status --porcelain --untracked-files=all --ignore-submodules=none)" ] \
  || fail 'Baseline has local changes. Keep customizations in the consumer project.'

# Reject symlink parents before inspecting or creating anything beneath them.
for dir in .agents .agents/skills .claude .claude/skills; do
  [ ! -L "$dir" ] || fail "$dir is a symlink; refusing to follow it."
  if [ -e "$dir" ]; then
    [ -d "$dir" ] || fail "$dir must be a directory."
  fi
done
for dir in global-agents templates templates/project skills; do
  [ -d "$baseline/$dir" ] && [ ! -L "$baseline/$dir" ] || fail "Invalid baseline directory: $dir"
done
for file in global-agents/AGENTS.md templates/project/shared-agents.md; do
  [ -f "$baseline/$file" ] && [ ! -L "$baseline/$file" ] || fail "Missing regular baseline file: $file"
  git_local -C "$baseline" ls-files --error-unmatch -- "$file" >/dev/null 2>&1 || fail "Untracked baseline file: $file"
done

manifest=.agents/foundry-skills
skills=()
if [ -e "$manifest" ] || [ -L "$manifest" ]; then
  [ -f "$manifest" ] && [ ! -L "$manifest" ] || fail "$manifest must be a regular file."
  while IFS= read -r name || [ -n "$name" ]; do
    skills+=("$name")
  done < "$manifest"
  if [ "$#" -gt 0 ]; then
    [ "$(printf '%s\n' "$@")" = "$(cat "$manifest")" ] \
      || fail "Selection differs from $manifest. Edit that file explicitly before rerunning."
  fi
else
  "$check" && fail "Missing $manifest; run setup first."
  if [ "$#" -gt 0 ]; then skills=("$@"); fi
fi

# Bash 3.2 (the macOS default) needs the guarded expansion for empty arrays.
seen=' '
for name in ${skills[@]+"${skills[@]}"}; do
  [[ "$name" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || fail "Invalid skill name: $name"
  case "$seen" in *" $name "*) fail "Duplicate skill: $name" ;; esac
  seen="$seen$name "
  [ -d "$baseline/skills/$name" ] && [ ! -L "$baseline/skills/$name" ] \
    && [ -f "$baseline/skills/$name/SKILL.md" ] && [ ! -L "$baseline/skills/$name/SKILL.md" ] \
    || fail "Missing regular baseline skill: $name"
  git_local -C "$baseline" ls-files --error-unmatch -- "skills/$name/SKILL.md" >/dev/null 2>&1 \
    || fail "Untracked baseline skill: $name"
  for client in .agents .claude; do
    link="$client/skills/$name"
    target="../../.agent-baseline/skills/$name"
    if [ -e "$link" ] || [ -L "$link" ]; then
      [ -L "$link" ] && [ "$(readlink "$link")" = "$target" ] || fail "Conflict at $link"
    else
      "$check" && fail "Missing skill link: $link"
    fi
  done
done

# Removed selections require explicit, reviewed unlinking; custom local skills stay.
for link in .agents/skills/* .claude/skills/*; do
  [ -L "$link" ] || continue
  case "$(readlink "$link")" in
    ../../.agent-baseline/skills/*)
      name="${link##*/}"
      case "$seen" in *" $name "*) ;; *) fail "Unselected Foundry link: $link. Review and unlink it explicitly." ;; esac ;;
  esac
done

# Literal Markdown backticks are part of the required instruction, not shell code.
# shellcheck disable=SC2016
directive='Read `.agent-baseline/global-agents/AGENTS.md` before applying the project rules below.'
[ ! -e AGENTS.override.md ] && [ ! -L AGENTS.override.md ] \
  || fail 'AGENTS.override.md can shadow AGENTS.md. Reconcile it manually before setup.'
if [ -e AGENTS.md ] || [ -L AGENTS.md ]; then
  [ -f AGENTS.md ] && [ ! -L AGENTS.md ] || fail 'AGENTS.md must be a project-owned regular file.'
  grep -qxF "$directive" AGENTS.md || fail "Preserve AGENTS.md and add this line manually: $directive"
else
  "$check" && fail 'Missing AGENTS.md.'
fi
# Imports load the reusable base, not the Foundry repository's root AGENTS.md.
imports="$(printf '%s\n' '@.agent-baseline/global-agents/AGENTS.md' '@AGENTS.md')"
for file in CLAUDE.md GEMINI.md; do
  if [ -e "$file" ] || [ -L "$file" ]; then
    if ! { [ -f "$file" ] && [ ! -L "$file" ] \
      && grep -qxF '@.agent-baseline/global-agents/AGENTS.md' "$file" \
      && grep -qxF '@AGENTS.md' "$file"; }; then
      fail "Conflict at $file; review and merge the two required imports manually."
    fi
  else
    "$check" && fail "Missing $file."
  fi
done

if "$check"; then
  printf 'PASS  project wiring matches staged baseline %s (client loading is not tested)\n' "$head"
  exit 0
fi

# Full preflight above; no files are replaced. Concurrent writers are unsupported.
mkdir -p .agents/skills .claude/skills
if [ ! -e "$manifest" ]; then
  (set -o noclobber; : > "$manifest")
  if [ "${#skills[@]}" -gt 0 ]; then printf '%s\n' "${skills[@]}" > "$manifest"; fi
fi
if [ ! -e AGENTS.md ]; then
  (set -o noclobber; cat "$baseline/templates/project/shared-agents.md" > AGENTS.md)
fi
for file in CLAUDE.md GEMINI.md; do
  if [ ! -e "$file" ]; then (set -o noclobber; printf '%s\n' "$imports" > "$file"); fi
done
for name in ${skills[@]+"${skills[@]}"}; do
  for client in .agents .claude; do
    link="$client/skills/$name"
    if [ ! -L "$link" ]; then ln -s "../../.agent-baseline/skills/$name" "$link"; fi
  done
done
printf '%s\n' 'Project wiring ready. Fill in AGENTS.md, review the diff, and commit the project files.'
