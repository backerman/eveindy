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

const (
	// Blueprints

	// Clear toon's blueprints.
	clearBlueprintsStmt = `
  DELETE FROM blueprints
  WHERE apiKey = $1 AND charID = $2
  `

	// Insert a blueprint.
	insertBlueprintStmt = `
  INSERT INTO blueprints
    (apiKey, charID, itemID, stationID, locationID, typeID, quantity, flag,
     materialEfficiency, timeEfficiency, numRuns, isOriginal)
   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
  `

	// Get user's blueprints.
	// Lowercase everything for sqlx.
	getBlueprintsStmt = `
  WITH availableCharacters AS (
    SELECT id
    FROM   characters
    WHERE  userid = $1
  )
  SELECT itemid, stationid, locationid, typeid, "typeName" typename, quantity,
         flag, materialefficiency, timeefficiency, numruns, isoriginal
  FROM   blueprints b
  JOIN   availableCharacters a on a.id = b.charID
  JOIN   "invTypes" t ON b.typeid = t."typeID"
  WHERE  charID = $2
  `

	// Unused salvage
	unusedSalvageStmt = `
  WITH salvagedrops AS (
    SELECT "typeID" typeid, "typeName" typename
    FROM   "invTypes" t
    JOIN   "invMarketGroups" mg USING("marketGroupID")
    WHERE  "marketGroupName" = 'Salvaged Materials'
  ), bpusage AS (
    SELECT ti."typeName" blueprint, ti."typeID" blueprintid,
           tm."typeID" inputmaterialid,
           iam."quantity" inputMaterialQty
    FROM   "industryActivityMaterials" iam
    JOIN   "invTypes" ti USING("typeID")
    JOIN   "invTypes" tm
    ON     iam."materialTypeID" = tm."typeID"
    JOIN   "ramActivities" USING("activityID")
    WHERE  "activityName" = 'Manufacturing'
  ), evolvesto AS (
    SELECT ti."typeID" inputblueprint, tyo."typeID" outputblueprint
    FROM   "invTypes" ti
    JOIN   "industryActivityProducts" iap
    ON     iap."typeID" = ti."typeID"
    JOIN   "ramActivities" USING("activityID")
    JOIN   "invTypes" tyo ON iap."productTypeID" = tyo."typeID"
    WHERE  "activityName" = 'Invention'
  ), availableCharacters AS (
			SELECT id
			FROM   characters
			WHERE  userid = $1
	)
	SELECT itemid, stationid, typeid, quantity, flag,
	 			 unpackaged
	FROM   assets b
	JOIN   availableCharacters a on a.id = b.charID
	WHERE  charID = $2
	AND    typeid IN (
    SELECT typeid
    FROM   assets
    JOIN   salvagedrops USING(typeid)
    EXCEPT (
      SELECT inputmaterialid
      FROM   blueprints, bpusage, evolvesto
      WHERE  bpusage.blueprintid = evolvesto.outputblueprint
      AND    evolvesto.inputblueprint = blueprints.typeid
      AND    blueprints.charid = $2
      UNION
      SELECT inputmaterialid
      FROM   blueprints
      JOIN   bpusage
      ON     bpusage.blueprintid = blueprints.typeid
      WHERE  blueprints.charid = $2
    )
  )
  `
)
