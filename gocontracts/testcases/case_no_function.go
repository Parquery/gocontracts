package testcases

// NoFunction tests that files without functions are left unchanged.
var NoFunction = Case{
	ID: "no_function",
	Text: `package somepkg
`,
	Expected: `package somepkg
`}
