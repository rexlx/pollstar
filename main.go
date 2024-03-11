package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var (
	questions        = flag.String("questions", "questions.json", "The file containing the questions")
	sessionKey       = flag.String("session-key", "thisisfine", "The session key")
	startInAdminMode = flag.Bool("admin", false, "Start in admin mode")
)

func main() {
	flag.Parse()
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}
	gateway := NewHTMXGateway(Modes{AdminMode: *startInAdminMode})
	gateway.Poll = NewPoll()
	err := gateway.Poll.LoadQuestions(*questions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Loaded questions, starting server")
	log.Fatal(http.ListenAndServe(":3000", gateway.Server))
}
