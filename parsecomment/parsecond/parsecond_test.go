package parsecond_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Parquery/gocontracts/parsecomment/parsecond"
)

func conditionEquals(
	t *testing.T, expected parsecond.Condition,
	got parsecond.Condition, description string) {

	msgs := []string{}

	if expected.Label != got.Label {
		msgs = append(msgs,
			fmt.Sprintf("expected label %#v, got %#v",
				expected.Label, got.Label))
	}

	if expected.InitStr != got.InitStr {
		msgs = append(msgs,
			fmt.Sprintf("expected initialization %#v, got %#v",
				expected.InitStr, got.InitStr))
	}

	if expected.CondStr != got.CondStr {
		msgs = append(msgs,
			fmt.Sprintf("expected condition string %#v, got %#v",
				expected.CondStr, got.CondStr))
	}

	if len(msgs) > 0 {
		for i := 0; i < len(msgs); i++ {
			msgs[i] = "\t" + msgs[i]
		}
		_, file, line, _ := runtime.Caller(1)

		fmt.Printf("\033[31m%s:%d: %s\n%s\033[39m\n",
			filepath.Base(file), line, description,
			strings.Join(msgs, "\n"))
		t.Fail()
	}
}

func TestParseCondition(t *testing.T) {
	type testCase struct {
		expected    parsecond.Condition
		text        string
		description string
	}

	cases := []testCase{
		{
			expected: parsecond.Condition{
				CondStr: "x < 100",
			},
			text:        " * x < 100",
			description: "only condition string",
		},
		{
			expected: parsecond.Condition{
				Label:   "some label",
				CondStr: "x < 100",
			},
			text:        " * some label: x < 100",
			description: "label, condition string",
		},
		{
			expected: parsecond.Condition{
				CondStr: "DoSomethingWithCopy(someArr[:]) == 0",
			},
			text:        " * DoSomethingWithCopy(someArr[:]) == 0",
			description: "colon in the expression",
		},
		{
			expected: parsecond.Condition{
				InitStr: "_, ok := someMap[3]",
				CondStr: "ok",
			},
			text:        " * _, ok := someMap[3]; ok",
			description: "short statement, condition string",
		},
		{
			expected: parsecond.Condition{
				Label:   "some label",
				InitStr: "_, ok := someMap[3]",
				CondStr: "ok",
			},
			text:        " * some label: _, ok := someMap[3]; ok",
			description: "label, short statement, condition string",
		},
		{
			expected: parsecond.Condition{
				Label:   "some label",
				InitStr: "_, ok := someMap[3]",
				CondStr: "ok",
			},
			text:        " \t *  \t some label: \t   _, ok := someMap[3];  \t ok  \t ",
			description: "extra whitespace ignored",
		},
		{
			expected: parsecond.Condition{
				CondStr: "x < 100",
			},
			text:        "* x < 100",
			description: "whitespace-prefix agnostic",
		},
	}

	for _, cs := range cases {
		got, err := parsecond.ToCondition(cs.text)
		if err != nil {
			t.Fatalf("Failed to parse condition: %s\n\tdue to: %s",
				cs.description, err.Error())
		}

		if got == nil {
			t.Errorf("Failed to match: %s\n\tfrom condition text: %#v",
				cs.description, cs.text)
		} else {
			conditionEquals(t, cs.expected, *got, cs.description)
		}
	}
}

func TestParseCondition_NoMatch(t *testing.T) {
	cond, err := parsecond.ToCondition("No condition could be matched.")
	if err != nil {
		t.Fatal(err.Error())
	}

	if cond != nil {
		t.Errorf("Expected to match no condition, but matched: %#v",
			*cond)
	}
}

func TestParseCondition_Unparsable(t *testing.T) {
	_, err := parsecond.ToCondition("* x ==")
	expected := `failed to parse the condition in the following playground:
package main

func main() {
	if x == {
		// Do something
	}
}

The error was: 4:10: expected operand, found '{' (and 7 more errors)`

	switch {
	case err == nil:
		t.Fatalf("Expected an error, but got nil")
	case err.Error() != expected:
		t.Fatalf("Expected an error %#v,\n\tbut got %#v",
			expected, err.Error())
	default:
		// Pass
	}
}

// TODO(marko): add more unit tests to cover up; other than that: ready to publish!tog
