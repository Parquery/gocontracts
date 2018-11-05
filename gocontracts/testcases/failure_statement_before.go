package testcases

// FailureStatementBefore tests that we detect when there is a statement preceding the condition checks.
var FailureStatementBefore = Failure{
	ID: "statement_before",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//
// SomeFunc ensures:
//  * result == "oi"
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	panic("hello")

	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	// Post-condition
	defer func() {
		if !(result == "oi") {
			panic("Violated: result == \"oi\"")
		}
	}();

	return
}
`,
	Error: "unexpected statement before the comment \"Pre-condition\" in function SomeFunc on line 13"}
