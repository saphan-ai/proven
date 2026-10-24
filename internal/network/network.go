// Package network is the air-gapped mode of ADR-0034. PROVEN_NETWORK=local
// makes Proven refuse, when a setting loads, every address that is not on the
// organisation's own network, and makes each HTTP client refuse the same
// addresses when it dials.
//
// Two checks carry the guarantee and neither stands in for the other. The load
// check names a wrong setting before a review starts. The dial check holds
// when a redirect or a later address would lead somewhere else.
package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// Environment variables of the network mode. The set is closed: Load reads
// these names and no others. An empty value means the default.
const (
	EnvNetwork = "PROVEN_NETWORK"
	EnvAllow   = "PROVEN_NETWORK_ALLOW"
)

// The two modes. The default is public: today's behaviour (ADR-0034).
const (
	ModePublic = "public"
	ModeLocal  = "local"
)

// Environment lists the network settings in the order the page shows.
// Each Name is the Env constant above.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvNetwork,
			Default:     envdoc.Tick(ModePublic),
			Description: "`public` or `local`. `local` is the air-gapped mode: a setting that names an address outside the organisation's network is refused at load, and every connection to such an address is refused when it is made",
		},
		{
			Name:        EnvAllow,
			Default:     "none",
			Description: "in `local` mode, host names that count as inside the network, comma-separated, for example `nightshift.corp.example`. A host name is never resolved to be judged: it is allowed by this list or it is refused. A listed name is resolved only through a name server inside the network, and an answer outside it is refused",
		},
	}
}

