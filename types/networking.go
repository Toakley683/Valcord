package types

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Converts data returned from valorant endpoints from JSON strings into maps with values as an interface for converting into required types

func GetJSON(res *http.Response) (map[string]interface{}, error) {

	data, err := io.ReadAll(res.Body)
	checkError(err)

	var returnValue map[string]interface{}

	if json.Valid(data) {

		err := json.Unmarshal(data, &returnValue)
		checkError(err)

		return returnValue, nil

	}

	return returnValue, errors.New("JSON was not valid")

}

var (
	client http.Client
)

func setup_networking() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = http.Client{Transport: tr}
}
