package testcases

// Case defines a use case and the expected output.
type Case struct {
	ID       string
	Text     string
	Expected string
}

// Failure defines a case when gocontracts should fail and the expected error.
type Failure struct {
	ID    string
	Text  string
	Error string
}
