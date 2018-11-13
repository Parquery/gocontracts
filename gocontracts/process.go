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
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var requiresRe = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z_0-9]*)\s+requires\s*:\s*$`)
var ensuresRe = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z_0-9]*)\s+ensures\s*:\s*$`)
var bulletRe = regexp.MustCompile(`\s*\*\s*(([^:]+)\s*:\s*)?(\s*.*\s*)$`)

type condition struct {
	label   string
	condStr string
	cond    ast.Expr
}

func parseContracts(name string, commentLines []string) (pres, posts []condition, err error) {
	const (
		stateText = iota
		stateRequires
		stateEnsures
	)

	pres = make([]condition, 0, 5)
	posts = make([]condition, 0, 5)

	state := stateText

	for _, line := range commentLines {
		// Ignore empty lines
		if len(strings.Trim(line, " \t")) == 0 {
			continue
		}

		mtchs := requiresRe.FindStringSubmatch(line)
		if len(mtchs) > 0 {
			gotName := mtchs[1]
			if name != gotName {
				err = fmt.Errorf("expected %#v in \"requires\" line, but got %#v",
					name, gotName)
				return
			}

			state = stateRequires
			continue
		}

		mtchs = ensuresRe.FindStringSubmatch(line)
		if len(mtchs) > 0 {
			gotName := mtchs[1]
			if name != gotName {
				err = fmt.Errorf("expected %#v in \"ensures\" line, but got %#v",
					name, gotName)
				return
			}

			state = stateEnsures
			continue
		}

		// Skip all the lines that are not preceded by a requires/ensures marker
		if state == stateText {
			continue
		}

		mtchs = bulletRe.FindStringSubmatch(line)
		if len(mtchs) > 0 {
			label := mtchs[2]
			exprStr := mtchs[3]

			// Check that the condition is parsable
			var expr ast.Expr
			expr, err = parser.ParseExpr(exprStr)
			if err != nil {
				return
			}

			cond := condition{label: label, condStr: exprStr, cond: expr}

			switch state {
			case stateRequires:
				pres = append(pres, cond)
			case stateEnsures:
				posts = append(posts, cond)
			case stateText:
				panic("expected to continue before if state == stateText, but got here")
			}
		} else {
			// If a line does not match the bullet pattern, assume a text paragraph starts
			state = stateText
		}
	}
	return
}

var preconditionRe = regexp.MustCompile(`^(Precondition|Pre-condition)s?\s*:?`)
var postconditionRe = regexp.MustCompile(`^(Postcondition|Post-condition)s?\s*:?`)

// condPositions indicates the start and the end positions of the condition blocks.
//
// -1 indicate that the block was not found.
type condPositions struct {
	preStart token.Pos
	preEnd   token.Pos

	postStart token.Pos
	postEnd   token.Pos

	// indicates the position of the first AST node after the contract conditions.
	// If there are no nodes in the function body after the conditions, nextNodePos is token.NoPos.
	nextNodePos token.Pos
}

// parsePreconditions parses the pre-conditions defined in the function body.
func parsePreconditions(
	fset *token.FileSet, fn *ast.FuncDecl, cmtGrp *ast.CommentGroup) (preStart, preEnd token.Pos, err error) {

	preStart = cmtGrp.Pos()

	cmtText := strings.Trim(cmtGrp.Text(), "\n \t")

	// Check that there are no statements between the pre-condition comment in the function body
	if len(fn.Body.List) > 0 && fn.Body.List[0].Pos() < preStart {
		err = fmt.Errorf("unexpected statement before the comment %#v in function %s on line %d",
			cmtText, fn.Name.String(), fset.Position(fn.Body.List[0].Pos()).Line)
		return
	}

	// Check that there is a block following the comment
	var stmtAfterCmt ast.Stmt
	for _, stmt := range fn.Body.List {
		if stmt.Pos() > preStart {
			stmtAfterCmt = stmt
			break
		}
	}

	if stmtAfterCmt == nil {
		err = fmt.Errorf("found no statement after the comment %v in function %s on line %d",
			cmtText, fn.Name.String(), fset.Position(preStart).Line)
		return
	}

	switch {
	case strings.HasPrefix(cmtText, "Pre-conditions"):
		// Expect multiple pre-conditions given the comment and hence a switch
		_, ok := stmtAfterCmt.(*ast.SwitchStmt)

		if !ok {
			err = fmt.Errorf(
				"expected a 'switch' statement after the comment %#v in function %s on line %d",
				cmtText, fn.Name.String(), fset.Position(stmtAfterCmt.Pos()).Line)
			return
		}

	case strings.HasPrefix(cmtText, "Pre-condition"):
		// Expect a single pre-condition given the comment and hence a
		_, ok := stmtAfterCmt.(*ast.IfStmt)
		if !ok {
			err = fmt.Errorf(
				"expected an 'if' statement after the comment %#v in function %s on line %d",
				cmtText, fn.Name.String(), fset.Position(stmtAfterCmt.Pos()).Line)
			return
		}
	default:
		panic(fmt.Sprintf("Unhandled comment text: %#v", cmtText))
	}

	preEnd = stmtAfterCmt.End()

	return
}

