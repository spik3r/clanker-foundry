#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT
baseline="$test_root/baseline source"
mkdir -p "$baseline/scripts" "$baseline/global-agents" "$baseline/templates/project" "$baseline/skills"
cp "$repo_dir/scripts/project-setup.sh" "$baseline/scripts/"
cp "$repo_dir/global-agents/AGENTS.md" "$baseline/global-agents/"
cp "$repo_dir/templates/project/shared-agents.md" "$baseline/templates/project/"
for name in engineer review-change; do cp -R "$repo_dir/skills/$name" "$baseline/skills/"; done
git -C "$baseline" init -q
git -C "$baseline" add .
git -C "$baseline" -c user.name=Test -c user.email=test@example.invalid commit -qm baseline

consumer="$test_root/consumer with spaces"
git init -q "$consumer"
git -C "$consumer" -c protocol.file.allow=always submodule add -q "$baseline" .agent-baseline
cd "$consumer"
setup() { bash .agent-baseline/scripts/project-setup.sh "$@"; }
snapshot() {
  find . -print | LC_ALL=C sort
  find . -type f -exec cksum {} \; | LC_ALL=C sort
  find . -type l -exec readlink {} \; | LC_ALL=C sort
}
reject_without_changes() {
  snapshot > "$test_root/before"
  if setup "$@" > "$test_root/output" 2>&1; then
    printf 'FAIL  accepted invalid setup: %s\n' "$*" >&2; exit 1
  fi
  snapshot > "$test_root/after"
  cmp "$test_root/before" "$test_root/after"
}

# Missing wiring, invalid names, a missing later skill and duplicates are read-only.
reject_without_changes --check
reject_without_changes engineer ../outside
reject_without_changes engineer missing-skill
reject_without_changes engineer engineer
reject_without_changes --unknown
test ! -e AGENTS.md
test ! -e .agents

# A gitlink without matching staged submodule metadata cannot survive a clean clone.
cp .gitmodules "$test_root/modules"
rm .gitmodules
reject_without_changes engineer
cp "$test_root/modules" .gitmodules
printf '\n# unstaged edit\n' >> .gitmodules
reject_without_changes engineer
cp "$test_root/modules" .gitmodules
git config --file .gitmodules submodule..agent-baseline.path wrong-path
git add .gitmodules
reject_without_changes engineer
cp "$test_root/modules" .gitmodules
git add .gitmodules
git config --file .gitmodules --add submodule..agent-baseline.path wrong-path
git add .gitmodules
reject_without_changes engineer
cp "$test_root/modules" .gitmodules
git add .gitmodules

# A late destination conflict must not create the earlier skill or any overlay.
mkdir -p .claude/skills
printf 'keep\n' > .claude/skills/review-change
reject_without_changes engineer review-change
test ! -e .agents
grep -qxF keep .claude/skills/review-change
rm .claude/skills/review-change

# Existing instructions are never replaced, including dangling symlinks.
printf 'local rules\n' > AGENTS.md
reject_without_changes engineer
rm AGENTS.md
ln -s "$test_root/missing" GEMINI.md
reject_without_changes engineer
rm GEMINI.md
printf 'shadow rules\n' > AGENTS.override.md
reject_without_changes engineer
rm AGENTS.override.md

# Reject parents pointing outside the consumer, even when the final link is absent.
mkdir "$test_root/outside"
ln -s "$test_root/outside" .agents
reject_without_changes engineer
test -z "$(ls -A "$test_root/outside")"
rm .agents
mkdir .agents
ln -s "$test_root/outside" .agents/skills
reject_without_changes engineer
rm .agents/skills

