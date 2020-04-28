package main

func cacheObjectTranslator(cacheObject sqlQuoteObject) quoteObjectStruct {
	return quoteObjectStruct{
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

func responsiveObjectTranslator(data responseObjectStruct) quoteObjectStruct {
	response := data[0].D.Quotes[0]
	return quoteObjectStruct{
		Timestamp:     response.Timestamp.UnixNano() / (1000 * 1000),
		ContractID:    response.ContractID,
		SessionVolume: response.Entries.TotalTradeVolume.Size,
		OpenInterest:  response.Entries.OpenInterest.Size,
		SessionPrices: sessionObject{
			OpeningPrice:    response.Entries.OpeningPrice.Price,
			HighPrice:       response.Entries.HighPrice.Price,
			SettlementPrice: response.Entries.SettlementPrice.Price,
			LowPrice:        response.Entries.LowPrice.Price,
		},
		Depth: depthObject{
			Bid: response.Entries.Bid,
			Ask: response.Entries.Offer,
		},
		Trade: response.Entries.Trade,
	}
}
