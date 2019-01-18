package testcases

// FailureCommentParse tests that the errors coming from comment parsing are propagated.
var FailureCommentParse = Failure{
	ID: "comment_parse",
	Text: `package somepkg

// SomeFunc does something.
//
// UnexpectedFunc ensures:
//  * x > 0
func SomeFunc(x int, y int) (result string, err error) {
	return
}
`,
	Error: "failed to parse comments of the function SomeFunc on line 3: " +
		"expected function name \"SomeFunc\" in post-condition block, " +
		"but got \"UnexpectedFunc\""}
