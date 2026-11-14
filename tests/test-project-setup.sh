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
# A clean tracked filename may contain text that resembles flags on later lines.
index_odd_path=$'notes\nh pretend-assume-unchanged\nS pretend-skip-worktree'
printf 'fixture notes\n' > "$baseline/$index_odd_path"
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

# Bash 3.2 locale ranges may include uppercase H in [a-z]. A normal cached
# entry must remain accepted under such a locale.
if LC_ALL=en_US.UTF-8 bash -c 'case "H normal" in [a-z]\ *) exit 0 ;; *) exit 1 ;; esac' 2>/dev/null; then
  LC_ALL=en_US.UTF-8 setup --check
fi

# Hidden index entries cannot certify a pinned baseline, even when unchanged.
# Both modes must preserve the working tree, index bytes and flag records.
for path in global-agents/AGENTS.md "$index_odd_path"; do
  cp ".agent-baseline/$path" "$test_root/flag-original"
  for flag in assume-unchanged skip-worktree; do
    for state in clean changed; do
      git -C .agent-baseline update-index "--$flag" -- "$path"
      if [ "$state" = changed ]; then
        printf '\nUnreviewed fixture change.\n' >> ".agent-baseline/$path"
      fi
      git -C .agent-baseline ls-files -v -z -- "$path" > "$test_root/flags-before"
      reject_without_changes
      grep -qF 'assume-unchanged or skip-worktree' "$test_root/output"
      reject_without_changes --check
      grep -qF 'assume-unchanged or skip-worktree' "$test_root/output"
      git -C .agent-baseline ls-files -v -z -- "$path" > "$test_root/flags-after"
      cmp "$test_root/flags-before" "$test_root/flags-after"
      git -C .agent-baseline update-index "--no-$flag" -- "$path"
      cp "$test_root/flag-original" ".agent-baseline/$path"
    done
  done
done
setup --check

# A Git error, including after valid NUL-delimited output, must not turn into a
# successful empty flag list. A failed empty status query cannot certify a clean
# tree either. The seam records calls outside the consumer tree.
INDEX_FLAGS_REAL_GIT="$(command -v git)"
export INDEX_FLAGS_REAL_GIT
export INDEX_FLAGS_TRACE="$test_root/index-flags-trace"
mkdir "$test_root/index-flags-bin"
cat > "$test_root/index-flags-bin/git" <<'FAULT'
#!/usr/bin/env bash
set -eu
case " $* " in
  *" ls-files -v -z "*)
    if [ "$INDEX_FLAGS_FAULT" = status ]; then exec "$INDEX_FLAGS_REAL_GIT" "$@"; fi
    printf 'queried\n' >> "$INDEX_FLAGS_TRACE"
    if [ "$INDEX_FLAGS_FAULT" = before ]; then exit 73; fi
    "$INDEX_FLAGS_REAL_GIT" "$@"
    exit 73 ;;
  *" status --porcelain --untracked-files=all --ignore-submodules=none "*)
    if [ "$INDEX_FLAGS_FAULT" = status ]; then
      printf 'status queried\n' >> "$INDEX_FLAGS_TRACE"
      exit 73
    fi ;;
esac
exec "$INDEX_FLAGS_REAL_GIT" "$@"
FAULT
chmod +x "$test_root/index-flags-bin/git"
for phase in before after status; do
  : > "$INDEX_FLAGS_TRACE"
  INDEX_FLAGS_FAULT="$phase" PATH="$test_root/index-flags-bin:$PATH" reject_without_changes
  test -s "$INDEX_FLAGS_TRACE"
  INDEX_FLAGS_FAULT="$phase" PATH="$test_root/index-flags-bin:$PATH" reject_without_changes --check
  test -s "$INDEX_FLAGS_TRACE"
done
setup
setup --check

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

