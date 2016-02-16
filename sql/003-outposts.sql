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

CREATE INDEX outposts_stationid ON outposts (stationid);

-- allstations: both outposts and stations, for foreign-key constraints
CREATE OR REPLACE VIEW eveindy.allstations AS
  -- some sovnull stations aren't actually outposts so will appear in both
  -- tables; we use COALESCE to merge them, preferring the conquerable stations
  -- table.
  SELECT COALESCE(o.stationID, s."stationID") "stationID",
         COALESCE(o.stationName, s."stationName") "stationName",
         ss."solarSystemID", ss."constellationID",
         ss."regionID",
         COALESCE(o.corporationID, s."corporationID") "corporationID",
         COALESCE(o.corporationName, n."itemName") "corporationName",
         COALESCE(s."reprocessingEfficiency", 0 :: double precision)
            "reprocessingEfficiency"
  FROM   outposts o
  FULL OUTER JOIN "staStations" s
  ON     o.stationID = s."stationID"
  -- Add in map information and corporation name for NPC stations.
  LEFT JOIN "invNames" n ON s."corporationID" = n."itemID"
  JOIN   "mapSolarSystems" ss
  ON ss."solarSystemID" = COALESCE(o.systemID, s."solarSystemID");
