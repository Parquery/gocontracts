package gocontracts

import (
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"fmt"
	"github.com/Parquery/gocontracts/gocontracts/testcases"
	"math/rand"
	"path/filepath"
)

var cases = []testcases.Case{
	testcases.NoPreviousConditions,
	testcases.HasPrecondition,
	testcases.HasPostcondition,
	testcases.CurlyBracketsOnSameLine,
	testcases.HasOnlyComment,
	testcases.SemicolonAfterPostcondition,
	testcases.CommentAfterPostcondition,
	testcases.MultipleFunctions,
	testcases.DoubleNegation,
	testcases.DoubleNegations,
	testcases.NoConditions,
	testcases.NoFunction,
	testcases.TypeDeclaration}

var failures = []testcases.Failure{
	testcases.FailureStatementInBetween,
	testcases.FailureStatementBefore,
	testcases.FailureNoSwitchInPrecondition,
	testcases.FailureNoIfInPrecondition,
	testcases.FailureNoDeferInPostcondition,
	testcases.FailureNoStatementAfterPrecondtion,
	testcases.FailureNoStatementAfterPostcondtion,
	testcases.FailureUnmatchedFunctionInPrecondition,
	testcases.FailureUnmatchedFunctionInPostcondition}

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
		updated, err := Process(cs.Text)
		if err != nil {
			t.Fatalf("Failed at case %s: %s", cs.ID, err.Error())
		}

		if cs.Expected != updated {
			i, found := lastCommon(cs.Expected, updated)

			lastChar := "N/A"
			if found {
				lastChar = string(cs.Expected[i])
			}

			t.Fatalf("Failed at case %s: "+
				"expected (len: %d):\n%s, got (len: %d):\n%s\nLast common character at %d: %s",
				cs.ID, len(cs.Expected), cs.Expected, len(updated), updated, i, lastChar)
		}
	}
}

func TestProcessFailures(t *testing.T) {
	for _, failure := range failures {
		_, err := Process(failure.Text)

		if err == nil {
			t.Fatalf("Expected an error in the failure case %s, but got nil", failure.ID)
		}

		if failure.Error != err.Error() {
			t.Fatalf("Expected a failure error %#v in the failure case %s, "+
				"but got %#v", failure.Error, failure.ID, err.Error())
		}
	}
}

func Test_kBulletRe(t *testing.T) {
	type bulletCase struct {
		line            string
		expectedMatches []string
	}

	bulletCases := []bulletCase{
		{line: " * x < 100",
			expectedMatches: []string{" * x < 100", "", "", "x < 100"}},

		{line: " * some label: x < 100",
			expectedMatches: []string{" * some label: x < 100", "some label: ", "some label", "x < 100"}},

		{line: " No-bullet text.", expectedMatches: nil}}

	for _, cs := range bulletCases {
		mtchs := bulletRe.FindStringSubmatch(cs.line)

		if !reflect.DeepEqual(cs.expectedMatches, mtchs) {
			t.Fatalf("Expected bullet regexp matches %#v, got %#v", cs.expectedMatches, mtchs)
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
		{condStr: "(x)", expected: "!(x)"},
		{condStr: "", expected: ""}}

	for _, tc := range testCases {
		got := notCondStr(tc.condStr)
		if got != tc.expected {
			t.Fatalf("Expected notCondStr %#v from condStr %#v, got: %#v", tc.expected, tc.condStr, got)
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

	err = ProcessInPlace(pth)
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
	failure := testcases.FailureUnmatchedFunctionInPostcondition

	pth := filepath.Join(tmpdir, "some_file.go")
	err = ioutil.WriteFile(pth, []byte(failure.Text), 0600)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = ProcessInPlace(pth)
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

	_, err := ProcessFile(fmt.Sprintf("/some/nonexisting/path-%s", randomStr))
	if err == nil {
		t.Fatal("Expected an error when processing a non-existing file, but got none")
	}

	expected := "failed to read: open /some/nonexisting/path-XVlBzgbaiC: no such file or directory"
	if expected != err.Error() {
		t.Fatalf("Expected an error %#v, but got %#v", expected, err.Error())
	}
}
