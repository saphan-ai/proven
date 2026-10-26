package generator_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Property: a live answer keeps the two facts the custody audit needs and the
// review used to drop: the model the request asked for, and the id of the
// response, which the Nightshift receipt names as its request id. The routed
// model stays what the receipt said, not what was asked.
// Recorded input: the provisioned Nightshift stream; the request asks for
// meterless/auto and the stream routes to a Nemotron model.
func TestLiveKeepsTheRequestedModelAndTheResponseID(t *testing.T) {
	body := streamFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	live := liveAt(t, srv.URL, nil)
	if _, err := live.Complete(context.Background(), "planner", "name the risks\n"); err != nil {
		t.Fatal(err)
	}
	u, ok := live.LastUse()
	if !ok {
		t.Fatal("no use reported")
	}
	if u.Requested != "meterless/auto" {
		t.Fatalf("requested %q, want meterless/auto", u.Requested)
	}
	if u.ResponseID != "chatcmpl-bbd961fe-3392-4b08-bab9-af9d24a2ae82" {
		t.Fatalf("response id %q", u.ResponseID)
	}
	if u.Model != routedNVIDIA {
		t.Fatalf("routed model %q, want %q", u.Model, routedNVIDIA)
	}
}
