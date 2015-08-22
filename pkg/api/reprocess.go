/*
Copyright © 2014 Brad Ackerman.

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

// ReprocessItems returns a handler function that takes as input an item list
// and returns the reprocessing output of each inventory line.
func ReprocessItems(db evego.Database) web.HandlerFunc {
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
		resultsJSON, _ := json.Marshal(results)
		w.Write(resultsJSON)
	}
}
