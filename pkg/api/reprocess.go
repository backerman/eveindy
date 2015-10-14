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
	"io/ioutil"
	"log"
	"math"
	"mime"
	"net/http"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/industry"
	"github.com/zenazn/goji/web"
)

type reproItem struct {
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
}

type reproQuery struct {
	StationYield    float64     `json:"stationYield"`
	TaxRate         float64     `json:"taxRate"`
	ScrapmetalSkill int         `json:"scrapmetalReprocessingSkill"`
	Items           []reproItem `json:"items"`
}

type reproResults struct {
	Items  map[string][]evego.InventoryLine `json:"items"`
	Prices map[string]responseItem          `json:"prices"`
}

// ReprocessItems returns a handler function that takes as input an item list
// and returns the reprocessing output of each inventory line.
func ReprocessItems(db evego.Database, mkt evego.Market) web.HandlerFunc {
	jita, err := db.StationForID(60003760) // Jita IV - Moon 4 - Caldari Navy Assembly Plant
	if err != nil {
		log.Fatalf("Seriously, guys, something's gone wrong with the database!")
	}
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		contentType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, "Bad request content type", http.StatusBadRequest)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		if contentType != "application/json" {
			http.Error(w, "Request must be of type application/json", http.StatusUnsupportedMediaType)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to process request body", http.StatusBadRequest)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		var req reproQuery
		err = json.Unmarshal(reqBody, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}

		// Convert 0..100 scale to 0..1.
		stationYield := math.Max(math.Min(1, req.StationYield*0.01), 0)
		taxRate := math.Max(math.Min(1, req.TaxRate*0.01), 0)
		reproSkills := industry.ReproSkills{
			ScrapmetalProcessing: req.ScrapmetalSkill,
		}
		results := make(map[string][]evego.InventoryLine)
		for _, i := range req.Items {
			item, err := db.ItemForName(i.ItemName)
			if err != nil {
				continue
			}
			itemResults, err := industry.ReprocessItem(db, item, i.Quantity, stationYield, taxRate, reproSkills)
			if err != nil {
				http.Error(w, "Unable to compute reprocessing output", http.StatusInternalServerError)
				w.Write([]byte(`{"status": "Error"}`))
				return
			}
			results[item.Name] = itemResults
		}
		prices := make(map[string]responseItem)

		// Loop over each item that was reprocessed.
		for _, itemOut := range results {
			// For each item in its component materials,
			for _, item := range itemOut {
				// Check if we already know its price in Jita
				itemName := item.Item.Name
				_, found := prices[itemName]
				if !found {
					// If not there, get its price.
					myPrices, err := getItemPrices(db, mkt, &[]queryItem{{Quantity: 1, ItemName: itemName}}, jita, "")
					if err != nil {
						http.Error(w, `{"status": "Error", "error": "Unable to look up prices (reprocessing)"}`,
							http.StatusInternalServerError)
						return
					}
					prices[itemName] = (*myPrices)[itemName]
				}
			}
		}

		response := reproResults{
			Items:  results,
			Prices: prices,
		}
		resultsJSON, _ := json.Marshal(response)
		w.Write(resultsJSON)
	}
}
