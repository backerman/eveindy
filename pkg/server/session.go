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

	"github.com/backerman/eveindy/pkg/db"
	"github.com/zenazn/goji/web"
)

const cookieName = "EVEINDY_SESSION"

type sessionizer struct {
	cookieDomain, cookiePath string
	isProduction             bool
	db                       db.LocalDB
}

// GetSessionizer returns a Sessionizer to be passed to handlers.
func GetSessionizer(cookieDomain, cookiePath string, isProduction bool, db db.LocalDB) Sessionizer {
	return &sessionizer{
		cookieDomain: cookieDomain,
		cookiePath:   cookiePath,
		isProduction: isProduction,
		db:           db,
	}
}

func (s *sessionizer) setCookie(w http.ResponseWriter, cookie string) {
	sessionCookie := &http.Cookie{
		Name:    cookieName,
		Value:   cookie,
		Domain:  s.cookieDomain,
		Path:    s.cookiePath,
		Expires: time.Now().Add(24 * 30 * 3 * time.Hour), // 3 months
		Secure:  s.isProduction,                          // HTTPS-only iff production system
	}
	http.SetCookie(w, sessionCookie)
}

func (s *sessionizer) GetSession(c *web.C, w http.ResponseWriter, r *http.Request) *db.Session {
	// Get my session cookie.
	var session db.Session
	var newSession bool
	sessionCookie, err := r.Cookie(cookieName)
	if err == nil {
		// Got a cookie; check to see if the corresponding session is available.
		oldCookie := sessionCookie.Value
		session, err = s.db.FindSession(oldCookie)
		if err == nil && session.Cookie != oldCookie {
			// This session didn't exist, so a new session has been created.
			newSession = true
		}
	} else {
		// The session cookie did not exist.
		session, err = s.db.NewSession()
		newSession = true
	}
	if err != nil {
		panic("OMG! " + err.Error())
	}
	if newSession {
		// Store a cookie.
		s.setCookie(w, session.Cookie)
	}
	c.Env["session"] = session

	return &session
}
