package feeds

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/source"
	"github.com/saphan-ai/proven/internal/store"
)

type Package struct{ Ecosystem, Name, Version string }
type Config struct {
	OSV, KEV, EPSS, Deps, Cache string
	Network                     *network.Policy
}
type Snapshot struct {
	Table      store.Table
	Additional []store.Table
	Refresh    source.Refresh
}
type Result struct {
	Packages int              `json:"packages"`
	Sources  []source.Refresh `json:"sources"`
	Notes    []string         `json:"notes,omitempty"`
}
type Session struct {
	Config     Config
	Client     *Client
	Packages   []Package
	Now        time.Time
	Advisories map[string][]byte
	CVEs       map[string]bool
}

// Connector acquires facts; the existing source.Adapter retains query and
// evidence rules. One provider may populate several explicitly named tables.
type Connector interface {
	Name() string
	Fetch(context.Context, *Session) []Snapshot
}

func Connectors() []Connector {
	return []Connector{osvConnector{}, kevConnector{}, epssConnector{}, depsConnector{}}
}

func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{Name: "PROVEN_SOURCE_CACHE", Default: "<user cache>/proven/feeds", Description: "persistent public-data HTTP cache; keep outside the reviewed checkout"},
		{Name: "PROVEN_OSV_URL", Default: "https://api.osv.dev/v1", Description: "OSV API root for incremental advisory refresh"},
		{Name: "PROVEN_KEV_URL", Default: "https://raw.githubusercontent.com/cisagov/kev-data/main/known_exploited_vulnerabilities.json", Description: "CISA-maintained KEV mirror JSON; conditional HTTP requests retain its catalogue date"},
		{Name: "PROVEN_EPSS_URL", Default: "https://api.first.org/data/v1", Description: "FIRST EPSS API root; requests only CVEs relevant to the selected dependencies"},
		{Name: "PROVEN_DEPS_URL", Default: "https://api.deps.dev/v3", Description: "deps.dev API root for package licences, project Scorecard data and CVSS scores"},
	}
}
func Load(env func(string) string) (Config, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	policy, err := network.Load(env)
	if err != nil {
		return Config{}, err
	}
	cache := env("PROVEN_SOURCE_CACHE")
	if cache == "" {
		root, err := os.UserCacheDir()
		if err != nil {
			return Config{}, err
		}
		cache = filepath.Join(root, "proven", "feeds")
	}
	c := Config{Cache: cache, Network: policy}
	for _, item := range []struct {
		name, defaultURL string
		dest             *string
	}{
		{"PROVEN_OSV_URL", "https://api.osv.dev/v1", &c.OSV}, {"PROVEN_KEV_URL", "https://raw.githubusercontent.com/cisagov/kev-data/main/known_exploited_vulnerabilities.json", &c.KEV},
		{"PROVEN_EPSS_URL", "https://api.first.org/data/v1", &c.EPSS}, {"PROVEN_DEPS_URL", "https://api.deps.dev/v3", &c.Deps},
	} {
		v := env(item.name)
		if v == "" {
			v = item.defaultURL
		}
		v = strings.TrimRight(v, "/")
		address, err := url.Parse(v)
		if err != nil || address.Hostname() == "" || (address.Scheme != "https" && address.Scheme != "http") || address.User != nil || address.RawQuery != "" || address.Fragment != "" {
			return Config{}, fmt.Errorf("%s must be an HTTP(S) source URL without embedded credentials, query or fragment", item.name)
		}
		if err := policy.CheckURL(item.name, v); err != nil {
			return Config{}, err
		}
		*item.dest = v
	}
	return c, nil
}
func Packages(hits []manifest.Hit) ([]Package, []string) {
	seen := map[string]bool{}
	var out []Package
	var notes []string
	for _, h := range hits {
		valid := false
		for _, r := range manifest.Readers() {
			if r.Ecosystem == h.Ecosystem && r.Compare != nil {
				_, valid = r.Compare(h.Version, h.Version)
				if valid {
					break
				}
			}
		}
		if !valid {
			notes = append(notes, "Cannot refresh unresolved version of "+h.Ecosystem+"/"+h.Package)
			continue
		}
		key := h.Ecosystem + "\x00" + h.Package + "\x00" + h.Version
		if !seen[key] {
			seen[key] = true
			out = append(out, Package{h.Ecosystem, h.Package, h.Version})
		}
	}
	sort.Slice(out, func(i, j int) bool { return fmt.Sprint(out[i]) < fmt.Sprint(out[j]) })
	return out, notes
}

// Sync uses fresh provider responses (or confirmed 304s). Failed sources never
// replace their old tables. Import is injected so backends remain adapters.
func Sync(ctx context.Context, c Config, packages []Package, client *Client, importTable func(context.Context, store.Table) error) (Result, error) {
	r := Result{Packages: len(packages), Sources: []source.Refresh{}}
	if len(packages) > 200 {
		return r, fmt.Errorf("source refresh is limited to 200 selected package versions per run")
	}
	if client == nil {
		client = NewClient(c.Network, c.Cache)
	}
	s := &Session{Config: c, Client: client, Packages: packages, Now: time.Now().UTC(), Advisories: map[string][]byte{}, CVEs: map[string]bool{}}
	failed := 0
	for _, connector := range Connectors() {
		for _, snapshot := range connector.Fetch(ctx, s) {
			status := snapshot.Refresh
			if status.Status != "error" && snapshot.Table.Name != "" {
				for _, table := range append([]store.Table{snapshot.Table}, snapshot.Additional...) {
					if err := boundedTable(table); err != nil {
						status.Status = "error"
						status.Issue = err.Error()
						break
					}
					table.Snapshot = status.Origin + " upstream=" + status.UpstreamAt
					if err := importTable(ctx, table); err != nil {
						status.Status = "error"
						status.Issue = "refreshed source could not be imported"
						break
					}
				}
			}
			if status.Status == "error" {
				failed++
			}
			r.Sources = append(r.Sources, status)
		}
	}
	if failed != 0 {
		return r, fmt.Errorf("%d sources did not refresh completely", failed)
	}
	return r, nil
}
func (s *Session) status(id, origin string, downloaded, reused int) source.Refresh {
	return source.Refresh{ID: id, Origin: origin, CheckedAt: s.Now.Format(time.RFC3339), Status: "updated", Downloaded: s.Client.Downloaded - downloaded, Reused: s.Client.Reused - reused}
}
func (s *Session) failure(id, origin string, err error) Snapshot {
	return Snapshot{Refresh: source.Refresh{ID: id, Origin: origin, CheckedAt: s.Now.Format(time.RFC3339), Status: "error", Issue: err.Error()}}
}

func boundedTable(t store.Table) error {
	if len(t.Rows) > 100000 {
		return fmt.Errorf("source table exceeded 100000 normalized rows")
	}
	total := 0
	for _, row := range t.Rows {
		for _, cell := range row {
			if len(cell) > 8192 {
				return fmt.Errorf("source table cell exceeded 8 KiB")
			}
			total += len(cell)
			if total > 64<<20 {
				return fmt.Errorf("source table exceeded the 64 MiB normalized content budget")
			}
		}
	}
	return nil
}
