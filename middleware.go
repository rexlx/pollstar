package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("sessionkey"))

// var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func VoteEnforcer(next http.Handler) http.Handler {
	fmt.Println("mw called")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session.id")
		if session.Values["voted"] == true {
			http.Error(w, "<-You have already voted", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
