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
	p.nextNodePos = token.NoPos

	for _, stmt := range fn.Body.List {
		if stmt.Pos() > conditionsEnd {
			p.nextNodePos = stmt.Pos()
			break
		}
	}

	// See if there is a comment before the conditions and the first next statement
	for _, cmtGrp := range bodyCmtMap.Comments() {
		if cmtGrp.Pos() > conditionsEnd &&
			(p.nextNodePos == token.NoPos || cmtGrp.Pos() < p.nextNodePos) {

			p.nextNodePos = cmtGrp.Pos()
			break
		}
	}

	return
}

// funcUpdate defines how a function should be updated.
type funcUpdate struct {
	pres      []condition
	posts     []condition
	fn        *ast.FuncDecl
	positions condPositions
}

func violationMsg(cond condition) string {
	msg := "Violated: "
	if len(cond.label) > 0 {
		msg += cond.label + ": "
	}
	msg += cond.condStr

	return strconv.Quote(msg)
}

var notExprWithParensRe = regexp.MustCompile(`^!\s*\((.+)\)$`)
var exprWithParensRe = regexp.MustCompile(`^\(\s*(.+)\s*\)$`)

func notCondStr(condStr string) string {
	if condStr == "" {
		return ""
	}

	mtch := notExprWithParensRe.FindStringSubmatch(condStr)
	if len(mtch) > 0 {
		return strings.Trim(mtch[1], " \t")
	}

	if condStr[0] == '!' {
		return strings.Trim(condStr[1:], " \t")
	}

	if exprWithParensRe.MatchString(condStr) {
		return fmt.Sprintf("!%s", strings.Trim(condStr, " \t"))
	}
	return fmt.Sprintf("!(%s)", strings.Trim(condStr, " \t"))
}

// generateCode generates the code of the two blocks.
//
// The first line of generated code is indented.
// The generated code does not end with a new-line character.
func generateCode(pres []condition, posts []condition) (code string) {
	blocks := []string{}

	if len(pres) > 0 {
		var buf bytes.Buffer

		if len(pres) == 1 {
			buf.WriteString(fmt.Sprintf(
				"\t// Pre-condition\n"+
					"\tif %s {\n"+
					"\t\tpanic(%s)\n"+
					"\t}", notCondStr(pres[0].condStr), violationMsg(pres[0])))
		} else {
			buf.WriteString("\t// Pre-conditions\n" +
				"\tswitch {\n")

			for _, pre := range pres {
				buf.WriteString(fmt.Sprintf(
					"\tcase %s:\n"+
						"\t\tpanic(%s)\n", notCondStr(pre.condStr), violationMsg(pre)))
			}
			buf.WriteString("\tdefault:\n" +
				"\t\t// Pass\n" +
				"\t}")
		}

		blocks = append(blocks, buf.String())
	}

	if len(posts) > 0 {
		var buf bytes.Buffer

		if len(posts) == 1 {
			buf.WriteString(fmt.Sprintf(
				"\t// Post-condition\n"+
					"\tdefer func() {\n"+
					"\t\tif %s {\n"+
					"\t\t\tpanic(%s)\n"+
					"\t\t}\n"+
					"\t}()", notCondStr(posts[0].condStr), violationMsg(posts[0])))
		} else {
			buf.WriteString(
				"\t// Post-conditions\n" +
					"\tdefer func() {\n" +
					"\t\tswitch {\n")

			for _, post := range posts {
				buf.WriteString(fmt.Sprintf(
					"\t\tcase %s:\n"+
						"\t\t\tpanic(%s)\n", notCondStr(post.condStr), violationMsg(post)))
			}
			buf.WriteString("\t\tdefault:\n" +
				"\t\t\t// Pass\n" +
				"\t\t}\n" +
				"\t}()")
		}

		blocks = append(blocks, buf.String())
	}

	return strings.Join(blocks, "\n\n")
}

