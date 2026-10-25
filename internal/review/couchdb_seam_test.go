package review

// The review-level proof that a live review's evidence names the form the
// store ran is TestReviewEvidenceNamesTheFormTheStoreRan. This file used to
// pin a seam that sent SQL and named CouchDB down on every query. That seam
// is gone: the review passes the template, and CouchDB runs the mango form
// (ADR-0047).
