/*database provide slow database

Slow database based on map, provide http server

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	db *SlowDB
)

type request struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// SlowDB is database
type SlowDB struct {
	DB  map[string][]byte
	mux sync.RWMutex
}

// NewSlowDB create a slow db
func NewSlowDB() (*SlowDB, error) {
	return &SlowDB{
		DB:  make(map[string][]byte),
		mux: sync.RWMutex{},
	}, nil
}

// Set value for SlowDB
func (db *SlowDB) Set(key string, value []byte) error {
	db.mux.Lock()
	db.mux.Unlock()
	db.DB[key] = value
	return nil
}

// Get value for SlowDB
func (db *SlowDB) Get(key string) ([]byte, error) {
	db.mux.Lock()
	db.mux.Unlock()

	time.Sleep(800 * time.Millisecond)

	v, ok := db.DB[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return v, nil
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

	key := req.Key
	value, err := db.Get(key)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, newResponse("fail", err.Error()))
		return
	}

	log.Infof("get %s's value: %s", key, value)
	WriteJSON(w, http.StatusOK, newResponse("succ", string(value)))
}

// Set store result
func Set(w http.ResponseWriter, r *http.Request) {
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

	if err := db.Set(req.Key, []byte(req.Value)); err != nil {
		WriteJSON(w, http.StatusBadRequest, newResponse("fail", "set database error"))
		return
	}
	log.Infof("set %s:%s", req.Key, req.Value)

	WriteJSON(w, http.StatusOK, newResponse("succ", "success set"))
}

func main() {
	port := flag.String("port", ":9000", "listen port for database")
	flag.Parse()

	var err error
	db, err = NewSlowDB()
	if err != nil {
		log.Errorf("create database error: %v", err)
		return
	}

	log.Infof("listen port: %v", *port)
	http.HandleFunc("/get", Get)
	http.HandleFunc("/set", Set)
	log.Panic(http.ListenAndServe(*port, nil))
}
