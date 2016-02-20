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
	"fmt"
	"log"
	"os"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/cache"
	"github.com/backerman/evego/pkg/dbaccess"
	"github.com/backerman/evego/pkg/eveapi"
	"github.com/backerman/evego/pkg/market"
	"github.com/backerman/evego/pkg/routing"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zenazn/goji"
	// Register PgSQL driver
	_ "github.com/lib/pq"
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

	if c.Dev {
		log.Printf("Configuration is: %+v", c)
	}

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

	setRoutes(sde, localdb, xmlAPI, eveCentralMarket, sessionizer, myCache)

	// Set up internal bits.

	// Start background jobs.
	server.StartJobs(localdb)

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
	// When dev is true, server only listens on localhost and will serve
	// external dependencies (e.g. AngularJS) from local disk instead of from
	// a CDN.
	rootCmd.Flags().Bool("Dev", false, "Set development mode.")
	rootCmd.Flags().String("Bind", ":8080", "The address and port to listen on.")
	viper.SetDefault("Dev", false)
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

	flags := []string{"Dev", "Bind"}
	for _, flag := range flags {
		viper.BindPFlag(flag, rootCmd.Flags().Lookup(flag))
	}
	rootCmd.Execute()
}