# Baseline modifications, a missing submodule and a checkout at the wrong pin fail.
printf 'dirty\n' >> .agent-baseline/global-agents/AGENTS.md
reject_without_changes engineer
git -C .agent-baseline restore global-agents/AGENTS.md
mv .agent-baseline "$test_root/parked-baseline"
ln -s "$test_root/parked-baseline" .agent-baseline
reject_without_changes engineer
rm .agent-baseline
mkdir .agent-baseline
reject_without_changes engineer
rmdir .agent-baseline
mv "$test_root/parked-baseline" .agent-baseline
old_pin="$(git -C .agent-baseline rev-parse HEAD)"
git -C .agent-baseline -c user.name=Test -c user.email=test@example.invalid commit --allow-empty -qm next
reject_without_changes engineer
git -C .agent-baseline checkout -q "$old_pin"

setup engineer review-change
for client in .agents .claude; do
  for name in engineer review-change; do
    test "$(readlink "$client/skills/$name")" = "../../.agent-baseline/skills/$name"
    test -f "$client/skills/$name/SKILL.md"
  done
done
grep -qxF '@.agent-baseline/global-agents/AGENTS.md' CLAUDE.md
grep -qxF '@AGENTS.md' GEMINI.md
if grep -q 'Managed orchestration scope' AGENTS.md; then
  printf '%s\n' 'FAIL  Foundry repository instructions leaked into the consumer' >&2
  exit 1
fi

# Project customizations survive reruns; verification also preserves Git metadata.
printf '\nLocal project rule.\n' >> AGENTS.md
printf '\nLocal client rule.\n' >> CLAUDE.md
snapshot > "$test_root/before"
setup
setup --check
snapshot > "$test_root/after"
cmp "$test_root/before" "$test_root/after"
reject_without_changes review-change

cp .agents/foundry-skills "$test_root/selection"
printf '../outside\n' > .agents/foundry-skills
reject_without_changes --check
printf 'engineer\n' > .agents/foundry-skills
reject_without_changes --check
reject_without_changes
cp "$test_root/selection" .agents/foundry-skills

# Explicit removal and readdition work without touching a project-owned skill.
mkdir .agents/skills/local-helper
printf 'local procedure\n' > .agents/skills/local-helper/SKILL.md
printf 'review-change\n' > .agents/foundry-skills
unlink .agents/skills/engineer
unlink .claude/skills/engineer
setup
setup --check
grep -qxF 'local procedure' .agents/skills/local-helper/SKILL.md
cp "$test_root/selection" .agents/foundry-skills
setup

# A deleted selected link is detected, then repaired; an unexpected link is refused.
rm .agents/skills/engineer
reject_without_changes --check
setup
rm .agents/skills/engineer
ln -s "$test_root/missing" .agents/skills/engineer
reject_without_changes
rm .agents/skills/engineer
setup

# A clean clone carries the selection and relative links. Initialization is explicit.
git add .gitmodules .agent-baseline .agents .claude AGENTS.md CLAUDE.md GEMINI.md
git -c user.name=Test -c user.email=test@example.invalid commit -qm consumer
git clone -q "$consumer" "$test_root/clean clone"
cd "$test_root/clean clone"
test ! -f .agent-baseline/scripts/project-setup.sh
git -c protocol.file.allow=always submodule update --init -q
bash .agent-baseline/scripts/project-setup.sh --check
test -z "$(git status --porcelain)"

# A reviewed update requires staging the new gitlink; verification never updates it.
git -C "$baseline" -c user.name=Test -c user.email=test@example.invalid commit --allow-empty -qm updated
new_pin="$(git -C "$baseline" rev-parse HEAD)"
git -C .agent-baseline fetch -q origin
git -C .agent-baseline checkout -q "$new_pin"
reject_without_changes --check
git add .agent-baseline
setup --check

# No skills is valid; selection is an empty regular file, with working verification.
cd "$test_root"
git init -q empty-selection
cd empty-selection
git -c protocol.file.allow=always submodule add -q "$baseline" .agent-baseline
setup
test -f .agents/foundry-skills
test ! -s .agents/foundry-skills
setup --check

printf '%s\n' 'PASS  project shared-base safety, clean-clone and update regressions'
