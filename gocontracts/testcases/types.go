package testcases

type Case struct {
	ID       string
	Text     string
	Expected string
}

type Failure struct {
	ID string
	Text string
	Error string
}