// ConfigError names a refused network setting and the values that are allowed.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the setting, the refused value, why, and what is allowed.
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := fmt.Sprintf("configuration refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// DialFunc opens one connection. It is the seam a test uses to see every dial.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// Policy is one validated network mode. A nil *Policy is the public mode, so a
// Settings built by hand and not by Load keeps today's behaviour.
type Policy struct {
	local bool
	allow map[string]bool
	// dial opens the connection once the address is allowed. Nil means net.Dialer.
	dial DialFunc
	// resolver answers the lookup of an allowed name. Nil means net.DefaultResolver.
	resolver *net.Resolver
}

// Load reads PROVEN_NETWORK and PROVEN_NETWORK_ALLOW. A wrong value is refused
// by name. Public mode returns nil: it has no policy to apply, and the allow
// list is still checked, so a typo is caught before the mode is switched on.
func Load(env func(string) string) (*Policy, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	mode := strings.TrimSpace(env(EnvNetwork))
	if mode == "" {
		mode = ModePublic
	}
	if mode != ModePublic && mode != ModeLocal {
		return nil, refuse(EnvNetwork, clip(mode), "not a network mode", ModePublic+" or "+ModeLocal)
	}
	allow, err := parseAllow(env(EnvAllow))
	if err != nil {
		return nil, err
	}
	if mode == ModePublic {
		return nil, nil
	}
	return &Policy{local: true, allow: allow}, nil
}

// Local reports whether the air-gapped mode is on. A nil policy is public.
func (p *Policy) Local() bool { return p != nil && p.local }

// WithDial returns a copy of the policy whose allowed connections are opened by
// dial. Tests use it to record every dial without opening a socket.
func (p *Policy) WithDial(dial DialFunc) *Policy {
	if p == nil {
		return nil
	}
	cp := *p
	cp.dial = dial
	return &cp
}

// WithResolver returns a copy of the policy whose allowed names are looked up
// by r. Tests use it to see every query a lookup would send.
func (p *Policy) WithResolver(r *net.Resolver) *Policy {
	if p == nil {
		return nil
	}
	cp := *p
	cp.resolver = r
	return &cp
}

// CheckURL refuses, in local mode, a configured address whose host is not on
// the organisation's network. setting names the variable the address came
// from, so the refusal says which one to fix. The host is judged by its text:
// no name is resolved.
func (p *Policy) CheckURL(setting, raw string) error {
	if !p.Local() {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return refuse(setting, clip(raw), "not an address with a host", "an address on the organisation's network")
	}
	if !p.allowedHost(u.Hostname()) {
		return refuse(setting, clip(raw), "host "+u.Hostname()+" is outside the organisation's network (PROVEN_NETWORK=local)", allowedForms())
	}
	return nil
}

// CheckDaemon refuses, in local mode, the address of a container daemon that is
// not on the organisation's network. A local socket (unix, npipe, fd) is always
// inside it; any other scheme is judged by its host like a setting address, and
// its host must be an IP literal, never a name.
// setting names where the address came from, so the refusal says what to fix.
func (p *Policy) CheckDaemon(setting, raw string) error {
	if !p.Local() {
		return nil
	}
	if u, err := url.Parse(raw); err == nil {
		switch strings.ToLower(u.Scheme) {
		case "unix", "npipe", "fd":
			return nil
		}
	}
	// ssh chooses its own route from the ssh configuration under HOME (ProxyJump,
	// ProxyCommand, Match rules), so the host in the address is not where it
	// connects. Proven cannot enforce or verify every hop, so it refuses the scheme.
	if u, err := url.Parse(raw); err == nil && strings.EqualFold(u.Scheme, "ssh") {
		return refuse(setting, clip(raw), "an ssh daemon route is chosen by the ssh configuration, which can pass through a host outside the organisation's network, and Proven cannot check every hop (PROVEN_NETWORK=local)", "a local socket, or tcp:// to the IP address of the daemon on the organisation's network")
	}
	if err := p.CheckURL(setting, raw); err != nil {
		return err
	}
	// The runtime, or ssh under it, resolves a host name with its own resolver:
	// the name servers of the host's configuration, its search list, and whatever
	// address comes back. Proven judges none of that, so a name is not handed to
	// it, listed or not. An IP literal has nothing to resolve.
	u, _ := url.Parse(raw)
	bare, _, _ := strings.Cut(u.Hostname(), "%")
	if _, err := netip.ParseAddr(bare); err != nil {
		return refuse(setting, clip(raw), "host "+u.Hostname()+" is a name, and the container runtime would resolve it with its own resolver (PROVEN_NETWORK=local)", "a local socket, or the IP address of the daemon on the organisation's network")
	}
	return nil
}

// Refuse names a setting that PROVEN_NETWORK=local forbids as a value, such as
// the guardrail's hosted mode.
func (p *Policy) Refuse(setting, value, reason string) error {
	if !p.Local() {
		return nil
	}
	return refuse(setting, clip(value), reason+" (PROVEN_NETWORK=local)", "a value that stays on the organisation's network")
}

// NewClient builds the HTTP client of one live adapter. It does not follow a
// redirect, so a key is not sent to another host. In local mode its dialer
// refuses every address the load check would have refused.
func (p *Policy) NewClient(timeout time.Duration) *http.Client {
	c := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if p.Local() {
		c.Transport = &http.Transport{
			DialContext: p.DialContext,
			// A proxy is an address the setting check never saw.
			Proxy: nil,
		}
	}
	return c
}

// LocalClient returns NewClient's client in local mode and nil in public mode,
// for a caller that has its own default client in public mode and must not
// change it (the OTLP exporters).
func (p *Policy) LocalClient(timeout time.Duration) *http.Client {
	if !p.Local() {
		return nil
	}
	return p.NewClient(timeout)
}

// DialContext opens a connection to addr only when its host is on the
// organisation's network. The check reads the address as written and comes
// before any lookup: a name off the list is never sent to a resolver. A name on
// the allow list is looked up by Lookup, and "localhost" is not looked up at all.
func (p *Policy) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("network: dial %q refused: not host:port", addr)
	}
	if p.Local() && !p.allowedHost(host) {
		return nil, fmt.Errorf("network: dial %q refused: host %s is outside the organisation's network (%s=%s)", addr, host, EnvNetwork, ModeLocal)
	}
	if p != nil && p.dial != nil {
		return p.dial(ctx, network, addr)
	}
	d := net.Dialer{}
	if !p.Local() {
		return d.DialContext(ctx, network, addr)
	}
	ips, err := p.Lookup(ctx, host)
	if err != nil {
		return nil, err
	}
	var last error
	for _, ip := range ips {
		// Lookup has judged every answer already, and ips comes from nowhere
		// else, so the loop holds no second check.
		conn, derr := d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if derr == nil {
			return conn, nil
		}
		last = derr
	}
	return nil, last
}

