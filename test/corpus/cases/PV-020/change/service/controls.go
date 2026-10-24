package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

// Outcome is one control group's result.
// Result is "passed", "failed", or "could not run".
// Reason is set for "could not run", and for "failed" it names the controls.
type Outcome struct {
	Result string
	Reason string
}

func (o Outcome) Line() string {
	if o.Reason == "" {
		return o.Result
	}
	return o.Result + ": " + o.Reason
}

// RunInProcess runs the limit controls in this process.
// They always run: they do not read a description and they do not listen.
func RunInProcess(svc *Service) Outcome {
	var fails []string
	fails = append(fails, inprocLimitPresent(svc)...)
	fails = append(fails, inprocAtLimitAccepted(svc)...)
	fails = append(fails, inprocOverLimitRefused(svc)...)
	fails = append(fails, inprocOverLimitExemptAccepted(svc)...)
	fails = append(fails, inprocExemptSet()...)
	return outcomeFrom(fails)
}

func inprocLimitPresent(svc *Service) []string {
	if svc.MaxBody() != DocumentedLimit {
		return []string{"inproc.limit-present"}
	}
	return nil
}

func inprocAtLimitAccepted(svc *Service) []string {
	body := bytes.Repeat([]byte("a"), DocumentedLimit)
	var fails []string
	for _, rt := range Routes {
		if rt.Exempt {
			continue
		}
		d := svc.Decide(rt.Method, rt.Path, body)
		if !d.Accepted || d.Reason != "accepted" {
			fails = append(fails, "inproc.at-limit "+rt.Method+" "+rt.Path)
		}
	}
	return fails
}

func inprocOverLimitRefused(svc *Service) []string {
	body := bytes.Repeat([]byte("a"), DocumentedLimit+1)
	var fails []string
	for _, rt := range Routes {
		if rt.Exempt {
			continue
		}
		d := svc.Decide(rt.Method, rt.Path, body)
		if d.Accepted || d.Reason != "body exceeds limit" {
			fails = append(fails, "inproc.over-limit-refused "+rt.Method+" "+rt.Path)
		}
	}
	return fails
}

func inprocOverLimitExemptAccepted(svc *Service) []string {
	body := bytes.Repeat([]byte("a"), DocumentedLimit+1)
	var fails []string
	for _, rt := range Routes {
		if !rt.Exempt {
			continue
		}
		d := svc.Decide(rt.Method, rt.Path, body)
		if !d.Accepted || d.Reason != "exempt upload" {
			fails = append(fails, "inproc.over-limit-exempt "+rt.Method+" "+rt.Path)
		}
	}
	return fails
}

func inprocExemptSet() []string {
	got := map[string]bool{}
	for _, rt := range Routes {
		if rt.Exempt {
			got[rt.Method+" "+rt.Path] = true
		}
	}
	if len(got) != len(DocumentedExempt) {
		return []string{"inproc.exempt-set"}
	}
	for _, name := range DocumentedExempt {
		if !got[name] {
			return []string{"inproc.exempt-set"}
		}
	}
	return nil
}

// errDescriptionAbsent is the stop the description controls return when
// the configured file is not there. The text is the contract's message.
var errDescriptionAbsent = errors.New("description not provisioned")

type apiRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Exempt bool   `json:"exempt"`
}

type apiDoc struct {
	MaxBodyBytes int        `json:"max_body_bytes"`
	Routes       []apiRoute `json:"routes"`
}

// RunDescription runs the two description controls.
// path is configuration: the file the controls read.
// An empty path or a missing file stops both with "description not provisioned".
func RunDescription(path string) Outcome {
	routesErr := descriptionRoutes(path)
	limitsErr := descriptionLimits(path)
	return outcomeFromErrs(routesErr, limitsErr)
}

func descriptionRoutes(path string) error {
	doc, err := loadDescription(path)
	if err != nil {
		return err
	}
	if !sameRoutes(doc) {
		return errors.New("description.routes")
	}
	return nil
}

func descriptionLimits(path string) error {
	doc, err := loadDescription(path)
	if err != nil {
		return err
	}
	if doc.MaxBodyBytes != Production().MaxBody() || !sameExempt(doc) {
		return errors.New("description.limits")
	}
	return nil
}

func loadDescription(path string) (apiDoc, error) {
	if path == "" {
		return apiDoc{}, errDescriptionAbsent
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return apiDoc{}, errDescriptionAbsent
		}
		return apiDoc{}, err
	}
	var doc apiDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return apiDoc{}, errors.New("description.unreadable")
	}
	return doc, nil
}

