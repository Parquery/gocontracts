package gocontracts

import (
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/Parquery/gocontracts/gocontracts/testcases"
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
	testcases.FailureStatementInbetween,
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
	tmp1.WriteString(expected)
	tmp1.Close()

	tmp2, _ := ioutil.TempFile("", "got")
	defer func() {
		err = os.Remove(tmp2.Name())
	}()
	tmp2.WriteString(got)
	tmp2.Close()

	err = exec.Command("meld", tmp1.Name(), tmp2.Name()).Run()
	return
}

func TestProcess(t *testing.T) {
	for _, cs := range cases {
		updated, err := Process(cs.Text)
		if err != nil {
			t.Fatalf("Failed at case %s: %s", cs.ID, err.Error())
		}

		if cs.Expected != updated {
			lastCommon := "N/A"
			lastCommonI := -1
			i := 0
			for i < len(cs.Expected) && i < len(updated) {
				if cs.Expected[i] == updated[i] {
					lastCommonI = i
					lastCommon = string(cs.Expected[i])
					i++
				} else {
					break
				}

			}
			meld(cs.Expected, updated)
			t.Fatalf("Failed at case %s: "+
				"expected (len: %d):\n%s, got (len: %d):\n%s\nLast common character at %d: %s",
				cs.ID, len(cs.Expected), cs.Expected, len(updated), updated, lastCommonI, lastCommon)
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
			t.Fatalf("Expected a failure error %#v, but got %#v", failure.Error, err.Error())
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
