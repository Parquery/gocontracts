package testcases

// FailureUnparsableFile tests that the errors coming from parsing a file are propagated.
var FailureUnparsableFile = Failure{
	ID: "unparsable_file",
	Text: `package somepkg

// SomeFunc does something.
//
// UnexpectedFunc ensures:
//  * x > 0
func SomeFunc(x int, y int) (result string, err error) {
	result =
}
`,
	Error: "unparsable_file:9:1: expected operand, found '}'"}
