# trex
[![goreport](https://goreportcard.com/badge/github.com/ziscky/trex)](https://goreportcard.com/report/github.com/ziscky/doom)
[![Build Status](https://travis-ci.org/ziscky/doom.svg?branch=master)](https://travis-ci.org/ziscky/doom)

*Under Heavy Development*

A Batteries included extensible torrent provider platform built on top of https://github.com/ziscky/Taipei-Torrent making it easy to share content through the BitTorrent protocol.

### Features

 - Torrent MetaInfo generator
 - Private Tracker
 - Initial and Supportive Seeders
 - Detailed introspection API endpoints
 - Minimal dependencies

 
### Getting Started

Basic directory structure
----------
```
Content/ - where content to be shared is stored.(files or folders)
Torrents/ - where torrent metainfo files (.torrent) will be stored
registry - file that ensures dynamic behavior of seeders and tracker when torrents are added or deleted.
```

Setup Instructions
----------
 1. Create directory where you will store content (anywhere in the filesystem as long as you have valid permissions) to be shared ( name anything you like)
 2. Create directory .torrent files will be stored (anywhere in the filesystem as long as you have valid permissions) to be shared ( name anything you like)
 3. Create an empty file, this will be the torrent registry

Running Instructions
----------
```
1.Generate .torrent files and populate registry
$./trex-content -content=/absolute/path/to/contentfolder -torrent=/absolute/path/to/torrentfolder -addr=addr:port -registry=/absolute/path/to/torrentregistry

//addr - address of tracker to be writtent into the .torrent file, exclude the protocol

2. Run the tracker
$./trex-tracker -registry=/absolute/path/to/registryfile addr=addr:port 

3. Run the initial seeders
$./trex-seeder -numSeeds=20 -registry=absolute/path/to/registryfile -port=18000

```
