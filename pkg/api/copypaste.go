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
	"errors"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/parsing"
	"github.com/zenazn/goji/web"
)

func getFormData(r *http.Request, fieldName string, results interface{}) (string, error) {
	contentType := r.Header.Get("Content-Type")
	contentType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", errors.New("Bad request content type")
	}
	var pasteContents string
	switch contentType {
	case "application/json":
		query := parseItemsQuery{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", errors.New("Unable to parse JSON")
		}
		json.Unmarshal(body, &query)
		pasteContents = query.Paste
	case "multipart/form-data":
		return "", errors.New("This endpoint does not support multipart forms.")
	default:
		err := r.ParseForm()
		if err != nil {
			return "", errors.New("Unable to parse form")
		}
		pasteValues := r.PostForm[fieldName]
		if len(pasteValues) == 0 {
			return "", errors.New("No request data")
		}
		pasteContents = pasteValues[0]
	}
	return pasteContents, nil
}

type parseItemsQuery struct {
	Paste string `json:"paste"`
}

// ParseItems returns a handler function that converts pasted
// item lists of any description into well-formed JSON.
// The paste should be passed as the form field "paste".
func ParseItems(db evego.Database) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		pasteContents, err := getFormData(r, "paste", &parseItemsQuery{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		items := parsing.ParseInventory(pasteContents, db)
		itemsJSON, _ := json.Marshal(items)
		w.Write(itemsJSON)
	}
}
