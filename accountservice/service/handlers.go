package service

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/callistaenterprise/goblog/accountservice/dbclient"
	"github.com/callistaenterprise/goblog/accountservice/model"
	cb "github.com/callistaenterprise/goblog/common/circuitbreaker"
	"github.com/callistaenterprise/goblog/common/messaging"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var DBClient dbclient.IBoltClient
var MessagingClient messaging.IMessagingClient
var isHealthy = true // NEW

var client = &http.Client{}

var fallbackQuote = model.Quote{
	Language: "en",
	ServedBy: "circuit-breaker",
	Text:     "May the source be with you, always."}

func init() {
	var transport http.RoundTripper = &http.Transport{
		DisableKeepAlives: true,
	}
	client.Transport = transport
}

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
	account.ServedBy = getIP()

	notifyVIP(account) // Send VIP notification concurrently.

	account.Quote = getQuote()
	account.ImageUrl = getImageUrl(accountId)

	// If found, marshal into JSON, write headers and content
	data, _ := json.Marshal(account)
	writeJsonResponse(w, http.StatusOK, data)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Since we're here, we already know that HTTP service is up. Let's just check the state of the boltdb connection
	dbUp := DBClient.Check()
	if dbUp && isHealthy {
		data, _ := json.Marshal(healthCheckResponse{Status: "UP"})
		writeJsonResponse(w, http.StatusOK, data)
	} else {
		data, _ := json.Marshal(healthCheckResponse{Status: "Database unaccessible"})
		writeJsonResponse(w, http.StatusServiceUnavailable, data)
	}
}

func writeJsonResponse(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

type healthCheckResponse struct {
	Status string `json:"status"`
}

func SetHealthyState(w http.ResponseWriter, r *http.Request) {
	// Read the 'state' path parameter from the mux map and convert to a bool
	var state, err = strconv.ParseBool(mux.Vars(r)["state"])
	// If we couldn't parse the state param, return a HTTP 400
	if err != nil {
		logrus.Infoln("Invalid request to SetHealthyState, allowed values are true or false")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Otherwise, mutate the package scoped "isHealthy" variable.
	isHealthy = state
	w.WriteHeader(http.StatusOK)
}

func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "error"
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	panic("Unable to determine local IP address (non loopback). Exiting.")
}

func getQuote() model.Quote {
	body, err := cb.CallUsingCircuitBreaker(
		"quotes-service",
		"http://quotes-service:8080/api/quote?strength=4",
		"GET")
	if err == nil {
		quote := model.Quote{}
		json.Unmarshal(body, &quote)
		return quote
	} else {
		return fallbackQuote
	}
}

func getImageUrl(accountId string) string {
	body, err := cb.CallUsingCircuitBreaker(
		"imageservice",
		"http://imageservice:7777/accounts/"+accountId,
		"GET")
	if err == nil {
		return string(body)
	} else {
		return "http://path.to.placeholder"
	}
}

// If our hard-coded "VIP" account, spawn a goroutine to send a message.
func notifyVIP(account model.Account) {
	if account.Id == "10000" {
		go func(account model.Account) {
			vipNotification := model.VipNotification{AccountId: account.Id, ReadAt: time.Now().UTC().String()}
			data, _ := json.Marshal(vipNotification)
			err := MessagingClient.PublishOnQueue(data, "vip_queue")
			if err != nil {
				logrus.Infoln(err.Error())
			}
		}(account)
	}
}
