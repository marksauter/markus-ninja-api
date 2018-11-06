package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/microcosm-cc/bluemonday"
	"github.com/writeas/go-strip-markdown"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func GetOptionalEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetRequiredEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("Env variable " + key + " required.")
}

func StringToInterface(strs []string) []interface{} {
	results := make([]interface{}, len(strs))
	for i, s := range strs {
		results[i] = s
	}
	return results
}

func MarkdownToHTML(input []byte) []byte {
	unsafe := blackfriday.Run(input)
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-z-A-Z0-9]+$")).OnElements("code")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^((caption|hint|main|overline|secondary|subtitle[1|2])(\s+|$))*$`)).OnElements("p")
	return p.SanitizeBytes(unsafe)
}

func MarkdownToText(s string) string {
	return stripmd.Strip(s)
}

func CompressString(s string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func DecompressString(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return "", err
	}
	bs, err := ioutil.ReadAll(r)
	return string(bs), err
}

func Split(s string, f func(rune) bool) []string {
	firstSplit := strings.FieldsFunc(s, f)
	secondSplit := []string{}
	for _, v := range firstSplit {
		camelcaseSplit := camelcase.Split(v)
		if len(camelcaseSplit) > 1 {
			secondSplit = append(secondSplit, v)
		}
		secondSplit = append(secondSplit, camelcaseSplit...)
	}
	for i, v := range secondSplit {
		secondSplit[i] = strings.ToLower(v)
	}
	return secondSplit
}

var rxHexColor = regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func IsHexColor(str string) bool {
	return rxHexColor.MatchString(str)
}

func ReplaceWithPadding(match, replace string) string {
	paddingLeft := ""
	paddingRight := ""
	switch match[0] {
	case ' ':
		paddingLeft = " "
	case '\n':
		paddingLeft = "\n"
	case '\t':
		paddingLeft = "\t"
	}
	switch match[len(match)-1] {
	case ' ':
		paddingRight = " "
	case '\n':
		paddingRight = "\n"
	case '\t':
		paddingRight = "\t"
	}
	return paddingLeft + replace + paddingRight
}

func RemoveEmptyStrings(strs []string) []string {
	noEmpties := make([]string, 0, len(strs))
	for _, s := range strs {
		if s != "" {
			noEmpties = append(noEmpties, s)
		}
	}
	return noEmpties
}
