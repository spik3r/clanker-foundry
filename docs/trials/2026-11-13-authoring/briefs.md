# Authoring trial briefs

For each brief produce files in the output directory named in your instructions.

## Brief 1: branch browser
We want a terminal screen listing the git branches of a repository. Each branch has a name, the short hash of its last commit, the author of that commit, the commit subject, and how far it is ahead/behind main (as text like "3 ahead, 1 behind"). The list shows branch names. Selecting a branch shows the other details, subject first. On terminals narrower than 90 columns show one pane at a time. The screen is read-only; the application supplies the data. It will live inside our application as a Go package named `branchbrowser`. Long commit subjects must be readable in full.

## Brief 2: expense approval process
Draw our expense approval process. An employee submits a claim. A manager reviews it. If the manager rejects it the employee is told and the process ends. If approved and the amount is over the finance threshold, finance must also check it; otherwise it goes straight to payment. Finance can send it back to the employee to fix and resubmit to the manager. After payment the process ends. Three roles: Employee, Manager, Finance.

## Brief 3: build health dashboard
A static weekly report page for the team. It shows: total builds this week (412) and the failure rate (6.3%); builds per day Monday to Friday as bars, split into passed and failed (Mon 80/5, Tue 78/7, Wed 85/3, Thu 70/9, Fri 73/2); the median build time in minutes per day as a line (Mon 11.5, Tue 12, Wed 10.75, Thu 14, Fri 11); and a table of the same daily figures.

## Brief 4: plugin manifest check
We need a function that checks a plugin manifest (already parsed JSON) before we load it. A valid manifest is an object with: a string `name`; a string `version`; a boolean `enabled`; an optional array `tags` of strings; `kind` which must be exactly one of the strings "scanner", "reporter" or "exporter"; and it must not contain any properties other than these.

## Brief 5: searchable service list
A terminal screen listing our internal services (name, owner team, on-call contact, status) where the user can type `/` and filter the list by name, and press `r` to reload the data from our service registry API.
