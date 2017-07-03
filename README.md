# Matchgrab: League of Legends match importer

Matchgrab retrieves League of Legends match data from [Riot's API](https://developer.riotgames.com). It starts with a seed user, starts downloading match data for that user, and then performs the same process for other summoners who played in the downloaded matches. It continues this process until you stop it.

You must have an [API key issued by Riot](https://developer.riotgames.com/) in order to use this tool, and you will only be able to download data at the rates that Riot allows. 

## Collecting data

To use matchgrab, first check out the configuration file to make sure you're happy with all defaults; one thing you'll definitely need to change is to provide your [Riot-issued API key](https://developer.riotgames.com/) in the `riot_api_key` field. You'll likely also want to specify where match data should be saved using the `match_store_location` field.

Once you're happy with the configuration, you can run the tool by executing the `grab` command. Note that this is a command-line tool and should be run from your system's terminal or command line.

## Config options

Matchgrab is designed to be simple when it can be, but there are a few configuration options that are important to be aware of.

```
http_timeout            : how long all network requests receive before timing out

match_store_location    : where downloaded match data should be stored

seed_account            : initial account ID to start from if no data is available

max_sim_requests        : the maximum number of simultaneous requests that will ever be attempted

requests_per_min        : the maximum number of requests that will be attempted in a minute

max_time_ago            : ignore matches beyond this age (default ~2 months)

riot_api_key            : copy and paste your API key from Riot here; used for all requests that require one
```

## Accessing data
Matchgrab records data to a [LevelDB database](https://github.com/google/leveldb) that contains the data described in [`structs.Match`](https://github.com/anyweez/matchgrab/blob/master/structs/match.go). The database will not contain any duplicate matches. At this point you can only read data using Golang because of the dependency on Go's `gob` library.

next version:

Matchgrab records data to a [LevelDB database](https://github.com/google/leveldb) that contains the data described in [Match.proto](https://github.com/anyweez/matchgrab/blob/master/Match.proto). In order to read the data, you'll need to find some libraries in your language of choice that allow you to read LevelDB databases and then decode the data stored there (encoded using [Google's protocol buffers](https://developers.google.com/protocol-buffers/)). A few recommendations include:

- For **Python**: [leveldb-py](https://github.com/jtolds/leveldb-py) and Google's [protobuf library](https://github.com/google/protobuf)
- For **Java**: [leveldbjni](https://github.com/fusesource/leveldbjni) and Google's [protobuf library](https://github.com/google/protobuf)
- For **Javascript**: [levelup](https://github.com/Level/levelup) and [protobuf.js](https://github.com/dcodeIO/ProtoBuf.js/)

## Statistical sampling

Note that the retrieval process described above will naturally be biased because it will download matches in 'neighborhoods' of players and will likely never provide a randomized sample of the League population, even distributions per tier, Challenger-only games, etc; you should plan on taking care of selecting relevant matches yourself once you've got the raw data.