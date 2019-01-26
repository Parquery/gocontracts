// Package parsecond parses the condition from the function documentation.
package parsecond

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// Condition defines a pre- or a post-condition of the function's contract.
type Condition struct {
	Label string

	// InitStr corresponds to initialization statement in conditions,
	// e.g., given
	// 	_, ok := someMap[3]; ok
	//
	// , the initialization is:
	//	_, ok := someMap[3]
	//
	// See https://golang.org/ref/spec#If_statements
	InitStr string

	CondStr string
	Cond    ast.Expr
}

var bulletRe = regexp.MustCompile(`^\s*\*\s*(.*)\s*$`)
var labelWithCondRe = regexp.MustCompile(
	`^([a-zA-Z0-9_;.\-=' ]+\s*:)([ \t^=]?.*)$`)

// ToCondition tries to parse the condition from text.
//
// If no condition could be matched, cond is nil.
// Err is set if the text matched the bullet format, but there was
// an error parsing the condition.
func ToCondition(text string) (cond *Condition, err error) {
	////
	// Parse the text as a bullet item
	////

	mtchs := bulletRe.FindStringSubmatch(text)

	if len(mtchs) == 0 {
		return
	}

	content := mtchs[1]

	////
	// Parse the content of the bullet as condition
	////

	mtchs = labelWithCondRe.FindStringSubmatch(content)

	var label string
	var parsable string
	if len(mtchs) == 0 {
		parsable = content
	} else {
		label = strings.TrimSuffix(
			strings.Trim(mtchs[1], " \t"),
			":")

		parsable = strings.Trim(mtchs[2], " \t")
	}

	////
	// Parse the Golang code
	////

	playground := fmt.Sprintf(`package main

func main() {
	if %s {
		// Do something
	}
}
`, parsable)

	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", playground, parser.AllErrors)

	if err != nil {
		err = fmt.Errorf(
			"failed to parse the condition "+
				"in the following playground:\n"+
				"%s\n"+
				"The error was: %s",
			playground, err.Error())
		return
	}

	funcDecl := f.Decls[0].(*ast.FuncDecl)
	ifStmt := funcDecl.Body.List[0].(*ast.IfStmt)

	var initStr string
	if ifStmt.Init != nil {
		s := fs.Position(ifStmt.Init.Pos()).Offset
		e := fs.Position(ifStmt.Cond.Pos()).Offset

		initStr = strings.Trim(playground[s:e], " \t")
		initStr = strings.TrimSuffix(initStr, ";")
	}

	s := fs.Position(ifStmt.Cond.Pos()).Offset
	e := fs.Position(ifStmt.Cond.End()).Offset
	condStr := strings.Trim(playground[s:e], " \t")

	// We need to re-parse CondStr so that the tokens of the parsed expression
	// correspond to CondStr and not to the playground.
	var expr ast.Expr
	expr, err = parser.ParseExpr(condStr)
	if err != nil {
		return
	}

	cond = &Condition{
		Label:   label,
		InitStr: initStr,
		CondStr: condStr,
		Cond:    expr}

	return
}
