package review

import "github.com/saphan-ai/proven/internal/change"

// These read a configuration under the process-wide git policy. They live in a test file so that
// production code in this package cannot call them: a review reads under the policy it was opened
// with (TestNoProductionCodeReachesTheProcessGitPolicy in internal/change names the census).

func loadConfigFile(path string) (fileSettings, error) {
	return loadConfigFileUnder(change.ProcessPolicy(), path, nil)
}

func loadSettingsAtBaseline(target, baseline string) (fileSettings, error) {
	return loadSettingsAtBaselineUnder(change.ProcessPolicy(), target, baseline, nil)
}

func loadSettingsInTarget(target string) (fileSettings, error) {
	return loadSettingsInTargetUnder(change.ProcessPolicy(), target, nil)
}
