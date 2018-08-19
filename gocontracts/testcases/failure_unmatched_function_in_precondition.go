package testcases

// FailureUnmatchedFunctionInPrecondition tests that we detect when the function name in the pre-condition
// documentation differs from the actual function name.
var FailureUnmatchedFunctionInPrecondition = Failure{
	ID: "unmatched_function_in_precondition",
	Text: `package somepkg

// SomeFunc does something.
//
// UnexpectedFunc requires:
// * x > 0
func SomeFunc(x int, y int) (result string, err error) {
	return
}
`,
	Error: "failed to parse comments of the function SomeFunc on line 3: expected \"SomeFunc\" " +
		"in \"requires\" line, but got \"UnexpectedFunc\""}
