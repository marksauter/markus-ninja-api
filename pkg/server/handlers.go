package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/marksauter/markus-ninja-api/pkg/myctx"
	"github.com/marksauter/markus-ninja-api/pkg/repo"
)

type RootRouteHandler struct{}

func (h RootRouteHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("hello, world!"))
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

type GraphQLHandler struct {
	Schema *graphql.Schema
}

func (h GraphQLHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	userRepo := new(repo.UserRepo)
	var ok bool
	ctx, ok = myctx.UserRepo.NewContext(ctx, userRepo)
	if !ok {
		http.Error(rw,
			"handlers.GraphQLHandler: expected type UserRepo for UserRepo.NewContext",
			http.StatusInternalServerError,
		)
		return
	}

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

	response := h.Schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(responseJSON)
}
