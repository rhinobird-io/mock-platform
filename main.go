package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"os"
	"regexp"
)

func handler(plugins map[string]int, pool *redis.Pool) http.HandlerFunc {
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
				conn := pool.Get()
				defer conn.Close()
				cookie, err := r.Cookie("Auth")
				if err != nil {
					w.WriteHeader(401)
					return
				}
				value, err := redis.String(conn.Do("GET", cookie.Value))
				if err != nil {
					w.WriteHeader(401)
					return
				} else {
					r.Header.Set("USER", value)
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

var (
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func main() {
	flag.Parse()

	log.Printf("Will connect redis server: %s", *redisAddress)
	log.Printf("Max connections: %d", *maxConnections)
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		return c, err
	}, *maxConnections)
	defer redisPool.Close()

	file, _ := os.Open("plugins.json")
	decoder := json.NewDecoder(file)
	plugins := map[string]int{}
	err := decoder.Decode(&plugins)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", handler(plugins, redisPool))
	http.ListenAndServe(":8000", nil)
}
