package service

// CopyRows copies query rows deeply. Nil stays nil, so "no rows" and "zero
// rows" stay distinct.
func CopyRows(rows [][]string) [][]string { return copyRows(rows) }

// copyRows copies query rows deeply. A nil input stays nil. Two earlier copies
// turned nil into an empty list, which invents a list the caller did not have.
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
