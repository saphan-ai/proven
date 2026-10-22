package testrun_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

func TestSampleReturnsTheRecordedTest(t *testing.T) {
	s := testrun.NewSample([]testrun.Recorded{{
		FindingID:     "F1",
		Path:          "billing/discount_test.go",
		Name:          "TestDiscounted",
		Command:       "go test ./billing -run TestDiscounted",
		Output:        "change: FAIL",
		FailsOnChange: true,
		PassesOnBase:  true,
	}})
	got, err := s.Run(context.Background(), testrun.Request{FindingID: "F1", Name: "TestDiscounted"})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Found || !got.FailsOnChange || !got.PassesOnBase {
		t.Fatalf("%#v", got)
	}
	missing, err := s.Run(context.Background(), testrun.Request{FindingID: "F1", Name: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	if missing.Found {
		t.Fatal("a different test name must not match")
	}
	withFile, err := s.Run(context.Background(), testrun.Request{
		FindingID: "F1", Name: "TestDiscounted", Content: "package billing\n", Path: "billing/extra_test.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !withFile.Found || !withFile.FailsOnChange || !withFile.PassesOnBase {
		t.Fatal("sample mode must ignore written content and execute nothing")
	}
}

func TestLiveNamesProverOff(t *testing.T) {
	_, err := testrun.Live{}.Run(context.Background(), testrun.Request{FindingID: "F1", Name: "TestDiscounted"})
	d, ok := service.As(err)
	if !ok {
		t.Fatal(err)
	}
	if d.Service != "prover" || d.Reason != testrun.ProverOffReason {
		t.Fatalf("%s %s", d.Service, d.Reason)
	}
}
