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

package server

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"github.com/zenazn/goji/web/mutil"
)

// Logger returns a Goji web middleware that logs web transactions in the format
// I want.
func Logger(c *web.C, h http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		fields := log.Fields{
			"requestId": middleware.GetReqID(*c),
			"client":    r.RemoteAddr,
			"url":       r.URL.String(),
			"method":    r.Method,
			"protocol":  r.Proto,
			"referer":   r.Referer(),
			"start":     startTime,
		}
		log.WithFields(fields).Debug("Request started.")
		lw := mutil.WrapWriter(w)
		h.ServeHTTP(lw, r)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		fields["duration"] = duration.Nanoseconds()
		status := lw.Status()
		fields["status"] = status
		switch {
		case status < 300: // 100- and 200-series statuses
			log.WithFields(fields).Info("OK")
		case status < 400: // 300-series redirection
			log.WithFields(fields).Info("Redirected")
		case status < 500: // 400-series client error
			log.WithFields(fields).Warn("Client error")
		default: // 500-series server erorr
			log.WithFields(fields).Error("Server error!")
		}
	}
	return http.HandlerFunc(fn)
}
