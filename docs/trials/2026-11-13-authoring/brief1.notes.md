# Brief 1 notes: branch browser

- Profile: `record-browser/v1` (registry summary: read-only list and detail browser, Go package). Chosen because the brief is a read-only list with a detail pane, all fields are text, and the app supplies the data.
- Covered: branch name as the list label; subject, author, commit hash and "ahead/behind" text as details, subject first; one pane below 90 columns (wide minimum width 90); Enter/Esc between panes; `detailText: wrap` so long subjects show in full; package mode with name `branchbrowser`.
- Gaps: none found for the brief's stated requirements. See the notes below for layout choices the brief did not specify.
- Assumed (not in the brief): wide height 6 and narrow minimum 24x6 (profile minimums, so the height rule cannot be turned off; terminals under 6 rows show the resize message); list width 28, so branch names longer than 27 cells are cut; identity is the branch name, which must be unique.
- Assumed: the details list omits the branch name, since the list already shows it.
- Unclear in the docs: the profile does not say whether a wide layout below the width minimum but above the height minimum counts as "one pane" by width alone; I read the PROFILE.md statement "below either wide minimum it shows only the focused pane" as applying.
- Cases: `model.canonicalSHA256` is the literal `UNKNOWN`. It fails the schema pattern and `tuicases` will refuse it until the digest is set after review. Cases were not run.
- The model was not validated: no generator was run.
