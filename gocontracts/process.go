package gocontracts

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Parquery/gocontracts/parsebody"
	"github.com/Parquery/gocontracts/parsecomment"
	"github.com/Parquery/gocontracts/parsecomment/parsecond"
)

// funcUpdate defines how a function should be updated.
type funcUpdate struct {
	contractInDoc  parsecomment.Contract
	fn             *ast.FuncDecl
	contractInBody parsebody.Contract
}

func violationMsg(c parsecond.Condition) string {
	parts := make([]string, 0, 6)

	parts = append(parts, "Violated: ")

	if len(c.Label) > 0 {
		parts = append(parts, fmt.Sprintf("%s: ", c.Label))
	}

	if len(c.InitStr) > 0 {
		parts = append(parts, fmt.Sprintf("%s; ", c.InitStr))
	}

	parts = append(parts, c.CondStr)
	msg := strings.Join(parts, "")

	return strconv.Quote(msg)
}

func negateNot(condStr string, unary *ast.UnaryExpr) string {
	if unary.Op != token.NOT {
		panic(fmt.Sprintf(
			"Expected NOT operator in the unary expression, but got: %d",
			unary.Op))
	}

	parenExpr, isParenExpr := unary.X.(*ast.ParenExpr)

	// Use .Pos() directly since the unary expression was parsed from condStr
	if !isParenExpr {
		s := unary.X.Pos() - 1
		e := unary.X.End() - 1

		return strings.Trim(condStr[s:e], " \t")
	}

	s := parenExpr.X.Pos() - 1
	e := parenExpr.X.End() - 1
	return strings.Trim(condStr[s:e], " \t")
}

// notCondStr negates the condition and returns it as a string representation of a Go boolean expression.
func notCondStr(c parsecond.Condition) string {
	switch v := c.Cond.(type) {
	case *ast.UnaryExpr:
		if v.Op == token.NOT {
			return negateNot(c.CondStr, v)
		}

		return fmt.Sprintf("!(%s)", strings.Trim(c.CondStr, " \t"))
	case *ast.ParenExpr:
		return fmt.Sprintf("!%s", strings.Trim(c.CondStr, " \t"))
	case *ast.Ident:
		return fmt.Sprintf("!%s", strings.Trim(c.CondStr, " \t"))
	}

	return fmt.Sprintf("!(%s)", strings.Trim(c.CondStr, " \t"))
}

// conditionToCode generates the condition as Golang code to be inserted
// into "if" and "switch" statements.
func conditionToCode(c parsecond.Condition) string {
	if c.InitStr == "" {
		return notCondStr(c)
	}

	return fmt.Sprintf("%s; %s", c.InitStr, notCondStr(c))
}

var tplPre = template.Must(
	template.New("preconditions").Funcs(
		template.FuncMap{
			"violationMsg":    violationMsg,
			"conditionToCode": conditionToCode,
		}).Parse(
		`{{$l := len .Pres }}{{ if eq $l 1 }}{{ $c := index .Pres 0 }}	// Pre-condition
	if {{ conditionToCode $c }} {
		panic({{ violationMsg $c }})
	}
{{- else }}	// Pre-conditions
	switch { {{- range .Pres }}
	case {{ conditionToCode . }}:
		panic({{ violationMsg . }})
{{- end }}
	default:
		// Pass
	}
{{- end }}`))

var tplPost = template.Must(
	template.New("postconditions").Funcs(
		template.FuncMap{
			"violationMsg":    violationMsg,
			"conditionToCode": conditionToCode,
		}).Parse(
		`{{$l := len .Posts }}{{ if eq $l 1 }}{{ $c := index .Posts 0 }}	// Post-condition
	defer func() {
		if {{ conditionToCode $c }} {
			panic({{ violationMsg $c }})
		}
	}()
{{- else }}	// Post-conditions
	defer func() {
		switch { {{- range .Posts }}
		case {{ conditionToCode . }}:
			panic({{ violationMsg . }})
		{{- end }}
		default:
			// Pass
		}
	}()
{{- end }}`))

