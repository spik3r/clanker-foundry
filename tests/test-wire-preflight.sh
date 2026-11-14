#!/usr/bin/env bash
# Isolated instruction-target checks; no client launch or global installation.
# An optional helper path permits negative controls against the unchanged helper.
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
helper="${1:-$repo_dir/global-agents/wire-global-agents.sh}"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

# Stop content tools before opening a forbidden fixture, even on the old helper.
# This seam makes an attempted FIFO access fail rather than block the test.
REAL_TOUCH="$(command -v touch)"
REAL_GREP="$(command -v grep)"
REAL_TAIL="$(command -v tail)"
export REAL_TOUCH REAL_GREP REAL_TAIL
mkdir -p "$test_root/bin"
cat > "$test_root/bin/content-tool" <<'TOOL'
#!/usr/bin/env bash
set -euo pipefail
for argument in "$@"; do
  if [ -n "$FORBIDDEN_TARGET" ] && [ "$argument" = "$FORBIDDEN_TARGET" ]; then
    printf '%s\n' "${0##*/}" >> "$CONTENT_TRACE"
    exit 72
  fi
done
case "${0##*/}" in
  touch) exec "$REAL_TOUCH" "$@" ;;
  grep) exec "$REAL_GREP" "$@" ;;
  tail) exec "$REAL_TAIL" "$@" ;;
esac
exit 74
TOOL
chmod +x "$test_root/bin/content-tool"
for tool in touch grep tail; do ln -s content-tool "$test_root/bin/$tool"; done
export PATH="$test_root/bin:$PATH"

new_case() {
  printf 'CHECK %s\n' "$1"
  case_root="$test_root/$1 with spaces"
  export HOME="$case_root/home"
  export XDG_CONFIG_HOME="$case_root/config"
  export CODEX_HOME="$case_root/codex"
  export CONTENT_TRACE="$case_root/content-calls"
  export FORBIDDEN_TARGET=''
  canon="$XDG_CONFIG_HOME/agents/AGENTS.md"
  claude="$HOME/.claude/CLAUDE.md"
  mkdir -p "${canon%/*}" "${claude%/*}"
  printf 'synthetic canonical instructions\n' > "$canon"
  : > "$CONTENT_TRACE"
}

run_helper() {
  local expected="$1" actual=0
  shift
  bash "$helper" "$@" > "$case_root/stdout" 2> "$case_root/stderr" || actual=$?
  if [ "$actual" -ne "$expected" ]; then
    cat "$case_root/stderr" >&2
    printf 'Expected exit %s, got %s.\n' "$expected" "$actual" >&2
    exit 1
  fi
}

assert_no_links() {
  test ! -e "$CODEX_HOME"
  test ! -e "$XDG_CONFIG_HOME/opencode"
  test ! -e "$HOME/.gemini"
}

for kind in fifo directory symlink dangling; do
  new_case "$kind-target"
  case "$kind" in
    fifo) mkfifo "$claude" ;;
    directory) mkdir "$claude"; printf 'keep\n' > "$claude/keep" ;;
    symlink)
      printf 'keep\n' > "$case_root/retained"
      ln -s "$case_root/retained" "$claude" ;;
    dangling) ln -s "$case_root/missing" "$claude" ;;
  esac
  export FORBIDDEN_TARGET="$claude"
  for mode in --check wire; do
    if [ "$mode" = --check ]; then run_helper 1 --check; else run_helper 1; fi
    grep -qF 'existing target must be a regular file' "$case_root/stderr"
    assert_no_links
    test ! -s "$CONTENT_TRACE"
  done
  case "$kind" in
    fifo) test -p "$claude" ;;
    directory) grep -qxF keep "$claude/keep" ;;
    symlink)
      test "$(readlink "$claude")" = "$case_root/retained"
      grep -qxF keep "$case_root/retained" ;;
    dangling) test "$(readlink "$claude")" = "$case_root/missing" ;;
  esac
done

for kind in absent empty unterminated terminated imported; do
  new_case "$kind-regular-target"
  case "$kind" in
    absent) ;;
    empty) : > "$claude" ;;
    unterminated) printf 'existing memory' > "$claude" ;;
    terminated) printf 'existing memory\n' > "$claude" ;;
    imported) printf 'existing memory\n@%s\n' "$canon" > "$claude" ;;
  esac
  if [ -e "$claude" ]; then cp "$claude" "$case_root/before"; fi
  run_helper 0 --check
  assert_no_links
  if [ "$kind" = absent ]; then
    test ! -e "$claude"
  else
    cmp "$case_root/before" "$claude"
  fi
  : > "$case_root/expected"
  case "$kind" in
    unterminated|terminated|imported) printf 'existing memory\n' > "$case_root/expected" ;;
  esac
  printf '@%s\n' "$canon" >> "$case_root/expected"
  run_helper 0
  cmp "$case_root/expected" "$claude"
  test "$(readlink "$CODEX_HOME/AGENTS.md")" = "$canon"
  test "$(readlink "$XDG_CONFIG_HOME/opencode/AGENTS.md")" = "$canon"
  test "$(readlink "$HOME/.gemini/GEMINI.md")" = "$canon"
  run_helper 0
  cmp "$case_root/expected" "$claude"
  test ! -s "$CONTENT_TRACE"
  grep -qxF 'synthetic canonical instructions' "$canon"
