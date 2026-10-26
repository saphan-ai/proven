package hmem_test

import (
	"encoding/json"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// TestReturnedRecordsShareNoStorage proves the property the Dream doc comment
// relies on: a Record handed out by a public call carries its own copy of every
// slice field. For each public path it edits every returned slice in place,
// then checks that the stored record is unchanged and that Dream still
// proposes no domain. Without the copy, Get(id).Tags[0] = "tech:go" changes the
// stored tags while the stored domain stays "general", and Dream proposes
// "tech" with no classifier change.
func TestReturnedRecordsShareNoStorage(t *testing.T) {
	paths := []struct {
		name string
		// returned adds one plain note, then returns the record a public call
		// handed out for it.
		returned func(t *testing.T, s *hmem.Store) hmem.Record
	}{
		{"Add", func(t *testing.T, s *hmem.Store) hmem.Record {
			rec, err := s.Add(hmem.AddInput{Content: "plain note", Tags: []string{"misc"}, Source: "t"})
			if err != nil {
				t.Fatal(err)
			}
			return rec
		}},
		{"Get", func(t *testing.T, s *hmem.Store) hmem.Record {
			rec := addPlain(t, s)
			got, ok := s.Get(rec.ID)
			if !ok {
				t.Fatal("record missing")
			}
			return got
		}},
		{"All", func(t *testing.T, s *hmem.Store) hmem.Record {
			addPlain(t, s)
			all := s.All()
			if len(all) != 1 {
				t.Fatalf("All returned %d records, want 1", len(all))
			}
			return all[0]
		}},
		{"Query", func(t *testing.T, s *hmem.Store) hmem.Record {
			addPlain(t, s)
			res, err := s.Query("plain note", 5, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Hits) != 1 {
				t.Fatalf("Query returned %d hits, want 1", len(res.Hits))
			}
			return res.Hits[0].Record
		}},
	}
	for _, p := range paths {
		t.Run(p.name, func(t *testing.T) {
			s := openApprove(t, t.TempDir(), t.TempDir())
			rec := p.returned(t, s)
			// The JSON text is a snapshot no slice can alias.
			before := snapshotJSON(t, s, rec.ID)
			scribble(rec)
			if after := snapshotJSON(t, s, rec.ID); after != before {
				t.Fatalf("editing a returned record changed the store:\nbefore %s\nafter  %s", before, after)
			}
			ps, err := s.Dream()
			if err != nil {
				t.Fatal(err)
			}
			for _, pr := range ps {
				if pr.Type == "domain" {
					t.Fatalf("a domain proposal arose from editing a returned record: %+v", pr)
				}
			}
		})
	}
}

// snapshotJSON is the stored record as JSON text.
func snapshotJSON(t *testing.T, s *hmem.Store, id string) string {
	t.Helper()
	rec, ok := s.Get(id)
	if !ok {
		t.Fatal("record missing")
	}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// addPlain stores the plain note the test edits through each public path.
func addPlain(t *testing.T, s *hmem.Store) hmem.Record {
	t.Helper()
	rec, err := s.Add(hmem.AddInput{Content: "plain note", Tags: []string{"misc"}, Source: "t"})
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// scribble edits every slice field of a returned record in place.
func scribble(rec hmem.Record) {
	rec.Tags[0] = "tech:go"
	for i := range rec.Embedding {
		rec.Embedding[i] = 9
	}
	for i := range rec.Entities {
		rec.Entities[i] = "edited"
	}
	for i := range rec.RelatedTo {
		rec.RelatedTo[i] = "edited"
	}
	for i := range rec.DerivedFrom {
		rec.DerivedFrom[i] = "edited"
	}
}
