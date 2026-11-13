# Brief 3 notes: build health dashboard

- Profile: `dashboard/v1`. Chosen because the brief is a static weekly page of stat tiles, bar and line charts and a table over five days of figures.
- Covered: two stat tiles (412 total builds, 6.3% failure rate, both as typed text); a line of median build time per day; a table of the daily figures (passed, failed, median); all in one self-contained page.
- Gap: the passed/failed bars are grouped, not stacked, because the profile has no stacked bars. One gap report in `brief3.gap.md`.
- Checked by hand, not computed by the model: 80+78+85+70+73 = 386 passed, 5+7+3+9+2 = 26 failed, 386+26 = 412, and 26/412 = 6.3%. The brief's figures agree.
- Assumed: the "split" bars are two series side by side (grouped), with stacking as the open question.
- Assumed: the page has no subtitle because the brief gives no week number or date range.
- Assumed: the line and table use the same median column, so the table holds three figures per day.
- Unclear in the docs: the profile does not say how a chart's y-axis reads when the values are close together (10.75 to 14); the chart rules imply the axis does not start at 0 for a line, but I did not check the output.
- The model was not validated: `diagramgen validate` was not run.
