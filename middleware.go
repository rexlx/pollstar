package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// sessionKey is the arg in main.go
var store = sessions.NewCookieStore([]byte(*sessionKey))

// you can also use env vars if you so please
// var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func VoteEnforcer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session.id")
		if session.Values["voted"] == true {
			http.Error(w, "You have already voted :)", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
