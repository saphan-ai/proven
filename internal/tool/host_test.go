package tool

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/testrun"
)

// TestContainerHostNamesAMissingLimit proves a host with no positive limit
// names the failure and does not fall back onto the test runner's limit.
func TestContainerHostNamesAMissingLimit(t *testing.T) {
	box, err := testrun.NewContainer(testrun.DefaultContainerConfig(), func(string) (string, error) {
		return "/bin/docker", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = (ContainerHost{Box: box}).Run(context.Background(), t.TempDir(), []string{"go", "test", "-race", "."})
	if err == nil || !strings.Contains(err.Error(), "time limit") {
		t.Fatalf("%v", err)
	}
	_, err = (ContainerHost{Timeout: time.Minute}).Run(context.Background(), t.TempDir(), []string{"go", "test", "-race", "."})
	if err == nil || !strings.Contains(err.Error(), "no container") {
		t.Fatalf("%v", err)
	}
}
