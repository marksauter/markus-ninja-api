package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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

var implicitFiguresRegexp = regexp.MustCompile(`<p><img src="(.+)" alt="(.*)"\s*/></p>`)

func implicitFigures(s string) string {
	result := implicitFiguresRegexp.FindStringSubmatch(s)
	if len(result) == 0 {
		return s
	}
	src := result[1]
	figcaption := result[2]

	figure := `<figure><img src="` + src + `" alt=""/>`
	if figcaption != "" {
		figure += `<figcaption>` + figcaption + `</figcaption>`
	}
	figure += `</figure>`

	return figure
}

func MarkdownImplicitFigures(input []byte) []byte {
	markdown := string(input)
	withFigures := implicitFiguresRegexp.ReplaceAllStringFunc(markdown, implicitFigures)

	return []byte(withFigures)
}

var codeBlockLanguageRegexp = regexp.MustCompile(`^language-[a-z-A-Z0-9]+$`)
var paragraphClassesRegexp = regexp.MustCompile(`^((caption|leading)(\s+|$))*$`)
var figureClassesRegexp = regexp.MustCompile(`^((left|right)(\s+|$))*$`)
var anchorTargetRegexp = regexp.MustCompile(`^(_blank)$`)

func MarkdownToHTML(input []byte) []byte {
	unsafe := blackfriday.Run(
		input,
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions|
				blackfriday.HardLineBreak|
				blackfriday.Footnotes,
		),
	)
	unsafeWithFigures := MarkdownImplicitFigures(unsafe)
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(codeBlockLanguageRegexp).OnElements("code")
	p.AllowAttrs("class").Matching(paragraphClassesRegexp).OnElements("p")
	p.AllowAttrs("class").Matching(figureClassesRegexp).OnElements("figure")
	p.AllowAttrs("target").Matching(anchorTargetRegexp).OnElements("a")
	return p.SanitizeBytes(unsafeWithFigures)
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
	default:
		paddingLeft = ""
	}
	switch match[len(match)-1] {
	case ' ':
		paddingRight = " "
	case '\n':
		paddingRight = "\n"
	case '\t':
		paddingRight = "\t"
	default:
		paddingRight = ""
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

func Trace(message string) string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d: %s", filepath.Base(frame.Function), frame.Line, message)
}

func NewBool(value bool) *bool {
	b := value
	return &b
}

func RemoveQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
