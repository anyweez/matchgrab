package structs

import (
  "io/ioutil"
  "os"
  "testing"
  "time"
)

// Make sure Count() increases when we add items.
func TestValidCount(t *testing.T) {
  countSize := 50

  dir, _ := ioutil.TempDir("", "test")

  defer os.RemoveAll(dir)
  defer os.RemoveAll(dir + SnapshotSuffix)

  store := NewMatchStore(dir)
  defer store.Close()

  for i := 0; i < countSize; i++ {
    store.Add(Match{
      GameID: RiotID(i),
    })
  }

  time.Sleep(1 * time.Second)

  if store.Count() != countSize {
    t.Fail()
  }
}

// Make sure we can retrieve items that we've added.
func TestRetrieval(t *testing.T) {
  countSize := 50

  dir, _ := ioutil.TempDir("", "test")

  defer os.RemoveAll(dir)
  defer os.RemoveAll(dir + SnapshotSuffix)

  store := NewMatchStore(dir)
  defer store.Close()

  for i := 0; i < countSize; i++ {
      store.Add(Match{
        GameID: RiotID(i),
      })
  }

  time.Sleep(1 * time.Second)

  // TODO: is retrieval order guaranteed here? Need to check out LevelDB guarantees.
  expectedNext := 0
  store.Each(func (m *Match) {
    if m.GameID != RiotID(expectedNext) {
      t.Fail()
    }

    expectedNext++
  })
}
