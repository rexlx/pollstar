package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

func main() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}
	gateway := NewHTMXGateway()
	gateway.Poll = NewPoll()
	// p.AddQuestion("What is your favorite programming language?", []string{"Python", "Go", "Java", "C"})
	err := gateway.Poll.LoadQuestions("questions.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	http.ListenAndServe(":3000", gateway.Server)
}
