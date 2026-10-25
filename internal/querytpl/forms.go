package querytpl

import "strings"

// The query languages a store can run (ADR-0047). Every template fills both:
// SQL for Zetaris and the local store, Mango for CouchDB.
const (
	FormSQL   = "sql"
	FormMango = "mango"
)

// Forms is one template built for each query language.
// SQL is the text Zetaris and the local store run. Mango is the text of the
// Mango query CouchDB runs, and Find is the same query as data, so a store
// sends it without reading it back out of text. Both forms come from the same
// validated values (ADR-0032, ADR-0047).
type Forms struct {
	SQL   string
	Mango string
	Find  Find
}

// BuildForms checks the request and returns each form.
// A refusal is the same *Refusal Build returns, and no form is filled. The
// request is checked once for each language, by the same checks, so a value
// one form refuses is refused by the other.
func BuildForms(req Request, rel Relations) (Forms, error) {
	sql, err := Build(req, rel)
	if err != nil {
		return Forms{}, err
	}
	// Build has accepted the template name and the parameters, so this
	// lookup and the parameter check below cannot fail; the Mango builder
	// still validates every value itself.
	find, err := buildFind(strings.TrimSpace(req.Template), req.Params, rel.resolved())
	if err != nil {
		return Forms{}, err
	}
	if find.Table == "" {
		return Forms{SQL: sql}, nil
	}
	return Forms{SQL: sql, Mango: find.text(), Find: find}, nil
}
