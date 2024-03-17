package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/sessions"
)

var (
	projectID        = flag.String("project-id", "default-project", "The project id")
	pollId           = flag.String("poll-id", "default-poll", "The poll id")
	addr             = flag.String("addr", ":3000", "The address to listen on")
	questions        = flag.String("questions", "questions.json", "The file containing the questions")
	sessionKey       = flag.String("session-key", "thisisfine", "The session key")
	startInAdminMode = flag.Bool("admin", false, "Start in admin mode")
	bucket           = flag.String("bucket", "", "The bucket to use for storing the poll")
	collapse         = flag.Bool("collapse", false, "Collapse the poll after voting")
	ttl              = flag.Int("ttl", 60, "how long for the poll to live")
	// sessionDur       = flag.Int("duration", 15, "The time to live for the session")
)

func main() {
	var modes Modes

	done := make(chan struct{})
	signals := make(chan os.Signal, 1)
	flag.Parse()

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	modes.AdminMode = *startInAdminMode
	modes.Collapse = *collapse

	gateway, err := NewHTMXGateway(modes, BucketConfig{Bucket: *bucket, Project: *projectID, Poll: *pollId})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gateway.Poll = NewPoll()
	gateway.Done = done

	err = gateway.Poll.LoadQuestions(*questions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Loaded questions, starting server")

	server := &http.Server{
		Addr:    *addr,
		Handler: gateway.Server,
	}

	go func() {
		<-done
		server.Shutdown(context.Background())
	}()
	go server.ListenAndServe()

	for {
		fmt.Println("Poll is running")
		select {
		case <-time.After(time.Duration(*ttl) * time.Second):
			if modes.Collapse {
				gateway.Collapse()
				fmt.Println("Poll has expired")
				close(done)
				os.Exit(0)
			}
		case <-signals:
			gateway.Collapse()
			fmt.Println("received signal to terminate")
			close(done)
			os.Exit(0)
		}
	}

}
