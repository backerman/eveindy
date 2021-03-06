-- Copyright © 2014–6 Brad Ackerman.
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
  charID integer NOT NULL,
  apikey integer NOT NULL,
  itemID bigint NOT NULL,
  stationID integer NOT NULL,
  locationID bigint NOT NULL,
  typeID integer NOT NULL,
  quantity integer NOT NULL,
  flag integer NOT NULL,
  materialEfficiency integer NOT NULL,
  timeEfficiency integer NOT NULL,
  numRuns integer,
  isOriginal boolean NOT NULL,

  FOREIGN KEY (charID) REFERENCES eveindy.characters (id)
    ON DELETE CASCADE DEFERRABLE,
  FOREIGN KEY (apikey) REFERENCES eveindy.apikeys (id)
    ON DELETE CASCADE DEFERRABLE,
  FOREIGN KEY (typeID) REFERENCES "invTypes" ("typeID") DEFERRABLE,
  FOREIGN KEY (flag) REFERENCES "invFlags" ("flagID") DEFERRABLE,
  CHECK (quantity > 0),
  CHECK (materialEfficiency BETWEEN 0 AND 10),
  CHECK (timeEfficiency BETWEEN 0 AND 20 AND timeEfficiency % 2 = 0),
  CHECK (isOriginal IS TRUE OR (numRuns IS NOT NULL AND numRuns > 0))
);

-- trigger on insert: ensure location is valid
CREATE OR REPLACE FUNCTION blueprints_insert_check() RETURNS TRIGGER AS $$
DECLARE
  station integer;
  parent bigint;
BEGIN
  SELECT "stationID" from eveindy.allstations
  WHERE  "stationID" = NEW.stationID
  INTO station;
  IF station IS NULL
  THEN
    RAISE EXCEPTION 'location ID % is invalid', NEW.stationID;
  END IF;
  IF NEW.stationID <> NEW.locationID
  THEN
    SELECT itemID from eveindy.assets
    WHERE itemID = NEW.locationID
    INTO parent;
    IF parent IS NULL
    THEN
      RAISE EXCEPTION 'parent ID % is invalid', NEW.locationID;
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER blueprints_insert AFTER INSERT OR UPDATE ON eveindy.blueprints
INITIALLY DEFERRED
FOR EACH ROW EXECUTE PROCEDURE blueprints_insert_check();
