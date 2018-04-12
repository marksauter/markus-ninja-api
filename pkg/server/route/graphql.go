package route

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
)

type GraphQLHandler struct {
	Logger *mylog.Logger
	Schema *graphql.Schema
}

func (h GraphQLHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if len(params.Query) > 2000 {
		http.Error(rw, "Query too large.", http.StatusBadRequest)
		return
	}

	response := h.Schema.Exec(req.Context(), params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(responseJSON)
}

func MarshalID(kind string, spec interface{}) graphql.ID {
	d, err := json.Marshal(spec)
	if err != nil {
		panic(fmt.Errorf("server.MarshalID: %s", err))
	}
	return graphql.ID(
		base64.URLEncoding.EncodeToString(
			append([]byte(kind+":"), d...),
		),
	)
}

func UnmarshalKind(id graphql.ID) string {
	s, err := base64.URLEncoding.DecodeString(string(id))
	if err != nil {
		return ""
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return ""
	}
	return string(s[:i])
}

func UnmarshalSpec(id graphql.ID, v interface{}) error {
	s, err := base64.URLEncoding.DecodeString(string(id))
	if err != nil {
		return err
	}
	i := strings.IndexByte(string(s), ':')
	if i == -1 {
		return errors.New("invalid graphql.ID")
	}
	return json.Unmarshal([]byte(s[i+1:]), v)
}