# A locally configured FSMonitor command must not run during helper queries.
cat > "$test_root/fsmonitor-hook" <<'HOOK'
#!/bin/sh
: > "$FS_MONITOR_MARKER"
printf 'token\000/\000'
HOOK
chmod +x "$test_root/fsmonitor-hook"
export FS_MONITOR_MARKER="$test_root/fsmonitor-executed"
git config core.fsmonitor "$test_root/fsmonitor-hook"
git -C .agent-baseline config core.fsmonitor "$test_root/fsmonitor-hook"
# Prove the fixture is effective before checking the helper's command override.
GIT_OPTIONAL_LOCKS=0 git -C .agent-baseline status --porcelain >/dev/null
test -f "$FS_MONITOR_MARKER"
rm "$FS_MONITOR_MARKER"
snapshot > "$test_root/before"
setup
setup --check
test ! -e "$FS_MONITOR_MARKER"
snapshot > "$test_root/after"
cmp "$test_root/before" "$test_root/after"
git config --unset core.fsmonitor
git -C .agent-baseline config --unset core.fsmonitor

# Exercise the complete helper, not an extracted comparison function. All blob
# writes, configured commands and fault injection below belong to these fixtures.
(
  accept_without_changes() {
    snapshot > "$test_root/metadata-before"
    setup "$@" > "$test_root/metadata-output" 2>&1
    snapshot > "$test_root/metadata-after"
    cmp "$test_root/metadata-before" "$test_root/metadata-after"
  }
  metadata_origin="$test_root/metadata origin"
  git init -q "$metadata_origin"
  cd "$metadata_origin"
  git -c protocol.file.allow=always submodule add -q "$baseline" .agent-baseline
  printf '.gitmodules text eol=lf\n' > .gitattributes
  git add .gitattributes .gitmodules .agent-baseline
  setup engineer
  git add AGENTS.md CLAUDE.md GEMINI.md .agents .claude
  git -c user.name=Test -c user.email=test@example.invalid commit -qm 'LF metadata consumer'
  git -c protocol.file.allow=always clone -q --recurse-submodules . "$test_root/metadata LF clone"
  cd "$test_root/metadata LF clone"
  accept_without_changes --check
  test -z "$(git status --porcelain)"
  cd "$metadata_origin"
  printf '.gitmodules text eol=crlf\n' > .gitattributes
  git add .gitattributes
  git -c user.name=Test -c user.email=test@example.invalid commit -qm 'CRLF metadata checkout'
  git -c protocol.file.allow=always clone -q --recurse-submodules . "$test_root/metadata CRLF clone"
  cd "$test_root/metadata CRLF clone"
  git ls-files --eol -- .gitmodules | grep -q 'i/lf.*w/crlf'
  test -z "$(git status --porcelain)"
  accept_without_changes --check
  git add .gitmodules
  accept_without_changes --check

  # The same Git-produced CRLF metadata also permits ordinary first setup.
  rm AGENTS.md CLAUDE.md GEMINI.md .agents/foundry-skills
  unlink .agents/skills/engineer
  unlink .claude/skills/engineer
  cp .gitmodules "$test_root/metadata-preserved"
  setup engineer
  cmp .gitmodules "$test_root/metadata-preserved"
  accept_without_changes --check

  # Construct LF and CRLF expectations independently from the production stream.
  # The two line-ending arguments allow deliberate mixed or unterminated cases.
  write_modules() {
    printf '# %s%s[submodule ".agent-baseline"]%s\tpath = %s%s\turl = %s%s' \
      "$5" "$1" "$1" "$3" "$1" "$4" "$2" > "$6"
  }
  lf=$'\n'; crlf=$'\r\n'
  staged="$test_root/metadata-staged"
  working="$test_root/metadata-working"
  stage_modules() {
    modules_oid="$(git hash-object --no-filters -w "$staged")"
    git update-index --add --cacheinfo "${1:-100644}" "$modules_oid" .gitmodules
  }
  native="$(printf 'native \200\377\303\251 \\ %%')"
  for ending in final missing extra; do
    stage_end="$lf"; work_end="$crlf"
    case "$ending" in missing) stage_end=''; work_end='' ;; extra) stage_end="$lf$lf$lf"; work_end="$crlf$crlf$crlf" ;; esac
    write_modules "$lf" "$stage_end" .agent-baseline "$baseline" "$native" "$staged"
    write_modules "$crlf" "$work_end" .agent-baseline "$baseline" "$native" "$working"
    stage_modules
    cp "$working" .gitmodules
    accept_without_changes --check
  done
  write_modules "$lf" "$lf" .agent-baseline "$baseline" original "$staged"
  write_modules "$crlf" "$crlf" .agent-baseline "$baseline" original "$working"
  stage_modules 100755
  cp "$working" .gitmodules
  accept_without_changes --check
  stage_modules

  for eol in "$lf" "$crlf"; do
    for change in url path comment whitespace; do
      url="$baseline"; path=.agent-baseline; comment=original
      case "$change" in url) url='../different-source' ;; path) path=other-baseline ;; comment) comment=changed ;; whitespace) comment='original ' ;; esac
      write_modules "$eol" "$eol" "$path" "$url" "$comment" .gitmodules
      reject_without_changes --check
    done
  done
  for ending in "$lf" $'\r' $'\r\r\n' ''; do
    write_modules "$crlf" "$ending" .agent-baseline "$baseline" original .gitmodules
    reject_without_changes --check
  done
  for extra in nul extra newline bom native; do
    cp "$working" .gitmodules
    case "$extra" in
      nul) printf '\000' >> .gitmodules ;;
      extra) printf x >> .gitmodules ;;
      newline) printf '\r\n' >> .gitmodules ;;
      bom) { printf '\357\273\277'; cat "$working"; } > .gitmodules ;;
      native) write_modules "$crlf" "$crlf" .agent-baseline "$baseline" "$(printf '\377')" .gitmodules ;;
    esac
    reject_without_changes --check
  done
  # A staged CR or NUL cannot be discarded to gain fallback eligibility.
  cp "$staged" "$test_root/metadata-normal"
  cp "$working" "$staged"
  stage_modules
  cp "$working" .gitmodules
  accept_without_changes --check # Existing exact-CRLF behavior stays unchanged.
  cp "$test_root/metadata-normal" .gitmodules
  reject_without_changes --check
  cp "$test_root/metadata-normal" "$staged"
  printf '# NUL\000\n' >> "$staged"
  stage_modules
  cp "$working" .gitmodules
  reject_without_changes --check
  cp "$test_root/metadata-normal" "$staged"
  stage_modules

  # Do not narrow a conflicting index listing to one selected entry.
  git update-index --cacheinfo 160000 "$(git -C .agent-baseline rev-parse HEAD)" .gitmodules
  reject_without_changes --check
  git update-index --force-remove .gitmodules
  reject_without_changes --check
  printf '100644 %s 1\t.gitmodules\n100644 %s 2\t.gitmodules\n' "$modules_oid" "$modules_oid" | git update-index --index-info
  reject_without_changes --check
  git update-index --force-remove .gitmodules
  stage_modules
  mv .gitmodules "$test_root/metadata-link-target"
  ln -s "$test_root/metadata-link-target" .gitmodules
  reject_without_changes --check
  unlink .gitmodules
  cp "$working" .gitmodules

  # An effective clean filter can hide changed values even from --no-textconv
  # diff. The helper must neither trust that result nor run configured commands.
  export MODULES_FILTER_MARKER="$test_root/metadata-filter-marker"
  export MODULES_FILTER_SOURCE="$staged"
  cat > "$test_root/metadata-command" <<'COMMAND'
