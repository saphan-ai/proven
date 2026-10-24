package querytpl

// The query languages a store can run (ADR-0047). SQL is filled for every
// template. Mango is the name CouchDB will use; no template fills it yet.
const (
	FormSQL   = "sql"
	FormMango = "mango"
)

// Forms is one template built for each query language.
// SQL is the text Zetaris and the local store run. Mango is empty until
// CouchDB (W49): the field is the room for that form, so a later wagon fills
// it without a new return type. Build still returns the SQL, which is the
// only form a store runs today.
type Forms struct {
	SQL   string
	Mango string
}

// BuildForms checks the request and returns each form.
// A refusal is the same *Refusal Build returns, and no form is filled.
func BuildForms(req Request, rel Relations) (Forms, error) {
	sql, err := Build(req, rel)
	if err != nil {
		return Forms{}, err
	}
	// ADR-0047: Zetaris and the local store speak SQL. The mango form stays
	// empty so a reader can see that no selector was built.
	return Forms{SQL: sql}, nil
}
