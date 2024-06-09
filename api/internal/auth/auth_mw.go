package auth

import (
	"context"
	"fmt"
	"github.com/moxicom/ws-server/internal/ws"
	"net/http"
)

func AuthMiddleware(hub *ws.Hub, f func(*ws.Hub, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middleware")

		token, ok := r.URL.Query()["bearer"]

		if ok && len(token) == 1 {
			user, err := validateToken(token[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else {
				ctx := context.WithValue(r.Context(), ws.UserContextKey, user)
				f(hub, w, r.WithContext(ctx))
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("login first"))
		}
	}
}
