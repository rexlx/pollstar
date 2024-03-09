package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	ID         string                `json:"id"`
	Questions  []Question            `json:"questions"`
	Selections map[PollSelection]int `json:"selections"`
}

func NewPoll() *Poll {
	return &Poll{
		Selections: make(map[PollSelection]int),
	}
}

func (p *Poll) AddQuestion(question string, options []string) {
	q := Question{
		ID:       "1",
		Question: question,
		Options:  options,
	}
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
	p.Selections[PollSelection{
		QuestionID: questionID,
		Selection:  selection,
	}]++
}

func (p *Poll) Results() map[string][]int {
	// fmt.Println("selections", p.Selections)
	results := make(map[string][]int)
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
	radioTmpl := `<label class="radio"><input type="radio" name="%s" value="%d">%s</label><hr><br>`
	for _, q := range p.Questions {
		out += "<h2>" + q.Question + "</h2><br>"
		for i, o := range q.Options {
			out += fmt.Sprintf(radioTmpl, q.ID, i, o)
		}
	}
	out += "</div>"
	return out
}