done

# A missing child below an existing non-directory is not a creatable path.
for parent in codex claude opencode gemini nested trailing dot canonical; do
  for kind in file dangling; do
    new_case "$parent-$kind-parent"
    case "$parent" in
      codex) blocker="$CODEX_HOME" ;;
      claude) blocker="$HOME/.claude"; rmdir "$blocker" ;;
      opencode) blocker="$XDG_CONFIG_HOME/opencode" ;;
      gemini) blocker="$HOME/.gemini" ;;
      nested) blocker="$case_root/blocked"; export CODEX_HOME="$blocker/child" ;;
      trailing) blocker="$case_root/blocked"; export CODEX_HOME="$blocker/" ;;
      dot) blocker="$case_root/blocked"; export CODEX_HOME="$blocker/." ;;
      canonical)
        blocker="$XDG_CONFIG_HOME/agents"
        rm "$canon"; rmdir "$blocker" ;;
    esac
    case "$kind" in
      file) printf 'retained parent bytes\n' > "$blocker" ;;
      dangling) ln -s "$case_root/missing" "$blocker" ;;
    esac
    before_dirs="$(find "$case_root" -type d | LC_ALL=C sort)"
    for mode in --check wire; do
      if [ "$mode" = --check ]; then run_helper 1 --check; else run_helper 1; fi
      grep -qF 'directory component is not a directory' "$case_root/stderr"
      test "$(find "$case_root" -type d | LC_ALL=C sort)" = "$before_dirs"
      test ! -s "$CONTENT_TRACE"
      for target in "$CODEX_HOME/AGENTS.md" "$HOME/.claude/CLAUDE.md" \
        "$XDG_CONFIG_HOME/opencode/AGENTS.md" "$HOME/.gemini/GEMINI.md"; do
        test ! -e "$target" && test ! -L "$target"
      done
      case "$kind" in
        file) test -f "$blocker"; grep -qxF 'retained parent bytes' "$blocker" ;;
        dangling) test "$(readlink "$blocker")" = "$case_root/missing" ;;
      esac
    done
  done
done

# Existing aliases to real directories remain supported, including the canonical
# parent. All aliases and their targets are owned by this isolated fixture.
new_case directory-symlink-parents
mkdir -p "$case_root/real/codex" "$case_root/real/claude" \
  "$case_root/real/opencode" "$case_root/real/gemini"
mv "$XDG_CONFIG_HOME/agents" "$case_root/real/agents"
ln -s "$case_root/real/agents" "$XDG_CONFIG_HOME/agents"
ln -s "$case_root/real/codex" "$CODEX_HOME"
rmdir "$HOME/.claude"
ln -s "$case_root/real/claude" "$HOME/.claude"
ln -s "$case_root/real/opencode" "$XDG_CONFIG_HOME/opencode"
ln -s "$case_root/real/gemini" "$HOME/.gemini"
before_dirs="$(find "$case_root" -type d | LC_ALL=C sort)"
run_helper 0 --check
test "$(find "$case_root" -type d | LC_ALL=C sort)" = "$before_dirs"
for target in "$CODEX_HOME/AGENTS.md" "$HOME/.claude/CLAUDE.md" \
  "$XDG_CONFIG_HOME/opencode/AGENTS.md" "$HOME/.gemini/GEMINI.md"; do
  test ! -e "$target" && test ! -L "$target"
done
run_helper 0
run_helper 0
test "$(readlink "$CODEX_HOME/AGENTS.md")" = "$canon"
test "$(readlink "$XDG_CONFIG_HOME/opencode/AGENTS.md")" = "$canon"
test "$(readlink "$HOME/.gemini/GEMINI.md")" = "$canon"
grep -qxF "@$canon" "$HOME/.claude/CLAUDE.md"
grep -qxF 'synthetic canonical instructions' "$canon"
test -L "$XDG_CONFIG_HOME/agents"
test -L "$CODEX_HOME"
test -L "$HOME/.claude"
test -L "$XDG_CONFIG_HOME/opencode"
test -L "$HOME/.gemini"

printf '%s\n' 'PASS  wiring target and directory preflight cases'
