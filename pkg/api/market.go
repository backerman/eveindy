/*
Copyright © 2014–5 Brad Ackerman.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/backerman/evego"
	"github.com/zenazn/goji/web"
)

type queryItem struct {
	Quantity int    `json:"quantity"`
	ItemName string `json:"itemName"`
}

// priceFloat overrides the marshalling of our prices to JSON,
// ensuring we get 00000000000.00 rather than 0000e+00.
type priceFloat float64

func (p priceFloat) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", p)), nil
}

type responseItem struct {
	// The EVE database ID of the item.
	ItemID int `json:"itemID"`
	// The name of the item.
	ItemName string `json:"itemName"`
	// The quantity of that item available for purchase in
	// the returned location.
	QuantityAvailable int `json:"quantityAvailable"`
	// The best buy price per unit found in the system/region.
	BestBuy priceFloat `json:"bestBuy"`
	// The best sell price per unit found in the system/region.
	BestSell priceFloat `json:"bestSell"`
	// Details of the best buy order
	BuyInfo orderInfo `json:"buyInfo"`
	// Details of the best sell order
	SellInfo orderInfo `json:"sellInfo"`
}

type orderInfo struct {
	// The quantity available to buy/sell on this order
	Quantity int `json:"quantity"`
	// This order's minimum quantity to buy (1 for sell)
	MinQuantity int `json:"minQuantity"`
	// Where the order is available
	Station station `json:"station"`
	// One of: station, system, region, or a number of jumps
	Within string `json:"within"`
}

func makeRangeString(order *evego.Order) string {
	var rangeStr string
	switch order.JumpRange {
	case evego.BuyRegion:
		rangeStr = "region"
	case evego.BuySystem:
		rangeStr = "system"
	case evego.BuyStation:
		rangeStr = "station"
	default:
		// fix upstream
		if order.NumJumps > 100 {
			rangeStr = "region"
		} else {
			rangeStr = fmt.Sprintf("%d jumps", order.NumJumps)
		}
	}
	return rangeStr
}

func makeOrderInfo(db evego.Database, order *evego.Order) orderInfo {
	return orderInfo{
		Quantity:    order.Quantity,
		MinQuantity: order.MinQuantity,
		// Cheat — it's an outpost if it has no reprocessing efficiency reported.
		Station: stationFromAPI(db, order.Station,
			order.Station.ReprocessingEfficiency == 0.0),
		Within: makeRangeString(order),
	}
}

// summarizeOrders takes as input the orders for a given type in a region/station
// and returns the corresponding responseItem struct filled out.
func summarizeOrders(db evego.Database, orders []evego.Order, dbItem *evego.Item) responseItem {
	var (
		quantity          int
		bestBuy, bestSell float64
		buyInfo, sellInfo orderInfo
	)

	for _, ord := range orders {
		if ord.Type == evego.Buy && ord.Price > bestBuy {
			bestBuy = ord.Price
			buyInfo = makeOrderInfo(db, &ord)
		} else if ord.Type == evego.Sell && (bestSell == 0 || ord.Price < bestSell) {
			bestSell = ord.Price
			sellInfo = makeOrderInfo(db, &ord)
		}
		if ord.Type == evego.Sell {
			quantity += ord.Quantity
		}
	}
	result := responseItem{
		ItemID:            dbItem.ID,
		ItemName:          dbItem.Name,
		QuantityAvailable: quantity,
		BestBuy:           priceFloat(bestBuy),
		BuyInfo:           buyInfo,
		BestSell:          priceFloat(bestSell),
		SellInfo:          sellInfo,
	}
	return result
}

func getItemPrices(
	db evego.Database,
	mkt evego.Market,
	req *[]queryItem,
	station *evego.Station,
	loc string) (*map[string]responseItem, error) {
	respItems := make(map[string]responseItem)
	for _, i := range *req {
		dbItem, err := db.ItemForName(i.ItemName)
		if err != nil {
			continue
		}
		var (
			item   responseItem
			orders *[]evego.Order
		)
		if station != nil {
			orders, err = mkt.OrdersInStation(dbItem, station)
		} else {
			orders, err = mkt.OrdersForItem(dbItem, loc, evego.AllOrders)
		}
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve order information for %v: %v", dbItem.Name, err)
		}
		item = summarizeOrders(db, *orders, dbItem)
		respItems[item.ItemName] = item
	}
	return &respItems, nil
}

// ItemsMarketValue returns a handler that takes as input a JSON
// array of items and their quantities, plus a specified station
// or region, and computes the items' value.
//
// FIXME: Should return all buy orders within range for the queried system.
func ItemsMarketValue(db evego.Database, mkt evego.Market, xmlAPI evego.XMLAPI) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		contentType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Bad request content type"}`,
				http.StatusBadRequest)
			return
		}
		if contentType != "application/json" {
			http.Error(w, `{"status": "Error", "error": "Request must be of type application/json"}`,
				http.StatusUnsupportedMediaType)
			return
		}
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to process request body"}`,
				http.StatusBadRequest)
			return
		}
		var req []queryItem
		err = json.Unmarshal(reqBody, &req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to process request JSON: %v", err), http.StatusBadRequest)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		loc := c.URLParams["location"]
		stationIDStr, isStation := c.URLParams["id"]
		var station *evego.Station
		if isStation {
			// Get station / outpost object.
			stationID, _ := strconv.Atoi(stationIDStr)
			station, err = db.StationForID(stationID)
			if err != nil {
				// Not a station; should be an outpost.
				station, err = xmlAPI.OutpostForID(stationID)
				if err != nil {
					http.Error(w, `{"status": "Error", "error": "Unable to identify location"}`,
						http.StatusBadRequest)
					return
				}
			}
		}
		respItems, err := getItemPrices(db, mkt, &req, station, loc)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to retrieve order information"}`,
				http.StatusBadRequest)
			return
		}
		respJSON, _ := json.Marshal(respItems)
		w.Write(respJSON)
	}
}

// Go only provides a float min, so do our own for ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const reprocessCacheKey = "reprocess-output-values"
const reprocessCacheTTL = time.Hour

// ReprocessOutputValues returns a web handler function that generates a list of
// possible output from reprocessing, along with the Jita sell and buy price of each.
func ReprocessOutputValues(db evego.Database, mkt evego.Market, xmlAPI evego.XMLAPI, cache evego.Cache) web.HandlerFunc {
	tickerboard := []queryItem{
		{Quantity: 1, ItemName: "Tritanium"},
		{Quantity: 1, ItemName: "Pyerite"},
		{Quantity: 1, ItemName: "Mexallon"},
		{Quantity: 1, ItemName: "Isogen"},
		{Quantity: 1, ItemName: "Nocxium"},
		{Quantity: 1, ItemName: "Megacyte"},
		{Quantity: 1, ItemName: "Zydrine"},
		{Quantity: 1, ItemName: "Megacyte"},
	}
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		// Check cached and use that instead if it's available.
		cached, found := cache.Get(reprocessCacheKey)
		if found {
			w.Write(cached)
			return
		}
		jita, err := db.StationForID(60003760) // Jita IV - Moon 4 - Caldari Navy Assembly Plant
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to retrieve ticker prices"}`,
				http.StatusInternalServerError)
			log.Printf("Error looking up Jita station info (???): %v", err)
			return
		}
		results, err := getItemPrices(db, mkt, &tickerboard, jita, "")
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to retrieve ticker prices"}`,
				http.StatusInternalServerError)
			log.Printf("Error getting Jita item prices: %v", err)
			return
		}

		resultsJSON, err := json.Marshal(results)
		// Write output to cache as well.
		cache.Put(reprocessCacheKey, resultsJSON, time.Now().Add(reprocessCacheTTL))
		w.Write(resultsJSON)
	}
}
