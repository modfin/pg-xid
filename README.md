# XID for Postgres, Globally Unique ID Generator

[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/modfin/pg-xid/master/LICENSE)


###### This project is a Postgres implementation of the Go Lang library found here: [https://github.com/rs/xid](https://github.com/rs/xid)

---

## Description

`Xid` is a globally unique id generator functions. They are small and ordered.

Xid uses the *Mongo Object ID* algorithm to generate globally unique ids with a different serialization (base32) to make it shorter when transported as a string:
https://docs.mongodb.org/manual/reference/object-id/

<table border="1">
<caption>Xid layout</caption>
<tr>
<td>0</td><td>1</td><td>2</td><td>3</td><td>4</td><td>5</td><td>6</td><td>7</td><td>8</td><td>9</td><td>10</td><td>11</td>
</tr>
<tr>
<td colspan="4">time</td><td colspan="3">machine id</td><td colspan="2">pid</td><td colspan="3">counter</td>
</tr>
</table>

- 4-byte value representing the seconds since the Unix epoch,
- 3-byte machine identifier,
- 2-byte process id, and
- 3-byte counter, starting with a random value.

The binary representation of the id is compatible with Mongo 12 bytes Object IDs.
The string representation is using [base32 hex (w/o padding)](https://tools.ietf.org/html/rfc4648#page-10) for better space efficiency when stored in that form (20 bytes). The hex variant of base32 is used to retain the
sortable property of the id.

`Xid`s simply offer uniqueness and speed, but they are not cryptographically secure. They are predictable and can be *brute forced* given enough time.

## Features
- Size: 12 bytes (96 bits), smaller than UUID, larger than [Twitter Snowflake](https://blog.twitter.com/2010/announcing-snowflake)
- Base32 hex encoded by default (20 chars when transported as printable string, still sortable)
- Configuration free: there is no need to set a unique machine and/or data center id
- K-ordered
- Embedded time with 1 second precision
- Unicity guaranteed for 16,777,216 (24 bits) unique ids per second and per host/process
- Lock-free (unlike UUIDv1 and v2)

## Comparison

| Name        | Binary Size | String Size    | Features
|-------------|-------------|----------------|----------------
| [UUID]      | 16 bytes    | 36 chars       | configuration free, not sortable
| [shortuuid] | 16 bytes    | 22 chars       | configuration free, not sortable
| [Snowflake] | 8 bytes     | up to 20 chars | needs machine/DC configuration, needs central server, sortable
| [MongoID]   | 12 bytes    | 24 chars       | configuration free, sortable
| xid         | 12 bytes    | 20 chars       | configuration free, sortable

[UUID]: https://en.wikipedia.org/wiki/Universally_unique_identifier
[shortuuid]: https://github.com/stochastic-technologies/shortuuid
[Snowflake]: https://blog.twitter.com/2010/announcing-snowflake
[MongoID]: https://docs.mongodb.org/manual/reference/object-id/

## Installation

```bash
cat ./xid.sql | psql your-database
```

## Usage
Generate a `xid` as base32Hex
```sql
SELECT xid() ;

-- +--------------------+
-- |xid                 |
-- +--------------------+
-- |c96r3a88u0k0b3e6hft0|
-- +--------------------+
```

Generate a `xid` as a base32Hex at a specific time
```sql
SELECT xid, xid_time(xid) FROM(
  SELECT xid( _at => CURRENT_TIMESTAMP - INTERVAL '10 YEAR')
) a

-- +--------------------+---------------------------------+
-- |xid                 |xid_time                         |
-- +--------------------+---------------------------------+
-- |9tvgr208u0k0b3e6hgb0|2012-04-06 15:36:40.000000 +00:00|
-- +--------------------+---------------------------------+

```


Use a xid as column type

```sql

CREATE TABLE users (
    id public.xid PRIMARY KEY DEFAULT xid(),
    email text
);

INSERT INTO users (email)
VALUES
('user1@example.com'),
('user2@example.com');

SELECT * FROM users;

-- +--------------------+-----------------+
-- |id                  |email            |
-- +--------------------+-----------------+
-- |c96r9eo8u0k0b3e6hgcg|user1@example.com|
-- |c96r9eo8u0k0b3e6hgd0|user2@example.com|
-- +--------------------+-----------------+


```

Decode a `xid` as byte array 
```sql
SELECT xid_decode(xid())
-- +---------------------------------------+
-- |xid_decode                             |
-- +---------------------------------------+
-- |{98,77,178,53,8,240,40,5,141,198,140,2}|
-- +---------------------------------------+

```

Encode byte array as `xid`
```sql
SELECT xid_encode('{98,77,178,53,8,240,40,5,141,198,140,2}'::INT[]);
-- +---------------------+
-- |xid_encode           |
-- +---------------------+
-- |c96r4d88u0k0b3e6hg10 |
-- +---------------------+


```


Inspect a `xid`
```sql
SELECT id, xid_time(id), xid_machine(id), xid_pid(id), xid_counter(id) 
FROM (
    SELECT xid() id FROM generate_series(1, 3)
) a

-- +--------------------+---------------------------------+-----------+-------+-----------+
-- |id                  |xid_time                         |xid_machine|xid_pid|xid_counter|
-- +--------------------+---------------------------------+-----------+-------+-----------+
-- |c96r2ho8u0k0b3e6hfr0|2022-04-06 15:27:03.000000 +00:00|{8,240,40} |1421   |13011958   |
-- |c96r2ho8u0k0b3e6hfrg|2022-04-06 15:27:03.000000 +00:00|{8,240,40} |1421   |13011959   |
-- |c96r2ho8u0k0b3e6hfs0|2022-04-06 15:27:03.000000 +00:00|{8,240,40} |1421   |13011960   |
-- +--------------------+---------------------------------+-----------+-------+-----------+
```


## Test
To run the tests you'll need to install docker along with go. Run `go test -v ./xid_test.go`

## Licenses
The source code is licensed under the [MIT License](https://github.com/modfin/pg-xid/master/LICENSE).