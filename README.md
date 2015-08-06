Legacy
======

Legacy is a utility for uploading Cassandra snapshots and incremental backups to S3.

Light-weight, simple tool that can be run on individual nodes.

It supports backing up snapshots and incremental backups, as well as tidying up incremental backups as it goes along.

Make sure to get the latest version from the [Releases Section](https://github.com/iamthemovie/legacy/releases) rather than master (which should be treated as unstable just in-case).

**Notes:**

Used in production [Cassandra 2.1+] in its current state - stable enough and provides the necessary functionality to keep a clean consistent backup of a multi-node, multi-terrabyte cluster.

Assuming the nodetool snapshot has --tags and the directory structure of the data/snapshots/backups directory is the same on 2.0 it should work.

See the to-do's below for a more descriptive list of features that Version 1 should / will comprise of.

- Memory - there is a possible memory leak in which during the upload of a 400GB set of SSTables memory peaked at ~ 1GB.

- CPU performance - during the uploads, Legacy can hog CPU (unsuprisingly) during 400GB uploads temporary spikes of 40% on one core (of 8). This didn't noticeably impact performance of the node though.

Usage
-----

```bash
legacy \
 -aws-secret="AWS SECRET" \
 -aws-access-key="AWS SECRET KEY" \
 -s3-bucket="S3-BUCKET-NAME" \
 -s3-base-path="BASE PATH" \
 -directories="/path-to-data-dir-1,/path-to-data-dir-2"
```


- Legacy will create a snapshot as soon as it is run via ```nodetool snapshot --tag legacy-{timestamp}```

- Checks if an initial snapshot has been created for each table, if not it will upload the snapshot made at initial execution of legacy.

- If a snapshot has been previously been uploaded to S3 it will upload the SSTables flushed into the incremental backups folder.
(it will delete the SSTable once successfully uploaded to S3)

Internals
---------

Legacy uses two S3 libraries, as the first doesn't support streaming files (rather than reading the entire contents into memory before hand).

#### Directory Structure ####

Legacy stores data in S3 per node like so:

`/s3-bucket/s3-base-path/{NODE-HOSTNAME}`

Legacy stores JSON files with small amounts of information to determine whether a snapshot has already been made for a table in:

`/s3-bucket-/s3-base-path/{NODE-HOSTNAME}/.legacy`

Backups are stored in the snapshot directory as to make it easier to restore if the need arises.

To do
------

Current tasks:

- Selecting specific keyspaces (rather than all of them)

- Request a brand new snapshot (via a cli option)

- Reduce Memory foot-print

- CPU + Network throttling options

- Built-in SSH / remote management support such as `legacy [options (inclusive of AWS keys)] --nodes=host1,host2`.

- Add cli option to control the SSTable deletion of incrementals (already written in just needs exposing through cli params)

- Creating a set of backup management methods - list, delete, info. These will give detailed information on the backups stored in S3.

- Backup compression (LZOP / Snappy)

- Some kind of automatic threshold management that will monitor flushed SSTables and compute the compaction spaced required. If it's over a certain threshold it will upload a new snapshot and remove the older backup files once completed.

- Better error handling / code structure / log support

Author
------

Jordan Appleson ([@jordanisonfire](https://twitter.com/jordanisonfire))

License
-------

Copyright (c) 2015 Jordan Appleson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