// generateCode generates the code of the two blocks.
//
// The first line of generated code is indented.
// The generated code does not end with a new-line character.
func generateCode(contract parsecomment.Contract) (code string, err error) {
	// Post-condition
	defer func() {
		if strings.HasSuffix(code, "\n") {
			panic("Violated: strings.HasSuffix(code, \"\\n\")")
		}
	}()

	blocks := []string{}

	if len(contract.Pres) > 0 {
		var buf bytes.Buffer
		err = tplPre.Execute(&buf, contract)
		if err != nil {
			return
		}

		blocks = append(blocks, buf.String())
	}

	if len(contract.Preamble) > 0 {
		// Since Golang package text/template does not contain "indent" filter,
		// we manually indent the code and do not use a template here.
		var buf bytes.Buffer
		buf.WriteString("\t// Preamble starts.\n")

		lines := strings.Split(contract.Preamble, "\n")
		for _, line := range lines {
			buf.WriteString("\t")
			buf.WriteString(line)
			buf.WriteString("\n")
		}

		buf.WriteString("\t// Preamble ends.")
		blocks = append(blocks, buf.String())
	}

	if len(contract.Posts) > 0 {
		var buf bytes.Buffer
		err = tplPost.Execute(&buf, contract)
		if err != nil {
			return
		}

		blocks = append(blocks, buf.String())
	}

	code = strings.Join(blocks, "\n\n")
	return
}

// updateEmptyFunc updates the function which contains no body.
// cursor points to the end of the function.
func updateEmptyFunc(fset *token.FileSet, up funcUpdate, code string, writer *bytes.Buffer) (cursor int) {
	// The function contains no statements except the conditions so we can simply fill it out.
	cursor = fset.Position(up.fn.Body.Rbrace).Offset // Move cursor to the end of the function

	if len(code) > 0 {
		writer.WriteRune('\n')
		writer.WriteString(code)
		writer.WriteRune('\n')
	}

	return
}

// updateSingleLineFunc updates the function whose body is a single line of statements separated by ';'.
// cursor points just after the right brace of the function definition.
func updateSingleLineFunc(
	fset *token.FileSet, up funcUpdate, code string, text string, writer *bytes.Buffer) (cursor int) {

	lbraceOffset := fset.Position(up.fn.Body.Lbrace).Offset
	rbraceOffset := fset.Position(up.fn.Body.Rbrace).Offset

	if len(code) > 0 {
		writer.WriteRune('\n')
		writer.WriteString(code)

		// Add an additional new line to nicely separate the contract conditions from the rest of the code
		writer.WriteString("\n\n")

		// Add an indention so that the statements on the single line are indented properly after the contract
		// conditions.
		writer.WriteRune('\t')

		// Write the function body

		fstStmtOffset := fset.Position(up.contractInBody.NextNodePos).Offset

		writer.WriteString(strings.TrimRight(text[fstStmtOffset:rbraceOffset], "\t "))

		// Write a new line so that the previous single-line function is nicely reformatted as a multi-line
		// function.
		writer.WriteString("\n}")
	} else {
		writer.WriteString(text[lbraceOffset+1 : rbraceOffset+1])
	}

	cursor = rbraceOffset + 1

	return
}

// updateMultilineFunc updates the function whose body after the conditions is not empty.
// cursor points to the next statement after the conditions.
func updateMultilineFunc(
	fset *token.FileSet, up funcUpdate, code string, text string, writer *bytes.Buffer) (cursor int) {

	lbraceOffset := fset.Position(up.fn.Body.Lbrace).Offset

	cursor = fset.Position(up.contractInBody.NextNodePos).Offset

	// Go back in order to include a farthest possible end of the last condition block
	for text[cursor] != '\n' && text[cursor] != ';' {
		cursor--

		if cursor == lbraceOffset {
			panic(fmt.Sprintf(
				"cursor reached the left brace of the function %s at %s:%d",
				up.fn.Name.Name, fset.File(up.fn.Pos()).Name(), lbraceOffset))
		}
	}

	if len(code) > 0 {
		writer.WriteRune('\n')
		writer.WriteString(code)

		if text[cursor] == ';' {
			// Keep the semi-colon at the same last line of the last condition block if it was already there in the
			// previous code.
		} else {
			// Add an additional new line to nicely separate the contract conditions from the rest of the code
			writer.WriteRune('\n')
		}
	}

	return
}

func update(text string, updates []funcUpdate, fset *token.FileSet) (updated string, err error) {
	writer := bytes.NewBufferString("")

	cursor := 0

	for _, up := range updates {
		lbraceOffset := fset.Position(up.fn.Body.Lbrace).Offset

		// Write the prefix
		writer.WriteString(text[cursor : lbraceOffset+1])

		var code string
		code, err = generateCode(up.contractInDoc)
		if err != nil {
			return
		}

		switch {
		case up.contractInBody.NextNodePos == token.NoPos:
			// The function contains no statements except the conditions so we can simply fill it out.

			cursor = updateEmptyFunc(fset, up, code, writer)

		case fset.Position(up.fn.Body.Lbrace).Line == fset.Position(up.fn.Body.Rbrace).Line:
			// The function contains statements on the same lines as the braces.

			// Assert
			if up.contractInBody.Start != token.NoPos {
				panic(fmt.Sprintf(
					"unexpected contract in a single-line function %s at %s:%d",
					up.fn.Name, fset.File(up.contractInBody.Start).Name(),
					fset.Position(up.contractInBody.Start).Line))
			}

			cursor = updateSingleLineFunc(fset, up, code, text, writer)

		default:
			// The function contains one or more statements and possibly a previously generated code to
			// verify contracts.

			cursor = updateMultilineFunc(fset, up, code, text, writer)
		}
	}

	// Write the suffix
	writer.WriteString(text[cursor:])

	updated = writer.String()
	return
}

