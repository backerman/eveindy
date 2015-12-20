# EVE industrial thingy

## Status

Alpha. More alpha than a full fleet of Catalysts, in fact.

## Requirements

This software requires:
* Node.js v0.12
* PostgreSQL v9.3 or higher
* Redis (developed using v3.0)
* A CREST API key.

It is tested on Linux, FreeBSD, and OS X. Windows is untested but should work.

## Installation

### Database

eveindy currently requires two PostgreSQL databases:

- one containing the [SDE dump][dump] as provided by [Steve Ronuken][steve]
(by default, `evetool`), and
- one for user data (`eveindy`).

I'll probably end up using two schemas in one database, but that hasn't happened yet.

[dump]: https://www.fuzzwork.co.uk/dump/
[steve]: https://www.fuzzwork.co.uk/

This application will require a user to connect to the database.

```
CREATE ROLE eveindy WITH
  LOGIN
  PASSWORD 'correct horse battery staple';
-- or a different authentication method.
```

Create a new database for the SDE dump using the instructions provided in
[evego][evego]. Grant the database user to be used by eveindy read permission
on everything in this database. (Yes, I know this part is all public data.)

[evego]: https://github.com/backerman/evego

```
BEGIN;

REVOKE ALL PRIVILEGES ON DATABASE evetool FROM public;
REVOKE ALL PRIVILEGES ON SCHEMA public FROM public;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM public;
GRANT CONNECT ON DATABASE evetool TO eveindy;
GRANT USAGE ON SCHEMA public TO eveindy;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO eveindy;

COMMIT;
```

Create a new database (we use `eveindy`) and run the provided SQL (in the `sql`
directory) to set it all up. You'll then need to configure this database's
security using something like the following:

```
BEGIN;

REVOKE ALL PRIVILEGES ON DATABASE eveindy FROM public;
REVOKE ALL PRIVILEGES ON SCHEMA public FROM public;
REVOKE ALL PRIVILEGES ON SCHEMA eveindy FROM public;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM public;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA eveindy FROM public;
GRANT CONNECT ON DATABASE eveindy TO eveindy;
GRANT USAGE ON SCHEMA eveindy TO eveindy;
GRANT SELECT ON ALL TABLES IN SCHEMA eveindy TO eveindy;

COMMIT;
```

### Cache (Redis)

## Operation & Maintenance

To update the local SDE copy, use the `sql/update_sde.py` script, e.g.:

```
update_sde.py /tmp/latest.dmp.bz2 myschema | psql -h somewhere mydatabase
```

## License

The contents of this repository are © 2014–5 Brad Ackerman and licensed under
the [Apache License 2.0][apache], the full text of which is in the LICENSE file.

[apache]: http://www.apache.org/licenses/LICENSE-2.0
