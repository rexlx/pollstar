package main

import (
	"fmt"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// func (p *Poll) CreateBarCharItems() []opts.BarData {
// 	var items []opts.BarData
// 	results := p.Results()
// 	for k, v := range results {
// 		items = append(items, opts.BarData{Value: v, Name: k})
// 	}
// 	return items
// }

// func (p *Poll) CreateBarItems(val int) []opts.BarData {
// 	var items []opts.BarData

// 	return items
// }

func (p *Poll) CreateBarChart() *charts.Bar {
	results := p.Results()
	var tickNames []string
	var categories [][]opts.BarData

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Poll Results",
		}),
	)

	for _, q := range p.Questions {
		tickNames = append(tickNames, q.Question)
		if v, ok := results[q.Question]; ok {
			categories = append(categories, CreateBarItems(v))
		}
	}

	bar.SetXAxis(tickNames)
	fmt.Println("tickNames", tickNames, results)
	for _, c := range categories {
		bar.AddSeries("a", c)
	}

	return bar
}

func CreateBarItems(results []int) []opts.BarData {
	var items []opts.BarData
	for _, v := range results {
		items = append(items, opts.BarData{Value: v})
	}
	return items
}

func (p *Poll) CreateBarXAxis() []string {
	var items []string
	p.Mem.RLock()
	defer p.Mem.RUnlock()
	// results := p.Results()
	for k, _ := range p.Selections {
		tickName := fmt.Sprintf("%v-%v", k.QuestionID, k.Selection)
		items = append(items, tickName)
	}
	return items
}
