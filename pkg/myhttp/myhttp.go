package myhttp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/marksauter/markus-ninja-api/pkg/model"
)

type Response interface {
	StatusHTTP() int
}

func WriteResponseTo(rw http.ResponseWriter, r Response) error {
	json, err := json.Marshal(r)
	if err != nil {
		return err
	}
	rw.WriteHeader(r.StatusHTTP())
	rw.Write(json)
	return nil
}

func UnmarshalRequestBody(req *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("myhttp: %v", err)
	}
	return json.Unmarshal(body, v)
}

func ValidateBasicAuthHeader(req *http.Request) (*model.UserCredentials, error) {
	auth := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		return nil, errors.New("Invalid credentials")
	}
	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return nil, errors.New("Invalid credentials")
	}
	userCredentials := model.UserCredentials{
		Login:    pair[0],
		Password: pair[1],
	}
	return &userCredentials, nil
}
