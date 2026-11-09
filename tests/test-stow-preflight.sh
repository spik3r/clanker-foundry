#!/usr/bin/env bash
# Test installer ordering with a recording seam, not GNU Stow link behavior.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

mkdir -p "$test_root/bin"
cat > "$test_root/bin/stow" <<'STOW'
#!/usr/bin/env bash
# Test-owned seam. It never creates package links or invokes GNU Stow.
set -euo pipefail
package=''
for argument in "$@"; do package="$argument"; done
printf '%s\n' "$package" >> "$STOW_TRACE"
if [ "$(wc -l < "$STOW_TRACE" | tr -d ' ')" -eq 1 ]; then
  case "$STOW_MODE" in
    fail-first) exit 72 ;;
    edit-canonical) printf 'late synthetic edit\n' >> "$STOW_CANON" ;;
    create-backup) printf 'late synthetic backup\n' > "$STOW_CANON.pre-stow" ;;
  esac
fi
# Stop before wiring: reaching this boundary is not a successful installation.
if [ "$package" = global-agents ]; then exit 73; fi
STOW
chmod +x "$test_root/bin/stow"
export PATH="$test_root/bin:$PATH"

new_case() {
  printf 'CHECK %s\n' "$1"
  case_root="$test_root/$1 with spaces"
  export HOME="$case_root/home"
  export XDG_CONFIG_HOME="$case_root/config"
  export CODEX_HOME="$case_root/codex"
  export STOW_TRACE="$case_root/stow-calls"
  export STOW_CANON="$XDG_CONFIG_HOME/agents/AGENTS.md"
  export STOW_MODE=record
  backup="$STOW_CANON.pre-stow"
  mkdir -p "$HOME" "$(dirname "$STOW_CANON")"
  cp "$repo_dir/global-agents/AGENTS.md" "$STOW_CANON"
  : > "$STOW_TRACE"
}

run_installer() {
  local expected="$1" actual
  if bash "$repo_dir/scripts/stow.sh" > "$case_root/stdout" 2> "$case_root/stderr"; then
    actual=0
  else
    actual=$?
  fi
  if [ "$actual" -ne "$expected" ]; then
    cat "$case_root/stderr" >&2
    printf 'Expected exit %s, got %s.\n' "$expected" "$actual" >&2
    exit 1
  fi
}

assert_calls() {
  test "$(wc -l < "$STOW_TRACE" | tr -d ' ')" -eq "$1"
}

assert_no_shared_targets() {
  test ! -e "$HOME/.agents"
  test ! -e "$HOME/.claude"
  test ! -e "$HOME/.gemini"
  test ! -e "$XDG_CONFIG_HOME/opencode"
}

assert_no_backup() {
  test ! -e "$backup"
  test ! -L "$backup"
}

new_case edited-copy
printf 'synthetic local edit\n' >> "$STOW_CANON"
cp "$STOW_CANON" "$case_root/expected"
run_installer 1
grep -qF 'contains local changes.' "$case_root/stderr"
assert_calls 0
assert_no_shared_targets
assert_no_backup
cmp "$case_root/expected" "$STOW_CANON"

for kind in file directory symlink dangling; do
  new_case "existing-backup-$kind"
  case "$kind" in
    file) printf 'synthetic backup\n' > "$backup" ;;
    directory) mkdir "$backup"; printf 'synthetic backup\n' > "$backup/keep" ;;
    symlink)
      printf 'synthetic backup\n' > "$case_root/retained"
      ln -s "$case_root/retained" "$backup" ;;
    dangling) ln -s "$case_root/missing" "$backup" ;;
  esac
  run_installer 1
  grep -qF 'backup already exists' "$case_root/stderr"
  assert_calls 0
  assert_no_shared_targets
  test ! -L "$STOW_CANON"
  cmp "$repo_dir/global-agents/AGENTS.md" "$STOW_CANON"
  case "$kind" in
    file) grep -qxF 'synthetic backup' "$backup" ;;
    directory) grep -qxF 'synthetic backup' "$backup/keep" ;;
    symlink)
      test "$(readlink "$backup")" = "$case_root/retained"
      grep -qxF 'synthetic backup' "$case_root/retained" ;;
    dangling) test "$(readlink "$backup")" = "$case_root/missing" ;;
  esac
done

new_case wiring-conflict
mkdir -p "$CODEX_HOME"
printf 'synthetic Codex content\n' > "$CODEX_HOME/AGENTS.md"
run_installer 1
assert_calls 0
assert_no_shared_targets
assert_no_backup
grep -qxF 'synthetic Codex content' "$CODEX_HOME/AGENTS.md"
cmp "$repo_dir/global-agents/AGENTS.md" "$STOW_CANON"

new_case matching-copy
run_installer 73
assert_calls 11
test "$(tail -n 1 "$STOW_TRACE")" = global-agents
test ! -e "$STOW_CANON"
cmp "$repo_dir/global-agents/AGENTS.md" "$backup"

for state in absent linked; do
  new_case "$state-canonical"
  rm "$STOW_CANON"
  if [ "$state" = linked ]; then
    ln -s "$repo_dir/global-agents/AGENTS.md" "$STOW_CANON"
  fi
  # An old backup only blocks migration of a regular legacy copy.
  printf 'retained backup\n' > "$backup"
  run_installer 73
  assert_calls 11
  test "$(tail -n 1 "$STOW_TRACE")" = global-agents
  grep -qxF 'retained backup' "$backup"
  if [ "$state" = linked ]; then
    test "$(readlink "$STOW_CANON")" = "$repo_dir/global-agents/AGENTS.md"
  else
    test ! -e "$STOW_CANON"
    test ! -L "$STOW_CANON"
  fi
done

new_case stow-failure
export STOW_MODE=fail-first
run_installer 72
assert_calls 1
assert_no_backup
cmp "$repo_dir/global-agents/AGENTS.md" "$STOW_CANON"

new_case late-canonical-edit
export STOW_MODE=edit-canonical
cp "$STOW_CANON" "$case_root/expected"
printf 'late synthetic edit\n' >> "$case_root/expected"
run_installer 1
assert_calls 10
assert_no_backup
cmp "$case_root/expected" "$STOW_CANON"
grep -qF 'contains local changes.' "$case_root/stderr"

new_case late-backup
export STOW_MODE=create-backup
run_installer 1
assert_calls 10
cmp "$repo_dir/global-agents/AGENTS.md" "$STOW_CANON"
grep -qxF 'late synthetic backup' "$backup"
grep -qF 'backup already exists' "$case_root/stderr"

printf '%s\n' 'PASS  12 installer ordering cases (recording seam, not GNU Stow acceptance)'
