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
  FOREIGN KEY (typeID) REFERENCES "invTypes" ("typeID"),
  FOREIGN KEY (flag) REFERENCES "invFlags" ("flagID"),
  CHECK (quantity > 0)
);

-- trigger on insert: ensure station is valid
-- can we guarantee everything has a valid container? depends on insert order.
CREATE OR REPLACE FUNCTION assets_insert_check() RETURNS TRIGGER AS $$
DECLARE
  station integer;
  parent integer;
BEGIN
  -- Ensure containing station exists.
  SELECT "stationID" from eveindy.allstations
  WHERE  "stationID" = NEW.stationID
  INTO station;
  IF station IS NULL
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
FOR EACH ROW EXECUTE PROCEDURE blueprints_insert_check();
