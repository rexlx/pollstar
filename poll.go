package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type PollSelection struct {
	QuestionID string `json:"question_id"`
	Selection  int    `json:"selection"`
}

type Question struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type Poll struct {
	Mem        *sync.RWMutex
	ID         string                `json:"id"`
	Questions  []Question            `json:"questions"`
	Selections map[PollSelection]int `json:"selections"`
}

func NewPoll() *Poll {
	mux := &sync.RWMutex{}
	return &Poll{
		Selections: make(map[PollSelection]int),
		Mem:        mux,
	}
}

func (p *Poll) AddQuestion(q Question) {
	p.Mem.Lock()
	defer p.Mem.Unlock()
	p.Questions = append(p.Questions, q)
}

func (p *Poll) LoadQuestions(fh string) error {
	contents, err := os.ReadFile(fh)
	if err != nil {
		return err
	}
	questions := make([]Question, 0)
	err = json.Unmarshal(contents, &questions)
	if err != nil {
		return err
	}
	p.Questions = questions
	return nil
}

func (p *Poll) AddSelection(questionID string, selection int) {
	p.Mem.Lock()
	defer p.Mem.Unlock()
	p.Selections[PollSelection{
		QuestionID: questionID,
		Selection:  selection,
	}]++
}

func (p *Poll) Results() map[string][]int {
	results := make(map[string][]int)
	p.Mem.RLock()
	defer p.Mem.RUnlock()
	for _, q := range p.Questions {
		var questionResults []int
		for i := 0; i < len(q.Options); i++ {
			questionResults = append(questionResults, 0)
		}
		for s, count := range p.Selections {
			if s.QuestionID == q.ID {
				questionResults[s.Selection] = count
			}
		}
		results[q.Question] = questionResults
	}
	return results
}

func (p *Poll) TotalVotes() int {
	total := 0
	for _, count := range p.Selections {
		total += count
	}
	return total
}

func (p *Poll) CreateQuestionHTML() string {
	out := `<div class="control">`
	radioTmpl := `<label class="radio has-text-primary"><input type="radio" name="%s" value="%d"> %s</label><hr><br>`
	for _, q := range p.Questions {
		out += fmt.Sprintf(`<h2 class="has-text-link">%s</h2><br>`, q.Question)
		for i, o := range q.Options {
			out += fmt.Sprintf(radioTmpl, q.ID, i, o)
		}
	}
	out += "</div>"
	return out
}

func (p *Poll) Clear() {
	p.Mem.Lock()
	defer p.Mem.Unlock()
	p.Selections = make(map[PollSelection]int)
	p.Questions = make([]Question, 0)
}

func (p *Poll) AdminCreateQuestionHTML() string {
	p.Mem.RLock()
	defer p.Mem.RUnlock()
	out := `<div class="card has-background-dark">`
	tmpl := `<div class="card-content"><h2 class="has-text-link">%s</h2><br>`
	for _, q := range p.Questions {
		out += fmt.Sprintf(tmpl, q.Question)
		for i, o := range q.Options {
			out += fmt.Sprintf(`<p class="has-text-white">%d. %s</p>`, i, o)
		}
		out += "</div>"
	}
	out += "</div>"
	return out
}
