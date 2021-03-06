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

// Package db interfaces to the local database (user data).
package db

import (
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"

	"github.com/backerman/evego"
	"github.com/backerman/evego/pkg/evesso"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type dbInterface struct {
	db                            *sqlx.DB
	getSessionStmt                *sqlx.Stmt
	getAPIKeysStmt                *sqlx.Stmt
	addAPIKeyStmt                 *sqlx.Stmt
	deleteAPIKeyStmt              *sqlx.Stmt
	setTokenStmt                  *sqlx.Stmt
	logoutSessionStmt             *sqlx.Stmt
	apiKeyInsertToonStmt          *sqlx.Stmt
	apiKeyListToonsStmt           *sqlx.Stmt
	apiKeyInsertSkillStmt         *sqlx.Stmt
	apiKeyClearSkillsStmt         *sqlx.Stmt
	getSkillStmt                  *sqlx.Stmt
	getSkillGroupStmt             *sqlx.Stmt
	apiKeyClearCorpStandingsStmt  *sqlx.Stmt
	apiKeyClearFacStandingsStmt   *sqlx.Stmt
	apiKeyInsertCorpStandingsStmt *sqlx.Stmt
	apiKeyInsertFacStandingsStmt  *sqlx.Stmt
	getStandingsStmt              *sqlx.Stmt
	deleteToonsStmt               *sqlx.Stmt
	clearOutpostsStmt             *sqlx.Stmt
	insertOutpostsStmt            *sqlx.Stmt
	searchStationsStmt            *sqlx.Stmt
	getStationStmt                *sqlx.Stmt
	clearBlueprintsStmt           *sqlx.Stmt
	insertBlueprintStmt           *sqlx.Stmt
	getBlueprintsStmt             *sqlx.Stmt
	clearAssetsStmt               *sqlx.Stmt
	insertAssetStmt               *sqlx.Stmt
	getAssetsStmt                 *sqlx.Stmt
	unusedSalvageStmt             *sqlx.Stmt

	// Need access to EVE APIs.
	xmlAPI evego.XMLAPI
}

// Interface returns an interface to the local data store. Currently, it assumes
// that the schema our tables and functions are in can be found in the search
// path, so you'll need to ensure that it's set in the provided resource.
//
// Example resource: "user=enoch dbname=evetool search_path=eveindy"
func Interface(driver, resource string, xmlAPI evego.XMLAPI) (LocalDB, error) {
	dbConn, err := sqlx.Connect(driver, resource)
	if err != nil {
		return nil, err
	}
	// Is resource a URL or the other thing?
	// Find out, then add/modify search_path parameter.
	d := &dbInterface{
		db:     dbConn,
		xmlAPI: xmlAPI,
	}
	// Prepare statements
	stmts := []struct {
		preparedStatement **sqlx.Stmt
		statementText     string
	}{
		// Pointer magic, stage 1: Pass the address of the pointer.
		{&d.getSessionStmt, getSessionStmt},
		{&d.setTokenStmt, setTokenStmt},
		{&d.getAPIKeysStmt, getAPIKeysStmt},
		{&d.addAPIKeyStmt, addAPIKeyStmt},
		{&d.deleteAPIKeyStmt, deleteAPIKeyStmt},
		{&d.logoutSessionStmt, logoutSessionStmt},
		{&d.apiKeyInsertToonStmt, apiKeyInsertToonStmt},
		{&d.apiKeyListToonsStmt, apiKeyListToonsStmt},
		{&d.apiKeyInsertSkillStmt, apiKeyInsertSkillStmt},
		{&d.apiKeyClearSkillsStmt, apiKeyClearSkillsStmt},
		{&d.getSkillStmt, getSkillStmt},
		{&d.getSkillGroupStmt, getSkillGroupStmt},
		{&d.apiKeyClearCorpStandingsStmt, apiKeyClearCorpStandingsStmt},
		{&d.apiKeyClearFacStandingsStmt, apiKeyClearFacStandingsStmt},
		{&d.apiKeyInsertCorpStandingsStmt, apiKeyInsertCorpStandingsStmt},
		{&d.apiKeyInsertFacStandingsStmt, apiKeyInsertFacStandingsStmt},
		{&d.getStandingsStmt, getStandingsStmt},
		{&d.deleteToonsStmt, deleteToonsStmt},
		{&d.clearOutpostsStmt, clearOutpostsStmt},
		{&d.insertOutpostsStmt, insertOutpostsStmt},
		{&d.searchStationsStmt, searchStationsStmt},
		{&d.getStationStmt, getStationStmt},
		{&d.clearBlueprintsStmt, clearBlueprintsStmt},
		{&d.insertBlueprintStmt, insertBlueprintStmt},
		{&d.getBlueprintsStmt, getBlueprintsStmt},
		{&d.clearAssetsStmt, clearAssetsStmt},
		{&d.insertAssetStmt, insertAssetStmt},
		{&d.getAssetsStmt, getAssetsStmt},
		{&d.unusedSalvageStmt, unusedSalvageStmt},
	}

	for _, s := range stmts {
		prepared, err := dbConn.Preparex(s.statementText)
		if err != nil {
			log.Fatalf("Unable to prepare statement: %v\n%v", err, s.statementText)
		}
		// Pointer magic, stage 2: Dereference the pointer to the pointer
		// and set it to point to the statement we just prepared.
		*s.preparedStatement = prepared
	}

	return d, nil
}

// NewSession returns a new session.
func (d *dbInterface) NewSession() (Session, error) {
	// Our implementation will return a new session if it can't find the queried
	// one, so just query for the empty string if we know there's no session.
	return d.FindSession("")
}

func (d *dbInterface) FindSession(cookie string) (Session, error) {
	row := d.getSessionStmt.QueryRowx(cookie)
	s := Session{
		// Initialize pointers in struct.
		Token: &oauth2.Token{},
	}
	// err := row.StructScan(&newSession)
	var tokenJSON []byte
	nullableUser := new(int)
	err := row.Scan(&nullableUser, &s.State, &s.Cookie, &tokenJSON, &s.LastSeen)
	if err != nil {
		return s, err
	}
	if nullableUser != nil {
		s.User = *nullableUser
	}
	json.Unmarshal(tokenJSON, s.Token)
	return s, err
}

func (d *dbInterface) AuthenticateSession(
	cookie string, token *oauth2.Token, charInfo *evesso.CharacterInfo) error {
	tokenJSON, err := json.Marshal(*token)
	if err != nil {
		return err
	}
	charInfoJSON, err := json.Marshal(*charInfo)
	if err != nil {
		return err
	}
	if err == nil {
		_, err = d.setTokenStmt.Exec(cookie, tokenJSON, charInfoJSON)
	}
	return err
}

func (d *dbInterface) LogoutSession(cookie string) error {
	// TODO: Update table, remove login.
	_, err := d.logoutSessionStmt.Exec(cookie)
	return err
}

func (d *dbInterface) SearchStations(search string) ([]evego.Station, error) {
	rows, err := d.searchStationsStmt.Queryx("%" + search + "%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stations []evego.Station
	for rows.Next() {
		station := evego.Station{}
		rows.StructScan(&station)
		stations = append(stations, station)
	}
	if len(stations) == 0 {
		err = sql.ErrNoRows
	}
	return stations, err
}
