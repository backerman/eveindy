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

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/cache"
	"github.com/backerman/evego/pkg/dbaccess"
	"github.com/backerman/evego/pkg/eveapi"
	"github.com/backerman/evego/pkg/evesso"
	"github.com/backerman/evego/pkg/market"
	"github.com/backerman/evego/pkg/routing"
	"github.com/backerman/eveindy/pkg/api"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zenazn/goji"
	// Register SQLite3 and PgSQL drivers
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var c config
var rootCmd *cobra.Command

type config struct {
	Dev                      bool
	DbDriver, DbPath         string
	Bind                     string
	XMLAPIEndpoint           string
	Router                   string
	Cache                    string
	RedisHost, RedisPassword string
	CookieDomain, CookiePath string
	ClientID, ClientSecret   string
	RedirectURL              string
}

func setRoutes(sde evego.Database, localdb db.LocalDB, xmlAPI evego.XMLAPI,
	eveCentral evego.Market, sessionizer server.Sessionizer) {
	assets := http.FileServer(http.Dir("dist"))
	bower := http.FileServer(http.Dir("bower_components"))
	if c.Dev {
		goji.Get("/bower_components/*", http.StripPrefix("/bower_components/", bower))
	}
	goji.Get("/autocomplete/system/:name", api.AutocompleteSystems(sde))
	goji.Get("/autocomplete/station/:name", api.AutocompleteStations(sde, xmlAPI))
	goji.Post("/pastebin", api.ParseItems(sde))
	marketHandler := api.ItemsMarketValue(sde, eveCentral, xmlAPI)
	// For now these do the same thing. That may change.
	goji.Post("/market/region/:location", marketHandler)
	goji.Post("/market/system/:location", marketHandler)
	goji.Post("/market/station/:id", marketHandler)
	goji.Post("/reprocess", api.ReprocessItems(sde))
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

	// Static assets
	goji.Get("/*", assets)
}

func mainCommand(cmd *cobra.Command, args []string) {
	err := viper.Unmarshal(&c)
	if err != nil {
		log.Fatalf("Unable to marshal configuration: %v", err)
	}
	if !viper.IsSet("dbpath") {
		log.Fatalf("Please set the dbpath configuration option or EVEINDY_DBPATH " +
			"environment variable to the database's path.")
	}

	if !(viper.IsSet("CookieDomain") && viper.IsSet("CookiePath")) {
		log.Fatalf("Please set the CookieDomain and CookiePath configuration options.")
	}

	if !(viper.IsSet("ClientID") && viper.IsSet("ClientSecret") &&
		viper.IsSet("RedirectURL")) {
		log.Fatalf("Please set the ClientID, ClientSecret, and RedirectURL configuration " +
			"options as registered with CCP.")
	}

	if c.Dev {
		log.Printf("Configuration is: %+v", c)
	}

	sde := dbaccess.SQLDatabase(c.DbDriver, c.DbPath)
	// if c.Dev {
	// 	ts := httptest.NewServer(
	// 		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 			respFile, _ := os.Open("./conquerable-stations.xml")
	// 			defer respFile.Close()
	// 			responseBytes, _ := ioutil.ReadAll(respFile)
	// 			responseBuf := bytes.NewBuffer(responseBytes)
	// 			responseBuf.WriteTo(w)
	// 		}))
	// 	c.XMLAPIEndpoint = ts.URL
	// }
	var myCache evego.Cache
	switch c.Cache {
	case "inproc":
		myCache = server.InMemCache()
	case "redis":
		if c.RedisPassword != "" {
			myCache = cache.RedisCache(c.RedisHost, c.RedisPassword)
		} else {
			myCache = cache.RedisCache(c.RedisHost)
		}
	default:
		log.Fatalf(
			"The Cache configuration option must be set to \"inproc\" (default) or \"redis\".")
	}

	xmlAPI := eveapi.XML(c.XMLAPIEndpoint, sde, myCache)
	localdb, err := db.Interface(c.DbDriver, c.DbPath, xmlAPI)
	if err != nil {
		log.Fatalf("Unable to connect to local database: %v", err)
	}
	var router evego.Router

	switch c.Router {
	case "evecentral":
		router = routing.EveCentralRouter(
			"http://api.eve-central.com/api/route", myCache)
	case "sql":
		router = routing.SQLRouter(c.DbDriver, c.DbPath, myCache)
	default:
		log.Fatalf(
			"The Router configuration option must be set to \"evecentral\" (default) or \"sql\".")
	}

	eveCentralMarket := market.EveCentral(sde, router, xmlAPI,
		"http://api.eve-central.com/api/quicklook", myCache)

	sessionizer := server.GetSessionizer(c.CookieDomain, c.CookiePath, !c.Dev, localdb)

	setRoutes(sde, localdb, xmlAPI, eveCentralMarket, sessionizer)
	// We like magic, but fix the magic some.
	bindArg := fmt.Sprintf("-bind=%s", c.Bind)
	if len(os.Args) > 1 {
		os.Args[1] = bindArg
		os.Args = os.Args[:2]
	} else {
		os.Args = append(os.Args, bindArg)
	}

	goji.Serve()
}

func main() {
	rootCmd = &cobra.Command{
		Use:   "server",
		Short: "The server component of eveindy",
		Run:   mainCommand,
	}

	// Configuration defaults

	// When DevMode is true, server only listens on localhost and will serve
	// external dependencies (e.g. AngularJS) from local disk instead of from
	// a CDN.
	viper.SetDefault("DevMode", false)
	// DBDriver and DBPath are the database driver name and resource path
	// as used by the Golang SQL library.
	viper.SetDefault("DBDriver", "sqlite3")
	// DBPath has no default. You must set it.
	// viper.SetDefault("DBPath", "")
	viper.SetDefault("Bind", "*:8080")
	viper.SetDefault("XMLAPIEndpoint", "https://api.eveonline.com")
	// Routing
	// Router: either "evecentral" or "sql".
	viper.SetDefault("Router", "evecentral")

	// Session cookies - you must set these explicitly.
	// viper.SetDefault("CookieDomain", "localhost")
	// viper.SetDefault("CookiePath", "/")

	// Cache
	// The default is an in-process cache, but you should probably use Redis
	// istead.
	viper.SetDefault("Cache", "inproc")
	viper.SetDefault("RedisHost", ":6379")
	viper.SetDefault("RedisPassword", "")

	// Set configuration file
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/eveindy")
	viper.AddConfigPath("$HOME/.eveindy")
	viper.ReadInConfig()

	// Environment variables
	viper.SetEnvPrefix("EVEINDY")
	viper.AutomaticEnv()

	rootCmd.Flags().Bool("dev", false, "Set development mode.")
	rootCmd.Flags().String("bind", ":8080", "The address and port to listen on.")

	flags := []string{"dev", "bind"}
	for _, flag := range flags {
		viper.BindPFlag(flag, rootCmd.Flags().Lookup(flag))
	}
	rootCmd.Execute()
}
