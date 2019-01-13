package testcases

// FailureBodyParse tests that the errors coming from body parsing are propagated.
var FailureBodyParse = Failure{
	ID: "comment_parse",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc ensures:
//  * x > 0
func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	// Post-condition

	return
}
`,
	Error: "expected a defer statement after the comment \"Post-condition\" " +
		"in function SomeFunc on line 15"}
