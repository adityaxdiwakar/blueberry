package main

import (
	"fmt"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"sync"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"database/sql"
 	_ "github.com/go-sql-driver/mysql"
)

type PriceWithSize struct {
	Price float64 `json:"price"`
	Size int `json:"size"`
}

type ResponseObject []struct {
	E string `json:"e"`
	D struct {
		Quotes []struct {
			ID         int       `json:"id"`
			Timestamp  time.Time `json:"timestamp"`
			ContractID int       `json:"contractId"`
			Entries    struct {
				Bid PriceWithSize `json:"Bid"`
				Offer PriceWithSize `json:"Offer"`
				Trade PriceWithSize `json:"Trade"`
				TotalTradeVolume struct {
					Size int `json:"size"`
				} `json:"TotalTradeVolume"`
				LowPrice struct {
					Price float64 `json:"price"`
				} `json:"LowPrice"`
				OpenInterest struct {
					Size int `json:"size"`
				} `json:"OpenInterest"`

				OpeningPrice struct {
					Price float64 `json:"price"`
				} `json:"OpeningPrice"`
				HighPrice struct {
					Price float64 `json:"price"`
				} `json:"HighPrice"`
				SettlementPrice struct {
					Price float64 `json:"price"`
				} `json:"SettlementPrice"`
			} `json:"entries"`
		} `json:"quotes"`
	} `json:"d"`
}

type SessionObject struct {
	OpeningPrice float64 `json:"open"`
	HighPrice float64 `json:"high"`
	SettlementPrice float64 `json:"settlement"`
	LowPrice float64 `json:"low"`
}  

type DepthObject struct {
	Bid PriceWithSize `json:"bid"`
	Ask PriceWithSize `json:"ask"`
}

type QuoteObject struct {
	Timestamp time.Time `json:"timestamp"`
	ContractID int `json:"contract_id"`
	SessionVolume int `json:"session_volume"`
	OpenInterest int `json:"open_interest"`
	SessionPrices SessionObject `json:"session_prices"`
	Depth DepthObject `json:"depth"`
	Trade PriceWithSize `json:"trade"`
}


var addr = flag.String("addr", "md-api.tradovate.com", "http service address")
var wg sync.WaitGroup

func main() {
	db, err := sql.Open("mysql", "blueberry:password@/md")
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	} else {
		log.Printf("Connected to DB Successfully!")
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loaded environment configuration")
	}

	// file, err := os.Open("quote-stream.txt")
 
	// if err != nil {
	// 	log.Fatalf("failed opening file: %s", err)
	// }

	// scanner := bufio.NewScanner(file)
	// scanner.Split(bufio.ScanLines)
	// var txtlines []string

	// for scanner.Scan() {
	// 	txtlines = append(txtlines, scanner.Text())
	// }
 
	// file.Close()
 
	// count := 0
	// for _, eachline := range txtlines {
	// 	count++
	// 	eachline = eachline[1:]
	// 	data := ResponseObject{}
	// 	json.Unmarshal([]byte(eachline), &data)
	// 	compressedResponse := data[0].D.Quotes[0]

	// 	statement := fmt.Sprintf("insert into quotes (contract_id, session_volume, open_interest, opening_price, high_price, settlement_price, low_price, bid_price, bid_size, ask_price, ask_size, trade_price, trade_size, timestamp) VALUES (%d, %d, %d, %f, %f, %f, %f, %f, %d, %f, %d, %f, %d, %d)", 
	// 	compressedResponse.ContractID, compressedResponse.Entries.TotalTradeVolume.Size, compressedResponse.Entries.OpenInterest.Size, compressedResponse.Entries.OpeningPrice.Price, compressedResponse.Entries.HighPrice.Price, compressedResponse.Entries.SettlementPrice.Price,
	// 	compressedResponse.Entries.LowPrice.Price, compressedResponse.Entries.Bid.Price, compressedResponse.Entries.Bid.Size, compressedResponse.Entries.Offer.Price, compressedResponse.Entries.Offer.Size, compressedResponse.Entries.Trade.Price, compressedResponse.Entries.Trade.Size,
	// 	compressedResponse.Timestamp.Unix())
	// 	quoteIn, err := db.Prepare(statement)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	_, err = quoteIn.Exec()
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	quoteIn.Close()

	// 	if count % 100 == 0 {
	// 		fmt.Println(compressedResponse.Timestamp)
	// 	}
	// }

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/v1/websocket?r=0.8840574374908023"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	//establish a connection
	accessToken := os.Getenv("ACCESS_TOKEN")
	c.WriteMessage(websocket.TextMessage, []byte("authorize\n2\n\n" + accessToken))

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}			
	
			r_message := string(message)
			switch {
			case strings.Contains(r_message, "quotes"):
				log.Printf("Quote received")
				r_message = r_message[1:]
				data := ResponseObject{}
				json.Unmarshal([]byte(r_message), &data)
				compressedResponse := data[0].D.Quotes[0]
				// quoteObject := QuoteObject{
				// 	Timestamp:     compressedResponse.Timestamp,
				// 	ContractID:    compressedResponse.ContractID,
				// 	SessionVolume: compressedResponse.Entries.TotalTradeVolume.Size,
				// 	OpenInterest:  compressedResponse.Entries.OpenInterest.Size,
				// 	SessionPrices: SessionObject{
				// 		OpeningPrice:    compressedResponse.Entries.OpeningPrice.Price,
				// 		HighPrice:       compressedResponse.Entries.HighPrice.Price,
				// 		SettlementPrice: compressedResponse.Entries.SettlementPrice.Price,
				// 		LowPrice:        compressedResponse.Entries.LowPrice.Price,
				// 	},
				// 	Depth: DepthObject{
				// 		Bid: compressedResponse.Entries.Bid,
				// 		Ask: compressedResponse.Entries.Offer,
				// 	},
				// 	Trade: compressedResponse.Entries.Trade,
				// }
				statement := fmt.Sprintf("insert into quotes (contract_id, session_volume, open_interest, opening_price, high_price, settlement_price, low_price, bid_price, bid_size, ask_price, ask_size, trade_price, trade_size, timestamp) VALUES (%d, %d, %d, %f, %f, %f, %f, %f, %d, %f, %d, %f, %d, %d)", 
										compressedResponse.ContractID, compressedResponse.Entries.TotalTradeVolume.Size, compressedResponse.Entries.OpenInterest.Size, compressedResponse.Entries.OpeningPrice.Price, compressedResponse.Entries.HighPrice.Price, compressedResponse.Entries.SettlementPrice.Price,
										compressedResponse.Entries.LowPrice.Price, compressedResponse.Entries.Bid.Price, compressedResponse.Entries.Bid.Size, compressedResponse.Entries.Offer.Price, compressedResponse.Entries.Offer.Size, compressedResponse.Entries.Trade.Price, compressedResponse.Entries.Trade.Size,
										compressedResponse.Timestamp.Unix())
				statement = strings.ReplaceAll(statement, " +0000 UTC", "")
				log.Printf(statement)
				quoteIn, err := db.Prepare(statement)
				if err != nil {
					panic(err.Error())
				}
				defer quoteIn.Close()

				_, err = quoteIn.Exec()
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	wg.Add(1)
	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, []byte("[]"))
				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-interrupt:
				log.Println("interrupt")
	
				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}	
	}()

	wg.Wait()
}