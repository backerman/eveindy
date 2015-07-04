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
  SELECT * FROM getSession($1)
  `

	// Associate a token with a session. The first argument should be the
	// session's cookie value, and the second the token (JSON text).
	setTokenStmt = `
  SELECT associateToken($1, $2)
  `
)
