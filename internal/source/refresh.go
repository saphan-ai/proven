package source

// Refresh records acquisition independently of the table's content hash.
// Checking a provider now does not mean its underlying observation is new.
type Refresh struct {
	ID         string `json:"id"`
	Origin     string `json:"origin"`
	CheckedAt  string `json:"checked_at"`
	UpstreamAt string `json:"upstream_at,omitempty"`
	Status     string `json:"status"`
	Rows       int    `json:"rows"`
	Downloaded int    `json:"downloaded"`
	Reused     int    `json:"reused"`
	Issue      string `json:"issue,omitempty"`
}
