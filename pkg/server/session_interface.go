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

	"github.com/backerman/eveindy/pkg/db"
	"github.com/zenazn/goji/web"
)

// Sessionizer is an object that provides a client's session.
type Sessionizer interface {
	// GetSession returns the client's session, creating a new one if necessary.
	GetSession(c *web.C, w http.ResponseWriter, r *http.Request) *db.Session
}
