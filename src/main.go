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
	"strings"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"database/sql"
	 _ "github.com/go-sql-driver/mysql"
	 
	"net/http"

	"github.com/gorilla/mux"
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

type SQLQuoteObject struct {
	ID int `json:"id"`
	ContractID int `json:"contract_id"`
	SessionVolume int `json:"session_volume"`
	OpenInterest int `json:"open_interest"`
	OpeningPrice float64 `json:"opening_price"`
	HighPrice float64 `json""high_price"`
	SettlementPrice float64 `json:"settlement_price"`
	LowPrice float64 `json:"low_price"`
	BidPrice float64 `json:"bid_price"`
	BidSize int `json:"bid_size"`
	AskPrice float64 `json:"ask_price"`
	AskSize int `json:"ask_size"`
	TradePrice float64 `json:"trade_price"`
	TradeSize int `json:"trade_size"`
	Timestamp int64 `json:"timestamp"`
}


var addr = flag.String("addr", "md-api.tradovate.com", "http service address")
var quoteObject QuoteObject;

func initCache(db *sql.DB) {
	results, err := db.Query("SELECT * FROM quotes ORDER BY id DESC LIMIT 1;")
	if err != nil {
		log.Fatal("Error ocurred! Could not populate cache!")
	}
	
	for results.Next() {
		cacheObject := SQLQuoteObject{}
		err = results.Scan(&cacheObject.ID, &cacheObject.ContractID, &cacheObject.SessionVolume, &cacheObject.OpenInterest, &cacheObject.OpeningPrice, &cacheObject.HighPrice, &cacheObject.SettlementPrice, &cacheObject.LowPrice,
						   &cacheObject.BidPrice, &cacheObject.BidSize, &cacheObject.AskPrice, &cacheObject.AskSize, &cacheObject.TradePrice, &cacheObject.TradeSize, &cacheObject.Timestamp)
		if err != nil {
			log.Fatal("Error occured population object", err)
		}

		quoteObject = QuoteObject{
			Timestamp:     cacheObject.Timestamp,
			ContractID:    cacheObject.ContractID,
			SessionVolume: cacheObject.SessionVolume,
			OpenInterest:  cacheObject.OpenInterest,
			SessionPrices: SessionObject{
				OpeningPrice:    cacheObject.OpeningPrice,
				HighPrice:       cacheObject.HighPrice,
				SettlementPrice: cacheObject.SettlementPrice,
				LowPrice:        cacheObject.LowPrice,
			},
			Depth: DepthObject{
				Bid: PriceWithSize{
					Price: cacheObject.BidPrice,
					Size: cacheObject.BidSize,
				},
				Ask: PriceWithSize{
					Price: cacheObject.AskPrice,
					Size: cacheObject.AskSize,
				},
			},
			Trade: PriceWithSize{
				Price: cacheObject.TradePrice,
				Size: cacheObject.TradeSize,
			},
		}
	}

}

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

	initCache(db)

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/v1/websocket?r=0.8840574374908023"}
	log.Printf("Connecting to Tradovate Market Data Socket")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	} else {
		log.Printf("Established connection with WS")
	}
	defer c.Close()

	//establish a connection
	accessToken := os.Getenv("ACCESS_TOKEN")
	c.WriteMessage(websocket.TextMessage, []byte("authorize\n2\n\n" + accessToken))

	done := make(chan struct{})

	quoteObject := QuoteObject{}
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Fatal("read:", err)
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
				quoteObject = QuoteObject{
					Timestamp:     compressedResponse.Timestamp.UnixNano()/(1000*1000),
					ContractID:    compressedResponse.ContractID,
					SessionVolume: compressedResponse.Entries.TotalTradeVolume.Size,
					OpenInterest:  compressedResponse.Entries.OpenInterest.Size,
					SessionPrices: SessionObject{
						OpeningPrice:    compressedResponse.Entries.OpeningPrice.Price,
						HighPrice:       compressedResponse.Entries.HighPrice.Price,
						SettlementPrice: compressedResponse.Entries.SettlementPrice.Price,
						LowPrice:        compressedResponse.Entries.LowPrice.Price,
					},
					Depth: DepthObject{
						Bid: compressedResponse.Entries.Bid,
						Ask: compressedResponse.Entries.Offer,
					},
					Trade: compressedResponse.Entries.Trade,
				}
				statement := fmt.Sprintf("insert into quotes (contract_id, session_volume, open_interest, opening_price, high_price, settlement_price, low_price, bid_price, bid_size, ask_price, ask_size, trade_price, trade_size, timestamp) VALUES (%d, %d, %d, %f, %f, %f, %f, %f, %d, %f, %d, %f, %d, %d)", 
										compressedResponse.ContractID, compressedResponse.Entries.TotalTradeVolume.Size, compressedResponse.Entries.OpenInterest.Size, compressedResponse.Entries.OpeningPrice.Price, compressedResponse.Entries.HighPrice.Price, compressedResponse.Entries.SettlementPrice.Price,
										compressedResponse.Entries.LowPrice.Price, compressedResponse.Entries.Bid.Price, compressedResponse.Entries.Bid.Size, compressedResponse.Entries.Offer.Price, compressedResponse.Entries.Offer.Size, compressedResponse.Entries.Trade.Price, compressedResponse.Entries.Trade.Size,
										compressedResponse.Timestamp.UnixNano()/(1000*1000))
				quoteIn, err := db.Prepare(statement)
				if err != nil {
					log.Println("Something went wrong preparing a SQL Insertion")
				}
				defer quoteIn.Close()

				_, err = quoteIn.Exec()
				if err != nil {
					log.Println("Something went wrong executing a SQL Insertion")
				}
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
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

	
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", rootPage)
	router.HandleFunc("/recent/", recentQuote)
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(":6009", router))
}

type HTTPTextResponse struct {
	Code int `json:"code"`
	Message string `json:"reason"`
}

type QuoteObject struct {
	Timestamp int64 `json:"timestamp"`
	ContractID int `json:"contract_id"`
	SessionVolume int `json:"session_volume"`
	OpenInterest int `json:"open_interest"`
	SessionPrices SessionObject `json:"session_prices"`
	Depth DepthObject `json:"depth"`
	Trade PriceWithSize `json:"trade"`
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Do stuff here
        log.Println("HTTP Request Received on", r.RequestURI)
        // Call the next handler, which can be another middleware in the chain, or the final handler.
        next.ServeHTTP(w, r)
    })
}

func rootPage(w http.ResponseWriter, r *http.Request) {
	response := HTTPTextResponse{
		Code: 200,
		Message: "You have reached the market data provisioner root endpoint, please see the documentation to access public facing endpoints.",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type PayloadResponse struct {
	Code int `json:"code"`
	Payload QuoteObject `json:"payload"`
}

func recentQuote(w http.ResponseWriter, r *http.Request) {
	response := PayloadResponse{
		Code: 200,
		Payload: quoteObject,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}