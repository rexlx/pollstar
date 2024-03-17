package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type HTMXGateway struct {
	Modes
	Storage   *Storage
	Addr      string
	Done      chan struct{}
	Poll      *Poll
	Style     BasicStyle
	StartTime time.Time
	Server    *http.ServeMux
}

type Modes struct {
	AdminMode    bool
	SaveToBucket bool
	Collapse     bool
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
	fmt.Fprint(w, div)
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
	if !h.AdminMode {
		fmt.Fprint(w, "Admin mode is not enabled, youve been reported")
		return
	}

	session, _ := store.Get(r, "session.admin")
	if session.Values["session.admin"] == nil {
		session.Values["session.admin"] = "supercowpower"
		session.Save(r, w)
	}

	questoinsDiv := h.Poll.AdminCreateQuestionHTML()
	configPage := fmt.Sprintf(configPage, questoinsDiv)
	fmt.Fprint(w, configPage)
}

func (h *HTMXGateway) QuestionHandler(w http.ResponseWriter, r *http.Request) {
	if !h.AdminMode {
		fmt.Fprint(w, "Admin mode is not enabled, youve been reported")
		return
	}

	session, _ := store.Get(r, "session.admin")
	if session.Values["session.admin"] == nil {
		session.Values["session.admin"] = "supercowpower"
		session.Save(r, w)
	}

	questoinsDiv := h.Poll.AdminCreateQuestionHTML()
	// configPage := fmt.Sprintf(configPage, questoinsDiv)
	fmt.Fprint(w, questoinsDiv)
}

func (h *HTMXGateway) ClearPollHandler(w http.ResponseWriter, r *http.Request) {
	if !h.AdminMode {
		http.Error(w, "Admin mode is not enabled", http.StatusForbidden)
		return
	}

	session, _ := store.Get(r, "session.admin")
	if session.Values["session.admin"] != "supercowpower" {
		http.Error(w, "You are not the one", http.StatusForbidden)
		return
	}

	h.Poll.Clear()
	fmt.Fprint(w, "Poll has been cleared")
}

func (h *HTMXGateway) AddOptionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<input class="input" type="text" name="options">`)
}

func (h *HTMXGateway) AddQuestionHandler(w http.ResponseWriter, r *http.Request) {
	var questionOptions []string
	if !h.AdminMode {
		http.Error(w, "Admin mode is not enabled", http.StatusForbidden)
		return
	}

	session, _ := store.Get(r, "session.admin")
	if session.Values["session.admin"] != "supercowpower" {
		http.Error(w, "You are not the one", http.StatusForbidden)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formData := r.Form
	question := formData.Get("question")
	options := formData["options"]
	if question == "" || len(options) < 2 {
		http.Error(w, "Invalid question", http.StatusBadRequest)
		return
	}
	for _, o := range options {
		if o != "" {
			questionOptions = append(questionOptions, o)
		}
	}
	id := uuid.New().String()
	q := Question{
		ID:       id,
		Question: question,
		Options:  questionOptions,
	}
	h.Poll.AddQuestion(q)
	// http.Redirect(w, r, "/config", http.StatusSeeOther)
	fmt.Fprintf(w, `<small class="has-text-success">question added (%v)</small><br>`, id)
}

func (h *HTMXGateway) AdminModeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.admin")
	if session.Values["session.admin"] == nil {
		session.Values["session.admin"] = "supercowpower"
		session.Save(r, w)
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	formData := r.Form
	if formData.Get("adminmode") == "true" {
		h.AdminMode = true
	} else {
		h.AdminMode = false
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=results.json")

	json.NewEncoder(w).Encode(results)
	fmt.Println("results downloaded", r.RemoteAddr)
}

func (h *HTMXGateway) Start() error {
	return http.ListenAndServe(h.Addr, h.Server)
}

func (h *HTMXGateway) Collapse() error {
	if h.Modes.SaveToBucket {
		return h.SaveToBucket()
	} else {
		return h.SaveToFile("test.json")
	}
}

func (h *HTMXGateway) SaveToFile(fname string) error {
	fmt.Println("saving to file", fname)
	out, err := h.Poll.JSON()
	if err != nil {
		return err
	}
	fh, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (h *HTMXGateway) SaveToBucket() error {
	return nil
}

func NewHTMXGateway(m Modes, bucket string) (*HTMXGateway, error) {
	h := &HTMXGateway{
		StartTime: time.Now(),
		Server:    http.NewServeMux(),
	}
	h.Modes = m
	protectedPoll := VoteEnforcer(http.HandlerFunc(h.HomeHandler))
	h.Server.HandleFunc("/poll", h.PollHandler)
	h.Server.Handle("/", protectedPoll)
	h.Server.HandleFunc("/results", h.ResultsHandler)
	h.Server.HandleFunc("/config", h.ConfigHandler)
	h.Server.HandleFunc("/download", h.DownloadHandler)
	h.Server.HandleFunc("/clear-poll", h.ClearPollHandler)
	h.Server.HandleFunc("/add-question", h.AddQuestionHandler)
	h.Server.HandleFunc("/add-option", h.AddOptionHandler)
	h.Server.HandleFunc("/questions", h.QuestionHandler)
	h.Server.HandleFunc("/admin-mode", h.AdminModeHandler)
	h.Poll = NewPoll()
	if bucket != "" {
		s, err := NewStorage(context.Background(), bucket)
		if err != nil {
			return nil, err
		}
		s.Active = true
		h.Storage = s
	}
	return h, nil
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

var configPage = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>pollstar</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">
	<script src="https://unpkg.com/htmx.org@1.9.6" integrity="sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni" crossorigin="anonymous"></script>
</head>
<body>
<nav class="navbar has-background-dark" role="navigation" aria-label="main navigation">
	<div class="navbar-brand">
		<a class="navbar-item" href="/">
			<h1 class="title is-1 has-text-info">pollstar</h1>
		</a>
	</div>
	<div class="navbar-menu">
		<div class="navbar-end">
			<div class="navbar-item">
				<a hx-get="/clear-poll" class="button is-danger has-text-black mr-2" hx-swap="none">clear poll</a>
				<a hx-get="/admin-mode" class="button is-danger has-text-black mr-2" hx-swap="none">admin mode</a>
			</div>
		</div>
	</div>
</nav>
<section class="section has-background-black">
<div class="container">
<form hx-post="/add-question" hx-swap="beforeend" hx-on::after-request="this.reset()">
	<div class="field">
		<label class="label has-text-white">question</label>
		<div class="control">
			<input class="input" type="text" name="question">
		</div>
	</div>
	<div class="field">
		<label class="label has-text-white">options</label>
		<button hx-get="/add-option" hx-swap="afterend" class="button is-info has-text-black">add</button>
		<div class="control">
			<input class="input" type="text" name="options">
			<input class="input" type="text" name="options">
			</div>
		</div>
		<div class="field">
			<div class="control">
			<button type="submit" class="button is-info has-text-black">submit</button>
			</div>
	</div>
	</form>
	<hr>
	<div class="container questions" hx-get="/questions" hx-trigger="every 2s">
	%v
	</div>
</section>
</body>
</html>`
