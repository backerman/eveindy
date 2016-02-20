/*
Copyright © 2014–6 Brad Ackerman.

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

package main

import (
	"net/http"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/evesso"
	"github.com/backerman/eveindy/pkg/api"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji"
)

func setRoutes(sde evego.Database, localdb db.LocalDB, xmlAPI evego.XMLAPI,
	eveCentral evego.Market, sessionizer server.Sessionizer, cache evego.Cache) {
	assets := http.FileServer(http.Dir("dist"))
	bower := http.FileServer(http.Dir("bower_components"))
	if c.Dev {
		goji.Get("/bower_components/*", http.StripPrefix("/bower_components/", bower))
	}
	goji.Get("/autocomplete/system/:name", api.AutocompleteSystems(sde))
	goji.Get("/autocomplete/station/:name", api.AutocompleteStations(sde, localdb, xmlAPI))
	goji.Post("/pastebin", api.ParseItems(sde))
	marketHandler := api.ItemsMarketValue(sde, eveCentral, xmlAPI)
	// For now these do the same thing. That may change.
	goji.Post("/market/region/:location", marketHandler)
	goji.Post("/market/system/:location", marketHandler)
	goji.Post("/market/station/:id", marketHandler)
	goji.Get("/market/jita", api.ReprocessOutputValues(sde, eveCentral, xmlAPI, cache))

	goji.Post("/reprocess", api.ReprocessItems(sde, eveCentral))
	// SSO!
	auth := evesso.MakeAuthenticator(evesso.Endpoint, c.ClientID, c.ClientSecret,
		c.RedirectURL, evesso.PublicData)
	goji.Get("/crestcallback", api.CRESTCallbackListener(localdb, auth, sessionizer))
	goji.Get("/authenticate", api.AuthenticateHandler(auth, sessionizer))
	goji.Get("/session", api.SessionInfo(auth, sessionizer, localdb))
	goji.Post("/logout", api.LogoutHandler(localdb, auth, sessionizer))

	// API keys
	listHandler, deleteHander, addHandler, refreshHandler := api.XMLAPIKeysHandlers(localdb, sessionizer)
	goji.Get("/apikeys/list", listHandler)
	goji.Post("/apikeys/delete/:keyid", deleteHander)
	goji.Post("/apikeys/add", addHandler)
	goji.Post("/apikeys/refresh", refreshHandler)

	// Standings and skills
	goji.Get("/standings/:charID/:npcCorpID", api.StandingsHandler(localdb, sessionizer))
	goji.Get("/skills/:charID/group/:skillGroupID", api.SkillsHandler(localdb, sessionizer))

	// Blueprints and industry
	_, getBPs := api.BlueprintsHandlers(localdb, sde, sessionizer)
	goji.Get("/blueprints/:charID", getBPs)
	goji.Get("/assets/unusedSalvage/:charID", api.UnusedSalvage(localdb, sde, sessionizer))

	// Static assets
	goji.Get("/*", assets)
}
