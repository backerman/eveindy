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

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/backerman/evego/pkg/evesso"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

type sessionInfo struct {
	Authenticated bool      `json:"authenticated"`
	OAuthURL      string    `json:"oauthURL"`
	CharName      string    `json:"characterName,omitempty"`
	OAuthExpiry   time.Time `json:"oauthExpiresAt,omitempty"`
}

// SessionInfo returns a web handler function that returns information about the
// current session.
func SessionInfo(auth evesso.Authenticator) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		curSession := server.GetSession(&c)
		returnInfo := sessionInfo{
			Authenticated: curSession.User != 0,
			OAuthURL:      auth.URL(curSession.State),
		}
		if curSession.Token != nil {
			returnInfo.OAuthExpiry = curSession.Token.Expiry
		}
		returnJSON, _ := json.Marshal(&returnInfo)
		w.Write(returnJSON)
	}
}

// AuthenticateHandler returns a web handler function that redirects to a
// session-specific authentication link.
func AuthenticateHandler(auth evesso.Authenticator) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		sess := server.GetSession(&c)
		url := auth.URL(sess.State)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

// LogoutHandler returns a web handler function that deletes the user's
// sessions.
func LogoutHandler(localdb db.LocalDB, auth evesso.Authenticator) web.HandlerFunc {
	successMsg := []byte("{ \"success\": true }")
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		sess := server.GetSession(&c)
		err := localdb.LogoutSession(sess.Cookie)
		if err != nil {
			http.Error(w, "Unable to find session", http.StatusTeapot)
			log.Printf("Error logging out: %v", err)
			return
		}
		// Serve some JSON that confirms success.
		w.Write(successMsg)
	}
}

// CRESTCallbackListener returns a web handler function that listens for a CREST
// SSO callback and accepts the results of authentication.
func CRESTCallbackListener(localdb db.LocalDB, auth evesso.Authenticator) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		// Verify state value.
		session := server.GetSession(&c)
		passedState := r.FormValue("state")
		if passedState != session.State {
			// CSRF attempt or session expired; reject.
			http.Error(w, "Returned state not valid for this user.", http.StatusBadRequest)
			log.Printf("Got state %#v, expected state %#v", passedState, session.State)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		// Extract code from query parameters.
		code := r.FormValue("code")
		// Exchange it for a token.
		tok, err := auth.Exchange(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		// Get character information.
		charInfo, err := auth.CharacterInfo(tok)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}

		// Update session in database.
		err = localdb.AuthenticateSession(session.Cookie, tok, charInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
				<head>
					<title>Authenticated</title>
				</head>
				<body>
					<p>OK.</p>
					<script type="text/javascript">
						window.onload = function() {
							window.opener.hasAuthenticated();
							window.close();
						}
					</script>
				</body>
			</html>
			`))
	}
}
