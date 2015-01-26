package main

import (
	"encoding/json"
	"fmt"
	"github.com/streamrail/concurrent-map"
	"log"
	"net/http"
	"os"
	"regexp"
)

var tokens = cmap.New()

func getAuth(token string) (string, bool) {
	tmp, ok := tokens.Get(token)
	if ok {
		userId := tmp.(string)
		return userId, ok
	} else {
		return "", ok
	}
}

func setAuth(token string, userId string) {
	tokens.Set(token, userId)
}

func handler(plugins map[string]int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reg := regexp.MustCompile(`/([^/]+)(/.*)`)
		result := reg.FindStringSubmatch(r.RequestURI)
		if len(result) < 3 {
			w.WriteHeader(404)
			return
		}
		subPath := result[1]
		tailingPath := result[2]
		if port, ok := plugins[subPath]; ok {
			if subPath != "platform" {
				cookie, err := r.Cookie("Auth")
				if err != nil {
					w.WriteHeader(401)
					return
				}
				value, ok := getAuth(cookie.Value)
				if ok {
					r.Header.Set("USER", value)
				} else {
					w.WriteHeader(401)
					return
				}
			}
			client := &http.Client{}
			r.RequestURI = ""
			r.Host = fmt.Sprintf("localhost:%d", port)
			r.URL.Scheme = "http"
			r.URL.Host = r.Host
			r.URL.Path = tailingPath
			resp, err := client.Do(r)
			if err != nil {
				log.Println(err)
				w.WriteHeader(502)
				return
			}
			resp.Write(w)
			return
		} else {
			w.WriteHeader(404)
			return
		}
	}
}

type authentication struct {
	Token  string
	UserId string
}

func auth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			decoder := json.NewDecoder(r.Body)
			var authInfo authentication
			err := decoder.Decode(authInfo)
			if err != nil {
				w.WriteHeader(400)
				return
			} else {
				setAuth(authInfo.Token, authInfo.UserId)
				if err != nil {
					w.WriteHeader(401)
					return
				} else {
					w.WriteHeader(200)
					return
				}
			}
		}
	}
}

func main() {
	file, er := os.Open("plugins.json")
	if er != nil {
		log.Fatal(er)
	}
	decoder := json.NewDecoder(file)
	plugins := map[string]int{}
	err := decoder.Decode(&plugins)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/auth", auth())
	http.HandleFunc("/", handler(plugins))
	http.ListenAndServe(":8000", nil)
}