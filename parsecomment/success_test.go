package parsecomment_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Parquery/gocontracts/parsecomment"
)

type expectedCondition struct {
	condStr string
	label   string
}

type expectedContract struct {
	pres     []expectedCondition
	posts    []expectedCondition
	preamble string
}

func checkContract(t *testing.T, exp expectedContract, got parsecomment.Contract) {
	msgs := []string{}

	if len(exp.pres) != len(got.Pres) {
		msgs = append(msgs,
			fmt.Sprintf("expected %d pre-condition(s), got %d", len(exp.pres), len(got.Pres)))
	} else {
		for i := range exp.pres {
			if exp.pres[i].condStr != got.Pres[i].CondStr {
				msgs = append(msgs,
					fmt.Sprintf("expected pre-condition %d to be parsed as %#v, got %#v",
						i+1, exp.pres[i].condStr, got.Pres[i].CondStr))
			}

			if exp.pres[i].label != got.Pres[i].Label {
				msgs = append(msgs,
					fmt.Sprintf("expected the label of the pre-condition %d to be parsed as %#v, got %#v",
						i+1, exp.pres[i].label, got.Pres[i].Label))
			}
		}
	}

	if len(exp.posts) != len(got.Posts) {
		msgs = append(msgs,
			fmt.Sprintf("expected %d post-condition(s), got %d", len(exp.posts), len(got.Posts)))
	} else {
		for i := range exp.posts {
			if exp.posts[i].condStr != got.Posts[i].CondStr {
				msgs = append(msgs,
					fmt.Sprintf("expected post-condition %d to be parsed as %#v, got %#v",
						i+1, exp.posts[i].condStr, got.Posts[i].CondStr))
			}

			if exp.posts[i].label != got.Posts[i].Label {
				msgs = append(msgs,
					fmt.Sprintf("expected the label of the post-condition %d to be parsed as %#v, got %#v",
						i+1, exp.posts[i].label, got.Posts[i].Label))
			}
		}
	}

	if exp.preamble != got.Preamble {
		msgs = append(msgs,
			fmt.Sprintf("expected a preamble %#v, got %#v", exp.preamble, got.Preamble))
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

func TestToContract_EmptyBlocks(t *testing.T) {
	lines := strings.Split(
		`SomeFunc does something.

SomeFunc requires:

SomeFunc ensures:

SomeFunc preamble:

Some text here.`, "\n")

	got, err := parsecomment.ToContract("SomeFunc", lines)
	if err != nil {
		t.Fatal(err.Error())
	}

	exp := expectedContract{}

	checkContract(t, exp, got)
}

func TestToContract_Preamble(t *testing.T) {
	lines := strings.Split(
		`SomeFunc does something.

SomeFunc preamble:
	x := 1984
	if x > 1 {
		panic("uh")
	}`, "\n")

	got, err := parsecomment.ToContract("SomeFunc", lines)
	if err != nil {
		t.Fatal(err.Error())
	}

	exp := expectedContract{
		preamble: "x := 1984\n" +
			"if x > 1 {\n" +
			"\tpanic(\"uh\")" +
			"\n}",
	}

	checkContract(t, exp, got)
}

func TestToContract_EmptyLinesAfterBlocks(t *testing.T) {
	lines := strings.Split(
		`SomeFunc does something.

SomeFunc requires:
 * x > 0
 * x < 100
 * some condition: y > 3

SomeFunc preamble:
	print("hello world")

SomeFunc ensures:
 * strings.HasPrefix(result, "hello")

Some text here.`, "\n")

	got, err := parsecomment.ToContract("SomeFunc", lines)
	if err != nil {
		t.Fatal(err.Error())
	}

	exp := expectedContract{
		pres: []expectedCondition{
			{condStr: "x > 0"},
			{condStr: "x < 100"},
			{condStr: "y > 3", label: "some condition"},
		},
		preamble: `print("hello world")`,
		posts: []expectedCondition{
			{condStr: "strings.HasPrefix(result, \"hello\")"},
		},
	}

	checkContract(t, exp, got)
}

func TestToContract_TextAfterBlocks(t *testing.T) {
	lines := strings.Split(
		`SomeFunc does something.

SomeFunc requires:
 * x > 0
bla bla bla
SomeFunc preamble:
	print("hello world")
bla bla bla
SomeFunc ensures:
 * strings.HasPrefix(result, "hello")
bla bla bla`, "\n")

	got, err := parsecomment.ToContract("SomeFunc", lines)
	if err != nil {
		t.Fatal(err.Error())
	}

	exp := expectedContract{
		pres: []expectedCondition{
			{condStr: "x > 0"},
		},
		preamble: `print("hello world")`,
		posts: []expectedCondition{
			{condStr: "strings.HasPrefix(result, \"hello\")"},
		},
	}

	checkContract(t, exp, got)
}