// Process automatically adds (or updates) the blocks for checking the pre and postconditions.
// If remove is set, the code to check the conditions is removed, but the conditions are left untouched
// in the comment.
func Process(text string, filename string, remove bool) (updated string, err error) {
	fset := token.NewFileSet()

	var node *ast.File
	node, err = parser.ParseFile(fset, filename, text, parser.ParseComments)
	if err != nil {
		return
	}

	cmtMap := ast.NewCommentMap(fset, node, node.Comments)

	updates := []funcUpdate{}

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		////
		// Parse comment
		////

		name := fn.Name.Name
		commentLines := strings.Split(fn.Doc.Text(), "\n")

		var contractInDoc parsecomment.Contract

		if !remove {
			contractInDoc, err = parsecomment.ToContract(name, commentLines)
			if err != nil {
				err = fmt.Errorf("failed to parse comments of the function %s on line %d: %s",
					name, fset.Position(fn.Doc.Pos()).Line, err)
				return
			}
		} else {
			// Remove is true, hence leave the pre and postconditions empty.
		}

		////
		// Parse body
		////

		bodyCmtMap := cmtMap.Filter(fn.Body)

		var contractInBody parsebody.Contract
		contractInBody, err = parsebody.ToContract(fset, fn, bodyCmtMap)
		if err != nil {
			return
		}

		////
		// Specify the update
		////

		// Update only if there is something to actually change.
		if len(contractInDoc.Pres) == 0 &&
			len(contractInDoc.Preamble) == 0 &&
			len(contractInDoc.Posts) == 0 &&
			contractInBody.Start == token.NoPos {
			continue
		}

		updates = append(updates,
			funcUpdate{
				contractInDoc:  contractInDoc,
				fn:             fn,
				contractInBody: contractInBody,
			})
	}

	if len(updates) == 0 {
		updated = text
		return
	}

	updated, err = update(text, updates, fset)
	if err != nil {
		return
	}

	return
}

// ProcessFile loads the Go file and processes it.
// If remove is set, the code to check the conditions is removed, but the conditions are left untouched
// in the comment.
func ProcessFile(pth string, remove bool) (updated string, err error) {
	data, err := ioutil.ReadFile(pth)
	if err != nil {
		err = fmt.Errorf("failed to read: %s", err)
		return
	}

	text := string(data)

	updated, err = Process(text, pth, remove)
	if err != nil {
		return
	}

	return
}

// ProcessInPlace loads the Go file in memory, proesses it and writes atomically back to the file.
// If remove is set, the code to check the conditions is removed, but the conditions are left untouched
// in the comment.
func ProcessInPlace(pth string, remove bool) (err error) {
	var updated string
	updated, err = ProcessFile(pth, remove)
	if err != nil {
		return
	}

	var tmp *os.File
	tmp, err = ioutil.TempFile(filepath.Dir(pth), "temporary-gocontracts-"+filepath.Base(pth))
	if err != nil {
		return
	}
	defer func() {
		_, statErr := os.Stat(tmp.Name())
		switch {
		case statErr == nil:
			removeErr := os.Remove(tmp.Name())
			if removeErr != nil {
				err = fmt.Errorf("failed to remove %s: %s", tmp.Name(), err.Error())
				return
			}
		case os.IsNotExist(statErr):
			// Pass

		default:
			err = fmt.Errorf("failed to stat %s: %s", tmp.Name(), statErr.Error())
			return
		}
	}()

	err = ioutil.WriteFile(tmp.Name(), []byte(updated), 0600)
	if err != nil {
		err = fmt.Errorf("failed to write to %s: %s", tmp.Name(), err.Error())
		return
	}

	err = os.Rename(tmp.Name(), pth)
	if err != nil {
		err = fmt.Errorf("failed to move %s to %s: %s", tmp.Name(), pth, err.Error())
		return
	}

	return
}
