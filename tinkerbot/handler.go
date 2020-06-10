package function

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/alexellis/tinkerbot/tinkerbot/cmd"
	"github.com/openfaas/openfaas-cloud/sdk"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	var query *url.Values
	if len(input) > 0 {
		q, err := url.ParseQuery(string(input))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query = &q
	}

	if val, ok := os.LookupEnv("validate"); ok && val == "true" {
		token, err := sdk.ReadSecret("validation-token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tokenSent := query.Get("token")
		if token != tokenSent {
			http.Error(w, fmt.Sprintf("Token: %s, invalid", tokenSent), http.StatusUnauthorized)
			return
		}
	}

	command := query.Get("command")
	text := query.Get("text")

	os.Stderr.Write([]byte(fmt.Sprintf("debug - command: %q, text: %q\n", command, text)))

	headerWritten := processCommand(w, r, command, text)

	if !headerWritten {
		http.Error(w, "Nothing to do", http.StatusBadRequest)
	}
}

func processCommand(w http.ResponseWriter, r *http.Request, command, text string) bool {
	if len(command) > 0 {

		switch command {
		case "/logs":
			if len(text) == 0 {
				w.Write([]byte("Please give an index from your ELK dashboard\n"))
				w.WriteHeader(http.StatusOK)
				return true
			}

			elkHost := os.Getenv("elk_host")
			logs, err := cmd.QueryLogs(text, elkHost)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return true
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(logs))
			return true
		}
	}
	return false
}
