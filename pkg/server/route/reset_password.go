package route

import (
	"net"
	"net/http"

	"github.com/marksauter/markus-ninja-api/pkg/data"
	"github.com/marksauter/markus-ninja-api/pkg/myhttp"
	"github.com/marksauter/markus-ninja-api/pkg/mylog"
	"github.com/marksauter/markus-ninja-api/pkg/mysmtp"
	"github.com/marksauter/markus-ninja-api/pkg/server/middleware"
	"github.com/rs/cors"
	"github.com/rs/xid"
)

func PasswordReset(
	mailSvc mysmtp.Mailer,
	pwrtSvc *data.PasswordResetTokenService,
	userSvc *data.UserService,
) http.Handler {
	passwordResetHandler := PasswordResetHandler{
		MailSvc: mailSvc,
		PWRTSvc: pwrtSvc,
		UserSvc: userSvc,
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

type PasswordResetHandler struct {
	MailSvc mysmtp.Mailer
	PWRTSvc *data.PasswordResetTokenService
	UserSvc *data.UserService
}

func (h PasswordResetHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	pwrt := &data.PasswordResetTokenModel{}
	token := xid.New()
	pwrt.Token.Set(token.String())

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

	user, err := h.UserSvc.GetByPrimaryEmail(reset.Email)
	switch err {
	case nil:
		pwrt.UserId = user.Id
	case data.ErrNotFound:
	default:
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	err = h.PWRTSvc.Create(pwrt)
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

	err = h.MailSvc.SendPasswordResetMail(reset.Email, token.String())
	if err != nil {
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	rw.WriteHeader(http.StatusOK)
	return
}
