package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/groupcache"
)

var (
	cache *groupcache.Group
)

type request struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// WriteJSON write json format for response
func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

// newResponse makes a new response
// res: succ, or fail
// msg: error msg or success return value
func newResponse(res, msg string) map[string]string {
	return map[string]string{"result": res, "message": msg}
}

// Get return result
func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, newResponse("fail", "bad request"))
		return
	}
	body, _ := ioutil.ReadAll(r.Body)
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		WriteJSON(w, http.StatusBadRequest, newResponse("fail", err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var data []byte
	if err := cache.Get(ctx, req.Key, groupcache.AllocatingByteSliceSink(&data)); err != nil {
		WriteJSON(w, http.StatusBadRequest, newResponse("fail", err.Error()))
		return
	}

	log.Infof("get %s's value: %s", req.Key, data)
	WriteJSON(w, http.StatusOK, newResponse("succ", string(data)))
}

func main() {
	addr := flag.String("addr", ":8001", "local address")
	port := flag.String("port", ":18001", "listen port")
	flag.Parse()

	peers := groupcache.NewHTTPPool("http://localhost" + *addr)
	peers.Set("http://localhost:8001", "http://localhost:8002")

	// init cache
	cache = groupcache.NewGroup("database_cache", 64<<20, groupcache.GetterFunc(cacheGetFunc))

	go http.ListenAndServe(*addr, http.HandlerFunc(peers.ServeHTTP))

	log.Infof("listen port: %v", *port)
	http.HandleFunc("/get", Get)
	log.Panic(http.ListenAndServe(*port, nil))
}

func cacheGetFunc(ctx context.Context, key string, dest groupcache.Sink) error {
	log.Infof("asking for key: %s", key)
	req := request{Key: key}
	data, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	// :9000 is our slow database
	resp, err := http.Post("http://localhost:9000/get", "application/json", bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("read response body error: %v", err)
		return err
	}
	log.Infof("read db response data: %s", data)

	if resp.StatusCode != http.StatusOK {
		dest.SetBytes([]byte("error http status"))
		return nil
	}

	dest.SetBytes(data)
	return nil
}
