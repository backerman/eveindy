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
  email text,
  authToon integer NOT NULL,
  authToonName text
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
  apikey integer NOT NULL REFERENCES eveindy.apikeys(id) ON DELETE CASCADE,
  name text NOT NULL,
  id integer NOT NULL,
  corp text NOT NULL,
  corpid integer NOT NULL,
  alliance text,
  allianceid integer
);

-- sessions: SSO sessions
CREATE UNLOGGED TABLE eveindy.sessions (
  id SERIAL PRIMARY KEY,
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