#!/bin/sh
printf '%s\n' invoked >> "$MODULES_FILTER_MARKER"
cat "$MODULES_FILTER_SOURCE"
COMMAND
  chmod +x "$test_root/metadata-command"
  printf '.gitmodules filter=metadata diff=metadata\n' > .git/info/attributes
  git config filter.metadata.clean "\"$test_root/metadata-command\""
  git config filter.metadata.smudge "\"$test_root/metadata-command\""
  git config diff.metadata.textconv "\"$test_root/metadata-command\""
  git config diff.external "\"$test_root/metadata-command\""
  write_modules "$lf" "$lf" .agent-baseline ../unreviewed original .gitmodules
  test "$(git hash-object --path=.gitmodules .gitmodules)" = "$modules_oid"
  git diff --quiet --no-ext-diff --no-textconv -- .gitmodules
  test -s "$MODULES_FILTER_MARKER"
  rm "$MODULES_FILTER_MARKER"
  reject_without_changes --check
  test ! -e "$MODULES_FILTER_MARKER"
  cp "$working" .gitmodules
  git config filter.metadata.process "\"$test_root/metadata-command\""
  accept_without_changes --check
  test ! -e "$MODULES_FILTER_MARKER"
  rm .git/info/attributes
  for key in filter.metadata.clean filter.metadata.smudge filter.metadata.process diff.metadata.textconv diff.external; do git config --unset "$key"; done

  # Replacement refs must not alter the staged object used for comparison.
  replacement="$(printf 'not the reviewed metadata\n' | git hash-object -w --stdin)"
  git replace "$modules_oid" "$replacement"
  accept_without_changes --check
  git replace -d "$modules_oid" >/dev/null

  # A real local promisor can supply the absent blob. Prove lazy fetch works in
  # a separate control copy, then require raw equality to retain its old result
  # and fallback to fail without any child process or repository write.
  promisor="$test_root/metadata-promisor.git"
  git init -q --bare "$promisor"
  write_modules "$lf" "$lf" .agent-baseline "$baseline" promisor "$staged"
  write_modules "$crlf" "$crlf" .agent-baseline "$baseline" promisor "$working"
  missing_oid="$(git -C "$promisor" hash-object --no-filters -w "$staged")"
  git -C "$promisor" update-ref refs/fixture/metadata "$missing_oid"
  git update-index --cacheinfo 100644 "$missing_oid" .gitmodules
  git config remote.fixture.url "$promisor"
  git config remote.fixture.promisor true
  git config extensions.partialClone fixture
  git config protocol.file.allow always
  test ! -e ".git/objects/${missing_oid:0:2}/${missing_oid:2}"
  cp -R . "$test_root/metadata-lazy-control"
  (
    unset GIT_NO_LAZY_FETCH GIT_ALLOW_PROTOCOL
    GIT_TRACE2_EVENT="$test_root/metadata-fetch-control.trace" git -C "$test_root/metadata-lazy-control" cat-file blob "$missing_oid" > "$test_root/metadata-fetched"
  )
  cmp "$staged" "$test_root/metadata-fetched"
  grep -q '"event":"child_start"' "$test_root/metadata-fetch-control.trace"
  cp "$staged" .gitmodules
  GIT_TRACE2_EVENT="$test_root/metadata-raw.trace" accept_without_changes --check
  trace_status=0
  grep -q '"event":"child_start"' "$test_root/metadata-raw.trace" || trace_status=$?
  if [ "$trace_status" -ne 1 ]; then
    printf 'FAIL  raw trace contains a child process or cannot be read (grep status %s)\n' "$trace_status" >&2
    exit 1
  fi
  cp "$working" .gitmodules
  GIT_TRACE2_EVENT="$test_root/metadata-fallback.trace" reject_without_changes --check
  trace_status=0
  grep -q '"event":"child_start"' "$test_root/metadata-fallback.trace" || trace_status=$?
  if [ "$trace_status" -ne 1 ]; then
    printf 'FAIL  fallback trace contains a child process or cannot be read (grep status %s)\n' "$trace_status" >&2
    exit 1
  fi
  test ! -e ".git/objects/${missing_oid:0:2}/${missing_oid:2}"
  git config --unset extensions.partialClone
  git config --remove-section remote.fixture
  git config --unset protocol.file.allow
  stage_modules

  # Valid output followed by a producer/hash/query error must not become success.
  METADATA_REAL_GIT="$(command -v git)"
  export METADATA_REAL_GIT
  export METADATA_FAULT_COUNT="$test_root/metadata-fault-count"
  mkdir "$test_root/metadata-fault-bin"
  cat > "$test_root/metadata-fault-bin/git" <<'FAULT'
