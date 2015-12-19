-- Copyright © 2014–5 Brad Ackerman.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
-- http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

-- outposts: conquerable outposts
CREATE TABLE eveindy.outposts (
  stationName text NOT NULL,
  stationID integer NOT NULL,
  systemID integer NOT NULL,
  corporationID integer NOT NULL,
  corporationName text NOT NULL,
  -- no reprocessing efficiency: this information is not publicly available
  FOREIGN KEY (systemID) REFERENCES "mapSolarSystems" ("solarSystemID") DEFERRABLE
);

-- allstations: both outposts and stations, for foreign-key constraints
CREATE OR REPLACE VIEW eveindy.allstations AS
  SELECT "stationID", "stationName", "solarSystemID", "constellationID",
         "regionID", "corporationID", "itemName" "corporationName",
         "reprocessingEfficiency"
  FROM   "staStations" s
  JOIN   "invNames" n ON n."itemID" = s."corporationID"
  UNION ALL
  SELECT stationID, stationName, systemID, "constellationID", "regionID",
         corporationID, corporationName,
         0 :: double precision reprocessingEfficiency
  FROM eveindy.outposts o
  JOIN "mapSolarSystems" s ON s."solarSystemID" = o.systemID
  ;
