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

package db

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strconv"
	"strings"

	"github.com/backerman/evego"
	"github.com/jmoiron/sqlx"
)

func (d *dbInterface) APIKeys(userID int) ([]XMLAPIKey, error) {
	// Use the unsafe statement - our key object has a list of characters,
	// which would otherwise trigger an error because it's not in this SQL
	// statement.
	unsafeStmt := d.getAPIKeysStmt.Unsafe()
	rows, err := unsafeStmt.Queryx(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]XMLAPIKey, 0, 2)
	for rows.Next() {
		key := XMLAPIKey{}
		err = rows.StructScan(&key)
		if err != nil {
			return nil, err
		}
		// Get the characters on this key.
		charRows, err := d.apiKeyListToonsStmt.Queryx(userID, key.ID)
		if err != nil {
			return nil, err
		}
		defer charRows.Close()
		for charRows.Next() {
			char := evego.Character{}
			err = charRows.StructScan(&char)
			if err != nil {
				return nil, err
			}
			key.Characters = append(key.Characters, char)
		}
		results = append(results, key)
	}
	return results, nil
}

func (d *dbInterface) DeleteAPIKey(userID, keyID int) error {
	_, err := d.deleteAPIKeyStmt.Exec(userID, keyID)
	return err
}

func (d *dbInterface) AddAPIKey(key XMLAPIKey) error {
	_, err := d.addAPIKeyStmt.Exec(key.User, key.ID, key.VerificationCode, key.Description)
	return err
}

