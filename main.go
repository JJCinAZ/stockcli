package main

import (
	"flag"
	"fmt"
	"github.com/jjcinaz/stock/quotes"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		pgmname := filepath.Base(os.Args[0])
		fmt.Printf("Format is: %s <ticker>{,<ticker>}\r\n", pgmname)
		os.Exit(1)
	}
	a := flag.Args()
	os.Exit(doquote(a))
}

func doquote(tickers []string) int {
	const (
		price = "latest"
		change = "change"
		percent = "percent"
	)
	var (
		okay bool
		errstring string
	)
	market := quotes.NewMarket()
	quotes := quotes.NewQuotes(market, tickers)
	if okay, errstring = market.Fetch().Ok(); !okay {
		fmt.Println(errstring)
		return 1
	}
	if okay, errstring = quotes.Fetch().Ok(); !okay {
		fmt.Println(errstring)
		return 1
	}
	fmt.Printf("DOW\t%s (%s/%s)\tNasdaq\t%s (%s/%s)\tS&P 500\t%s (%s/%s)\r\n",
		market.Dow[price], market.Dow[change], market.Dow[percent],
		market.Nasdaq[price], market.Nasdaq[change], market.Nasdaq[percent],
		market.Sp500[price], market.Sp500[change], market.Sp500[percent],
	)
	for _, q := range quotes.Stocks {
		var (
			l, h, c float64
		)
		l, _ = strconv.ParseFloat(q.Low52, 64)
		h, _ = strconv.ParseFloat(q.High52, 64)
		c, _ = strconv.ParseFloat(q.LastTrade, 64)
		p := math.Round((c - l) / (h - l) * 100 / 2)
		if p > 49 {
			p = 49
		}
		dotline := make([]byte, 50)
		for i := range dotline {
			dotline[i] = '.'
		}
		dotline[int(p)] = '*'
		fmt.Printf("%s\t%s (%s/%s%%) on %s shares\r\n\t52-wk: %s%s%s\r\n",
			q.Ticker, q.LastTrade, q.Change, q.ChangePct, q.Volume, q.Low52, dotline, q.High52)
	}
	return 0
}
