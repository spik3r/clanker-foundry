# Brief 2 notes: expense approval process

- Profile: `process-diagram/v1`. Chosen because the brief is a lane-based business process with decisions, a rework loop and an end state. Other registered profiles (sequence, class, dashboard) do not describe processes.
- Covered: three lanes (Employee, Manager, Finance); the claim start; manager review and decision; the rejection path ending in "Claim rejected"; finance check only above the threshold; finance send-back to the employee, who resubmits to the manager (a cycle, which the profile allows); payment and end.
- Gap: the threshold rule (the condition on the route and the amount it depends on). One gap report in `brief2.gap.md`.
- Assumed: the "tell the employee" step is a task in the Manager lane, since the brief does not say who tells the employee. Messages to the employee are not drawn as message flows, because the profile has none.
- Assumed: the threshold gateway sits in the Manager lane, since the brief places the threshold check after manager approval and does not name a lane.
- Assumed: "goes straight to payment" means the Pay task is reached from the threshold gateway's "no" branch, as the profile has no other way to join the paths.
- Unclear: the brief gives no threshold amount and no decision owner for the threshold.
- The model was not validated: `diagramgen validate` was not run. Degree and reachability rules were checked by hand against the profile document only.
