package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElkSearchResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     json.RawMessage `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

// QueryLogs uses the ELK search API to fetch a recent
// number of logs for the index sent via indexName.
// Valid indicies are listed in the Kibana UI such as
// nginx, worker, tink-server
func QueryLogs(indexName, elkHost string) (string, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{elkHost},
	}

	es, _ := elasticsearch.NewClient(cfg)

	res, err := es.Search(es.Search.WithIndex(indexName))

	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := ioutil.ReadAll(res.Body)
	searchRes := ElkSearchResponse{}
	err = json.Unmarshal(body, &searchRes)
	if err != nil {
		return "", err
	}

	if len(searchRes.Hits.Hits) < 1 {
		return "No results found in index\n", nil
	}

	out := ""
	for _, hit := range searchRes.Hits.Hits {
		var s Source
		if err := json.Unmarshal(hit.Source, &s); err != nil {
			log.Printf("Unmarshal error: %s\n", err)
			continue
		}

		out = out + "[" + s.Container + "] " + s.Log + "\n"

	}

	return out, nil
}

type Source struct {
	Timestamp *time.Time `json:"@timestamp"`
	Log       string     `json:"log"`
	Container string     `json:"container_name"`

	// "@timestamp": "2020-06-01T19:37:49.000Z",
	// "log": "time=\"2020-06-01T19:37:49.611479907Z\" level=warning msg=\"error authorizing context: basic authentication challenge for realm \"Registry Realm\": invalid authorization credential\" go.version=go1.11.2 http.request.host=192.168.1.1 http.request.id=42d9b7e2-4479-461f-848d-9ecde7ba05d6 http.request.method=GET http.request.remoteaddr=\"192.168.1.1:55230\" http.request.uri=\"/v2/\" http.request.useragent=\"docker/19.03.9 go/go1.13.10 git-commit/9d988398e7 kernel/4.15.0-101-generic os/linux arch/amd64 UpstreamClient(Docker-Client/19.03.9 \\(linux\\))\" ",
	// "container_id": "845742670c02d58ede72aa0759ef0a5b50af0006aa4678a0fb2d387be60af209",
	// "container_name": "/deploy_registry_1",
	// "source": "stderr"
}