func update(text string, updates []funcUpdate, fset *token.FileSet) (updated string) {
	if len(updates) == 0 {
		updated = text
		return
	}

	writer := bytes.NewBufferString("")

	cursor := 0

	for _, up := range updates {
		if len(up.pres) == 0 && len(up.posts) == 0 {
			continue
		}

		if up.positions.nextNodePos == token.NoPos {
			// The function contains no statements except the conditions so we can simply fill it out.

			lbraceOffset := fset.Position(up.fn.Body.Lbrace).Offset
			rbraceOffset := fset.Position(up.fn.Body.Rbrace).Offset

			// Write the prefix
			writer.WriteString(text[cursor : lbraceOffset+1])
			cursor = rbraceOffset // Move cursor to the end of the function

			writer.WriteRune('\n')
			writer.WriteString(generateCode(up.pres, up.posts))
			writer.WriteRune('\n')
		} else {
			// The function contains one or more statements after the contract conditions.

			lbraceOffset := fset.Position(up.fn.Body.Lbrace).Offset
			nodeOffset := fset.Position(up.positions.nextNodePos).Offset

			// Go back in order to include a farthest possible end of the post-condition
			conditionsEnd := nodeOffset
			for text[conditionsEnd] != '\n' && text[conditionsEnd] != ';' {
				conditionsEnd--

				if conditionsEnd == lbraceOffset {
					panic(fmt.Sprintf(
						"conditionsEnd reached the left brace of the function at offset: %d", lbraceOffset))
				}
			}

			writer.WriteString(text[cursor : lbraceOffset+1])
			cursor = conditionsEnd

			writer.WriteRune('\n')
			writer.WriteString(generateCode(up.pres, up.posts))

			switch text[conditionsEnd] {
			case ';':
				// Pass, keep the semi-colon at the same last line of the post-condition if it was already there in the
				// previous code
			case '\n':
				// Add an additional new line to nicely separate the contract conditions from the rest of the code
				writer.WriteString("\n")
			default:
				panic(fmt.Sprintf("Unexpected rune at the end of the contract condition(s) at offset %d: %c",
					conditionsEnd, text[conditionsEnd]))
			}
		}
	}

	// Write the suffix
	writer.WriteString(text[cursor:])

	updated = writer.String()
	return
}

// Process automatically adds (or updates) the blocks for checking the pre- and post-conditions.
func Process(text string) (updated string, err error) {
	fset := token.NewFileSet()

	var node *ast.File
	node, err = parser.ParseFile(fset, "", text, parser.ParseComments)
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
		pres, posts, err = parseContracts(name, commentLines)
		if err != nil {
			err = fmt.Errorf("failed to parse comments of the function %s on line %d: %s",
				name, fset.Position(fn.Doc.Pos()).Line, err)
			return
		}

		var positions condPositions
		positions, err = findBlocks(fset, fn, cmtMap)
		if err != nil {
			return
		}

		updates = append(updates,
			funcUpdate{pres: pres, posts: posts, fn: fn, positions: positions})
	}

	updated = update(text, updates, fset)

	return
}

// ProcessFile loads the Go file and processes it.
func ProcessFile(pth string) (updated string, err error) {
	data, err := ioutil.ReadFile(pth)
	if err != nil {
		err = fmt.Errorf("failed to read: %s", err)
		return
	}

	text := string(data)

	updated, err = Process(text)
	if err != nil {
		return
	}

	return
}

// ProcessInPlace loads the Go file in memory, proesses it and writes atomically back to the file.
func ProcessInPlace(pth string) (err error) {
	var updated string
	updated, err = ProcessFile(pth)
	if err != nil {
		return
	}

	var tmp *os.File
	tmp, err = ioutil.TempFile("", "gocontracts-"+filepath.Base(pth))
	if err != nil {
		return
	}
	defer func() {
		_, err = os.Stat(tmp.Name())
		switch {
		case err == nil:
			err = os.Remove(tmp.Name())
			if err != nil {
				err = fmt.Errorf("failed to remove %s: %s", tmp.Name(), err.Error())
				return
			}
		case os.IsNotExist(err):
			// Pass
			err = nil

		default:
			err = fmt.Errorf("failed to stat %s: %s", tmp.Name(), err.Error())
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
