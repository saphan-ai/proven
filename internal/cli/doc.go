// Package cli is Proven's command-line shell: the part a pipeline touches.
//
// It turns `proven scan [TARGET]` and `proven ci` into one review.Request,
// hands it to the reviewer through internal/review, and turns the
// review.Result into output and an exit code (text today; JSON and SARIF
// use the same Formatter seam). The text form also prints the section
// "Agreed by two models, not proven" apart from the findings (ADR-0041).
// It does no reviewing of its own: the six agents (Planner, Investigator, Prover,
// Skeptic, Judge and Writer) live behind review.Open, and this package
// never reaches past that seam.
//
// The interface is ADR-0020's, so a team moving from Semgrep or an
// Opengrep action changes as few pipeline lines as possible (ADR-0016):
// the same flag names, a variable for each flag, and the same exit codes.
// Every setting resolves in one order, flag over variable over default,
// and a value outside its allowed set is refused at load with the allowed
// list (ADR-0013), never guessed.
package cli
