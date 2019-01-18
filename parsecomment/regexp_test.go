package parsecomment

import (
	"reflect"
	"testing"
)

func Test_bulletRe(t *testing.T) {
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
