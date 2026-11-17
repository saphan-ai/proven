package source

// Observation applies a registered data rule without a model hypothesis. A
// matching rule is a data observation, not automatically a proved code defect.
type Observation struct {
	Adapter   string     `json:"adapter"`
	Template  string     `json:"template"`
	Ecosystem string     `json:"ecosystem"`
	Package   string     `json:"package"`
	Version   string     `json:"version"`
	Path      string     `json:"path"`
	Line      int        `json:"line"`
	Matched   bool       `json:"matched"`
	Complete  bool       `json:"complete"`
	SQL       string     `json:"query,omitempty"`
	Form      string     `json:"form,omitempty"`
	Engine    string     `json:"engine,omitempty"`
	Snapshot  string     `json:"snapshot,omitempty"`
	Columns   []string   `json:"columns,omitempty"`
	Rows      [][]string `json:"rows,omitempty"`
	Issue     string     `json:"issue,omitempty"`
}

func (o Observation) Redacted(text func(string) string) Observation {
	o.Adapter, o.Template, o.Ecosystem = text(o.Adapter), text(o.Template), text(o.Ecosystem)
	o.Package, o.Version, o.Path = text(o.Package), text(o.Version), text(o.Path)
	o.SQL, o.Form, o.Engine, o.Snapshot, o.Issue = text(o.SQL), text(o.Form), text(o.Engine), text(o.Snapshot), text(o.Issue)
	o.Columns = append([]string(nil), o.Columns...)
	for i := range o.Columns {
		o.Columns[i] = text(o.Columns[i])
	}
	rows := make([][]string, len(o.Rows))
	for i, row := range o.Rows {
		rows[i] = make([]string, len(row))
		for j, v := range row {
			rows[i][j] = text(v)
		}
	}
	o.Rows = rows
	return o
}
