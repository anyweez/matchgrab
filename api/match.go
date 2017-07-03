package api

import (
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

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	cb(raw)

	return nil
}
