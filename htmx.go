package main

import (
	"encoding/json"
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

// this should be protected by middleware to prevent multiple votes
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
	if session.Values["voted"] == true {
		http.Error(w, "You have already voted", http.StatusForbidden)
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
	session.Save(r, w)
	div := `<div class="notification is-success">Thanks for voting! <a href="/results">results</a> | <a href="/download">raw data</a></div>`
	fmt.Fprintf(w, div)
}

func (h *HTMXGateway) ResultsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")

	if session.Values["voted"] != true {
		http.Error(w, "You must vote first", http.StatusForbidden)
		return
	}

	barChart := h.Poll.CreateBarChart()
	for _, c := range barChart {
		c.Render(w)
	}

}

func (h *HTMXGateway) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "config not configured")
}

func (h *HTMXGateway) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if session.Values["session.id"] == nil {
		fmt.Println("no session id for downloader", r.RemoteAddr)
		http.Error(w, "No session ID", http.StatusBadRequest)
		return
	}
	h.Poll.Mem.RLock()
	results := h.Poll.Results()
	h.Poll.Mem.RUnlock()
	// out, err := json.Marshal(results)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=results.json")
	// w.Header().Set("Content-Length", strconv.Itoa(len(results)))
	json.NewEncoder(w).Encode(results)
	fmt.Println("results downloaded", r.RemoteAddr)
}

func NewHTMXGateway() *HTMXGateway {
	h := &HTMXGateway{
		StartTime: time.Now(),
		Server:    http.NewServeMux(),
	}
	protectedPoll := VoteEnforcer(http.HandlerFunc(h.HomeHandler))
	h.Server.HandleFunc("/poll", h.PollHandler)
	h.Server.Handle("/", protectedPoll)
	h.Server.HandleFunc("/results", h.ResultsHandler)
	h.Server.HandleFunc("/config", h.ConfigHandler)
	h.Server.HandleFunc("/download", h.DownloadHandler)
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
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>pollstar</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">
	<script src="https://unpkg.com/htmx.org@1.9.6" integrity="sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni" crossorigin="anonymous"></script>
</head>
<body>
<section class="section has-background-dark">
	<div class="container">
	<form hx-post="/poll">
		%s
		<button type="submit" class="button is-info has-text-black">submit</button>
	</form>
	</div>
</section>
</body>
</html>`
