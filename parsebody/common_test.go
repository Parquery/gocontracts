package parsebody_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Parquery/gocontracts/parsebody"
)

func parse(text string) (
	fset *token.FileSet, fn *ast.FuncDecl, bodyCmtMap ast.CommentMap, err error) {

	fset = token.NewFileSet()

	var node *ast.File
	node, err = parser.ParseFile(fset, "", text, parser.ParseComments)
	if err != nil {
		return
	}

	cmtMap := ast.NewCommentMap(fset, node, node.Comments)

	if len(node.Decls) != 1 {
		err = fmt.Errorf(
			"expected a single declaration in the function code, but got: %d",
			len(node.Decls))
		return
	}

	fn, ok := node.Decls[0].(*ast.FuncDecl)
	if !ok {
		err = fmt.Errorf(
			"expected the declaration in the function code to be a function, but got: %#v",
			node.Decls[0])
		return
	}

	bodyCmtMap = cmtMap.Filter(fn.Body)

	return
}

func checkContract(t *testing.T, text string, expected parsebody.Contract) {
	////
	// Parse test
	////

	fset, fn, bodyCmtMap, err := parse(text)
	if err != nil {
		t.Fatal(err.Error())
	}

	////
	// Parse contract from the AST
	////

	got, err := parsebody.ToContract(fset, fn, bodyCmtMap)
	if err != nil {
		t.Fatal(err.Error())
	}

	////
	// Check
	////

	msgs := []string{}

	if expected.Start != got.Start {
		msgs = append(msgs,
			fmt.Sprintf(
				"expected Start %d (%s), got %d (%s)",
				expected.Start, fset.Position(expected.Start).String(),
				got.Start, fset.Position(got.Start).String()))
	}

	if expected.End != got.End {
		msgs = append(msgs,
			fmt.Sprintf(
				"expected End %d (%s), got %d (%s)",
				expected.End, fset.Position(expected.End).String(),
				got.End, fset.Position(got.End).String()))
	}

	if expected.NextNodePos != got.NextNodePos {
		msgs = append(msgs,
			fmt.Sprintf(
				"expected NextNodePos %d (%s), got %d (%s)",
				expected.NextNodePos, fset.Position(expected.NextNodePos).String(),
				got.NextNodePos, fset.Position(got.NextNodePos).String()))
	}

	if len(msgs) > 0 {
		for i := 1; i < len(msgs); i++ {
			msgs[i] = "\t" + msgs[i]
		}
		_, file, line, _ := runtime.Caller(1)

		t.Errorf("\033[31m%s:%d: %s\033[39m\n",
			filepath.Base(file), line, strings.Join(msgs, "\n"))
	}
}
