package types

import (
	"errors"
	"fmt"
	"net/http"
)

type EntitlementsTokenResponse struct {
	accessToken  string
	entitlements []any
	issuer       string
	subject      string
	token        string
}

// Uses access token from lockfile, allows us to get information on the player and a access token for the valorant public endpoints

func GetEntitlementsToken(lockfile Lockfile_type) EntitlementsTokenResponse {

	req, err := http.NewRequest("GET", "https://127.0.0.1:"+lockfile.Port+"/entitlements/v1/token", nil)
	checkError(err)

	req.Header.Add("Authorization", "Basic "+BasicAuth("riot", lockfile.Password))

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var entitlement map[string]interface{}

	entitlement, err = GetJSON(res)
	checkError(err)

	if entitlement["errorCode"] != nil {

		if entitlement["message"].(string) == "Invalid URI format" {

			fmt.Println("Riot Client local webserver not open; Please restart riot client.")

		}

		checkError(errors.New(entitlement["message"].(string)))

	}

	return EntitlementsTokenResponse{
		accessToken:  entitlement["accessToken"].(string),
		entitlements: entitlement["entitlements"].([]any),
		issuer:       entitlement["issuer"].(string),
		subject:      entitlement["subject"].(string),
		token:        entitlement["token"].(string),
	}

}
