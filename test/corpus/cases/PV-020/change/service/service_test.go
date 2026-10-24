package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExemptHTTPAcceptsPastFormerReadCeiling(t *testing.T) {
	for _, path := range []string{"/v1/uploads/media", "/v1/uploads/bundle"} {
		t.Run(path, func(t *testing.T) {
			body := bytes.NewReader(bytes.Repeat([]byte("a"), (8<<20)+1))
			req := httptest.NewRequest(http.MethodPost, path, body)
			res := httptest.NewRecorder()
			Production().Handler().ServeHTTP(res, req)
			var got Decision
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if res.Code != http.StatusOK || !got.Accepted || got.Reason != "exempt upload" {
				t.Fatalf("status=%d decision=%+v", res.Code, got)
			}
		})
	}
}
