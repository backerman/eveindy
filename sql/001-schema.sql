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

CREATE SCHEMA eveindy;

-- users: accounts on this system.
CREATE TABLE eveindy.users (
  id SERIAL PRIMARY KEY,
  email text
);

-- apikeys: keys used to retrieve character information from the XML API.
CREATE TABLE eveindy.apikeys (
  userid integer REFERENCES eveindy.users(id) ON DELETE CASCADE,
  label text,
  id integer NOT NULL PRIMARY KEY,
  vcode text NOT NULL
);

-- characters: EVE toons
CREATE TABLE eveindy.characters (
  userid integer NOT NULL REFERENCES eveindy.users(id) ON DELETE CASCADE,
  apikey integer REFERENCES eveindy.apikeys(id) ON DELETE CASCADE DEFERRABLE,
  name text NOT NULL UNIQUE,
  id integer NOT NULL UNIQUE,
  corp text NOT NULL,
  corpid integer NOT NULL,
  alliance text,
  allianceid integer
);

-- skills: a character's trained skills
CREATE TABLE eveindy.skills (
  charid integer NOT NULL REFERENCES eveindy.characters(id) ON DELETE CASCADE DEFERRABLE,
  id integer NOT NULL REFERENCES "invTypes"("typeID"),
  groupid integer NOT NULL REFERENCES "invGroups"("groupID"),
  level integer NOT NULL CHECK (level >= 0 AND level <= 5)
);

-- corpStandings: a character's standings with NPC corporations (before skills)
CREATE TABLE eveindy.corpStandings (
  charid integer REFERENCES eveindy.characters(id) ON DELETE CASCADE DEFERRABLE,
  corp integer REFERENCES "crpNPCCorporations"("corporationID"),
  standing float NOT NULL CHECK (standing >= -10.0 AND standing <= 10.0),
  PRIMARY KEY (charid, corp)
);

-- facStandings: a character's standings with NPC factions (before skills)
CREATE TABLE eveindy.facStandings (
  charid integer REFERENCES eveindy.characters(id) ON DELETE CASCADE DEFERRABLE,
  faction integer REFERENCES "chrFactions"("factionID"),
  standing float NOT NULL CHECK (standing >= -10.0 AND standing <= 10.0),
  PRIMARY KEY (charid, faction)
);

-- sessions: SSO sessions
CREATE UNLOGGED TABLE eveindy.sessions (
  -- state and cookie are 256 bits of random, which is 44 characters of base64.
  state char(44) NOT NULL UNIQUE,
  -- user will be null if this session hasn't authenticated yet
  userid integer REFERENCES eveindy.users(id) ON DELETE CASCADE,
  cookie char(44) NOT NULL UNIQUE,
  -- lastSeen is the last time this session's user was active.
  lastSeen timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- Just in case something isn't set to Zulu, we'll also save the timezone.
  tokenExpiry timestamp with time zone,
  token jsonb
);
