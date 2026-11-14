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

check_model_map() {
  local file="$1" prefix nul_line diagnostics diagnostic
  [ -f "$file" ] || { err "$file: missing or not a regular file"; return; }
  # Check before awk: BSD awk treats NUL as a string terminator. Comments and
  # otherwise ignored bytes must obey the same NUL rule as the model resolver.
  if LC_ALL=C IFS= read -r -d '' prefix < "$file"; then
    nul_line="$(printf '%s' "$prefix" | LC_ALL=C wc -l)"
    err "$file:$((nul_line + 1)): NUL byte is not allowed"
    return
  fi
  if ! diagnostics="$(LC_ALL=C awk -v label="$file" '
    function bad(message) {
      printf "%s:%d: %s\n", label, NR, message
      failed = 1
      exit
    }
    /^[[:space:]]*(#|$)/ { next }
    {
      eq = index($0, "=")
      name = substr($0, 1, eq - 1)
      value = substr($0, eq + 1)
      if (eq == 0 || name !~ /^(claude|codex)\.(fast|balanced|flagship)$/ ||
          value !~ /^[[:graph:]]+$/ || index(value, "=")) {
        bad("invalid or empty model mapping")
      }
      if (name in first_line) {
        bad("duplicate key " name " (first defined on line " first_line[name] ")")
      }
      first_line[name] = NR
      entries++
    }
    END {
      if (!failed && entries == 0) printf "%s: no model mappings\n", label
    }
  ' "$file")"; then
    err "$file: could not read model mappings"
    return
  fi
  if [ -n "$diagnostics" ]; then
    while IFS= read -r diagnostic; do err "$diagnostic"; done <<< "$diagnostics"
  fi
}

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
check_model_map "$map_file"
if [ -e agents/model-map.local.conf ] || [ -L agents/model-map.local.conf ]; then
  check_model_map agents/model-map.local.conf
fi
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

# Tool-specific root pointers must remain thin links to the one source of truth.
[ "$(cat CLAUDE.md)" = '@AGENTS.md' ] \
  || err "CLAUDE.md: expected exactly '@AGENTS.md'"
# Match literal Markdown backticks, not shell command substitution.
# shellcheck disable=SC2016
grep -q 'Read `AGENTS.md`' GEMINI.md \
  || err "GEMINI.md: missing AGENTS.md pointer"

# ---------------------------------------------------------------------------
info "== Shell syntax =="
for shell_file in scripts/*.sh global-agents/*.sh tests/*.sh; do
  bash -n "$shell_file" || err "$shell_file: shell syntax check failed"
done

# ---------------------------------------------------------------------------
info "== Relative links =="
# Extract Markdown links, skip external/anchor/mail, resolve the rest on disk.
# Inline-code paths and reference-style links are outside this check's scope.
# Set decoded_path directly: command substitution would strip filename newlines.
# Decode valid triplets once; malformed escapes remain literal filesystem bytes.
decode_link_path() {
  local raw="$1" hex byte LC_ALL=C
  decoded_path=''
  while [ -n "$raw" ]; do
    if [[ "$raw" == %??* ]]; then
      hex="${raw:1:2}"
      case "$hex" in
        [0-9a-fA-F][0-9a-fA-F])
          [ "$hex" != 00 ] || return 1
          printf -v byte '\\%03o' "$((16#$hex))"
          printf -v byte '%b' "$byte"
          decoded_path+="$byte"
          raw="${raw:3}"
          continue ;;
      esac
    fi
    decoded_path+="${raw:0:1}"
    raw="${raw:1}"
  done
}

list_md() {
  # NUL records preserve literal paths instead of Git's quoted display names.
  git ls-files -z --cached --others --exclude-standard -- '*.md' 2>/dev/null \
    || find . -type f -name '*.md' -not -path './.git/*' -print0
}
while IFS= read -r -d '' md; do
  [ -f "$md" ] || continue
  # Command substitution would strip trailing newlines from a parent directory.
  case "$md" in */*) dir="${md%/*}" ;; *) dir=. ;; esac
  grep -oE '\]\([^)]+\)' -- "$md" 2>/dev/null | sed -E 's/^\]\(//; s/\)$//' | while IFS= read -r target; do
    # Keep the extracted spelling for diagnostics; unwrap only before decoding.
    destination="$target"
    case "$destination" in
      \<*\>) destination="${destination#<}"; destination="${destination%>}" ;;
    esac
    case "$destination" in http://*|https://*|mailto:*|\#*|"") continue ;; esac
    path="${destination%%#*}"
    path="${path%%\?*}"
    [ -z "$path" ] && continue
    if ! decode_link_path "$path"; then
      err "$md: NUL byte in link path -> $target"
      continue
    fi
    path="$decoded_path"
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
