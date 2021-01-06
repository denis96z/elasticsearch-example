package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func main() {
	if len(os.Args) != 7 {
		log.Fatalf("app -p BIND_PORT --esh HOST --esp PORT")
	}

	bindPort := os.Args[2]

	esHost := os.Args[4]
	esPort := os.Args[6]

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("http://%s:%s", esHost, esPort),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		evt := r.URL.Query().Get("event")
		uid := r.URL.Query().Get("user_id")
		if evt == "" || uid == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, err := strconv.ParseUint(uid, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		req := esapi.IndexRequest{
			Index:      "event",
			Body:       strings.NewReader(fmt.Sprintf(
				`{"event":%q,"user_id":%d}`, evt, userID,
			)),
		}

		res, err := req.Do(context.TODO(), es)
		if err != nil {
			log.Printf("Error getting response: %s", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			_ = res.Body.Close()
		}()

		if res.IsError() {
			log.Printf("[%s] Error indexing document", res.Status())

			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			var r map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				log.Printf("Error parsing the response body: %s", err)

				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				log.Printf("OK: %v", r)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	_ = http.ListenAndServe(":" + bindPort, nil)
}