// parsePostconditions parses the post-conditions defined in the function body.
func parsePostconditions(
	fset *token.FileSet, fn *ast.FuncDecl, cmtGrp *ast.CommentGroup) (postStart, postEnd token.Pos, err error) {
	postStart = cmtGrp.Pos()

	cmtText := strings.Trim(cmtGrp.Text(), "\n \t")

	// Check that there is a defer following the comment
	var stmtAfterCmt ast.Stmt
	for _, stmt := range fn.Body.List {
		if stmt.Pos() > postStart {
			stmtAfterCmt = stmt
			break
		}
	}

	if stmtAfterCmt == nil {
		err = fmt.Errorf("found no statement after the comment %#v in function %s on line %d",
			cmtText, fn.Name.String(), fset.Position(cmtGrp.Pos()).Line)
		return
	}

	deferStmt, ok := stmtAfterCmt.(*ast.DeferStmt)
	if !ok {
		err = fmt.Errorf("expected a defer statement after the comment %#v in function %s on line %d",
			cmtText, fn.Name.String(), fset.Position(stmtAfterCmt.Pos()).Line)
		return
	}

	postEnd = deferStmt.End()
	return
}

// parseConditions parses the pre and post-conditions from the function body.
// bodyCmtMap should contain only the comments written in the function body.
func parseConditions(
	fset *token.FileSet, fn *ast.FuncDecl,
	bodyCmtMap ast.CommentMap) (preStart, preEnd, postStart, postEnd token.Pos, err error) {

	for _, cmtGrp := range bodyCmtMap.Comments() {
		cmtText := strings.Trim(cmtGrp.Text(), "\n \t")
		switch {
		case preconditionRe.MatchString(cmtText):
			preStart, preEnd, err = parsePreconditions(fset, fn, cmtGrp)
			if err != nil {
				return
			}

		case postconditionRe.MatchString(cmtText):
			postStart, postEnd, err = parsePostconditions(fset, fn, cmtGrp)
			if err != nil {
				return
			}

		default:
			// pass
		}
	}

	return
}

func findNextNodePos(bodyCmtMap ast.CommentMap, body *ast.BlockStmt, conditionsEnd token.Pos) (nextNodePos token.Pos) {
	nextNodePos = token.NoPos

	for _, stmt := range body.List {
		if stmt.Pos() > conditionsEnd {
			nextNodePos = stmt.Pos()
			break
		}
	}

	// See if there is a comment before the conditions and the first next statement
	for _, cmtGrp := range bodyCmtMap.Comments() {
		if cmtGrp.Pos() > conditionsEnd &&
			(nextNodePos == token.NoPos || cmtGrp.Pos() < nextNodePos) {

			nextNodePos = cmtGrp.Pos()
			break
		}
	}

	return
}

// findBlocks searches for the start and end of the pre-condition and post-condition block, respectively.
func findBlocks(
	fset *token.FileSet, fn *ast.FuncDecl, cmtMap ast.CommentMap) (p condPositions, err error) {

	bodyCmtMap := cmtMap.Filter(fn.Body)

	p.preStart, p.preEnd, p.postStart, p.postEnd, err = parseConditions(fset, fn, bodyCmtMap)

	// Check that there are no statements between pre-end and post-start
	if p.preStart != -1 && p.postStart != -1 {
		for _, stmt := range fn.Body.List {
			if stmt.Pos() >= p.preEnd && stmt.Pos() < p.postStart {
				err = fmt.Errorf(
					"unexpected statement between the pre- and post-condition blocks in function %s on line %d",
					fn.Name.String(), fset.Position(stmt.Pos()).Line)
				return
			}
		}
	}

	conditionsEnd := token.NoPos
	switch {
	case p.postEnd != token.NoPos:
		conditionsEnd = p.postEnd
	case p.preEnd != token.NoPos:
		conditionsEnd = p.preEnd
	default:
		// pass, conditionsEnd is expected to be a NoPos
	}

	// Find the next node in statements
	p.nextNodePos = findNextNodePos(bodyCmtMap, fn.Body, conditionsEnd)

	return
}

// funcUpdate defines how a function should be updated.
type funcUpdate struct {
	pres      []condition
	posts     []condition
	fn        *ast.FuncDecl
	positions condPositions
}

func (c condition) ViolationMsg() string {
	msg := "Violated: "
	if len(c.label) > 0 {
		msg += c.label + ": "
	}
	msg += c.condStr

	return strconv.Quote(msg)
}

