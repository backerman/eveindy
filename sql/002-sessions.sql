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

-- Session-related functions/views.

-- pgcrypto is required for random-number generation.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- eveindy.token_expiry: calculate a token's expiry time, assuming that it was
-- just issued.
CREATE OR REPLACE FUNCTION eveindy.tokenExpiry(
  token jsonb
) RETURNS timestamp with time zone AS $$
DECLARE
  valid_duration interval;
BEGIN
  valid_duration := make_interval(secs := (token ->> 'expires_in') :: int);
  RETURN CURRENT_TIMESTAMP + valid_duration;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- okay...
-- find a session... if it exists.
CREATE OR REPLACE FUNCTION eveindy.getSession(sessionid text)
RETURNS eveindy.sessions AS $$
DECLARE
  session eveindy.sessions;
BEGIN
  SELECT * FROM eveindy.sessions WHERE cookie = sessionid INTO session;
  IF session IS NULL
  THEN
    -- Need to get a new one.
    RETURN eveindy.newSession();
  ELSE
    UPDATE eveindy.sessions
       SET lastSeen = CURRENT_TIMESTAMP
     WHERE cookie = sessionid;
    RETURN session;
  END IF;
END;
$$ LANGUAGE plpgsql;

-- new_session: Start a new session (because the client has no cookie.)
CREATE OR REPLACE FUNCTION eveindy.newSession()
RETURNS eveindy.sessions AS $$
DECLARE
  theSession eveindy.sessions;
BEGIN
  INSERT INTO eveindy.sessions( state, cookie, lastSeen )
         VALUES (eveindy.genCookie(), eveindy.genCookie(), CURRENT_TIMESTAMP)
  RETURNING * INTO theSession;
  RETURN theSession;
END;
$$ LANGUAGE plpgsql;

-- Generate a random string to use for a cookie or state.
CREATE OR REPLACE FUNCTION eveindy.genCookie()
RETURNS char(44) AS $$
BEGIN
  RETURN encode(gen_random_bytes(32), 'base64');
END;
$$ LANGUAGE plpgsql;

-- Associate a token with a session.
CREATE OR REPLACE FUNCTION eveindy.associateToken(
  aCookie text,
  jsonToken text,
  charInfo text)
RETURNS VOID AS $$
DECLARE
  myToken jsonb;
  charJson jsonb;
  charID integer;
  siteuser integer;
BEGIN
  myToken := jsonToken::json;
  charJson := charInfo::json;
  charID := (charJson ->> 'CharacterID')::integer;
  -- Do we have a site user for this toon? If not, create one.
  SELECT userid
  FROM   eveindy.characters
  WHERE  id = charID
  INTO   siteuser;
  IF siteuser IS NULL
  THEN
    -- Create a new site user and add this toon to it.
    INSERT INTO eveindy.users(email) VALUES(null)
    RETURNING id INTO siteuser;
    INSERT INTO eveindy.characters(userid, name, id)
    VALUES (siteuser, charJson ->> 'CharacterName', charID);
  END IF;
  UPDATE eveindy.sessions
  SET    token = myToken, tokenExpiry = eveindy.tokenExpiry(myToken),
         userid = siteuser
  WHERE  cookie = aCookie;
END;
$$ LANGUAGE plpgsql;