func sameRoutes(doc apiDoc) bool {
	got := map[string]bool{}
	for _, rt := range doc.Routes {
		got[rt.Method+" "+rt.Path] = true
	}
	if len(got) != len(Routes) || len(doc.Routes) != len(Routes) {
		return false
	}
	for _, rt := range Routes {
		if !got[rt.Method+" "+rt.Path] {
			return false
		}
	}
	return true
}

func sameExempt(doc apiDoc) bool {
	got := map[string]bool{}
	for _, rt := range doc.Routes {
		got[rt.Method+" "+rt.Path] = rt.Exempt
	}
	for _, rt := range Routes {
		ex, ok := got[rt.Method+" "+rt.Path]
		if !ok || ex != rt.Exempt {
			return false
		}
	}
	return true
}

// RunListener starts a real local listener and sends real streamed requests.
// Two controls run when the listener starts: a non-exempt route must refuse
// an oversize stream, and an exempt route must accept one.
// If the listen fails with "operation not permitted", both could not run.
// The limit is not changed on that path.
func RunListener(svc *Service) Outcome {
	ln, err := listenLocal()
	if err != nil {
		if errors.Is(err, syscall.EPERM) {
			return Outcome{Result: "could not run", Reason: "operation not permitted"}
		}
		return Outcome{Result: "could not run", Reason: err.Error()}
	}
	srv := &http.Server{
		Handler:           svc.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	base := "http://" + ln.Addr().String()
	var fails []string
	if err := listenerNonExemptStreamRefused(base); err != nil {
		fails = append(fails, err.Error())
	}
	if err := listenerExemptStreamAccepted(base); err != nil {
		fails = append(fails, err.Error())
	}
	return outcomeFrom(fails)
}

// listenLocal binds 127.0.0.1 on an ephemeral port.
// PV020_SIMULATE_LISTEN_REFUSAL=1 fails the call with EPERM before any bind,
// which is how a restricted environment refuses it. The product limit is untouched.
// A real EPERM from the host is classified the same way.
func listenLocal() (net.Listener, error) {
	if os.Getenv("PV020_SIMULATE_LISTEN_REFUSAL") == "1" {
		return nil, &net.OpError{Op: "listen", Net: "tcp", Err: syscall.EPERM}
	}
	return net.Listen("tcp", "127.0.0.1:0")
}

func listenerNonExemptStreamRefused(base string) error {
	d, err := postStream(base+"/v1/items", DocumentedLimit+1)
	if err != nil {
		return fmt.Errorf("listener.non-exempt-stream-refused: %s", err.Error())
	}
	if d.Accepted || d.Reason != "body exceeds limit" {
		return fmt.Errorf("listener.non-exempt-stream-refused: accepted=%t reason=%s", d.Accepted, d.Reason)
	}
	return nil
}

func listenerExemptStreamAccepted(base string) error {
	d, err := postStream(base+"/v1/uploads/media", DocumentedLimit+1)
	if err != nil {
		return fmt.Errorf("listener.exempt-stream-accepted: %s", err.Error())
	}
	if !d.Accepted || d.Reason != "exempt upload" {
		return fmt.Errorf("listener.exempt-stream-accepted: accepted=%t reason=%s", d.Accepted, d.Reason)
	}
	return nil
}

func postStream(rawURL string, n int) (Decision, error) {
	req, err := http.NewRequest(http.MethodPost, rawURL, &chunkReader{left: n, chunk: 64})
	if err != nil {
		return Decision{}, err
	}
	// 0 and -1 both send a chunked body. -1 states that the length is unknown.
	req.ContentLength = -1
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Decision{}, err
	}
	defer resp.Body.Close()
	var d Decision
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return Decision{}, err
	}
	return d, nil
}

// chunkReader yields the body in fixed pieces so the client streams it.
type chunkReader struct {
	left  int
	chunk int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if c.left == 0 {
		return 0, io.EOF
	}
	n := c.chunk
	if n > len(p) {
		n = len(p)
	}
	if n > c.left {
		n = c.left
	}
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	c.left -= n
	return n, nil
}

func outcomeFrom(fails []string) Outcome {
	if len(fails) == 0 {
		return Outcome{Result: "passed"}
	}
	reason := fails[0]
	for _, f := range fails[1:] {
		reason += ", " + f
	}
	return Outcome{Result: "failed", Reason: reason}
}

func outcomeFromErrs(errs ...error) Outcome {
	var fails []string
	blocked := 0
	for _, err := range errs {
		if err == nil {
			continue
		}
		if errors.Is(err, errDescriptionAbsent) {
			blocked++
			continue
		}
		fails = append(fails, err.Error())
	}
	if len(fails) > 0 {
		return outcomeFrom(fails)
	}
	if blocked > 0 {
		return Outcome{Result: "could not run", Reason: errDescriptionAbsent.Error()}
	}
	return Outcome{Result: "passed"}
}
