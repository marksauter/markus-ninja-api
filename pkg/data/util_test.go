package data_test

import (
	"testing"

	"github.com/marksauter/markus-ninja-api/pkg/data"
)

var toTsQueryTests = []struct {
	s        string
	expected string
}{
	{
		"foo",
		"foo:*",
	},
	{
		"foo bar",
		"foo:* & bar:*",
	},
	{
		"foo-bar_baz qux",
		"foo:* & bar:* & baz:* & qux:*",
	},
	{
		"fooBar baz-qux",
		"fooBar:* & baz:* & qux:*",
	},
}

func TestToTsQuery(t *testing.T) {
	for _, tt := range toTsQueryTests {
		actual := data.ToTsQuery(tt.s)
		if actual != tt.expected {
			t.Errorf(
				"ToTsQuery(%s) expected %s actual %s",
				tt.s,
				tt.expected,
				actual,
			)
		}
	}
}

var ToPatternQueryTests = []struct {
	s        string
	expected []string
}{
	{
		"foo",
		[]string{"foo%"},
	},
	{
		"foo bar",
		[]string{"foo%", "bar%"},
	},
	{
		"foo-bar_baz qux",
		[]string{"foo%", "bar%", "baz%", "qux%"},
	},
	{
		"fooBar baz-qux",
		[]string{"fooBar%", "baz%", "qux%"},
	},
}

func TestToPatternQuery(t *testing.T) {
	for _, tt := range ToPatternQueryTests {
		actual := data.ToPatternQuery(tt.s)
		if len(actual) != len(tt.expected) {
			t.Errorf(
				"ToPatternQuery(%s) expected %v actual %v",
				tt.s,
				tt.expected,
				actual,
			)
		} else {
			for i, expected := range tt.expected {
				if actual[i] != expected {
					t.Errorf(
						"ToPatternQuery(%s) expected %v actual %v",
						tt.s,
						tt.expected,
						actual,
					)
				}
			}
		}
	}
}
