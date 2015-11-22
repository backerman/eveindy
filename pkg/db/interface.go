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
	"database/sql"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/evesso"
	"golang.org/x/oauth2"
)

// LocalDB is an interface to this application's local data.
type LocalDB interface {
	// NewSession generates a new session.
	NewSession() (Session, error)

	// FindSession attempts to retrieve an existing session from the database;
	// if it was not found, a new session will be returned.
	FindSession(cookie string) (Session, error)

	// AuthenticateSession associates a session with an EVE character authenticated
	// by OAuth2.
	AuthenticateSession(cookie string, token *oauth2.Token, charInfo *evesso.CharacterInfo) error

	// APIKeys returns the user's API keys that have been registered in this application.
	APIKeys(userID int) ([]XMLAPIKey, error)

	// LogoutSession deletes all of a user's sessions.
	LogoutSession(cookie string) error

	// DeleteAPIKey deletes the specified API key.
	DeleteAPIKey(userID, keyID int) error

	// AddAPIKey adds the specified API key.
	AddAPIKey(key XMLAPIKey) error

	// GetAPICharacters adds the characters on an API key to the database.
	GetAPICharacters(userid int, key XMLAPIKey) ([]evego.Character, error)

	// GetAPISkills adds the skills on a character to the database.
	GetAPISkills(key XMLAPIKey, charID int) error

	// CharacterSkill returns the specified skill's level, or 0 if it has not
	// been injected.
	CharacterSkill(userID, charID, skillID int) (int, error)

	// CharacterSkill returns the levels of all injected skills in the specified
	// group.
	CharacterSkillGroup(userID, charID, skillGroupID int) ([]evego.Skill, error)

	// GetAPIStandings adds a character's standings with NPC entities to the database.
	GetAPIStandings(key XMLAPIKey, charID int) error

	// CharacterStandings queries a character's standings (corporation and faction)
	// with an NPC corporation.
	CharacterStandings(userID, charID, corpID int) (corpStanding, factionStanding sql.NullFloat64, err error)

	// RepopulateOutposts updates outpost information in the local database.
	RepopulateOutposts() error

	// SearchStations searches outpost and station names for the provided term,
	// adding %s on either side.
	SearchStations(search string) ([]evego.Station, error)

	// StationForID gets the station or outpost corresponding to the passed
	// ID.
	StationForID(stationID int) (*evego.Station, error)

	// GetBlueprints retrieves a character's blueprints and adds them to the
	// database.
	GetBlueprints(key XMLAPIKey, charID int) error

	// GetAssets retrieves a character's assets and adds them to the
	// database.
	GetAssets(key XMLAPIKey, charID int) error

	// CharacterBlueprints returns a character's blueprints from the local
	// database.
	CharacterBlueprints(userID, charID int) ([]evego.BlueprintItem, error)

	// UnusedSalvage returns a character's salvage inventory that is not used
	// by any blueprint he owns.
	UnusedSalvage(userid, characterID int) ([]evego.InventoryItem, error)
}
