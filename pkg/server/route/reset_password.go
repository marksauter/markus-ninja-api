package route

import (
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/pgtype"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/myjwt"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/oid"
	"github.com/marksauter/markus-ninja-api/pkg/passwd"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func ResetPassword(svcs *service.Services) http.Handler {
	passwordResetHandler := ResetPasswordHandler{
		Svcs: svcs,
	}
	return middleware.CommonMiddleware.Append(
		passwordResetCors.Handler,
	).Then(passwordResetHandler)
}

var passwordResetCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type ResetPasswordHandler struct {
	Svcs *service.Services
}

func (h ResetPasswordHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	var resetPassword struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	err := myhttp.UnmarshalRequestBody(req, &resetPassword)
	if err != nil {
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	pwrt, err := h.Svcs.PWRT.GetByPK(resetPassword.Token)
	if err == data.ErrNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if _, ok := pwrt.UserId.Get().(oid.OID); !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if pwrt.EndedAt.Status == pgtype.Present {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	user := &data.User{}
	user.Id = pwrt.UserId
	password, err := passwd.New(resetPassword.Password)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create password")
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	if err := password.CheckStrength(passwd.VeryWeak); err != nil {
		mylog.Log.Error("password failed strength check")
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	user.Password.Set(password.Hash())

	err = h.Svcs.User.Update(user)
	if err != nil {
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			mask := net.CIDRMask(len(ip)*8, len(ip)*8)
			ipNet := &net.IPNet{IP: ip, Mask: mask}
			pwrt.EndIP.Set(&ipNet)
		}
	}
	pwrt.EndedAt.Set(time.Now())
	err = h.Svcs.PWRT.Update(pwrt)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	userId, _ := user.Id.Get().(oid.OID)

	exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	payload := myjwt.Payload{Exp: exp, Iat: time.Now().Unix(), Sub: userId.String}
	jwt, err := h.Svcs.Auth.SignJWT(&payload)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	response := &TokenSuccessResponse{
		AccessToken: jwt.String(),
		ExpiresIn:   jwt.Payload.Exp,
	}
	myhttp.WriteResponseTo(rw, response)
}
