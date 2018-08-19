package testcases

var FailureUnmatchedFunctionInPostcondition = Failure{
	ID: "unmatched_function_in_postcondition",
	Text: `package somepkg

// SomeFunc does something.
//
// UnexpectedFunc ensures:
// * x > 0
func SomeFunc(x int, y int) (result string, err error) {
	return
}
`,
	Error: "failed to parse comments of the function SomeFunc on line 3: expected \"SomeFunc\" in \"ensures\" line, " +
		"but got \"UnexpectedFunc\""}
