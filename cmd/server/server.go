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
	"net"
	"net/http"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/cache"
	"github.com/backerman/evego/pkg/dbaccess"
	"github.com/backerman/evego/pkg/eveapi"
	"github.com/backerman/evego/pkg/market"
	"github.com/backerman/evego/pkg/routing"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	// Register PgSQL driver
	_ "github.com/lib/pq"
)

var c config

type config struct {
	Dev                      bool
	DbDriver, DbPath         string
	Bind                     string
	BindProtocol             string
	XMLAPIEndpoint           string
	Router                   string
	Cache                    string
	RedisHost, RedisPassword string
	CookieDomain, CookiePath string
	ClientID, ClientSecret   string
	RedirectURL              string
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
	// workaround for viper bug
	// c.Dev = viper.GetBool("Dev")

	sde := dbaccess.SQLDatabase(c.DbDriver, c.DbPath)
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

	mux := newMux()
	setRoutes(mux, sde, localdb, xmlAPI, eveCentralMarket, sessionizer, myCache)

	// Set up internal bits.

	// Start background jobs.
	server.StartJobs(localdb)

	serve(mux, c.BindProtocol, c.Bind)
}

func serve(mux *web.Mux, bindProtocol, bindPort string) {
	// For now, this is completely lifted from goji's default handler.
	http.Handle("/", mux)
	log.Printf("Starting on %v/%v", bindProtocol, bindPort)
	graceful.HandleSignals()
	listener, err := net.Listen(bindProtocol, bindPort)
	if err != nil {
		log.Fatalf("Couldn't open socket on %v/%v: %v", bindProtocol, bindPort, err)
	}
	graceful.PreHook(func() { log.Info("Received signal, gracefully stopping.") })
	graceful.PostHook(func() { log.Info("Stopped.") })
	err = graceful.Serve(listener, http.DefaultServeMux)
	if err != nil {
		log.Fatalf("Couldn't serve on %v/%v: %v", bindProtocol, bindPort, err)
	}
	graceful.Wait()
}
