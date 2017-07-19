package structs

import (
  "io/ioutil"
  "encoding/json"
  "math"
  "math/rand"
)

func randRiotID() RiotID {
  return RiotID(rand.Int63n(math.MaxInt64))
}

func fakeMatch() Match {
  return Match{
    GameID: randRiotID(),
    SeasonID: int(rand.Int31()),
    GameCreation: rand.Int63n(math.MaxInt64),
    GameDuration: int(rand.Int31()),
  }
}

func rawSamples() []APIMatch {
  files, _ := ioutil.ReadDir("../sample")
  matches := make([]APIMatch, 0, len(files))

  for _, file := range files {
    raw, _ := ioutil.ReadFile("../sample/" + file.Name())

    var match APIMatch

    json.Unmarshal(raw, &match)
    matches = append(matches, match)
  }

  return matches
}
