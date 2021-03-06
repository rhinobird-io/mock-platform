package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"strconv"

	"github.com/streamrail/concurrent-map"
	"github.com/go-fsnotify/fsnotify"
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
				websocketProxy(w, r, host)
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
	if r.Header.Get("Accept") == "text/event-stream" {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}
		for {
			buffer := make([]byte, 100000)
            cBytes, err := resp.Body.Read(buffer)
            if err == io.EOF {
                    break
            }
			w.Write(buffer[0:cBytes])
			flusher.Flush()
		}
	} else {
		io.Copy(w, resp.Body)
    }
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

func websocketProxy(w http.ResponseWriter, r *http.Request, host string) {
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

	connToPlugin, err := net.Dial("tcp", host)
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

func loadConfig(plugins map[string]string) {
	file, err := os.Open("plugins.json")
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&plugins)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	plugins := map[string]string{}
	loadConfig(plugins)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if (event.Name == "./plugins.json") {
						log.Println("The config plugins.json changed")
						loadConfig(plugins)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Add("./")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/auth", auth())
	http.HandleFunc("/", handler(plugins))

	log.Println("Listening on 8000")
	http.ListenAndServe(":8000", nil)
}
