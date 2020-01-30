package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/callistaenterprise/goblog/accountservice/dbclient"
	"github.com/gorilla/mux"
)

var DBClient dbclient.IBoltClient

func GetAccount(w http.ResponseWriter, r *http.Request) {
	// Read the 'accountId' path parameter from the mux map
	var accountId = mux.Vars(r)["accountId"]
	// Read the account struct BoltDB
	account, err := DBClient.QueryAccount(accountId)
	// If err, return a 404
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
