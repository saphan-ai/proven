package review

// This file writes a review's findings to CouchDB as evidence records and
// gives each stored finding the address of its record in the viewer page. The
// documents, the views and the page are in internal/store; this file is the
// seam between them and the Result.

import (
	"context"
	"time"

	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/store"
)

// recordsService names CouchDB in a review's failed services when its records
// could not be written or its viewer could not be read.
const recordsService = "couchdb"

// envGitHubRef is the variable the Action's runner sets to refs/pull/<n>/merge
// on a pull request. It is how a record learns its pull request: the review's
// request carries commits, not a pull request number.
const envGitHubRef = "GITHUB_REF"

// recorder writes one review's records. It is nil when PROVEN_COUCHDB_RECORDS
// is unset, and a nil recorder writes nothing (ADR-0013).
type recorder struct {
	// writer writes the records. Its account is the writing one.
	writer *store.Recorder
	// reader opens the viewer page as the read-only account, to see that the
	// reader's own account is served it before a link to it is shown. It is
	// nil when the read-only account is not configured, and the link is then
	// shown unchecked, with a failed service that says so.
	reader *store.CouchDB
	// linkBase is PROVEN_EVIDENCE_LINK_BASE_URL; empty means no link.
	linkBase string
	// pr is the pull request number, "" outside a pull request.
	pr  string
	now func() time.Time
}

// publish stores the Result's findings and sets each stored finding's
// EvidenceRecordURL. A failure is a failed service named couchdb: the review
// goes on, and a finding whose record was not stored has no link, so a link
// never points at nothing.
func (rc *recorder) publish(ctx context.Context, req Request, res *Result) {
	if rc == nil || len(res.Findings) == 0 {
		return
	}
	// The review's own deadline may be over by now (a partial review); the
	// records are still owed for what it proved.
	ctx = context.WithoutCancel(ctx)
	run := store.RunStamp(rc.now())
	recs := make([]store.EvidenceRecord, len(res.Findings))
	for i, f := range res.Findings {
		recs[i] = recordOf(f, rc.pr, req.HeadCommit, run)
	}
	ids, err := rc.writer.Publish(ctx, recs)
	if err != nil {
		res.FailedServices = append(res.FailedServices, ServiceFailure{Service: recordsService, Reason: "records: " + reasonOf(err)})
	}
	if rc.linkBase == "" {
		return
	}
	if rc.reader == nil {
		// The link is shown, but the page could not be checked as the read-only account because
		// that account is not configured. Say so, so a reader is not left to find out by a dead link.
		res.FailedServices = append(res.FailedServices, ServiceFailure{Service: recordsService, Reason: "viewer: link not checked, the read-only account is not configured"})
	} else {
		if cerr := rc.reader.CheckViewer(ctx, rc.writer.Database()); cerr != nil {
			res.FailedServices = append(res.FailedServices, ServiceFailure{Service: recordsService, Reason: "viewer: " + reasonOf(cerr)})
			return
		}
	}
	for i, f := range res.Findings {
		if id, ok := ids[f.ID]; ok {
			res.Findings[i].EvidenceRecordURL = store.EvidenceURL(rc.linkBase, rc.writer.Database(), id)
		}
	}
}

// reasonOf is the text a failed service shows: the store's own reason for a
// down store, which never holds a credential.
func reasonOf(err error) string {
	if d, ok := service.As(err); ok {
		return d.Reason
	}
	return "failed"
}

// recordOf copies one finding into the document that stores it.
func recordOf(f Finding, pr, commit, run string) store.EvidenceRecord {
	rec := store.EvidenceRecord{
		PR: pr, Commit: commit, Run: run, FindingID: f.ID, RuleID: f.RuleID,
		Title: f.Title, Message: f.Message, Severity: string(f.Severity),
		Path: f.Location.Path, StartLine: f.Location.StartLine, EndLine: f.Location.EndLine,
		ProofLevel: string(f.ProofLevel), Unresolved: f.Unresolved,
	}
	seenCVE, seenPkg := map[string]bool{}, map[string]bool{}
	for _, e := range f.Evidence {
		p := store.RecordProof{Kind: e.Kind}
		switch {
		case e.Query != nil:
			q := e.Query
			p.Form, p.Query, p.Engine, p.Snapshot = q.Form, q.SQL, q.Engine, q.Snapshot
			p.Columns, p.Rows = q.Columns, q.Rows
			cves, pkgs := store.QueryKeys(q.Columns, q.Rows)
			for _, c := range cves {
				if !seenCVE[c] {
					seenCVE[c] = true
					rec.CVEs = append(rec.CVEs, c)
				}
			}
			for _, k := range pkgs {
				if !seenPkg[k] {
					seenPkg[k] = true
					rec.Packages = append(rec.Packages, k)
				}
			}
		case e.Test != nil:
			t := e.Test
			p.Path, p.Name, p.Command, p.Output = t.Path, t.Name, t.Command, t.Output
			p.FailsOnChange, p.PassesOnBase = t.FailsOnChange, t.PassesOnBase
		case e.Tool != nil:
			t := e.Tool
			p.Tool, p.Version, p.Command, p.Output, p.BaseOutput = t.Tool, t.Version, t.Command, t.Output, t.BaseOutput
			p.PresentOnChange, p.PresentOnBase = t.PresentOnChange, t.PresentOnBase
		}
		rec.Evidence = append(rec.Evidence, p)
	}
	return rec
}
