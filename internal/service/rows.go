package service

// CopyRows copies query rows deeply. Nil stays nil, so "no rows" and "zero
// rows" stay distinct. Review and the pipeline keep this.
func CopyRows(rows [][]string) [][]string { return copyRows(rows) }

// ListRows copies query rows and turns nil into an empty list, so evidence
// prints [] and not null. EPSS and OSV keep this.
func ListRows(rows [][]string) [][]string {
	out := copyRows(rows)
	if out == nil {
		return [][]string{}
	}
	return out
}

// copyRows copies query rows deeply. A nil input stays nil.
func copyRows(rows [][]string) [][]string {
	if rows == nil {
		return nil
	}
	out := make([][]string, len(rows))
	for i, row := range rows {
		out[i] = append([]string(nil), row...)
	}
	return out
}
