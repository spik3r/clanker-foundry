#!/usr/bin/env bash
# Validate the pack against the conventions in AGENTS.md.
# Checks skills, agent profiles, and relative-link integrity across all Markdown.
# Exit status is non-zero if any check fails. Safe to run locally or in CI.
set -uo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir" || exit 1

failures="$(mktemp)"
trap 'rm -f "$failures"' EXIT

err()  { printf 'FAIL  %s\n' "$1" >&2; echo x >> "$failures"; }
info() { printf '%s\n' "$1"; }

# Read a single-line frontmatter field from a file's leading `---` block.
frontmatter_field() {
  local file="$1" key="$2"
  awk -v key="$key" '
    NR == 1 && $0 != "---" { exit }
    NR > 1 && $0 == "---"  { exit }
    NR > 1 {
      if (match($0, "^" key ":[[:space:]]*")) {
        print substr($0, RLENGTH + 1)
        exit
      }
    }
  ' "$file"
}

# ---------------------------------------------------------------------------
info "== Skills =="
for dir in skills/*/; do
  [ -d "$dir" ] || continue
  name_dir="$(basename "$dir")"
  skill="$dir/SKILL.md"
  readme="$dir/README.md"

  if [ ! -f "$skill" ]; then
    err "$name_dir: missing SKILL.md"
    continue
  fi

  fm_name="$(frontmatter_field "$skill" name)"
  fm_desc="$(frontmatter_field "$skill" description)"

  [ -n "$fm_name" ] || err "$name_dir: SKILL.md has no 'name' frontmatter"
  [ -n "$fm_desc" ] || err "$name_dir: SKILL.md has no 'description' frontmatter"
  [ "$fm_name" = "$name_dir" ] || err "$name_dir: frontmatter name '$fm_name' != directory"

  printf '%s' "$name_dir" | grep -Eq '^[a-z0-9]+(-[a-z0-9]+)*$' \
    || err "$name_dir: name must be lowercase letters, digits, and hyphens"

  lines="$(wc -l < "$skill" | tr -d ' ')"
  [ "$lines" -lt 500 ] || err "$name_dir: SKILL.md is $lines lines (limit 500)"

  [ -f "$readme" ] || err "$name_dir: missing README.md"
done

# ---------------------------------------------------------------------------
info "== Agent profiles =="
valid_tiers="fast balanced flagship"
map_file="agents/model-map.conf"
model_tier() {
  awk '
    /^Suggested model tier: / {
      tier = $0
      sub(/^Suggested model tier: /, "", tier)
      tick = sprintf("%c", 96)
      if (tier ~ "^" tick "[^" tick "]+" tick "$") {
        gsub(tick, "", tier)
        print tier
        exit
      }
    }
  ' "$1"
}

for profile in agents/*.md; do
  [ -f "$profile" ] || continue
  base="$(basename "$profile")"
  [ "$base" = "README.md" ] && continue

  [ -n "$(frontmatter_field "$profile" name)" ]        || err "$base: no 'name' frontmatter"
  [ -n "$(frontmatter_field "$profile" description)" ] || err "$base: no 'description' frontmatter"

  tier="$(model_tier "$profile")"
  if [ -z "$tier" ]; then
    err "$base: no 'Suggested model tier' line"
    continue
  fi
  case " $valid_tiers " in
    *" $tier "*) ;;
    *) err "$base: tier '$tier' is not one of: $valid_tiers" ;;
  esac
  for client in claude codex; do
    grep -q "^$client\.$tier=" "$map_file" \
      || err "$base: no mapping '$client.$tier' in $map_file"
  done
done

# ---------------------------------------------------------------------------
info "== Relative links =="
# Extract [text](target) links, skip external/anchor/mail, resolve the rest on disk.
list_md() { git ls-files '*.md' 2>/dev/null || find . -name '*.md' -not -path './.git/*'; }
while IFS= read -r md; do
  [ -f "$md" ] || continue
  dir="$(dirname "$md")"
  grep -oE '\]\([^)]+\)' "$md" 2>/dev/null | sed -E 's/^\]\(//; s/\)$//' | while IFS= read -r target; do
    case "$target" in http://*|https://*|mailto:*|\#*|"") continue ;; esac
    path="${target%%#*}"
    [ -z "$path" ] && continue
    case "$path" in
      /*) resolved="$repo_dir$path" ;;
      *)  resolved="$dir/$path" ;;
    esac
    [ -e "$resolved" ] || err "$md: broken link -> $target"
  done
done < <(list_md)

# ---------------------------------------------------------------------------
info "== skills-ref (optional) =="
if command -v skills-ref >/dev/null 2>&1; then
  for dir in skills/*/; do
    skills-ref validate "$dir" || err "skills-ref rejected $dir"
  done
else
  info "skills-ref not installed; skipping format validation"
fi

# ---------------------------------------------------------------------------
count="$(wc -l < "$failures" | tr -d ' ')"
printf '\n'
if [ "$count" -eq 0 ]; then
  info "PASS  no problems found"
  exit 0
fi
printf 'FAILED  %s problem(s)\n' "$count" >&2
exit 1