#!/usr/bin/env bash
set -u
kind=''
case " $* " in
  *" cat-file blob "*) kind=cat ;;
  *" hash-object --no-filters --stdin "*) kind=hash ;;
  *" ls-files --stage -- .gitmodules "*) kind=entry ;;
esac
if [ "$kind" = "$METADATA_FAULT_KIND" ]; then
  count=0
  if [ -f "$METADATA_FAULT_COUNT" ]; then read -r count < "$METADATA_FAULT_COUNT"; fi
  count=$((count + 1)); printf '%s\n' "$count" > "$METADATA_FAULT_COUNT"
  "$METADATA_REAL_GIT" "$@" || exit $?
  if [ "$count" -eq "$METADATA_FAULT_WHICH" ]; then exit 73; fi
  exit 0
fi
exec "$METADATA_REAL_GIT" "$@"
FAULT
  chmod +x "$test_root/metadata-fault-bin/git"
  for fault in cat:1 cat:2 hash:1 hash:2 entry:1; do
    rm -f "$METADATA_FAULT_COUNT"
    METADATA_FAULT_KIND="${fault%:*}" METADATA_FAULT_WHICH="${fault#*:}" \
      PATH="$test_root/metadata-fault-bin:$PATH" reject_without_changes --check
    test -s "$METADATA_FAULT_COUNT"
  done
  printf 'PASS  strict metadata representation, effective filters, promisor and failure controls (%s)\n' "$(git --version)"
)

