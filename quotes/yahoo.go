package quotes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const quotesURLv7 = `https://query1.finance.yahoo.com/v7/finance/quote?symbols=%s`
const quotesURLv7QueryParts = `&range=1d&interval=5m&indicators=close&includeTimestamps=false&includePrePost=false&corsDomain=finance.yahoo.com&.tsrc=finance`

// Stock stores quote information for the particular stock ticker. The data
// for all the fields except 'Advancing' is fetched using Yahoo market API.
type Stock struct {
	Ticker     string `json:"symbol"` // Stock ticker.
	LastTrade  string `json:"regularMarketPrice"` // l1: last trade.
	Change     string `json:"regularMarketChange"` // c6: change real time.
	ChangePct  string `json:"regularMarketChangePercent"` // k2: percent change real time.
	Open       string `json:"regularMarketOpen"`  // o: market open price.
	Low        string `json:"regularMarketDayLow"` // g: day's low.
	High       string `json:"regularMarketDayHigh"` // h: day's high.
	Low52      string `json:"fiftyTwoWeekLow"` // j: 52-weeks low.
	High52     string `json:"fiftyTwoWeekHigh"` // k: 52-weeks high.
	Volume     string `json:"regularMarketVolume"` // v: volume.
	AvgVolume  string `json:"averageDailyVolume10Day"` // a2: average volume.
	PeRatio    string `json:"trailingPE"` // r2: P/E ration real time.
	Dividend   string `json:"trailingAnnualDividendRate"` // d: dividend.
	Yield      string `json:"trailingAnnualDividendYield"` // y: dividend yield.
	MarketCap  string `json:"marketCap"` // j3: market cap real time.
	Currency   string `json:"currency"` // String code for currency of stock.
	Advancing  bool   // True when change is >= $0.
}

// Quotes stores relevant pointers as well as the array of stock quotes for
// the Tickers we are tracking.
type Quotes struct {
	market  *Market  // Pointer to Market.
	Tickers []string // List of stock Tickers to display.
	Stocks  []Stock  // Array of stock quote data.
	errors  string   // Error string if any.
}

// Sets the initial values and returns new Quotes struct.
func NewQuotes(market *Market, tickers []string) *Quotes {
	return &Quotes{
		market:  market,
		Tickers: tickers,
		errors:  ``,
	}
}

// Fetch the latest stock quotes and parse raw fetched data into array of
// []Stock structs.
func (quotes *Quotes) Fetch() (self *Quotes) {
	self = quotes // <-- This ensures we return correct quotes after recover() from panic().
	if quotes.isReady() {
		defer func() {
			if err := recover(); err != nil {
				quotes.errors = fmt.Sprintf("Error fetching quotes...\r\n%s", err)
			}
		}()

		url := fmt.Sprintf(quotesURLv7, strings.Join(quotes.Tickers, `,`))
		response, err := http.Get(url + quotesURLv7QueryParts)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		quotes.parse2(body)
	}

	return quotes
}

// Ok returns two values: 1) boolean indicating whether the error has occured,
// and 2) the error text itself.
func (quotes *Quotes) Ok() (bool, string) {
	return quotes.errors == ``, quotes.errors
}

// isReady returns true if we haven't fetched the quotes yet *or* the stock
// market is still open and we might want to grab the latest quotes. In both
// cases we make sure the list of requested Tickers is not empty.
func (quotes *Quotes) isReady() bool {
	return (quotes.Stocks == nil || !quotes.market.IsClosed) && len(quotes.Tickers) > 0
}

// this will parse the json objects
func (quotes *Quotes) parse2(body []byte) (*Quotes, error) {
	// response -> quoteResponse -> result|error (array) -> map[string]interface{}
	// Stocks has non-int things
	// d := map[string]map[string][]Stock{}
	// some of these are numbers vs strings
	// d := map[string]map[string][]map[string]string{}
	d := map[string]map[string][]map[string]interface{}{}
	err := json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	results := d["quoteResponse"]["result"]

	quotes.Stocks = make([]Stock, len(results))
	for i, raw := range results {
		result := map[string]string{}
		for k, v := range raw {
			switch v.(type) {
			case string:
				result[k] = v.(string)
			case float64:
				result[k] = float2Str(v.(float64))
			default:
				result[k] = fmt.Sprintf("%v", v)
			}

		}
		quotes.Stocks[i].Ticker = result["symbol"]
		quotes.Stocks[i].LastTrade = result["regularMarketPrice"]
		quotes.Stocks[i].Change = result["regularMarketChange"]
		quotes.Stocks[i].ChangePct = result["regularMarketChangePercent"]
		quotes.Stocks[i].Open = result["regularMarketOpen"]
		quotes.Stocks[i].Low = result["regularMarketDayLow"]
		quotes.Stocks[i].High = result["regularMarketDayHigh"]
		quotes.Stocks[i].Low52 = result["fiftyTwoWeekLow"]
		quotes.Stocks[i].High52 = result["fiftyTwoWeekHigh"]
		quotes.Stocks[i].Volume = result["regularMarketVolume"]
		quotes.Stocks[i].AvgVolume = result["averageDailyVolume10Day"]
		quotes.Stocks[i].PeRatio = result["trailingPE"]
		quotes.Stocks[i].Dividend = result["trailingAnnualDividendRate"]
		quotes.Stocks[i].Yield = result["trailingAnnualDividendYield"]
		quotes.Stocks[i].MarketCap = result["marketCap"]
		quotes.Stocks[i].Currency = result["currency"]

		/*
			fmt.Println(i)
			fmt.Println("-------------------")
			for k, v := range result {
				fmt.Println(k, v)
			}
			fmt.Println("-------------------")
		*/
		adv, err := strconv.ParseFloat(quotes.Stocks[i].Change, 64)
		if err == nil {
			quotes.Stocks[i].Advancing = adv >= 0.0
		}
	}
	return quotes, nil
}

func float2Str(v float64) string {
	unit := ""
	switch {
	case v > 1.0e12:
		v = v / 1.0e12
		unit = "T"
	case v > 1.0e9:
		v = v / 1.0e9
		unit = "B"
	case v > 1.0e6:
		v = v / 1.0e6
		unit = "M"
	case v > 1.0e5:
		v = v / 1.0e3
		unit = "K"
	default:
		unit = ""
	}
	return fmt.Sprintf("%0.3f%s", v, unit)
}

