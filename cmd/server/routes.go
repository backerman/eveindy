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
	"github.com/zenazn/goji/web"
)

func setRoutes(mux *web.Mux, sde evego.Database, localdb db.LocalDB, xmlAPI evego.XMLAPI,
	eveCentral evego.Market, sessionizer server.Sessionizer, cache evego.Cache) {
	assets := http.FileServer(http.Dir("dist"))
	bower := http.FileServer(http.Dir("bower_components"))
	if c.Dev {
		mux.Get("/bower_components/*", http.StripPrefix("/bower_components/", bower))
	}
	mux.Get("/autocomplete/system/:name", api.AutocompleteSystems(sde))
	mux.Get("/autocomplete/station/:name", api.AutocompleteStations(sde, localdb, xmlAPI))
	mux.Post("/pastebin", api.ParseItems(sde))
	marketHandler := api.ItemsMarketValue(sde, eveCentral, xmlAPI)
	// For now these do the same thing. That may change.
	mux.Post("/market/region/:location", marketHandler)
	mux.Post("/market/system/:location", marketHandler)
	mux.Post("/market/station/:id", marketHandler)
	mux.Get("/market/jita", api.ReprocessOutputValues(sde, eveCentral, xmlAPI, cache))

	mux.Post("/reprocess", api.ReprocessItems(sde, eveCentral))
	// SSO!
	auth := evesso.MakeAuthenticator(evesso.Endpoint, c.ClientID, c.ClientSecret,
		c.RedirectURL, evesso.PublicData)
	mux.Get("/crestcallback", api.CRESTCallbackListener(localdb, auth, sessionizer))
	mux.Get("/authenticate", api.AuthenticateHandler(auth, sessionizer))
	mux.Get("/session", api.SessionInfo(auth, sessionizer, localdb))
	mux.Post("/logout", api.LogoutHandler(localdb, auth, sessionizer))

	// API keys
	listHandler, deleteHander, addHandler, refreshHandler := api.XMLAPIKeysHandlers(localdb, sessionizer)
	mux.Get("/apikeys/list", listHandler)
	mux.Post("/apikeys/delete/:keyid", deleteHander)
	mux.Post("/apikeys/add", addHandler)
	mux.Post("/apikeys/refresh", refreshHandler)

	// Standings and skills
	mux.Get("/standings/:charID/:npcCorpID", api.StandingsHandler(localdb, sessionizer))
	mux.Get("/skills/:charID/group/:skillGroupID", api.SkillsHandler(localdb, sessionizer))

	// Blueprints and industry
	_, getBPs := api.BlueprintsHandlers(localdb, sde, sessionizer)
	mux.Get("/blueprints/:charID", getBPs)
	mux.Get("/assets/unusedSalvage/:charID", api.UnusedSalvage(localdb, sde, sessionizer))

	// Static assets
	mux.Get("/*", assets)
}
