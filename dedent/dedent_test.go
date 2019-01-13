package dedent_test

import (
	"strings"
	"testing"

	"github.com/Parquery/gocontracts/dedent"
)

func TestTrimEmptyLines(t *testing.T) {
	type TestCase struct {
		name     string
		input    string
		expected string
	}

	cases := []TestCase{
		{
			name:     "no empty lines",
			input:    "no\nempty\nlines",
			expected: "no\nempty\nlines"},
		{
			name:     "only empty lines",
			input:    "\n\n",
			expected: ""},
		{
			name:     "empty lines in-between",
			input:    "in\n\nbetween",
			expected: "in\n\nbetween"},
		{
			name:     "empty lines before",
			input:    "\n\nbefore\nbut not after",
			expected: "before\nbut not after"},
		{
			name:     "empty lines after",
			input:    "after\nbut not before\n\n",
			expected: "after\nbut not before"},
		{
			name:     "empty lines before and after",
			input:    "\nempty lines\nbefore and after\n\n",
			expected: "empty lines\nbefore and after"},
	}

	for _, tc := range cases {
		lines := strings.Split(tc.input, "\n")
		result := dedent.TrimEmptyLines(lines)
		got := strings.Join(result, "\n")

		if got != tc.expected {
			t.Fatalf("test case %#v failed: expected %#v, got %#v", tc.name, tc.expected, got)
		}
	}
}

func TestDedent(t *testing.T) {
	type TestCase struct {
		name     string
		input    string
		expected string
	}

	cases := []TestCase{
		{
			name:     "no dedention possible",
			input:    "test me\n some\n\tmore",
			expected: "test me\n some\n\tmore"},
		{
			name:     "space dedention",
			input:    " test me\n some\n more",
			expected: "test me\nsome\nmore"},
		{
			name:     "dedention of a whitespace line",
			input:    " test me\n \n more",
			expected: "test me\n\nmore"},
		{
			name:     "tab dedention",
			input:    "\ttest me\n\tsome\n\tmore",
			expected: "test me\nsome\nmore"},
		{
			name:     "dedention of a tab line",
			input:    "\ttest me\n\t\n\tmore",
			expected: "test me\n\nmore"},
		{
			name:     "empty line ignore in the dedention",
			input:    "\ttest me\n\n\tmore",
			expected: "test me\n\nmore"},
		{
			name:     "no lines",
			input:    "",
			expected: ""},
		{
			name:     "single line",
			input:    "  testme",
			expected: "testme"},
		{
			name:     "all empty lines",
			input:    "\n\n\n",
			expected: "\n\n\n"},
	}

	for _, tc := range cases {
		lines := strings.Split(tc.input, "\n")
		result := dedent.Dedent(lines)
		got := strings.Join(result, "\n")

		if got != tc.expected {
			t.Fatalf("test case %#v failed: expected %#v, got %#v", tc.name, tc.expected, got)
		}
	}
}
