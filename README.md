simple polling web application. See dockerfile if running in docker, otherwise run

```shell
go build .
./pollstar -questions questions.json -admin
```

visit http://localhost:3000/config to configure. toggle admin mode when finished.

routes:
- h.Server.HandleFunc("/poll", h.PollHandler)
- h.Server.Handle("/", protectedPoll)
- h.Server.HandleFunc("/results", h.ResultsHandler)
- h.Server.HandleFunc("/config", h.ConfigHandler)
- h.Server.HandleFunc("/download", h.DownloadHandler)
- h.Server.HandleFunc("/clear-poll", h.ClearPollHandler)
- h.Server.HandleFunc("/add-question", h.AddQuestionHandler)
- h.Server.HandleFunc("/add-option", h.AddOptionHandler)
- h.Server.HandleFunc("/questions", h.QuestionHandler)
- h.Server.HandleFunc("/admin-mode", h.AdminModeHandler)