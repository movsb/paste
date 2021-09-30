# Paste

A dead-simple HTTP-based cross-device text clipboard/syncer/paste bin.

## Docker

```shell-session
$ docker run -d taocker/paste:latest
```

## Browser

Open <http://localhost:7962>
from multiple browsers you can see the same thing.
They are automatically synchronized every 1 second.

## Terminal

To get the content:

```shell-session
$ curl localhost:7962
```

To update the content:

```shell-session
$curl -X POST localhost:7962 --data-binary @file.txt
```

## Multiple Paths

Different URL path serves different content.
You can append arbitrary path to url to get a fresh new content.

* localhost:7962
* localhost:7962/1
* localhost:7962/share/a.txt

## How sync works

It's dead simple:
whoever sends an update request to the paste server latest wins the race.

So, it's *not* an alternative to any collaborate-editing tool.

It's just for you, and me.

## Others

* If a content is not accessed within 24h, it will be removed.
* The maximum content size is 1 megabytes.
* The default port is 7962, but you can change it with `-p` flag.
