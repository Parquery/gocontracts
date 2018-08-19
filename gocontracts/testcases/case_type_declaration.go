package testcases

var TypeDeclaration = Case{
	ID: "type_declaration",
	Text: `package somepkg

// SomeStruct defines a struct.
//
// SomeStruct requires:
// * x > 3
type SomeStruct struct {
	int x
}
`,
	Expected: `package somepkg

// SomeStruct defines a struct.
//
// SomeStruct requires:
// * x > 3
type SomeStruct struct {
	int x
}
`}
