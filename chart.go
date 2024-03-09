package main

import (
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

func (p *Poll) CreateBarItems(k string, vals []int) []opts.BarData {
	var items []opts.BarData
	for _, v := range vals {
		items = append(items, opts.BarData{Value: v, Name: k})
	}
	return items
}

func (p *Poll) CreateBarChart() *charts.Bar {
	results := p.Results()
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Poll Results",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
			Data: p.CreateBarXAxis(),
		}),
		charts.WithYAxisOpts(opts.YAxis{}),
	)

	for k, v := range results {
		bar.AddSeries(k, p.CreateBarItems(k, v))
	}

	return bar
}

func (p *Poll) CreateBarXAxis() []string {
	var items []string
	for _, q := range p.Questions {
		items = append(items, q.ID)
	}
	return items
}
