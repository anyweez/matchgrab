package api

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"testing"
	"time"

	"bytes"

	"os"

	"github.com/anyweez/matchgrab/config"
)

var nextPort = 12001

type imitationResponse func(*http.ResponseWriter, *http.Request, []byte) []byte

// Imitate the server's response and check the result of Get()
func imitateServer(fn imitationResponse, chk func(e error, c int), t *testing.T) {
	os.Setenv("RIOT_API_KEY", "abcde")
	config.Setup()

	port := nextPort
	nextPort++

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body := fn(&w, r, []byte("success"))

			fmt.Println(fmt.Sprintf("responding with %d bytes on port %d", len(body), port))
			w.Write(body)
		}),
	}

	go func() {
		time.Sleep(250 * time.Millisecond)

		// TODO: need to eval certain errors
		chk(Get(fmt.Sprintf("http://localhost:%d", port), func(r []byte) {
			fmt.Println(fmt.Sprintf("Received response on %d", port))

			if string(r) != "success" {
				t.Fail()
			}

		}))

		srv.Shutdown(nil)
	}()

	srv.ListenAndServe()
}

func noop(e error, c int) {}

// TestBasicRequest : Make sure a basic request works (no gzip)
func TestBasicRequest(t *testing.T) {
	imitateServer(func(w *http.ResponseWriter, r *http.Request, s []byte) []byte {
		return s
	}, noop, t)
}

// TestGzip : Ensure that Get() correctly decodes gzip-encoded information.
func TestGzip(t *testing.T) {
	imitateServer(func(w *http.ResponseWriter, r *http.Request, msg []byte) []byte {
		(*w).Header().Add("Content-Encoding", "gzip")

		var buf bytes.Buffer
		zip := gzip.NewWriter(&buf)
		zip.Write(msg)

		zip.Close()

		return buf.Bytes()
	}, noop, t)
}

// TestRiotKey : Ensures that something is sent in the X-Riot-Token header.
func TestRiotKey(t *testing.T) {
	imitateServer(func(w *http.ResponseWriter, r *http.Request, msg []byte) []byte {
		if r.Header.Get("X-Riot-Token") != "abcde" {
			t.Fail()
		}

		return msg
	}, noop, t)
}

// TestRateLimitWithRetry : Ensure we get an appropriate retry timer if specified by the server.
func TestRateLimitWithRetry(t *testing.T) {
	imitateServer(func(w *http.ResponseWriter, r *http.Request, msg []byte) []byte {
		(*w).Header().Add("X-Rate-Limit-Type", "100")
		(*w).Header().Add("Retry-After", "100")

		return msg
	}, func(e error, c int) {
		if c != 100 {
			t.Fail()
		}
	}, t)
}

// TestRateLimitNoRetry : Ensure we get an appropriate retry timer if notspecified by the server.
func TestRateLimitNoRetry(t *testing.T) {
	imitateServer(func(w *http.ResponseWriter, r *http.Request, msg []byte) []byte {
		(*w).Header().Add("X-Rate-Limit-Type", "100")

		return msg
	}, func(e error, c int) {
		if c != DefaultWaitSeconds {
			t.Fail()
		}
	}, t)
}
