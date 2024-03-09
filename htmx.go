package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type HTMXGateway struct {
	Poll      *Poll
	Style     BasicStyle
	StartTime time.Time
	Server    *http.ServeMux
}

type BasicStyle struct {
	BodyBG   string
	BodyText string
	H1       string
	Btn      string
	BtnText  string
}

func (h *HTMXGateway) HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if session.Values["session.id"] == nil {
		session.Values["session.id"] = createSessionID()
		session.Save(r, w)
	}

	questionsDiv := h.Poll.CreateQuestionHTML()

	fmt.Fprintf(w, homePage, questionsDiv)
}

func (h *HTMXGateway) PollHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if session.Values["session.id"] == nil {
		http.Error(w, "No session ID", http.StatusBadRequest)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formData := r.Form
	for k, v := range formData {
		i, err := strconv.Atoi(v[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.Poll.AddSelection(k, i)
	}
	if session.Values["voted"] == true {
		http.Error(w, "You have already voted", http.StatusForbidden)
		return
	}
	session.Values["voted"] = true
	fmt.Fprintf(w, "Thanks for voting!")
}

func (h *HTMXGateway) ResultsHandler(w http.ResponseWriter, r *http.Request) {
	// session, _ := store.Get(r, "session.id")
	// if session.Values["voted"] != true {
	// 	http.Error(w, "You must vote first", http.StatusForbidden)
	// 	return
	// }

	barChart := h.Poll.CreateBarChart()
	for _, c := range barChart {
		c.Render(w)
	}

}

func (h *HTMXGateway) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	// ...
}

func NewHTMXGateway() *HTMXGateway {
	h := &HTMXGateway{
		StartTime: time.Now(),
		Server:    http.NewServeMux(),
	}
	protectedPoll := VoteEnforcer(http.HandlerFunc(h.PollHandler))
	h.Server.Handle("/poll", protectedPoll)
	h.Server.HandleFunc("/", h.HomeHandler)
	h.Server.HandleFunc("/results", h.ResultsHandler)
	h.Server.HandleFunc("/config", h.ConfigHandler)
	h.Poll = NewPoll()
	return h
}

func createSessionID() string {
	out := uuid.New().String()
	return out
}

var homePage = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>poll</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">
	<script src="https://unpkg.com/htmx.org@1.9.6" integrity="sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni" crossorigin="anonymous"></script>
</head>
<body>
<div class="container">
  <form hx-post="/poll">
    %s
    <button type="submit" class="button is-info">submit</button>
  </form>
</div>
</body>
</html>`
