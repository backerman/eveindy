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

-- blueprints: blueprints owned by users
CREATE TABLE eveindy.blueprints (
  userID integer NOT NULL,
  charID integer NOT NULL,
  apikey integer NOT NULL,
  itemID integer NOT NULL,
  locationID integer NOT NULL,
  typeID integer NOT NULL,
  quantity integer NOT NULL,
  flag integer NOT NULL,
  materialEfficiency integer NOT NULL,
  timeEfficiency integer NOT NULL,
  numRuns integer,
  isOriginal boolean NOT NULL,

  FOREIGN KEY (charID) REFERENCES eveindy.characters (id)
    ON DELETE CASCADE DEFERRABLE,
  FOREIGN KEY (userID) REFERENCES eveindy.users (id)
    ON DELETE CASCADE DEFERRABLE,
  FOREIGN KEY (apikey) REFERENCES eveindy.apikeys (id)
    ON DELETE CASCADE DEFERRABLE,
  --  can't FK to a view!
  -- FOREIGN KEY (locationID) REFERENCES allstations (stationID)
  --   ON DELETE NO ACTION DEFERRABLE,
  FOREIGN KEY (typeID) REFERENCES "invTypes" ("typeID"),
  FOREIGN KEY (flag) REFERENCES "invFlags" ("flagID"),
  CHECK (quantity > 0),
  CHECK (materialEfficiency BETWEEN 0 AND 10 AND materialEfficiency % 2 = 0),
  CHECK (timeEfficiency BETWEEN 0 AND 20 AND timeEfficiency % 2 = 0),
  CHECK (isOriginal IS TRUE OR (numRuns IS NOT NULL AND numRuns > 0))
);

-- trigger on insert: ensure location is valid
CREATE OR REPLACE FUNCTION blueprints_insert_check() RETURNS TRIGGER AS $$
DECLARE
  station integer;
BEGIN
  SELECT stationID from eveindy.allstations
  WHERE  stationID = NEW.stationID
  INTO station;
  IF station IS NULL
  THEN
    RAISE EXCEPTION 'location ID % is invalid', NEW.stationID;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER blueprints_insert AFTER INSERT OR UPDATE ON eveindy.blueprints
FOR EACH ROW EXECUTE PROCEDURE blueprints_insert_check();
