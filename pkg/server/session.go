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

package server

import (
	"net/http"
	"time"

	"github.com/backerman/eveindy/pkg/db"
	"github.com/zenazn/goji/web"
)

const cookieName = "EVEINDY_SESSION"

// SessionHandler is a middleware that maps the request to its corresponding
// session.
func SessionHandler(d db.LocalDB, cookieDomain, cookiePath string, isProduction bool) func(*web.C, http.Handler) http.Handler {
	aHandler := func(c *web.C, h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Get my session cookie.
			var session db.Session
			var newSession bool
			sessionCookie, err := r.Cookie(cookieName)
			if err == nil {
				// Got a cookie; check to see if the corresponding session is available.
				oldCookie := sessionCookie.Value
				session, err = d.FindSession(oldCookie)
				if err == nil && session.Cookie != oldCookie {
					// This session didn't exist, so a new session has been created.
					newSession = true
				}
			} else {
				// The session cookie did not exist.
				session, err = d.NewSession()
				newSession = true
			}
			if err != nil {
				panic("OMG! " + err.Error())
			}
			if newSession {
				// Store a cookie.
				sessionCookie = &http.Cookie{
					Name:    cookieName,
					Value:   session.Cookie,
					Domain:  cookieDomain,
					Path:    cookiePath,
					Expires: time.Now().Add(168 * time.Hour), // 1 week
					Secure:  isProduction,                    // HTTPS-only iff production system
				}
				http.SetCookie(w, sessionCookie)
			}
			c.Env["session"] = session
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
	return aHandler
}

// GetSession returns the context's application session information.
func GetSession(c *web.C) *db.Session {
	userSession, ok := c.Env["session"].(db.Session)
	if !ok {
		return nil
	}
	return &userSession
}
