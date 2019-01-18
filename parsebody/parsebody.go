// Package parsebody parses the contract sections from the function body.
package parsebody

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"
)

// section defines a section of the tokens belonging to a contract block (e.g., pre-conditions).
type section struct {
	// Start of a section.
	// Start set to token.NoPos means that the section does not appear in the function.
	start token.Pos

	// End of the section, inclusive
	end token.Pos
}

// parsePreconditions parses the pre-conditions defined in the function body.
func parsePreconditions(
	fset *token.FileSet, fn *ast.FuncDecl, cmtGrp *ast.CommentGroup) (s section, err error) {

	s.start = cmtGrp.Pos()

	cmtText := strings.Trim(cmtGrp.Text(), "\n \t")

	// Check that there is a block following the comment
	var stmtAfterCmt ast.Stmt
	for _, stmt := range fn.Body.List {
		if stmt.Pos() > s.start {
			stmtAfterCmt = stmt
			break
		}
	}

	if stmtAfterCmt == nil {
		err = fmt.Errorf("found no statement after the comment %v in function %s on line %d",
			cmtText, fn.Name.String(), fset.Position(s.start).Line)
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

	s.end = stmtAfterCmt.End()

	return
}

// parsePostconditions parses the post-conditions defined in the function body.
func parsePostconditions(
	fset *token.FileSet, fn *ast.FuncDecl, cmtGrp *ast.CommentGroup) (s section, err error) {
	s.start = cmtGrp.Pos()

	cmtText := strings.Trim(cmtGrp.Text(), "\n \t")

	// Check that there is a defer following the comment
	var stmtAfterCmt ast.Stmt
	for _, stmt := range fn.Body.List {
		if stmt.Pos() > s.start {
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

	s.end = deferStmt.End()
	return
}

// validatePreambleSection validates that the preamble markers are well-positioned.
func validatePreambleSection(fset *token.FileSet, fn *ast.FuncDecl, preamble section) (err error) {
	if preamble.start == token.NoPos && preamble.end == token.NoPos {
		return
	}

	if preamble.start != token.NoPos && preamble.end == token.NoPos {
		err = fmt.Errorf("preamble start marker without the end marker in function %s on line %d",
			fn.Name.String(), fset.Position(preamble.start).Line)
		return
	}

	if preamble.start == token.NoPos && preamble.end != token.NoPos {
		err = fmt.Errorf("preamble end marker without the start marker in function %s on line %d",
			fn.Name.String(), fset.Position(preamble.end).Line)
		return
	}

	if preamble.start > preamble.end {
		err = fmt.Errorf("preamble start marker after the end marker in function %s on line %d",
			fn.Name.String(), fset.Position(preamble.start).Line)
		return
	}

	return
}

type parsedPositions struct {
	// Pre-conditions
	pre section

	preamble section

	// Post-conditions
	post section
}

var preconditionRe = regexp.MustCompile(`^(Precondition|Pre-condition)s?\s*:?\s*$`)
var preambleStartsRe = regexp.MustCompile(`^Preamble\s+starts.?\s*$`)
var preambleEndsRe = regexp.MustCompile(`^Preamble\s+ends.?\s*$`)
var postconditionRe = regexp.MustCompile(`^(Postcondition|Post-condition)s?\s*:?\s*$`)

// parseContract parses the contract blocks from the function body.
// bodyCmtMap is expected to contain only the comments written in the function body.
func parseContract(
	fset *token.FileSet, fn *ast.FuncDecl,
	bodyCmtMap ast.CommentMap) (p parsedPositions, err error) {

	for _, cmtGrp := range bodyCmtMap.Comments() {
		cmtText := strings.Trim(cmtGrp.Text(), "\n \t")

		switch {
		case preconditionRe.MatchString(cmtText):
			if p.pre.start != token.NoPos {
				err = fmt.Errorf("duplicate pre-condition block found in function %s on line %d",
					fn.Name.String(), fset.Position(cmtGrp.Pos()).Line)
				return
			}

			p.pre, err = parsePreconditions(fset, fn, cmtGrp)
			if err != nil {
				return
			}

		// This is an anomalous case when the statements
		// between the preamble start and end marker are erased, so that the
		// preamble markers are fused into a single comment.
		case cmtText == "Preamble starts.\nPreamble ends.":
			p.preamble.start = cmtGrp.Pos()
			p.preamble.end = cmtGrp.Pos()

		// This is an anomalous case when the statements
		// between the inverted preamble end and start marker are erased,
		// so that the inverted preamble markers are fused into a single comment.
		case cmtText == "Preamble ends.\nPreamble starts.":
			err = fmt.Errorf(
				"preamble start marker after the end marker in function %s on line %d",
				fn.Name.String(), fset.Position(cmtGrp.Pos()).Line)
			return

		case preambleStartsRe.MatchString(cmtText):
			if p.preamble.start != token.NoPos {
				err = fmt.Errorf("duplicate preamble start found in function %s on line %d",
					fn.Name.String(), fset.Position(cmtGrp.Pos()).Line)
				return
			}

			p.preamble.start = cmtGrp.Pos()

		case preambleEndsRe.MatchString(cmtText):
			if p.preamble.end != token.NoPos {
				err = fmt.Errorf("duplicate preamble end found in function %s on line %d",
					fn.Name.String(), fset.Position(cmtGrp.Pos()).Line)
				return
			}

			p.preamble.end = cmtGrp.Pos()

		case postconditionRe.MatchString(cmtText):
			if p.post.start != token.NoPos {
				err = fmt.Errorf("duplicate post-condition block found in function %s on line %d",
					fn.Name.String(), fset.Position(cmtGrp.Pos()).Line)
				return
			}

			p.post, err = parsePostconditions(fset, fn, cmtGrp)
			if err != nil {
				return
			}

		default:
			// pass
		}
	}

	err = validatePreambleSection(fset, fn, p.preamble)
	if err != nil {
		return
	}

	return
}

func findNextNodePos(bodyCmtMap ast.CommentMap, body *ast.BlockStmt, contractEnd token.Pos) (nextNodePos token.Pos) {
	nextNodePos = token.NoPos

	for _, stmt := range body.List {
		if stmt.Pos() > contractEnd {
			nextNodePos = stmt.Pos()
			break
		}
	}

	// See if there is a comment before the contract blocks and the first next statement
	for _, cmtGrp := range bodyCmtMap.Comments() {
		if cmtGrp.Pos() > contractEnd &&
			(nextNodePos == token.NoPos || cmtGrp.Pos() < nextNodePos) {

			nextNodePos = cmtGrp.Pos()
			break
		}
	}

	return
}

func validateNoBlockOverlap(fset *token.FileSet, fn *ast.FuncDecl, p parsedPositions) (err error) {
	sections := make([]section, 0, 3)
	if p.pre.start != token.NoPos {
		sections = append(sections, p.pre)
	}
	if p.preamble.start != token.NoPos {
		sections = append(sections, p.preamble)
	}
	if p.post.start != token.NoPos {
		sections = append(sections, p.post)
	}

	if len(sections) > 1 {
		// Quadratic time complexity is fine as long as there are few sections.
		for i, section := range sections {
			for j, other := range sections {
				if i != j && section.start <= other.end && other.start <= section.end {
					err = fmt.Errorf(
						"unexpected overlap in contract blocks in function %s "+
							"starting on lines %d and %d, respectively",
						fn.Name.String(), fset.Position(section.start).Line, fset.Position(other.start).Line)
					return
				}
			}
		}
	}

	return
}

func (p parsedPositions) asSection() (s section) {
	if p.pre.start != token.NoPos {
		s.start = p.pre.start
		s.end = p.pre.end
	}

	if p.preamble.start != token.NoPos {
		if s.start == token.NoPos || s.start > p.preamble.start {
			s.start = p.preamble.start
		}

		if s.end == token.NoPos || s.end < p.preamble.end {
			s.end = p.preamble.end
		}
	}

	if p.post.start != token.NoPos {
		if s.start == token.NoPos || s.start > p.post.start {
			s.start = p.post.start
		}

		if s.end == token.NoPos || s.end < p.post.end {
			s.end = p.post.end
		}
	}

	return
}

// validateNoStmts validates that there are no statements before the contract and between the contract blocks,
// respectively.
func validateNoStmts(fset *token.FileSet, fn *ast.FuncDecl, p parsedPositions) (err error) {
	s := p.asSection()

	// Check that there are no statements before the contract start in the function body
	if s.start != token.NoPos && len(fn.Body.List) > 0 && fn.Body.List[0].Pos() < s.start {
		err = fmt.Errorf(
			"unexpected statement before the contract in function %s on line %d",
			fn.Name.String(), fset.Position(fn.Body.List[0].Pos()).Line)
		return
	}

	// Check that there are no statements between the blocks
	sections := make([]section, 0, 3)
	if p.pre.start != token.NoPos {
		sections = append(sections, p.pre)
	}
	if p.preamble.start != token.NoPos {
		sections = append(sections, p.preamble)
	}
	if p.post.start != token.NoPos {
		sections = append(sections, p.post)
	}

	if len(sections) > 1 {
		for i := 1; i < len(sections); i++ {
			prevS := sections[i-1]
			curS := sections[i]

			for _, stmt := range fn.Body.List {
				if stmt.Pos() >= prevS.end && stmt.Pos() < curS.start {
					err = fmt.Errorf(
						"unexpected statement between the contract blocks in function %s on line %d",
						fn.Name.String(), fset.Position(stmt.Pos()).Line)
					return
				}
			}

		}
	}

	return
}

// Contract indicates the token section corresponding to the contract blocks.
type Contract struct {
	// Start indicates the first node of the contract.
	// If Start == token.NoPos, there is no contract in the function body.
	Start token.Pos
	End   token.Pos

	// NextNodePos indicates the position of the first AST node just after the contract.
	// If there are no nodes in the function body after the contract, NextNodePos is token.NoPos.
	NextNodePos token.Pos
}

// ToContract searches for the start and end of the contract in the function body.
func ToContract(
	fset *token.FileSet, fn *ast.FuncDecl, bodyCmtMap ast.CommentMap) (c Contract, err error) {

	////
	// Parse
	////

	p, err := parseContract(fset, fn, bodyCmtMap)
	if err != nil {
		return
	}

	////
	// Validate the parsed blocks
	////

	err = validateNoBlockOverlap(fset, fn, p)
	if err != nil {
		return
	}

	err = validateNoStmts(fset, fn, p)
	if err != nil {
		return
	}

	////
	// Find the start and end of the contract
	////

	s := p.asSection()
	c.Start = s.start
	c.End = s.end

	////
	// Find the next node in the statements
	////

	c.NextNodePos = findNextNodePos(bodyCmtMap, fn.Body, c.End)

	return
}
