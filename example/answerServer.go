package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"roob.re/tcache"
	"time"
)

func main() {
	// Create new cache
	hashCache := tcache.New(tcache.NewMapStorage())

	// Simple HTTP handler, which receives a question and answers with yes/no
	// Since figuring out the answer takes a long time, but many people want to request "Am I loved?",
	// we cache the answers using tcache
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		je := json.NewEncoder(rw)
		rw.Header().Add("content-type", "application/json")

		question := r.FormValue("q")
		if question == "" {
			je.Encode(map[string]string{"error": "you need to specify a question using ?q=Question"})
			return
		}

		type Answer struct {
			Question string
			Anwser   string // yes or no
		}

		// Find "passowrd" in collection "passwords"
		err := hashCache.Access(question, 8*time.Hour, tcache.Handler{
			Then: func(cacheReader io.Reader) error {
				// Found in cache stored in json, dump it
				_, err := io.Copy(rw, cacheReader)
				return err
			},
			Else: func(cacheWriter io.Writer) error {
				// Think about the answer very carefully...
				time.Sleep(4 * time.Second)
				var answerstr string
				if rand.Int()%2 == 0 {
					answerstr = "Yes!"
				} else {
					answerstr = "No :("
				}

				a := Answer{
					Question: question,
					Anwser:   answerstr,
				}

				// Store answer in cache and reply at the same time
				err := json.NewEncoder(io.MultiWriter(rw, cacheWriter)).Encode(&a)
				return err
			},
		})

		if err != nil {
			log.Println(err)
			return
		}
	})

	http.ListenAndServe(":8888", http.DefaultServeMux)
}
