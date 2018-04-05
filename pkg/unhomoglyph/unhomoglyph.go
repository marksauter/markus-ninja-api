package unhomoglyph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func Unhomoglyph(s string) string {
	if !imported {
		importData()
	}
	escape := regexp.MustCompile(`/([.?*+^$[\]\\(){}|-])/g`)
	escapedKeys := make([]string, len(data))
	i := 0
	for k, _ := range data {
		escapedKeys[i] = string(escape.ReplaceAll([]byte(k), []byte(`\\$1`)))
		i++
	}
	replaceRegExp := regexp.MustCompile(strings.Join(escapedKeys[:], "|"))

	return string(replaceRegExp.ReplaceAllFunc(
		[]byte(s),
		func(match []byte) []byte {
			return []byte(data[string(match)])
		},
	))
}

var data map[string]string
var imported bool = false

func importData() {
	file, err := ioutil.ReadFile("static/confusables.json")
	if err != nil {
		panic(fmt.Errorf("Unhomoglyph: %v\n", err))
	}
	err = json.Unmarshal(file, &data)
	if err != nil {
		panic(fmt.Errorf("Unhomoglyph: %v\n", err))
	}
	imported = true
}
