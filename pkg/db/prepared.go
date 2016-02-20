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

	// Add an API key to the database.
	addAPIKeyStmt = `
	INSERT
	INTO   apikeys(userid, id, vcode, label)
	VALUES ($1, $2, $3, $4)
	`

	// Delete an API key.
	deleteAPIKeyStmt = `
	DELETE
	FROM   apikeys
	WHERE  userid = $1 AND id = $2
	`

	// Delete user's sessions.
	logoutSessionStmt = `
	DELETE FROM sessions
	WHERE userid IN (SELECT userid FROM sessions WHERE cookie = $1)
	`

	// Add an API key's characters to the database.
	apiKeyInsertToonStmt = `
	INSERT INTO characters(userid, apikey, name, id, corp, corpid, alliance, allianceid)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	apiKeyListToonsStmt = `
	SELECT name, id, corp AS corporation, corpid AS corporationid, alliance, allianceid
	FROM characters
	WHERE userid = $1 and apikey = $2
	`

	apiKeyClearSkillsStmt = `
	DELETE FROM skills
	WHERE charid = $1
	`

	apiKeyInsertSkillStmt = `
	INSERT INTO skills(charid, id, groupid, level)
	VALUES ($1, $2, $3, $4)
	`

	// Get a character's specified skill (by ID); if the character hasn't injected
	// this skill, return zero.
	getSkillStmt = `
	SELECT COALESCE(
		(SELECT level
		FROM   skills s
		JOIN characters c ON c.id = s.charID
		WHERE c.userID = $1 AND c.id = $2 AND s.id = $3), 0
	)
	`

	// Get a character's skills of the specified group.
	getSkillGroupStmt = `
	SELECT t."typeName" "name", s.id typeID, g."groupName" "group", s.groupID,
	       s.level, true published
	FROM   skills s
	JOIN   "invTypes" t on t."typeID" = s.id
	JOIN   "invGroups" g on g."groupID" = s.groupID
	JOIN   characters c ON c.id = s.charID
	WHERE  c.userID = $1 AND c.id = $2 AND s.groupID = $3
	`

	apiKeyClearCorpStandingsStmt = `
	DELETE FROM corpStandings
	WHERE charid = $1
	`

	apiKeyClearFacStandingsStmt = `
	DELETE FROM facStandings
	WHERE charid = $1
	`

	apiKeyInsertCorpStandingsStmt = `
	INSERT INTO corpStandings(charid, corp, standing)
	VALUES ($1, $2, $3)
	`

	apiKeyInsertFacStandingsStmt = `
	INSERT INTO facStandings(charid, faction, standing)
	VALUES ($1, $2, $3)
	`

	// Get NPC corporation and faction standings for a character.
	getStandingsStmt = `
	WITH availableCharacters AS (
	  SELECT id
	  FROM   characters
	  WHERE  userid = $1
	)
	SELECT    corp.standing corp_standing, fac.standing fac_standing
	FROM      availableCharacters c, "crpNPCCorporations" npcCorps
	LEFT JOIN corpStandings corp ON corp.corp = npcCorps."corporationID"
	AND       corp.charID = $2
	LEFT JOIN facStandings fac ON fac.faction = npcCorps."factionID"
	AND       fac.charID = $2
	WHERE     npcCorps."corporationID" = $3
	AND       (c.id = corp.charID OR c.id = fac.charID
             OR (corp.charID IS NULL AND fac.charID IS NULL));
	`

	// Delete all characters that either came from this API key or are in the
	// provided list. (The former condition is required to handle the case where
	// an API key previously, but no longer, provided access to a given character.)
	deleteToonsStmt = `
	DELETE FROM characters
	WHERE userid = $1
	AND   (apikey = $2 OR id = ANY ($3::int[]))
	`

	// Outpost list update

	// Clear outposts.
	// Not using TRUNCATE because it's not MVCC-safe.
	clearOutpostsStmt = `
	DELETE FROM outposts WHERE 1=1
	`

	// Insert outposts into the database.
	insertOutpostsStmt = `
	INSERT INTO outposts(stationName, stationID, systemID, corporationID, corporationName)
	VALUES ($1, $2, $3, $4, $5)
	`

	// Search outposts/stations from the database
	searchStationsStmt = `
	SELECT   "stationName", "stationID", "solarSystemID", "corporationID", "corporationName",
	         "constellationID", "regionID", "reprocessingEfficiency"
	FROM     allStations o
	WHERE    LOWER("stationName") LIKE LOWER($1)
	ORDER BY "stationName"
	LIMIT    10
	`

	// Get a specified station or outpost by ID.
	getStationStmt = `
	SELECT   "stationName", "stationID", "solarSystemID", "constellationID", "regionID",
					 "corporationID", "corporationName", "reprocessingEfficiency"
	FROM     allStations o
	WHERE    "stationID" = $1
	`

	// Assets

	// Clear toon's assets.
	clearAssetsStmt = `
	DELETE FROM assets
	WHERE apiKey = $1 AND charID = $2
	`

	// Insert an asset.
	insertAssetStmt = `
	INSERT INTO assets
		(apiKey, charID, itemID, locationID, stationID, typeID, quantity, flag,
		 unpackaged)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Get user's assets.
	// Lowercase everything for sqlx.
	getAssetsStmt = `
	WITH availableCharacters AS (
		SELECT id
		FROM   characters
		WHERE  userid = $1
	)
	SELECT apikey, itemid, locationid, stationid, typeid, quantity, flag,
	 			 unpackaged
	FROM   assets b
	JOIN   availableCharacters a on a.id = b.charID
	WHERE  charID = $2
	`
)
