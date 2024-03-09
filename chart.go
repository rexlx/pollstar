package main

import (
	"fmt"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func (p *Poll) CreateBarChart() []*charts.Bar {
	results := p.Results()
	// var tickNames []string
	var chartz []*charts.Bar

	p.Mem.RLock()
	defer p.Mem.RUnlock()
	for _, q := range p.Questions {
		var xAxis []string
		// tickNames = append(tickNames, q.Question)
		if v, ok := results[q.Question]; ok {
			for _, o := range q.Options {
				xAxis = append(xAxis, o)
			}
			newChart := charts.NewBar()
			newChart.SetGlobalOptions(
				charts.WithTitleOpts(opts.Title{
					Title: q.Question,
				}),
			)
			newChart.SetXAxis(xAxis)
			newChart.AddSeries("selections", CreateBarItems(v))
			chartz = append(chartz, newChart)
		}
	}

	return chartz
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
