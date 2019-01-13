package gocontracts

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"fmt"
	"github.com/Parquery/gocontracts/gocontracts/testcases"
	"github.com/Parquery/gocontracts/parsecomment"
	"go/parser"
	"math/rand"
	"path/filepath"
)

var cases = []testcases.Case{
	testcases.UnchangedIfNoContracts,
	testcases.NoPreviousConditions,
	testcases.HasPrecondition,
	testcases.HasPreamble,
	testcases.HasPreambleWithoutStatements,
	testcases.HasPostcondition,
	testcases.CurlyBracketsOnSameLine,
	testcases.OneLineFunction,
	testcases.HasOnlyComment,
	testcases.SemicolonAfterPostcondition,
	testcases.CommentAfterPostcondition,
	testcases.MultipleFunctions,
	testcases.DoubleNegation,
	testcases.DoubleNegations,
	testcases.NoConditions,
	testcases.NoFunction,
	testcases.TypeDeclaration,
	testcases.ConditionsRemovedInCommentOfEmptyFunction,
	testcases.ConditionsRemovedInComment,
	testcases.RemoveInCode,
	testcases.RemoveInCodeOfEmptyFunction,
	testcases.RemoveInCodeWithSemicolon,
	testcases.FromReadme}

var failures = []testcases.Failure{
	testcases.FailureCommentParse,
	testcases.FailureBodyParse,
	testcases.FailureUnparsableFile}

// meld runs a meld to compare the expected against the got.
func meld(expected string, got string) (err error) {
	tmp1, _ := ioutil.TempFile("", "expected")
	defer func() {
		err = os.Remove(tmp1.Name())
	}()
	_, err = tmp1.WriteString(expected)
	if err != nil {
		return
	}

	err = tmp1.Close()
	if err != nil {
		return
	}

	tmp2, _ := ioutil.TempFile("", "got")
	defer func() {
		err = os.Remove(tmp2.Name())
	}()
	_, err = tmp2.WriteString(got)
	if err != nil {
		return
	}

	err = tmp2.Close()
	if err != nil {
		return
	}

	err = exec.Command("meld", tmp1.Name(), tmp2.Name()).Run()
	return
}

// lastCommonChar searches for the last character of the common prefix in the 'expected' and the 'got'.
// If they don't have a common prefix, found is false.
func lastCommon(expected string, got string) (i int, found bool) {
	found = false

	i = 0
	for i < len(expected) && i < len(got) {
		if expected[i] == got[i] {
			found = true
			i++
		} else {
			break
		}
	}

	return
}

func TestProcess(t *testing.T) {
	for _, cs := range cases {
		updated, err := Process(cs.Text, cs.ID, cs.Remove)

		switch {
		case err != nil:
			t.Errorf("Failed at case %s: %s", cs.ID, err.Error())
		case cs.Expected != updated:
			i, found := lastCommon(cs.Expected, updated)

			lastChar := "N/A"
			if found {
				lastChar = string(cs.Expected[i])
			}

			t.Errorf("Failed at case %s: "+
				"expected (len: %d):\n"+
				"%s, got (len: %d):\n"+
				"%s\n"+
				"Last common character at %d: %s",
				cs.ID, len(cs.Expected), cs.Expected, len(updated), updated, i, lastChar)
		default:
			// pass
		}
	}
}

func TestProcessFailures(t *testing.T) {
	for _, failure := range failures {
		_, err := Process(failure.Text, failure.ID, false)

		switch {
		case err == nil:
			t.Errorf("Expected an error in the failure case %s, but got nil", failure.ID)
		case failure.Error != err.Error():
			t.Errorf("Expected a failure error %#v in the failure case %#v, "+
				"but got %#v", failure.Error, failure.ID, err.Error())
		default:
			// pass
		}
	}
}

func TestNotCondStr(t *testing.T) {
	type testCase struct {
		condStr  string
		expected string
	}

	testCases := []testCase{
		{condStr: "!x", expected: "x"},
		{condStr: "!  x", expected: "x"},
		{condStr: "!(x)", expected: "x"},
		{condStr: "! ( x )", expected: "x"},
		{condStr: "x", expected: "!(x)"},

		// go/parser package can parse this even though it is invalid Go code.
		// We decided to handle this case gracefully in order not to complicate
		// the template code.
		{condStr: "-x", expected: "!(-x)"},

		{condStr: "(x)", expected: "!(x)"},
		{condStr: "!x || y != 3", expected: "!(!x || y != 3)"}}

	for _, tc := range testCases {
		expr, err := parser.ParseExpr(tc.condStr)
		if err != nil {
			t.Fatalf("Failed to parse the condition string %#v: %s", tc.condStr, err)
		}

		c := parsecomment.Condition{Cond: expr, CondStr: tc.condStr}
		got := notCondStr(c)
		if got != tc.expected {
			t.Fatalf("Expected NotCondStr %#v from condStr %#v, got: %#v", tc.expected, tc.condStr, got)
		}
	}
}

func TestProcessInPlace(t *testing.T) {
	tmp, err := ioutil.TempFile("", "gocontracts-process_test-")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		err = os.Remove(tmp.Name())
		if err != nil {
			t.Fatal(err.Error())
		}
	}()

	// Pick an arbitrary test case
	cs := testcases.DoubleNegation

	pth := tmp.Name()
	err = ioutil.WriteFile(pth, []byte(cs.Text), 0600)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = ProcessInPlace(pth, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	var data []byte
	data, err = ioutil.ReadFile(pth)
	if err != nil {
		t.Fatal(err.Error())
	}

	updated := string(data)
	if updated != cs.Expected {
		i, found := lastCommon(cs.Expected, updated)

		lastChar := "N/A"
		if found {
			lastChar = string(cs.Expected[i])
		}

		t.Fatalf("Failed to process in-place the case %s: "+
			"expected (len: %d):\n%s, updated (len: %d):\n%s\nLast common character at %d: %s",
			cs.ID, len(cs.Expected), cs.Expected, len(updated), updated, i, lastChar)
	}
}

func TestProcessInPlace_Failure(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "process_test-")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		err = os.RemoveAll(tmpdir)
		if err != nil {
			t.Fatal(err.Error())
		}
	}()

	// Pick an arbitrary failure case
	failure := testcases.FailureCommentParse

	pth := filepath.Join(tmpdir, "some_file.go")
	err = ioutil.WriteFile(pth, []byte(failure.Text), 0600)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = ProcessInPlace(pth, false)
	if err == nil {
		t.Fatalf("Expected an error when processing the failure case %s in-place, but got nil", failure.ID)
	}

	if err == nil {
		t.Fatalf("Expected an error when processing the failure case %s in-place, but got nil", failure.ID)
	}

	if failure.Error != err.Error() {
		t.Fatalf("Expected a failure error %#v in the failure case %s, "+
			"but got %#v", failure.Error, failure.ID, err.Error())
	}
}

func TestProcessFile_NonExisting(t *testing.T) {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	randomData := make([]byte, 10)
	for i := range randomData {
		randomData[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	randomStr := string(randomData)

	pth := fmt.Sprintf("/some/nonexisting/path-%s", randomStr)
	_, err := ProcessFile(pth, false)
	if err == nil {
		t.Fatal("Expected an error when processing a non-existing file, but got none")
	}

	expected := fmt.Sprintf("failed to read: open %s: no such file or directory", pth)
	if expected != err.Error() {
		t.Fatalf("Expected an error %#v, but got %#v", expected, err.Error())
	}
}
