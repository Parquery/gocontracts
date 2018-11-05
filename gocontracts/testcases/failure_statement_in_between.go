package testcases

// FailureStatementInBetween tests that we detect when there is a statement between the pre- and post-condition blocks.
var FailureStatementInBetween = Failure{
	ID: "statement_in_between",
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
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	panic("hello")

	// Post-condition
	defer func() {
		if !(result == "oi") {
			panic("Violated: result == \"oi\"")
		}
	}();

	return
}
`,
	Error: "unexpected statement between the pre- and post-condition blocks in function SomeFunc on line 18"}
