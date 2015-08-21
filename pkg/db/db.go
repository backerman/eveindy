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

// Package db interfaces to the local database (user data).
package db

import (
	"encoding/json"
	"log"

	"github.com/backerman/evego/pkg/evesso"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type dbInterface struct {
	db                *sqlx.DB
	getSessionStmt    *sqlx.Stmt
	getAPIKeysStmt    *sqlx.Stmt
	deleteAPIKeyStmt  *sqlx.Stmt
	setTokenStmt      *sqlx.Stmt
	logoutSessionStmt *sqlx.Stmt
}

// Interface returns an interface to the local data store. Currently, it assumes
// that the schema our tables and functions are in can be found in the search
// path, so you'll need to ensure that it's set in the provided resource.
//
// Example resource: "user=enoch dbname=evetool search_path=eveindy"
func Interface(driver, resource string) (LocalDB, error) {
	dbConn, err := sqlx.Connect(driver, resource)
	if err != nil {
		return nil, err
	}
	// Is resource a URL or the other thing?
	// Find out, then add/modify search_path parameter.
	d := &dbInterface{db: dbConn}
	// Prepare statements
	stmts := []struct {
		preparedStatement **sqlx.Stmt
		statementText     string
	}{
		// Pointer magic, stage 1: Pass the address of the pointer.
		{&d.getSessionStmt, getSessionStmt},
		{&d.setTokenStmt, setTokenStmt},
		{&d.getAPIKeysStmt, getAPIKeysStmt},
		{&d.deleteAPIKeyStmt, deleteAPIKeyStmt},
		{&d.logoutSessionStmt, logoutSessionStmt},
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
