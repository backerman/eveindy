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

// LocalDB is an interface to this application's local data.
type LocalDB interface {
	// NewSession generates a new session.
	NewSession() (Session, error)

	// FindSession attempts to retrieve an existing session from the database;
	// if it was not found, a new session will be returned.
	FindSession(cookie string) (Session, error)
}
