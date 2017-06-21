# Matchgrab: League of Legends match importer

Matchgrab retrieves League of Legends match data from [Riot's API](https://developer.riotgames.com). It started with a seed user, downloads all match data, then begins searching for matches from summoners that participated in these games. The process continues for as long as you let the application run.

You must have an API key issues by Riot in order to use this tool. If you have Go installed, you can run with:

```
RIOT_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx go run grab.go
```

If not, I'll have a binary available under [Releases](https://github.com/anyweez/matchgrab/releases) soon.

Matchgrab records data to a [LevelDB database](https://github.com/google/leveldb) that contains the data described in [`structs.Match`](https://github.com/anyweez/matchgrab/blob/master/structs/match.go). The database will not contain any duplicate matches. At this point you can only read data using Golang because of the dependency on Go's `gob` library.

## Statistical sampling

Note that the retrieval process described above will naturally be biased because it will download matches in 'neighborhoods' of players and will likely never provide a randomized sample of the League population, even distributions per tier, Challenger-only games, etc; you should plan on taking care of this yourself once you've got the raw data.

If you want to mitigate the effect of 'neighborhoods' you can occassionally restart the tool; the order of summoners
is randomized when the application is started.