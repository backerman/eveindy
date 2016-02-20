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
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd *cobra.Command

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
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	rootCmd.Execute()
}
