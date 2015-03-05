package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"strconv"

	"github.com/streamrail/concurrent-map"
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

func handler(plugins map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("In: %s", r.RequestURI)
		reg := regexp.MustCompile(`/([^/]+)(/.*)`)
		result := reg.FindStringSubmatch(r.URL.Path)
		if len(result) < 3 {
			if r.RequestURI == "/" {
				http.Redirect(w, r, r.URL.Host+"/platform/", 301)
			} else {
				w.WriteHeader(404)
			}
			return
		}

		subPath := result[1]
		tailingPath := result[2]
		if host, ok := plugins[subPath]; ok {
			cookie, err := r.Cookie("Auth")
			flag := true
			if err != nil {
				if subPath != "platform" {
					log.Printf("Auth fail")
					w.WriteHeader(401)
					return
				} else {
					flag = false
				}
			}
			if flag {
				value, ok := getAuth(cookie.Value)
				if ok {
					r.Header.Set("X-USER", value)
				} else if subPath != "platform" {
					log.Printf("Auth fail")
					w.WriteHeader(401)
					return
				}
			}

			r.RequestURI = ""
			r.Host = host
			r.URL.Scheme = "http"
			r.URL.Host = r.Host
			r.URL.Path = tailingPath
			log.Printf("Out: %s", r.URL.String())
			_, portStr, _ := net.SplitHostPort(host)
			port, err := strconv.Atoi(portStr)
			if isWebSocket(r) {
				websocketProxy(w, r, port)
			} else {
				httpProxy(w, r, port)
			}
		} else {
			w.WriteHeader(404)
		}
	}
}

func httpProxy(w http.ResponseWriter, r *http.Request, port int) {
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(502)
		return
	}
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func isWebSocket(r *http.Request) bool {
	connection := r.Header["Connection"]
	if len(connection) > 0 && strings.ToLower(connection[0]) == "upgrade" {
		upgrade := r.Header["Upgrade"]
		if len(upgrade) > 0 && strings.ToLower(upgrade[0]) == "websocket" {
			return true
		}
	}
	return false
}

func websocketProxy(w http.ResponseWriter, r *http.Request, port int) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Print("Not support hijacker")
		return
	}
	connToClient, _, err := hj.Hijack()
	if err != nil {
		log.Println(err)
		return
	}
	defer connToClient.Close()

	connToPlugin, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Println(err)
		return
	}
	defer connToPlugin.Close()

	err = r.Write(connToPlugin)
	if err != nil {
		log.Println(err)
		return
	}

	errchan := make(chan error, 2)
	forward := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errchan <- err
	}

	go forward(connToClient, connToPlugin)
	go forward(connToPlugin, connToClient)
	<-errchan
}

type authentication struct {
	Token  string
	UserId string
}

func auth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			decoder := json.NewDecoder(r.Body)
			authInfo := new(authentication)
			err := decoder.Decode(authInfo)
			if err != nil {
				log.Println(err)
				w.WriteHeader(400)
			} else {
				setAuth(authInfo.Token, authInfo.UserId)
				w.WriteHeader(200)
			}
		}
	}
}

func main() {
	file, err := os.Open("plugins.json")
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	plugins := map[string]string{}
	err = decoder.Decode(&plugins)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/auth", auth())
	http.HandleFunc("/", handler(plugins))
	http.ListenAndServe(":8000", nil)
}