func (d *dbInterface) GetAPICharacters(userid int, key XMLAPIKey) ([]evego.Character, error) {
	k := &evego.XMLKey{
		KeyID:            key.ID,
		VerificationCode: key.VerificationCode,
	}
	// Using the EVE XML API, get the characters on this account.
	toons, err := d.xmlAPI.AccountCharacters(k)
	if err != nil {
		return nil, err
	}
	tx, err := d.db.Beginx()
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// Defer constraints until end of transaction - this only affects those that
	// have been declared DEFERRABLE, and prevents API-derived skill information
	// from being deleted if the character is still on the key.
	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// Delete the toons returned by this key—avoid attempting to add a toon that's
	// already there. (PostgreSQL doesn't support upsert until 9.5 is released.)
	toonIDs := make([]string, 0, 3)
	for _, toon := range toons {
		toonIDs = append(toonIDs, strconv.Itoa(toon.ID))
	}
	// toonIDs is an array of toon IDs in string format; make it into a SQL
	// array string representation (which will be cast to an array by the
	// prepared statement).
	toonIDsString := fmt.Sprintf("{%s}", strings.Join(toonIDs, ", "))
	deleteStmt := tx.Stmtx(d.deleteToonsStmt)
	_, err = deleteStmt.Exec(userid, k.KeyID, toonIDsString)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// Now insert them list.
	insertStmt := tx.Stmtx(d.apiKeyInsertToonStmt)
	for _, toon := range toons {
		_, err := insertStmt.Exec(
			userid,
			key.ID,
			toon.Name,
			toon.ID,
			toon.Corporation,
			toon.CorporationID,
			toon.Alliance,
			toon.AllianceID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	return toons, tx.Commit()
}

func (d *dbInterface) GetAPISkills(key XMLAPIKey, charID int) error {
	k := &evego.XMLKey{
		KeyID:            key.ID,
		VerificationCode: key.VerificationCode,
	}
	charsheet, err := d.xmlAPI.CharacterSheet(k, charID)
	if err != nil {
		return err
	}
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Stmtx(d.apiKeyClearSkillsStmt).Exec(charID)
	if err != nil {
		return err
	}
	insertStmt := tx.Stmtx(d.apiKeyInsertSkillStmt)
	for _, skill := range charsheet.Skills {
		_, err := insertStmt.Exec(charID, skill.TypeID, skill.GroupID, skill.Level)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *dbInterface) GetAPIStandings(key XMLAPIKey, charID int) error {
	k := &evego.XMLKey{
		KeyID:            key.ID,
		VerificationCode: key.VerificationCode,
	}
	standings, err := d.xmlAPI.CharacterStandings(k, charID)
	if err != nil {
		return err
	}
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	// Clear standings before inserting the API's information.
	_, err = tx.Stmtx(d.apiKeyClearCorpStandingsStmt).Exec(charID)
	if err != nil {
		return err
	}
	_, err = tx.Stmtx(d.apiKeyClearFacStandingsStmt).Exec(charID)
	if err != nil {
		return err
	}
	var insertStmt *sqlx.Stmt
	for _, standing := range standings {
		switch standing.EntityType {
		case evego.NPCCorporation:
			insertStmt = d.apiKeyInsertCorpStandingsStmt
		case evego.NPCFaction:
			insertStmt = d.apiKeyInsertFacStandingsStmt
		default:
			// Agent standings - we don't handle those, so skip.
			continue
		}
		insertStmt = tx.Stmtx(insertStmt)
		_, err := insertStmt.Exec(charID, standing.ID, standing.Standing)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *dbInterface) CharacterStandings(userID, charID, corpID int) (corpStanding, factionStanding sql.NullFloat64, err error) {
	err = d.getStandingsStmt.QueryRow(userID, charID, corpID).Scan(&corpStanding, &factionStanding)
	return
}

func (d *dbInterface) CharacterSkill(userID, charID, skillID int) (int, error) {
	var skillLevel int
	err := d.getSkillStmt.QueryRow(userID, charID, skillID).Scan(&skillLevel)
	return skillLevel, err
}

func (d *dbInterface) CharacterSkillGroup(userID, charID, skillGroupID int) ([]evego.Skill, error) {
	skills := make([]evego.Skill, 0, 20)
	rows, err := d.getSkillGroupStmt.Unsafe().Queryx(userID, charID, skillGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		skill := evego.Skill{}
		err = rows.StructScan(&skill)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

func (d *dbInterface) RepopulateOutposts() error {
	outpostList := d.xmlAPI.DumpOutposts()
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	// clear existing
	_, err = tx.Stmtx(d.clearOutpostsStmt).Exec()
	if err != nil {
		return err
	}
	insertStmt := tx.Stmtx(d.insertOutpostsStmt)
	for _, o := range outpostList {
		_, err = insertStmt.Exec(o.Name, o.ID, o.SystemID, o.CorporationID, o.Corporation)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *dbInterface) StationForID(stationID int) (*evego.Station, error) {
	stn := &evego.Station{}
	row := d.getStationStmt.QueryRowx(stationID)
	err := row.StructScan(stn)
	return stn, err
}

func (d *dbInterface) GetBlueprints(key XMLAPIKey, charID int) error {
	k := &evego.XMLKey{
		KeyID:            key.ID,
		VerificationCode: key.VerificationCode,
	}
	blueprints, err := d.xmlAPI.Blueprints(k, charID)
	if err != nil {
		return err
	}
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	// Clear standings before inserting the API's information.
	_, err = tx.Stmtx(d.clearBlueprintsStmt).Exec(key.ID, charID)
	if err != nil {
		return err
	}
	insertStmt := tx.Stmtx(d.insertBlueprintStmt)
	for _, bp := range blueprints {
		_, err := insertStmt.Exec(key.ID, charID, bp.ItemID, bp.StationID, bp.LocationID,
			bp.TypeID, bp.Quantity, bp.Flag, bp.MaterialEfficiency, bp.TimeEfficiency,
			bp.NumRuns, bp.IsOriginal)
		if err != nil {
			log.Printf("Failed to insert blueprint %+v", bp)
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *dbInterface) CharacterBlueprints(userID, charID int) ([]evego.BlueprintItem, error) {
	rows, err := d.getBlueprintsStmt.Queryx(userID, charID)
	if err != nil {
		return nil, err
	}
	results := make([]evego.BlueprintItem, 0, 10)
	for rows.Next() {
		result := evego.BlueprintItem{}
		err = rows.StructScan(&result)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (d *dbInterface) GetAssets(key XMLAPIKey, charID int) error {
	k := &evego.XMLKey{
		KeyID:            key.ID,
		VerificationCode: key.VerificationCode,
	}
	// assetParent is a map of item IDs to their parent container.
	assetParent := make(map[int]int)

	assets, err := d.xmlAPI.Assets(k, charID)
	if err != nil {
		log.Printf("Unable to obtain assets for character %v: %v", charID, err)
		return err
	}
	tx, err := d.db.Beginx()
	if err != nil {
		log.Printf("Unable to acquire transaction: %v", err)
		return err
	}
	// Clear assets before inserting the API's information.
	_, err = tx.Stmtx(d.clearAssetsStmt).Exec(key.ID, charID)
	if err != nil {
		log.Printf("Unable to clear assets for key %v, character %v", key.ID, charID)
		return err
	}
	insertStmt := tx.Stmtx(d.insertAssetStmt)
	// Set up queue of assets.
	assetQueue := assets
	var a evego.InventoryItem
	for len(assetQueue) > 0 {
		a, assetQueue = assetQueue[0], assetQueue[1:]
		parentID, found := assetParent[a.ItemID]
		if !found {
			parentID = a.StationID
		}
		_, err := insertStmt.Exec(key.ID, charID, a.ItemID, parentID, a.StationID,
			a.TypeID, a.Quantity, a.Flag, a.Unpackaged)
		if err != nil {
			log.Printf("Failed to insert asset %+v", a)
			tx.Rollback()
			return err
		}
		// If there are containers in this asset, prepend them to the queue.
		if a.Contents != nil && len(a.Contents) > 0 {
			// First, set their parents in our lookup map.
			for _, item := range a.Contents {
				assetParent[item.ItemID] = a.ItemID
			}
			assetQueue = append(a.Contents, assetQueue...)
		}
	}
	return tx.Commit()
}
