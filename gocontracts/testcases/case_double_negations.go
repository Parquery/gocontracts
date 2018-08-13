package testcases

var DoubleNegations = Case{
	ID: "double_negations",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * !strings.HasPrefix(x, "something")
// * !strings.HasPrefix(x, "another")
//
// SomeFunc ensures:
// * !strings.HasSuffix(result, "smth else")
// * !strings.HasSuffix(result, "yet another")
func SomeFunc(x string) (result string, err error) {}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * !strings.HasPrefix(x, "something")
// * !strings.HasPrefix(x, "another")
//
// SomeFunc ensures:
// * !strings.HasSuffix(result, "smth else")
// * !strings.HasSuffix(result, "yet another")
func SomeFunc(x string) (result string, err error) {
	// Pre-conditions
	switch {
	case strings.HasPrefix(x, "something"):
		panic("Violated: !strings.HasPrefix(x, \"something\")")
	case strings.HasPrefix(x, "another"):
		panic("Violated: !strings.HasPrefix(x, \"another\")")
	default:
		// Pass
	}

	// Post-conditions
	defer func() {
		switch {
		case strings.HasSuffix(result, "smth else"):
			panic("Violated: !strings.HasSuffix(result, \"smth else\")")
		case strings.HasSuffix(result, "yet another"):
			panic("Violated: !strings.HasSuffix(result, \"yet another\")")
		default:
			// Pass
		}
	}()
}
`}