# Execute the guide's actual first Bash block in each section without adding shell
# options to the extracted code: the published block must enforce its own failures.
guide_snippet() {
  awk -v heading="$1" '
    $0 == heading { section = 1; next }
    section && /^## / { exit }
    section && $0 == "```bash" { code = 1; next }
    code && $0 == "```" { exit }
    code { print }
  ' "$repo_dir/docs/project-shared-base.md"
}
setup_example="$(guide_snippet '## First setup')"
update_example="$(guide_snippet '## Updates and customizations')"
test -n "$setup_example"
test -n "$update_example"
guide_source='https://github.com/spik3r/clanker-foundry.git'
guide_placeholder='<full-reviewed-commit-id>'
[[ "$setup_example" == *"$guide_source"* ]]
[[ "$setup_example" == *"$guide_placeholder"* ]]
[[ "$update_example" == *"$guide_placeholder"* ]]

# A fixture-only wrapper records any invocation, including read-only --check.
# The actual implementation and all snippet commands still run unchanged.
mv "$baseline/scripts/project-setup.sh" "$baseline/scripts/project-setup-implementation.sh"
cat > "$baseline/scripts/project-setup.sh" <<'WRAPPER'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GUIDE_HELPER_LOG"
exec bash "$(dirname "${BASH_SOURCE[0]}")/project-setup-implementation.sh" "$@"
WRAPPER
git -C "$baseline" add scripts
git -C "$baseline" -c user.name=Test -c user.email=test@example.invalid commit -qm 'guide invocation fixture'
guide_pin="$(git -C "$baseline" rev-parse HEAD)"
git -C "$baseline" -c user.name=Test -c user.email=test@example.invalid commit --allow-empty -qm 'different default tip'
default_pin="$(git -C "$baseline" rev-parse HEAD)"
test "$guide_pin" != "$default_pin"
missing_pin=0000000000000000000000000000000000000000
short_pin="${guide_pin:0:12}"
export GUIDE_HELPER_LOG="$test_root/guide-helper.log"

