package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"net/http"

	"github.com/gorilla/mux"
)

func getAccessToken(username string, password string) string {
	payload := tradovateCredentials{
		Username: username,
		Password: password,
	}
	payloadByteArray, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("Something went wrong when loading the access token!")
	}
	req, err := http.NewRequest("POST", "https://live-api.tradovate.com/v1/auth/accesstokenrequest", bytes.NewBuffer(payloadByteArray))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	respObj := tradovateKeyResponse{}
	json.Unmarshal([]byte(string(body)), &respObj)
	log.Printf("Retrieved Access Token with Credentials")
	return respObj.MdAccessToken
}

var addr = flag.String("addr", "md-api.tradovate.com", "http service address")
var quoteObject quoteObjectStruct

var db *sql.DB

func initCache() {
	err := db.Ping()
	if err != nil {
		panic(err.Error())
	} else {
		log.Printf("Connected to DB Successfully!")
	}
	results, err := db.Query("SELECT * FROM quotes ORDER BY id DESC LIMIT 1;")
	if err != nil {
		log.Fatal("Error ocurred! Could not populate cache!")
	}

	for results.Next() {
		cacheObject := sqlQuoteObject{}
		err = results.Scan(&cacheObject.ID, &cacheObject.ContractID, &cacheObject.SessionVolume, &cacheObject.OpenInterest, &cacheObject.OpeningPrice, &cacheObject.HighPrice, &cacheObject.SettlementPrice, &cacheObject.LowPrice,
			&cacheObject.BidPrice, &cacheObject.BidSize, &cacheObject.AskPrice, &cacheObject.AskSize, &cacheObject.TradePrice, &cacheObject.TradeSize, &cacheObject.Timestamp)
		if err != nil {
			log.Fatal("Error occured population object", err)
		}

		quoteObject = quoteObjectStruct{
			Timestamp:     cacheObject.Timestamp,
			ContractID:    cacheObject.ContractID,
			SessionVolume: cacheObject.SessionVolume,
			OpenInterest:  cacheObject.OpenInterest,
			SessionPrices: sessionObject{
				OpeningPrice:    cacheObject.OpeningPrice,
				HighPrice:       cacheObject.HighPrice,
				SettlementPrice: cacheObject.SettlementPrice,
				LowPrice:        cacheObject.LowPrice,
			},
			Depth: depthObject{
				Bid: priceWithSize{
					Price: cacheObject.BidPrice,
					Size:  cacheObject.BidSize,
				},
				Ask: priceWithSize{
					Price: cacheObject.AskPrice,
					Size:  cacheObject.AskSize,
				},
			},
			Trade: priceWithSize{
				Price: cacheObject.TradePrice,
				Size:  cacheObject.TradeSize,
			},
		}
	}
}

func main() {
	var err error
	db, err = sql.Open("mysql", "blueberry:password@/md")
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
	} else {
		log.Printf("Loaded environment variables successfully")
	}

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/v1/websocket?r=0.8840574374908023"}
	log.Printf("Connecting to Tradovate Market Data Socket")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	} else {
		log.Printf("Received 'Hello' from WS")
	}
	defer c.Close()

	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	accessToken := getAccessToken(username, password)

	//establish a connection
	c.WriteMessage(websocket.TextMessage, []byte("authorize\n2\n\n"+accessToken))

	done := make(chan struct{})

	go func() {
		defer close(done)
		count := 0
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Fatal("read:", err)
				return
			}
			count++
			if count == 2 {
				c.WriteMessage(websocket.TextMessage, []byte("md/subscribequote\n7\n\n{\"symbol\":1717695}"))
			}

			rMessage := string(message)
			switch {
			case strings.Contains(rMessage, "quotes"):
				log.Printf("Quote received")
				rMessage = rMessage[1:]
				data := responseObjectStruct{}
				json.Unmarshal([]byte(rMessage), &data)
				compressedResponse := data[0].D.Quotes[0]
				quoteObject = quoteObjectStruct{
					Timestamp:     compressedResponse.Timestamp.UnixNano() / (1000 * 1000),
					ContractID:    compressedResponse.ContractID,
					SessionVolume: compressedResponse.Entries.TotalTradeVolume.Size,
					OpenInterest:  compressedResponse.Entries.OpenInterest.Size,
					SessionPrices: sessionObject{
						OpeningPrice:    compressedResponse.Entries.OpeningPrice.Price,
						HighPrice:       compressedResponse.Entries.HighPrice.Price,
						SettlementPrice: compressedResponse.Entries.SettlementPrice.Price,
						LowPrice:        compressedResponse.Entries.LowPrice.Price,
					},
					Depth: depthObject{
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println("HTTP Request Received on", r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func rootPage(w http.ResponseWriter, r *http.Request) {
	response := httpTextResponse{
		Code:    200,
		Message: "You have reached the market data provisioner root endpoint, please see the documentation to access public facing endpoints.",
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func recentQuote(w http.ResponseWriter, r *http.Request) {
	response := payloadResponse{
		Code:    200,
		Payload: quoteObject,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
