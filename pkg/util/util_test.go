package util_test

import (
	"reflect"
	"testing"

	"github.com/marksauter/markus-ninja-api/pkg/util"
)

var splitIntoWordsTests = []struct {
	s        string
	expected []string
}{
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
		[]string{"Foo", "BARB", "az", "QUX"},
	},
	{
		"foo bar_bazQux",
		[]string{"foo", "bar", "baz", "Qux"},
	},
}

func TestSplitIntoWords(t *testing.T) {
	for _, tt := range splitIntoWordsTests {
		actual := util.SplitIntoWords(tt.s)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf(
				"TestSplitIntoWords(%s): expected %v, actual %v",
				tt.s,
				tt.expected,
				actual,
			)
		}
	}
}