run_guide() {
  local snippet="$1" pin="$2"
  snippet="${snippet//"$guide_source"/\"$baseline\"}"
  snippet="${snippet//"$guide_placeholder"/$pin}"
  printf '%s\n' "$snippet" > "$test_root/guide-example.sh"
  # Only the fixture's local transport is enabled; no global config is changed.
  GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=protocol.file.allow GIT_CONFIG_VALUE_0=always \
    bash "$test_root/guide-example.sh" > "$test_root/guide-output" 2>&1
}
reject_guide() {
  rm -f "$GUIDE_HELPER_LOG"
  if run_guide "$@"; then
    cat "$test_root/guide-output" >&2
    printf 'FAIL  guide accepted pin: %s\n' "$2" >&2; exit 1
  fi
  test ! -e "$GUIDE_HELPER_LOG"
}
assert_pin() {
  test "$(git -C .agent-baseline rev-parse --verify HEAD)" = "$1"
  test "$(git ls-files --stage -- .agent-baseline)" = "$(printf '160000 %s 0\t.agent-baseline' "$1")"
}
for bad_pin in "$missing_pin" "$short_pin"; do
  git init -q "$test_root/guide-rejected-$bad_pin"
  cd "$test_root/guide-rejected-$bad_pin"
  reject_guide "$setup_example" "$bad_pin"
  test ! -e AGENTS.md
  test ! -e .agents
  test ! -e .claude
done

git init -q "$test_root/guide consumer with spaces"
cd "$test_root/guide consumer with spaces"
run_guide "$setup_example" "$guide_pin"
assert_pin "$guide_pin"
test "$(cat "$GUIDE_HELPER_LOG")" = engineer
test -f .agents/skills/engineer/SKILL.md
test -f .claude/skills/engineer/SKILL.md
git add .gitmodules .agent-baseline .agents .claude AGENTS.md CLAUDE.md GEMINI.md
git -c user.name=Test -c user.email=test@example.invalid commit -qm 'guide consumer'

# Create the reviewed update after setup so the guide must fetch it first.
printf '\nGuide fixture update.\n' >> "$baseline/skills/engineer/SKILL.md"
git -C "$baseline" add skills/engineer/SKILL.md
git -C "$baseline" -c user.name=Test -c user.email=test@example.invalid commit -qm 'reviewed guide update'
update_pin="$(git -C "$baseline" rev-parse HEAD)"
for bad_pin in "$missing_pin" "${update_pin:0:12}"; do
  prior_index="$(git write-tree)"
  reject_guide "$update_example" "$bad_pin"
  assert_pin "$guide_pin"
  test "$(git write-tree)" = "$prior_index"
  test -z "$(git status --porcelain)"
done
# Even with a valid available commit, a blocked checkout must stop before --check.
printf '\nLocal baseline edit.\n' >> .agent-baseline/skills/engineer/SKILL.md
cp .agent-baseline/skills/engineer/SKILL.md "$test_root/local-skill"
prior_index="$(git write-tree)"
reject_guide "$update_example" "$update_pin"
assert_pin "$guide_pin"
test "$(git write-tree)" = "$prior_index"
cmp "$test_root/local-skill" .agent-baseline/skills/engineer/SKILL.md
git -C .agent-baseline restore -- skills/engineer/SKILL.md
run_guide "$update_example" "$update_pin"
assert_pin "$update_pin"
test "$(cat "$GUIDE_HELPER_LOG")" = --check
printf '%s\n' 'PASS  actual guide snippets reject unavailable/abbreviated pins, preserve failed updates, and set exact reviewed pins'

printf '%s\n' 'PASS  project shared-base safety, clean-clone and update regressions'
