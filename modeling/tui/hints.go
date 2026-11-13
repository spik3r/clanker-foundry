package tui

// HintSpec documents one KeyHint ID of the generated browser: when it appears
// and its default key and action text. The IDs are stable within the
// record-browser/v1 profile: a new hint may add an ID, an existing ID is never
// renamed or reused. Key and Action wording may change, so a host maps on ID.
//
// This list is the one place the vocabulary is written down in code. The
// generator emits the IDs into the generated package (keyHintIDs), the
// generated acceptance tests compare what the browser returns with that list,
// and a generator test checks README.md and PROFILE.md against it.
type HintSpec struct {
	ID     string
	When   string
	Key    string
	Action string
}

// KeyHintVocabulary lists every KeyHint ID, in the order the docs show them.
var KeyHintVocabulary = []HintSpec{
	{"select", "list focused, a record selected", "j/k", "select"},
	{"scroll", "detail focused", "j/k", "scroll"},
	{"pane", "a record selected", "tab", "pane"},
	{"open", "list focused, a record selected", "enter", "open"},
	{"back", "detail focused", "esc", "back"},
	{"help.open", "help closed", "?", "help"},
	{"help.scroll", "help open", "j/k", "scroll"},
	{"help.close", "help open", "esc/?", "close"},
	{"filter.open", "model has a filter; help and filter input closed", "/", "filter"},
	{"filter.keep", "filter input open", "enter", "keep"},
	{"filter.cancel", "filter input open", "esc", "clear"},
	{"filter.clear", "filter kept (text not empty), list focused", "esc", "clear"},
	{"quit", "standalone browser only; help and filter input closed", "q", "quit"},
}

// KeyHintIDs returns the vocabulary's IDs in documented order.
func KeyHintIDs() []string {
	ids := make([]string, len(KeyHintVocabulary))
	for i, h := range KeyHintVocabulary {
		ids[i] = h.ID
	}
	return ids
}
