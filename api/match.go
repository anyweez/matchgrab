package api

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"

	"github.com/anyweez/matchgrab/config"
	"github.com/anyweez/matchgrab/utils"
)

// Get : Make a request to the Riot API and call the specified function on success.
func Get(url string, cb func(body []byte)) error {
	client := http.Client{
		Timeout: config.Config.HTTPTimeout,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.Log(err.Error())
		return err
	}
	request.Header.Set("X-Riot-Token", config.Config.RiotAPIKey)
	request.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Decode the body first if its gzip'd.
	src := resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		src, _ = gzip.NewReader(src)
	}

	raw, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}

	cb(raw)

	return nil
}
