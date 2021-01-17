// If I have time to implement gRPC I'd use this package to start/control
// REST and gRPC packages, but for now it only supports REST, which doens't have its
// own package yet.
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

// I'd like to put these in a separate file
const (
	dbPath                  string = "GeoLite2-Country.mmdb"
	initDBReader            string = "Initializing Country Reader..."
	taskDone                string = "Done."
	initRest                string = "Setting up REST server"
	restAddress             string = "/api/ipWhiteListed"
	restPort                string = ":3000"
	printFatalReader        string = "Unable to init reader, ensure path is correct and file is valid"
	codeRetrieve            string = "Getting country code for ip address: "
	returnBlank             string = ""
	contentType             string = "Content-Type"
	jsonParam               string = "application/json"
	errorWhitelist          string = "IP not within the whitelist"
	errorMissingCountryCode string = "No country codes present for whitelist parameter"
	errorIPCountryCode      string = "Couldn't find country code"
	errorIPInvalid          string = "Invalid IP address"
	parameterWhiteList      string = "whiteListValues"
	parameterIP             string = "ip"
)

var countryReader *geoip2.CountryReader

func init() {
	// want to make sure we never have to wait on this once the request
	// server is running
	countryReader = initializeDBReader(dbPath)
}

func main() {
	// right now we're only open for REST requests
	handleRequests()
}

// Would like to separate the next three functions
// into their own package that this main file imports
func handleRequests() {
	fmt.Println(initRest)
	router := mux.NewRouter().StrictSlash(true)
	router.Queries(parameterIP, parameterWhiteList)
	router.HandleFunc(restAddress, checkIPIsWhitelisted)
	fmt.Println(taskDone)
	log.Fatal(http.ListenAndServe(restPort, router))
}

func checkIPIsWhitelisted(response http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	ipAddress := r.FormValue(parameterIP)
	responseData := responseData{IPAddress: ipAddress}
	if !isIpv4Net(ipAddress) {
		responseData.ErrorReason = errorIPInvalid
		returnResponse(response, responseData, 400)
		return
	}

	countryCode := getIPCountry(ipAddress, countryReader)
	if countryCode == "" {
		responseData.ErrorReason = errorIPCountryCode
		returnResponse(response, responseData, 400)
		return
	}
	responseData.CountryCode = countryCode

	whiteList, present := r.Form[parameterWhiteList]
	if !present || len(whiteList) == 0 {
		responseData.ErrorReason = errorMissingCountryCode
		returnResponse(response, responseData, 400)
		return
	}

	responseData.ValidCode = countryIsWhiteListed(countryCode, whiteList)
	if responseData.ValidCode == false {
		responseData.ErrorReason = errorWhitelist
	}
	returnResponse(response, responseData, 200)
}

func returnResponse(response http.ResponseWriter, responseData responseData, responseHeader int) {
	response.Header().Set(contentType, jsonParam)
	response.WriteHeader(responseHeader)
	json.NewEncoder(response).Encode(responseData)
}

// These next four functions could stay in main, as they are separate from
// REST vs gRPC implementations
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
	fmt.Println(initDBReader)
	reader, err := geoip2.NewCountryReaderFromFile(dbPath)
	if err != nil {
		log.Fatal(printFatalReader)
	}
	fmt.Println(taskDone)
	return reader
}

func getIPCountry(ipAddress string, reader *geoip2.CountryReader) string {
	fmt.Println(codeRetrieve, ipAddress)
	record, err := reader.Lookup(net.ParseIP(ipAddress))
	if err != nil {
		return returnBlank
	}
	fmt.Println(taskDone)
	return record.Country.ISOCode
}
