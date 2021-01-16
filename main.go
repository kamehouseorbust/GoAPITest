package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/IncSW/geoip2"
	"github.com/gorilla/mux"
)

type responseData struct {
	IPAddress   string `json:"IPAddress"`
	CountryCode string `json:"CountryCode"`
	ValidCode   bool   `json:"ValidCode"`
	ErrorReason string `json:"ErrorReason"`
}

var dbPath string
var countryReader *geoip2.CountryReader

func init() {
	dbPath = "/home/hdr/code/GoAPITest/GeoLite2-Country.mmdb"
	countryReader = initializeDBReader(dbPath)
}

func main() {
	handleRequests()
	fmt.Println("Initializing the db....")
}

func handleRequests() {
	fmt.Println("Setting up request server")
	router := mux.NewRouter().StrictSlash(true)
	router.Queries("ip", "whiteList")
	router.HandleFunc("/api/ipWhiteListed", checkIPIsWhitelisted)
	log.Fatal(http.ListenAndServe(":3000", router))
}

func checkIPIsWhitelisted(response http.ResponseWriter, r *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	r.ParseForm()
	ipAddress := r.FormValue("ip")
	responseData := responseData{IPAddress: ipAddress}
	if !isIpv4Net(ipAddress) {
		responseData.ErrorReason = "Not valid ip"
		returnResponse(response, responseData, 400)
		return
	}

	countryCode := getIPCountry(ipAddress, countryReader)
	if countryCode == "" {
		responseData.ErrorReason = "Couldn't find country code"
		returnResponse(response, responseData, 400)
		return
	}
	responseData.CountryCode = countryCode

	whiteList, present := r.Form["whiteListValues"]
	if !present || len(whiteList) == 0 {
		responseData.ErrorReason = "No country codes present request for whitelist parameter"
		returnResponse(response, responseData, 400)
		return
	}

	responseData.ValidCode = countryIsWhiteListed(countryCode, whiteList)
	returnResponse(response, responseData, 200)
}

func returnResponse(response http.ResponseWriter, responseData responseData, responseHeader int) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(responseHeader)
	json.NewEncoder(response).Encode(responseData)
}

func isIpv4Net(ipAddress string) bool {
	return net.ParseIP(ipAddress) != nil
}

func countryIsWhiteListed(countryCode string, whiteList []string) bool {
	for _, code := range whiteList {
		if code == countryCode {
			return true
		}
	}
	return false
}

func initializeDBReader(dbPath string) *geoip2.CountryReader {
	fmt.Println("Initializing Country Reader")
	reader, err := geoip2.NewCountryReaderFromFile(dbPath)
	if err != nil {
		log.Fatal("Unable to initialize country reader, ensure path string is correct and the file is valid")
	}
	return reader
}

func getIPCountry(ipAddress string, reader *geoip2.CountryReader) string {
	fmt.Println("Getting country code for ip Address:", ipAddress)
	record, err := reader.Lookup(net.ParseIP("162.192.104.160"))
	if err != nil {
		return ""
	}
	return record.Country.ISOCode
}
