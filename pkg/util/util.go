package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"os"
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
	return bluemonday.UGCPolicy().SanitizeBytes(unsafe)
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
