package testcases

// Case defines a use case and the expected output.
type Case struct {
	ID string

	// Input
	Text string

	// The value of remove argument to Process
	Remove bool

	// Expected code after the Text was processed
	Expected string
}

// Failure defines a case when gocontracts should fail and the expected error.
type Failure struct {
	ID    string
	Text  string
	Error string
}
