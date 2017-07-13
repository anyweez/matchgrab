# Matchgrab: League of Legends match importer

[![Build Status](https://travis-ci.org/anyweez/matchgrab.svg?branch=master)](https://travis-ci.org/anyweez/matchgrab)
[![Coverage Status](https://coveralls.io/repos/github/anyweez/matchgrab/badge.svg)](https://coveralls.io/github/anyweez/matchgrab)

Matchgrab retrieves League of Legends match data from [Riot's API](https://developer.riotgames.com) and stores it locally. It starts with a seed user, starts downloading match data for that user, and then performs the same process for other summoners who played in the downloaded matches. It continues this process until you stop it.

You must have an [API key issued by Riot](https://developer.riotgames.com/) in order to use this tool, and you will only be able to download data at the rates that Riot allows. The default values in the configuration file should be safe for development keys.

## Collecting data

Check out the configuration file to make sure you're happy with all defaults; one thing you'll definitely need to change is to provide your [Riot-issued API key](https://developer.riotgames.com/) in the `riot_api_key` field. You'll likely also want to specify where match data should be saved using the `match_store_location` field as well.

Once you're happy with the configuration, you can run the tool by executing the `grab` command. Note that this is a command-line tool and should be run from your system's terminal or command line.

## Config options

Matchgrab is designed to be simple when it can be, but there are a few configuration options that are important to be aware of.

```
http_timeout            : how long all network requests receive before timing out (seconds)

match_store_location    : where downloaded match data should be stored (relative or absolute)

seed_account            : initial account ID to start from if no existing data is available

max_sim_requests        : the maximum number of simultaneous requests allowed

requests_per_min        : the maximum number of requests per minute (note Riot's rate limits)

max_time_ago            : ignore matches beyond this age (default ~2 months)

riot_api_key            : copy and paste your API key from Riot here; used for all requests that require one
```

## Accessing data

Matchgrab records data to a [LevelDB database](https://github.com/google/leveldb) that contains the data described in [Match.proto](https://github.com/anyweez/matchgrab/blob/master/Match.proto). In order to read the data, you'll need to find some libraries in your language of choice that allow you to read LevelDB databases and then decode the data stored there (encoded using [Google's protocol buffers](https://developers.google.com/protocol-buffers/)). A few recommendations include:

- For **Python**: [leveldb-py](https://github.com/jtolds/leveldb-py) and Google's [protobuf library](https://github.com/google/protobuf)
- For **Java**: [leveldbjni](https://github.com/fusesource/leveldbjni) and Google's [protobuf library](https://github.com/google/protobuf)
- For **Javascript**: [levelup](https://github.com/Level/levelup) and [protobuf.js](https://github.com/dcodeIO/ProtoBuf.js/)

## Rate limits

You are solely responsible for ensuring that you don't violate Riot's rate limits. Matchgrab will respect Riot's headers if their responses indicate that you have exceeded your rate limit, but it's still possible to get banned if you aren't careful. Again, **do not leave this or any other program running with your API key** until you can ensure that it's fetching data at an acceptable pace.

## Statistical sampling

Note that the retrieval process described above will naturally be biased because it will download matches in 'neighborhoods' of players and will likely never provide a randomized sample of the League population, even distributions per tier, Challenger-only games, etc; you should plan on taking care of selecting relevant matches for your use case yourself once you've got the raw data.

## License and contributions

This code is available under the MIT license. See the attached LICENSE.txt for more information.

Contributions are welcome. Please file issues or create PR's.
