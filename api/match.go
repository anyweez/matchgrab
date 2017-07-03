package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/anyweez/matchgrab/config"
	"github.com/anyweez/matchgrab/utils"
)

// Default amount of time to wait if Riot doesn't tell us.
const DefaultWaitSeconds = 30

// Get : Make a request to the Riot API and call the specified function on success. If anything
// goes wrong with the request the first return value will provide more information. If the error
// is rate limit related, the second value will be the number of seconds you should wait before
// attempting another request. If the error is not rate limit-related then the value of the second
// argument will always be 0.
func Get(url string, cb func(body []byte)) (error, int) {
	client := http.Client{
		Timeout: config.Config.HTTPTimeout,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.Log(err.Error())
		return err, 0
	}
	request.Header.Set("X-Riot-Token", os.Getenv("RIOT_API_KEY"))

	resp, err := client.Do(request)

	// Note: rate limits (420 or 429) don't cause errors but need to be handled (see below).
	if err != nil {
		return nil, 0
	}
	defer resp.Body.Close()

	// Check for rate limit warnings from Riot. There are two potential headers that they
	// may send, detailed here:
	//
	//    https://developer.riotgames.com/rate-limiting.html
	//
	// We'll pause automatically any time they send `X-Rate-Limit-Type` with any value. Ideally
	// they tell us how long to pause and we'll follow that instruction. Otherwise we'll wait for
	// a while and try again later.
	if resp.Header.Get("X-Rate-Limit-Type") != "" {
		retryAfter := resp.Header.Get("Retry-After")

		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)

			if err != nil {
				return errors.New("Rate limit exceeded; pausing..."), DefaultWaitSeconds
			}

			return errors.New("Rate limit exceeded; pausing..."), seconds
		}

		return errors.New("Rate limit exceeded; pausing..."), DefaultWaitSeconds
	}

	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, 0
	}

	cb(raw)

	return nil, 0
}
