package testcases

// DoubleNegation tests that condition expressions are correctly checked in If statement with double negation.
var DoubleNegation = Case{
	ID: "double_negation",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * !strings.HasPrefix(x, "something")
//
// SomeFunc ensures:
// * !strings.HasSuffix(result, "smth else")
func SomeFunc(x string) (result string, err error) {}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * !strings.HasPrefix(x, "something")
//
// SomeFunc ensures:
// * !strings.HasSuffix(result, "smth else")
func SomeFunc(x string) (result string, err error) {
	// Pre-condition
	if strings.HasPrefix(x, "something") {
		panic("Violated: !strings.HasPrefix(x, \"something\")")
	}

	// Post-condition
	defer func() {
		if strings.HasSuffix(result, "smth else") {
			panic("Violated: !strings.HasSuffix(result, \"smth else\")")
		}
	}()
}
`}
