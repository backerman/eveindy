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

-- assets: characters' assets
CREATE TABLE eveindy.assets (
  charID integer NOT NULL,
  apikey integer NOT NULL,
  itemID bigint NOT NULL,
  locationID bigint NOT NULL,
  stationID integer NOT NULL,
  typeID integer NOT NULL,
  quantity integer NOT NULL,
  flag integer NOT NULL,
  unpackaged boolean NOT NULL,
  FOREIGN KEY (charID) REFERENCES eveindy.characters (id)
    ON DELETE CASCADE DEFERRABLE,
  FOREIGN KEY (apikey) REFERENCES eveindy.apikeys (id)
    ON DELETE CASCADE DEFERRABLE,
  FOREIGN KEY (typeID) REFERENCES "invTypes" ("typeID") DEFERRABLE,
  FOREIGN KEY (flag) REFERENCES "invFlags" ("flagID") DEFERRABLE,
  CHECK (quantity > 0)
);

-- trigger on insert: ensure location is valid
CREATE OR REPLACE FUNCTION assets_insert_check() RETURNS TRIGGER AS $$
DECLARE
  location bigint;
  parent bigint;
BEGIN
  -- Ensure containing station or solar system exists.
  SELECT COALESCE(
    (SELECT "stationID" from eveindy.allstations
    WHERE  "stationID" = NEW.stationID),
    (SELECT "solarSystemID" from "mapSolarSystems"
    WHERE  "solarSystemID" = NEW.stationID)
  )
  INTO location;

  IF location IS NULL
  THEN
    RAISE EXCEPTION 'location ID % is invalid', NEW.stationID;
  END IF;
  -- Ensure parent container exists
  IF NEW.locationID <> NEW.StationID
  THEN
    SELECT locationID from eveindy.assets
    WHERE  locationID = NEW.locationID
    INTO parent;
    IF parent IS NULL
    THEN
      RAISE EXCEPTION 'parent item ID % is invalid', NEW.locationID;
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER assets_insert AFTER INSERT OR UPDATE ON eveindy.assets
INITIALLY DEFERRED
FOR EACH ROW EXECUTE PROCEDURE assets_insert_check();