func negateNot(condStr string, unary *ast.UnaryExpr) string {
	if unary.Op != token.NOT {
		panic(fmt.Sprintf("Expected NOT operator in the unary expression, but got: %d", unary.Op))
	}

	parenExpr, isParenExpr := unary.X.(*ast.ParenExpr)

	// Use .Pos() directly since the unary expression was parsed from condStr

	if !isParenExpr {
		return strings.Trim(condStr[unary.X.Pos()-1:unary.X.End()-1], " \t")
	}

	return strings.Trim(condStr[parenExpr.X.Pos()-1:parenExpr.X.End()-1], " \t")
}

// NotCondStr negates the condition and returns it as a string representation of a Go boolean expression.
func (c condition) NotCondStr() string {
	if c.condStr == "" {
		return ""
	}

	switch v := c.cond.(type) {
	case *ast.UnaryExpr:
		if v.Op == token.NOT {
			return negateNot(c.condStr, v)
		}

		return fmt.Sprintf("!(%s)", strings.Trim(c.condStr, " \t"))
	case *ast.ParenExpr:
		return fmt.Sprintf("!%s", strings.Trim(c.condStr, " \t"))
	}

	return fmt.Sprintf("!(%s)", strings.Trim(c.condStr, " \t"))
}

var tplPre = template.Must(template.New("preconditions").Parse(
	`{{$l := len .Pres }}{{ if eq $l 1 }}{{ $c := index .Pres 0 }}	// Pre-condition
	if {{ $c.NotCondStr }} {
		panic({{ $c.ViolationMsg }})
	}
{{- else }}	// Pre-conditions
	switch { {{- range .Pres }}
	case {{ .NotCondStr }}:
		panic({{ .ViolationMsg }})
{{- end }}
	default:
		// Pass
	}
{{- end }}`))

var tplPost = template.Must(template.New("postconditions").Parse(
	`{{$l := len .Posts }}{{ if eq $l 1 }}{{ $c := index .Posts 0 }}	// Post-condition
	defer func() {
		if {{ $c.NotCondStr }} {
			panic({{ $c.ViolationMsg }})
		}
	}()
{{- else }}	// Post-conditions
	defer func() {
		switch { {{- range .Posts }}
		case {{ .NotCondStr }}:
			panic({{ .ViolationMsg }})
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
func generateCode(pres []condition, posts []condition) (code string, err error) {
	blocks := []string{}

	if len(pres) > 0 {
		data := struct{ Pres []condition }{Pres: pres}

		var buf bytes.Buffer
		err = tplPre.Execute(&buf, data)
		if err != nil {
			return
		}

		blocks = append(blocks, buf.String())
	}

	if len(posts) > 0 {
		data := struct{ Posts []condition }{Posts: posts}

		var buf bytes.Buffer
		err = tplPost.Execute(&buf, data)
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

		fstStmtOffset := fset.Position(up.positions.nextNodePos).Offset

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

	cursor = fset.Position(up.positions.nextNodePos).Offset

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
		code, err = generateCode(up.pres, up.posts)
		if err != nil {
			return
		}

		switch {
		case up.positions.nextNodePos == token.NoPos:
			// The function contains no statements except the conditions so we can simply fill it out.

			cursor = updateEmptyFunc(fset, up, code, writer)

		case fset.Position(up.fn.Body.Lbrace).Line == fset.Position(up.fn.Body.Rbrace).Line:
			// The function contains statements on the same lines as the braces.

			if up.positions.preStart != token.NoPos {
				panic(fmt.Sprintf(
					"Unexpected to have found a precondition block in a single-line function %s at %s:%d",
					up.fn.Name, fset.File(up.positions.preStart).Name(),
					fset.Position(up.positions.preStart).Line))
			}

			if up.positions.postStart != token.NoPos {
				panic(fmt.Sprintf(
					"Unexpected to have found a postcondition block in a single-line function %s at %s:%d",
					up.fn.Name, fset.File(up.positions.postStart).Name(),
					fset.Position(up.positions.postStart).Line))
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

		name := fn.Name.Name
		commentLines := strings.Split(fn.Doc.Text(), "\n")

		var pres []condition
		var posts []condition

		if !remove {
			pres, posts, err = parseContracts(name, commentLines)
			if err != nil {
				err = fmt.Errorf("failed to parse comments of the function %s on line %d: %s",
					name, fset.Position(fn.Doc.Pos()).Line, err)
				return
			}
		} else {
			// Remove is true, hence leave the pre and postconditions empty.
		}

		var positions condPositions
		positions, err = findBlocks(fset, fn, cmtMap)
		if err != nil {
			return
		}

		// Update only if there is something to actually change.
		if len(pres) == 0 && len(posts) == 0 && positions.preStart == token.NoPos &&
			positions.postStart == token.NoPos {
			continue
		}

		updates = append(updates,
			funcUpdate{pres: pres, posts: posts, fn: fn, positions: positions})
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
