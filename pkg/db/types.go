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

package db

import (
	"time"

	"golang.org/x/oauth2"
)

// Session is a user's session (logged in or otherwise).
type Session struct {
	// User is the application user's ID. (This is a local identifier and is not
	// related to the authenticating account or character.)
	User int `db:"userid"`

	// State is a random value that will be passed to CCP's servers when making
	// an OAuth request.
	State string `db:"state"`

	// Cookie is a random value that is stored on the client's system (as a cookie)
	// to identify the client across requests.
	Cookie string `db:"cookie"`

	// Token is the OAuth token returned from CCP's servers and can be used to
	// access the CREST API.
	Token *oauth2.Token `db:"-"`

	// LastSeen is the time at which the session owner used this website.
	LastSeen time.Time `db:"lastseen"`
}

// XMLAPIKey is a user-provided key for the EVE XML API.
type XMLAPIKey struct {
	// User is the application user's ID. (This is a local identifier and is not
	// related to the authenticating account or character.)
	User int `db:"userid" json:"userid"`

	// ID is the key's ID assigned by CCP.
	ID int `db:"id" json:"id"`

	// VerificationCode is the key's verification code assigned by CCP.
	VerificationCode string `db:"vcode" json:"vcode"`

	// Description is a user-entered label for the key; it is not the one that
	// the user entered in EVE's API managment interface.
	Description string `db:"label" json:"label"`
}
