# Gap report 1: typing `/` to filter the list by name

## Requirement

> where the user can type `/` and filter the list by name

Source: brief 5 (searchable service list).

## Profile considered

- Profile: `record-browser/v1`
- Registry entry and `cannotExpress` line that applies: none names search directly. The `cannotExpress` line "editable form fields" is the nearest, and the profile document's fixed-behaviour list does not include filtering.
- Profile document or schema section read: `modeling/tui/PROFILE.md` ("Fixed behavior" and "Semantic input"); `modeling/tui/README.md` was not read, as it is outside the allowed reading list; `modeling/tui/CASES.md` (the list of documented keys).

## Missing construct

The profile fixes the keys: `tab`, `shift+tab`, `enter`, `esc`, `?`, `q`, `ctrl+c`, `j`, `k`, `up`, `down`, `pgup`, `pgdown`, `home`, `end`. None starts a filter, and there is no text-input state. The model has no member for a search field, a filter key or a match rule.

## What was not done as a result

- The list cannot be filtered by name. The model shows every service.
- No case checks filtering; the cases only cover the browsing behaviour the profile fixes.

## Options

- Extend the profile: add a filter input state and a filter key (for example `/`) with a fixed match rule on a named field, to the profile, schema, cases vocabulary and generator.
- Hand-written seam owned by a person: a filter written in application code around the generated browser. This would need a second input path and its own tests.
- Leave the feature out: the screen lists all services and the user scrolls them.

Decision owner: unknown

Decision: (blank until the owner decides)

---

# Gap report 2: pressing `r` to reload from the service registry API

## Requirement

> and press `r` to reload the data from our service registry API.

Source: brief 5 (searchable service list).

## Profile considered

- Profile: `record-browser/v1`
- Registry entry and `cannotExpress` line that applies: "custom key actions or shell commands"; "background jobs or loading data from files or services".
- Profile document or schema section read: `modeling/tui/PROFILE.md` ("Semantic input" and "Fixed behavior": "Models cannot ... invoke a service"; "There are no ... background jobs. Application code supplies typed records and adapts its own loading and error states"); `modeling/tui/CASES.md` (the `replace` step).

## Missing construct

The model cannot bind `r` to an action, and cannot name a service or API to call. The browser has no command or background-job member. The only data path is the application's `replace` snapshot (`loading`, `ready`, `failed`), which application code must drive itself.

## What was not done as a result

- Nothing in the generated browser reloads data or listens for `r`. Pressing `r` does nothing.
- The application code that calls the registry API and supplies the snapshot is not covered by the model and is not written here.
- The `replace` behaviour is covered by a case (S_replace_keeps_selected_service), but that case does not show the reload itself.

## Options

- Extend the profile: add a named action with a declared key and a data source to the profile and generator. This is a change of profile scope and needs an owner decision.
- Hand-written seam owned by a person: application code handles `r`, calls the registry API and passes a `replace` snapshot to the generated browser. The owner needs to decide where that code lives and who reviews it.
- Leave the feature out: the screen shows the data loaded when it opens, and reloading needs a restart.

Decision owner: unknown

Decision: (blank until the owner decides)
