# EVE industrial thingy

## Requirements

This software requires:
* Node.js v0.12
* PostgreSQL v9.3 or higher
* Redis (version?)
* A CREST API key.

It is tested on Linux, FreeBSD, and OS X. Windows is untested but should work.

## Installation

### Database

eveindy requires two PostgreSQL databases:

- one containing the [SDE dump][dump] as provided by [Steve Ronuken][steve]
(by default, `evetool`), and
- one for user data (`eveindy`).

[dump]: https://www.fuzzwork.co.uk/dump/
[steve]: https://www.fuzzwork.co.uk/

Create a new database for the SDE dump using the instructions provided in
[evego][evego]. Grant the database user to be used by eveindy read permission
on everything in this database. (Of course, it's all public...)

[evego]: https://github.com/backerman/evego

```
BEGIN;
CREATE ROLE eveindy WITH
  LOGIN
  PASSWORD 'correct horse battery staple';

REVOKE ALL PRIVILEGES ON DATABASE evetool FROM public;
REVOKE ALL PRIVILEGES ON SCHEMA public FROM public;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM public;
GRANT CONNECT ON DATABASE evetool TO eveindy;
GRANT USAGE ON SCHEMA public TO eveindy;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO eveindy;
COMMIT;
```

Create a database for eveindy's configuration, users, etc.

### Cache (Redis)

## Operation & Maintenance

## License

The contents of this repository are © 2014–5 Brad Ackerman and licensed under
the [Apache License 2.0][apache], the full text of which is in the LICENSE file.

[apache]: http://www.apache.org/licenses/LICENSE-2.0
