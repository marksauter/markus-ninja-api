package route

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/marksauter/markus-ninja-api/pkg/service"
	"github.com/rs/cors"
)

func RequestPasswordReset(svcs *service.Services) http.Handler {
	requestPasswordResetHandler := RequestPasswordResetHandler{
		Svcs: svcs,
	}
	return middleware.CommonMiddleware.Append(
		requestPasswordResetCors.Handler,
	).Then(requestPasswordResetHandler)
}

var requestPasswordResetCors = cors.New(cors.Options{
	AllowedHeaders: []string{"Content-Type"},
	AllowedMethods: []string{http.MethodOptions, http.MethodPost},
	AllowedOrigins: []string{"ma.rkus.ninja", "localhost:3000"},
})

type RequestPasswordResetHandler struct {
	Svcs *service.Services
}

func (h RequestPasswordResetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	routeVars := mux.Vars(req)

	pwrt := &data.PasswordResetTokenModel{}

	if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			mask := net.CIDRMask(len(ip)*8, len(ip)*8)
			ipNet := &net.IPNet{IP: ip, Mask: mask}
			pwrt.RequestIP.Set(&ipNet)
		}
	}

	var reset struct {
		Email string `json:"email"`
	}

	err := myhttp.UnmarshalRequestBody(req, &reset)
	if err != nil {
		response := myhttp.InvalidRequestErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}
	if reset.Email == "" {
		response := &myhttp.ErrorResponse{
			Error:            myhttp.InvalidCredentials,
			ErrorDescription: "The request email was invalid",
		}
		myhttp.WriteResponseTo(rw, response)
	}
	if len(reset.Email) > 40 {
		response := &myhttp.ErrorResponse{
			Error:            myhttp.InvalidCredentials,
			ErrorDescription: "The request email must be less than or equal to 40 characters",
		}
		myhttp.WriteResponseTo(rw, response)
		return
	}

	pwrt.Email.Set(reset.Email)

	user, err := h.Svcs.User.GetCredentialsByEmail(reset.Email)
	switch err {
	case nil:
		pwrt.UserId = user.Id
	case data.ErrNotFound:
	default:
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	err = h.Svcs.PWRT.Create(pwrt)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if user == nil {
		mylog.Log.WithField(
			"email",
			reset.Email,
		).Warn("Password reset requested for missing email")
		return
	}

	login := routeVars["login"]
	err = h.Svcs.Mail.SendPasswordResetMail(reset.Email, login, pwrt.Token.String)
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.WriteHeader(http.StatusOK)
	return
}
