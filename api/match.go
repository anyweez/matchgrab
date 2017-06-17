package api

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/anyweez/kickoff/utils"
	"github.com/anyweez/matchgrab/config"
)

func Get(url string, cb func(body []byte)) {
	client := http.Client{
		Timeout: config.Config.HTTPTimeout,
	}

	request, err := http.NewRequest("GET", url, nil)
	// resp, err := client.Get(url)
	if err != nil {
		utils.Log(err.Error())
		return
	}
	request.Header.Set("X-Riot-Token", os.Getenv("RIOT_API_KEY"))

	resp, err := client.Do(request)
	if err != nil {
		utils.Log(err.Error())
		return
	}

	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Log(err.Error())
		return
	}

	cb(raw)
}
