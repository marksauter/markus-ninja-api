package util_test

import (
	"reflect"
	"testing"
	"unicode"

	"github.com/marksauter/markus-ninja-api/pkg/util"
)

var splitTests = []struct {
	s        string
	expected []string
}{
	{
		"foo",
		[]string{"foo"},
	},
	{
		"foo bar baz qux",
		[]string{"foo", "bar", "baz", "qux"},
	},
	{
		"FooBarBazQux",
		[]string{"Foo", "Bar", "Baz", "Qux"},
	},
	{
		"fooBarBazQux",
		[]string{"foo", "Bar", "Baz", "Qux"},
	},
	{
		"foo-bar-baz-qux",
		[]string{"foo", "bar", "baz", "qux"},
	},
	{
		"foo_bar_baz_qux",
		[]string{"foo", "bar", "baz", "qux"},
	},
	{
		"FooBARBazQUX",
		[]string{"Foo", "BAR", "Baz", "QUX"},
	},
	{
		"foo bar_bazQux",
		[]string{"foo", "bar", "baz", "Qux"},
	},
	{
		"foo  bar--baz__qux",
		[]string{"foo", "bar", "baz", "qux"},
	},
	{
		"foo123bar456baz789qux",
		[]string{"foo", "123", "bar", "456", "baz", "789", "qux"},
	},
}

func testDelimiter(r rune) bool {
	return unicode.IsSpace(r) ||
		r == '-' ||
		r == '_'
}

func TestSplit(t *testing.T) {
	for _, tt := range splitTests {
		actual := util.Split(tt.s, testDelimiter)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf(
				"TestSplit(%s): expected %v, actual %v",
				tt.s,
				tt.expected,
				actual,
			)
		}
	}
}
