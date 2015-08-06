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

// Prepared statements.

const (
	// Find an existing session; return it if it still exists or a new session
	// otherwise.
	getSessionStmt = `
  SELECT userid, state, cookie, token, lastseen FROM getSession($1)
  `

	// Associate a token with a session. The first argument should be the
	// session's cookie value, the second the token (JSON text), and the third
	// the character's information object as returned from the SSO API (JSON text).
	setTokenStmt = `
  SELECT associateToken($1, $2, $3)
  `

	// Get all API keys that have been registered for a user.
	getAPIKeysStmt = `
	SELECT userid, id, vcode, label
	FROM   apikeys
	WHERE  userid = $1
	`

	// Delete user's sessions.
	logoutSessionStmt = `
	DELETE FROM sessions
	WHERE userid IN (SELECT userid FROM sessions WHERE cookie = $1)
	`
)
