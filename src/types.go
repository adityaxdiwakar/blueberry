package main

import "time"

type priceWithSize struct {
	Price float64 `json:"price"`
	Size  int     `json:"size"`
}

type responseObjectStruct []struct {
	E string `json:"e"`
	D struct {
		Quotes []struct {
			ID         int       `json:"id"`
			Timestamp  time.Time `json:"timestamp"`
			ContractID int       `json:"contractId"`
			Entries    struct {
				Bid              priceWithSize `json:"Bid"`
				Offer            priceWithSize `json:"Offer"`
				Trade            priceWithSize `json:"Trade"`
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

type sessionObject struct {
	OpeningPrice    float64 `json:"open"`
	HighPrice       float64 `json:"high"`
	SettlementPrice float64 `json:"settlement"`
	LowPrice        float64 `json:"low"`
}

type depthObject struct {
	Bid priceWithSize `json:"bid"`
	Ask priceWithSize `json:"ask"`
}

type sqlQuoteObject struct {
	ID              int     `json:"id"`
	ContractID      int     `json:"contract_id"`
	SessionVolume   int     `json:"session_volume"`
	OpenInterest    int     `json:"open_interest"`
	OpeningPrice    float64 `json:"opening_price"`
	HighPrice       float64 `json:"high_price"`
	SettlementPrice float64 `json:"settlement_price"`
	LowPrice        float64 `json:"low_price"`
	BidPrice        float64 `json:"bid_price"`
	BidSize         int     `json:"bid_size"`
	AskPrice        float64 `json:"ask_price"`
	AskSize         int     `json:"ask_size"`
	TradePrice      float64 `json:"trade_price"`
	TradeSize       int     `json:"trade_size"`
	Timestamp       int64   `json:"timestamp"`
}

type tradovateCredentials struct {
	Username string `json:"name"`
	Password string `json:"password"`
}

type tradovateKeyResponse struct {
	AccessToken    string    `json:"accessToken"`
	MdAccessToken  string    `json:"mdAccessToken"`
	ExpirationTime time.Time `json:"expirationTime"`
	UserStatus     string    `json:"userStatus"`
	UserID         int       `json:"userId"`
	Name           string    `json:"name"`
	HasLive        bool      `json:"hasLive"`
	OutdatedTaC    bool      `json:"outdatedTaC"`
	HasFunded      bool      `json:"hasFunded"`
	HasMarketData  bool      `json:"hasMarketData"`
}

type httpTextResponse struct {
	Code    int    `json:"code"`
	Message string `json:"reason"`
}

type quoteObjectStruct struct {
	Timestamp     int64         `json:"timestamp"`
	ContractID    int           `json:"contract_id"`
	SessionVolume int           `json:"session_volume"`
	OpenInterest  int           `json:"open_interest"`
	SessionPrices sessionObject `json:"session_prices"`
	Depth         depthObject   `json:"depth"`
	Trade         priceWithSize `json:"trade"`
}

type payloadResponse struct {
	Code    int               `json:"code"`
	Payload quoteObjectStruct `json:"payload"`
}