// Lookup returns the addresses of host. In local mode the name is judged first:
// one that is not on the organisation's network is refused before any query is
// sent. An IP literal is returned as it is, and "localhost" is answered with
// the loopback addresses without a query, so no resolver is asked for it.
func (p *Policy) Lookup(ctx context.Context, host string) ([]netip.Addr, error) {
	if p.Local() && !p.allowedHost(host) {
		return nil, fmt.Errorf("network: lookup of %q refused: host is outside the organisation's network (%s=%s)", host, EnvNetwork, ModeLocal)
	}
	name := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if a, err := netip.ParseAddr(name); err == nil {
		return []netip.Addr{a}, nil
	}
	if name == "localhost" && p.Local() {
		return []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")}, nil
	}
	r := net.DefaultResolver
	if p.Local() {
		r = p.checkedResolver()
	}
	if p != nil && p.resolver != nil {
		r = p.resolver
	}
	if !p.Local() {
		return r.LookupNetIP(ctx, "ip", name)
	}
	ips, err := r.LookupNetIP(ctx, "ip", absoluteName(name))
	if err != nil {
		return nil, err
	}
	// The name is on the allow list; its answer is not. An address outside the
	// organisation's network is refused, whole answer and not only that address,
	// so a dial never meets it.
	for _, ip := range ips {
		if !p.allowedHost(ip.String()) {
			return nil, outsideAnswer(name, ip)
		}
	}
	return ips, nil
}

// absoluteName returns name with a trailing dot. The dot makes the name
// absolute, so the resolver's search list is not appended and no other name is
// queried: a listed name is asked once, as itself.
func absoluteName(name string) string { return name + "." }

// outsideAnswer names the host and the address a DNS answer asked Proven to
// connect to.
func outsideAnswer(host string, ip netip.Addr) error {
	return fmt.Errorf("network: lookup of %q refused: the answer %s is outside the organisation's network (%s=%s)", host, ip, EnvNetwork, ModeLocal)
}

// checkedResolver is the resolver of local mode. It is the pure-Go resolver, so
// the host's C library never sends a query Proven did not see, and each
// connection to a name server goes through resolverDial.
func (p *Policy) checkedResolver() *net.Resolver {
	return &net.Resolver{PreferGo: true, Dial: p.resolverDial}
}

// resolverDial opens the connection to a name server. The server's address is
// the one the host configuration names (resolv.conf, or the loopback default
// where the platform has none). A server outside the organisation's network is
// refused by name and address before any query is sent; the resolver then tries
// the next server, and the lookup fails when none is inside.
func (p *Policy) resolverDial(ctx context.Context, network, address string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil || !p.allowedHost(host) {
		return nil, fmt.Errorf("network: resolver %q refused: the name server is outside the organisation's network (%s=%s); point the host's resolver at an address on the network", address, EnvNetwork, ModeLocal)
	}
	if p.dial != nil {
		return p.dial(ctx, network, address)
	}
	d := net.Dialer{}
	return d.DialContext(ctx, network, address)
}

// allowedHost reports whether host is on the organisation's network: an IP
// literal that is loopback, private or link-local, "localhost", or a name on
// the allow list. Any other name is refused without a lookup.
func (p *Policy) allowedHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	// A zone on an IPv6 literal does not change which address it is.
	bare, _, _ := strings.Cut(host, "%")
	if a, err := netip.ParseAddr(bare); err == nil {
		a = a.Unmap()
		return a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() || p.allow[host]
	}
	return host == "localhost" || p.allow[host]
}

// parseAllow reads the allow list: comma-separated host names. An entry with a
// scheme, a path, a port or a space is refused, so the list cannot hold an
// address in a form the dial check would not match.
func parseAllow(raw string) (map[string]bool, error) {
	allow := map[string]bool{}
	if strings.TrimSpace(raw) == "" {
		return allow, nil
	}
	for _, part := range strings.Split(raw, ",") {
		name := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(part), "."))
		if !hostName(name) {
			return nil, refuse(EnvAllow, clip(raw), "not a list of host names", "comma-separated host names with no scheme, port or path, for example nightshift.corp.example")
		}
		allow[name] = true
	}
	return allow, nil
}

// hostName accepts letters, digits, dots, hyphens and underscores, and the
// colons and brackets of an IP literal are left out on purpose.
func hostName(s string) bool {
	if s == "" || len(s) > 253 {
		return false
	}
	for _, r := range s {
		if r == '.' || r == '-' || r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' {
			continue
		}
		return false
	}
	return true
}

func allowedForms() string {
	return "a loopback or private address, localhost, or a host listed in " + EnvAllow
}

func refuse(field, value, reason, allowed string) *ConfigError {
	return &ConfigError{Field: field, Value: value, Reason: reason, Allowed: []string{allowed}}
}

// As reports whether err is a refusal of this package, for a caller that maps
// it onto its own ConfigError.
func As(err error) (*ConfigError, bool) {
	var ce *ConfigError
	if errors.As(err, &ce) {
		return ce, true
	}
	return nil, false
}

// clip keeps a refused value short enough to print in an error.
func clip(value string) string {
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
