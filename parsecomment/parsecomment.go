// Package parsecomment parses the contract from the function comment.
package parsecomment

import (
	"fmt"
	"go/ast"
	"go/parser"
	"regexp"
	"strings"

	"github.com/Parquery/gocontracts/dedent"
)

var requiresRe = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z_0-9]*)\s+requires\s*:\s*$`)
var ensuresRe = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z_0-9]*)\s+ensures\s*:\s*$`)
var bulletRe = regexp.MustCompile(`\s*\*\s*(([^:]+)\s*:\s*)?(\s*.*\s*)$`)
var preambleRe = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z_0-9]*)('s)?\s+preamble\s*:\s*$`)

// Line tokens are obtained by tokenizing each line of the the function description as a whole.

type lineToken interface {
	text() string
}

type requiresToken struct {
	aText string
	name  string
}

func (r *requiresToken) text() string {
	return r.aText
}

type ensuresToken struct {
	aText string
	name  string
}

func (e *ensuresToken) text() string {
	return e.aText
}

type preambleToken struct {
	aText string
	name  string
}

func (p *preambleToken) text() string {
	return p.aText
}

type textToken struct {
	aText string
}

func (u *textToken) text() string {
	return u.aText
}

func tokenizeComment(commentLines []string) (tokens []lineToken) {
	tokens = make([]lineToken, 0, len(commentLines))

	for _, line := range commentLines {
		mtchs := requiresRe.FindStringSubmatch(line)
		if len(mtchs) > 0 {
			tokens = append(tokens, &requiresToken{aText: line, name: mtchs[1]})
			continue
		}

		mtchs = ensuresRe.FindStringSubmatch(line)
		if len(mtchs) > 0 {
			tokens = append(tokens, &ensuresToken{aText: line, name: mtchs[1]})
			continue
		}

		mtchs = preambleRe.FindStringSubmatch(line)
		if len(mtchs) > 0 {
			tokens = append(tokens, &preambleToken{aText: line, name: mtchs[1]})
			continue
		}

		tokens = append(tokens, &textToken{aText: line})
		continue
	}

	return
}

// Condition defines a pre-condition or a post-condition of the function's contract.
type Condition struct {
	Label   string
	CondStr string
	Cond    ast.Expr
}

// parseCondition tries to parse the condition from text.
//
// If no condition could be parsed, cond is nil.
// Err is set if the text matched the bullet format, but there was an error parsing the condition
// as Go code.
func parseCondition(text string) (cond *Condition, err error) {
	mtchs := bulletRe.FindStringSubmatch(text)
	if len(mtchs) > 0 {
		label := mtchs[2]
		exprStr := mtchs[3]

		// Check that the condition is parsable
		var expr ast.Expr
		expr, err = parser.ParseExpr(exprStr)
		if err != nil {
			return
		}

		cond = &Condition{Label: label, CondStr: exprStr, Cond: expr}
	}
	return
}

// Contract bundles the conditions and the preamble of the function's contract.
type Contract struct {
	Pres     []Condition
	Posts    []Condition
	Preamble string
}

// ToContract parses the contract from the function's documentation.
func ToContract(name string, commentLines []string) (c Contract, err error) {
	tokens := tokenizeComment(commentLines)

	requiresCount := 0
	ensuresCount := 0
	preambleCount := 0
	for _, token := range tokens {
		switch token.(type) {
		case *requiresToken:
			requiresCount++
		case *ensuresToken:
			ensuresCount++
		case *preambleToken:
			preambleCount++
		default:
			// pass
		}
	}

	if requiresCount > 1 {
		err = fmt.Errorf("multiple pre-condition blocks")
		return
	}
	if ensuresCount > 1 {
		err = fmt.Errorf("multiple post-condition blocks")
		return
	}
	if preambleCount > 1 {
		err = fmt.Errorf("multiple preambles")
		return
	}

	const (
		stateText     = 0
		stateRequires = 1
		stateEnsures  = 2
		statePreamble = 3
	)

	c.Pres = make([]Condition, 0, 5)
	c.Posts = make([]Condition, 0, 5)

	preambleLines := make([]string, 0, 5)

	state := stateText

	for _, token := range tokens {
		switch t := token.(type) {
		case *requiresToken:
			if name != t.name {
				err = fmt.Errorf("expected function name %#v in pre-condition block, but got %#v",
					name, t.name)
				return
			}

			state = stateRequires
			continue

		case *ensuresToken:
			if name != t.name {
				err = fmt.Errorf("expected function name %#v in post-condition block, but got %#v",
					name, t.name)
				return
			}

			state = stateEnsures
			continue

		case *preambleToken:
			if name != t.name {
				err = fmt.Errorf("expected function name %#v in preamble block, but got %#v",
					name, t.name)
				return
			}

			state = statePreamble
			continue

		case *textToken:
			switch state {
			case stateText:
				// pass

			case stateRequires:
				if len(strings.Trim(token.text(), " \t")) == 0 {
					// Empty line ends a pre-condition block.
					state = stateText
				} else {
					var cond *Condition
					cond, err = parseCondition(token.text())
					if err != nil {
						err = fmt.Errorf("failed to parse a pre-condition: %s", err.Error())
						return
					}
					if cond != nil {
						c.Pres = append(c.Pres, *cond)
					} else {
						// Unmatched condition ends a pre-condition block.
						state = stateText
					}
				}

			case stateEnsures:
				if len(strings.Trim(token.text(), " \t")) == 0 {
					// Empty line ends a post-condition block.
					state = stateText
				} else {
					var cond *Condition
					cond, err = parseCondition(token.text())
					if err != nil {
						err = fmt.Errorf("failed to parse a post-condition: %s", err.Error())
						return
					}
					if cond != nil {
						c.Posts = append(c.Posts, *cond)
					} else {
						// Unmatched condition ends a post-condition block.
						state = stateText
					}
				}

			case statePreamble:
				if len(token.text()) > 0 &&
					token.text()[0] != '\t' &&
					token.text()[0] != ' ' {

					// Un-indented non-empty line ends a preamble block.
					state = stateText
				} else {
					preambleLines = append(preambleLines, token.text())
				}

			default:
				panic(fmt.Sprintf("unhandled state: %d", state))
			}

		default:
			panic(fmt.Sprintf("unhandled token: %#v", token))
		}
	}

	if len(preambleLines) > 0 {
		c.Preamble = strings.Join(
			dedent.TrimEmptyLines(
				dedent.Dedent(preambleLines)),
			"\n")
	}

	return
}
