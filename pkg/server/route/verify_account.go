package route

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func VerifyAccount(svcs *service.Services) http.Handler {
	requestAccountVerificationHandler := RequestAccountVerificationHandler{
		Svcs: svcs,
	}
	return middleware.CommonMiddleware.Append(
		requestAccountVerificationCors.Handler,
	).Then(requestAccountVerificationHandler)
}

var requestAccountVerificationCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type RequestAccountVerificationHandler struct {
	Svcs *service.Services
}

func (h RequestAccountVerificationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	var verifyAccount struct {
		Token string `json:"token"`
	}

	err := myhttp.UnmarshalRequestBody(req, &verifyAccount)
	if err != nil {
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	avt, err := h.Svcs.AVT.GetByPK(verifyAccount.Token)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if avt.UserId.Status != pgtype.Present {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if avt.VerifiedAt.Status == pgtype.Present {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if avt.ExpiresAt.Time.Before(time.Now()) {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	err = h.Svcs.Role.GrantUser(avt.UserId.String, data.UserRole)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	err = avt.VerifiedAt.Set(time.Now())
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	err = h.Svcs.AVT.Update(avt)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.WriteHeader(http.StatusOK)
	return
}
