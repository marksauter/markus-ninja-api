package server_test

// import (
//   "context"
//   "net/http"
//   "net/http/httptest"
//   "testing"
//
//   "github.com/marksauter/markus-ninja-api/pkg/myaws"
//   "github.com/marksauter/markus-ninja-api/pkg/server"
// )
//
// var mockJwtKms = myaws.NewMockJwtKms()
//
// func TestSetUserToken(t *testing.T) {
//   req, err := http.NewRequest("POST", "/login", nil)
//   if err != nil {
//     t.Fatal(err)
//   }
//
//   testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//     if val, ok := r.Context().Value("user.token").(string); !ok {
//       t.Errorf("user.token not in request context: got %q", val)
//     }
//   })
//
//   rr := httptest.NewRecorder()
//   handler := server.SetUserToken(testHandler)
//
//   ctx := context.WithValue(req.Context(), "user", server.User{})
//   ctx = context.WithValue(ctx, "JwtKms", mockJwtKms)
//   handler.ServeHTTP(rr, req.WithContext(ctx))
// }
