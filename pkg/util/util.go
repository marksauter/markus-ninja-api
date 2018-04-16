package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func LoadEnv() error {
	// Load env vars from .env
	return godotenv.Load()
}

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

func CompressString(s string) string {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func DecompressString(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadAll(r)
	return string(bs)
}
