package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WebSocketURL = "wss://ws.finnhub.io?token="
	APIKey       = "abc123_BOGUS_BOGUS_BOGUS_xyz789" // Get your own hey here: https://finnhub.io/
)

var (
	Symbols []string
)

func init() {
	Symbols = []string{"AAPL", "AMZN", "BINANCE:BTCUSDT", "IC MARKETS:1"}
}

/* Example trade:
{
  "data": [
    {
      "p": 7296.89,
      "s": "BINANCE:BTCUSDT",
      "t": 1575526691134,
      "v": 0.011467
    }
  ],
  "type": "trade"
}
*/

type Data struct {
	Price     float64 `json:"p"`
	Symbol    string  `json:"s"`
	Timestamp int64   `json:"t"` // Unix milliseconds
	Volume    float64 `json:"v"`
}

type Trade struct {
	Data []Data `json:"data"`
	Type string `json:"type"`
}

func main() {
	var ws *websocket.Conn
	var err error

	if ws, _, err = websocket.DefaultDialer.Dial(WebSocketURL+APIKey, nil); err != nil {
		panic(err)
	}
	defer ws.Close()

	//
	// Query for the given list of symbols
	//
	var msg []byte
	for _, s := range Symbols {
		if msg, err = json.Marshal(map[string]any{"type": "subscribe", "symbol": s}); err != nil {
			panic(err)
		}
		ws.WriteMessage(websocket.TextMessage, msg)
	}

	//
	// Listen to any relevant trades
	//
	var trade Trade
	for {
		if err = ws.ReadJSON(&trade); err != nil {
			panic(err)
		}
		for _, d := range trade.Data {
			ts := time.UnixMilli(d.Timestamp).Format(time.RFC3339Nano)
			fmt.Printf("%v|%v|%v|%v\n", ts, d.Symbol, d.Price, d.Volume)
		}
	}
}
