package types

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type PWInfo struct {
	cng_at     float64
	reset      bool
	must_reset bool
}

type ACCTInfo struct {
	account_type float64
	state        string
	adm          bool
	game_name    string
	tag_line     string
	created_at   float64
}

type BanInfo struct {
	restrictions []interface{}
}

type Version struct {
	manifestID        string
	branch            string
	version           string
	buildVersion      string
	engineVersion     string
	riotClientVersion string
	riotClientBuild   string
}

type PlayerInfo struct {
	puuid                 string
	acct                  ACCTInfo
	ban                   BanInfo
	country               string
	country_at            float64
	email_verified        bool
	jti                   string
	original_platform_id  string
	phone_number_verified bool
	player_locale         string
	preferred_username    string
	pw                    PWInfo
	sub                   string
	username              string
	client_platform       string
	version               Version
}

// Uses the public access token from Entitlements, gets player data such as username, tagline and etc.

func GetPlayerInfo() PlayerInfo {

	entitlement := GetEntitlementsToken(GetLockfile(true))

	req, err := http.NewRequest("GET", "https://auth.riotgames.com/userinfo", nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var player_info map[string]interface{}

	player_info, err = GetJSON(res)
	checkError(err)

	if player_info == nil {
		fmt.Println(player_info)
		checkError(errors.New("player Info not found"))
	}

	ACCTInfoMap := player_info["acct"].(map[string]interface{})
	BansMap := player_info["ban"].(map[string]interface{})
	PwMap := player_info["pw"].(map[string]interface{})

	var ACCTInfoData = ACCTInfo{
		account_type: ACCTInfoMap["type"].(float64),
		state:        ACCTInfoMap["state"].(string),
		adm:          ACCTInfoMap["adm"].(bool),
		game_name:    ACCTInfoMap["game_name"].(string),
		tag_line:     ACCTInfoMap["tag_line"].(string),
		created_at:   ACCTInfoMap["created_at"].(float64),
	}

	var BansData = BanInfo{
		restrictions: BansMap["restrictions"].([]interface{}),
	}

	var PWData = PWInfo{
		cng_at:     PwMap["cng_at"].(float64),
		reset:      PwMap["reset"].(bool),
		must_reset: PwMap["must_reset"].(bool),
	}

	var original_platform_id interface{}

	if player_info["original_platform_id"] == nil {
		original_platform_id = ""
	} else {
		original_platform_id = player_info["original_platform_id"].(string)
	}

	var client_platform_temp = map[string]string{
		"platformType":      "PC",
		"platformOS":        "Windows",
		"platformOSVersion": "10.0.19042.1.256.64bit",
		"platformChipset":   "Unknown",
	}

	client_platform_json, err := json.Marshal(client_platform_temp)
	checkError(err)

	client_platform_base64 := base64.StdEncoding.EncodeToString(client_platform_json)

	req, err = http.NewRequest("GET", "https://valorant-api.com/v1/version", nil)
	checkError(err)

	res, err = Client.Do(req)
	checkError(err)

	var version_info map[string]interface{}

	version_info, err = GetJSON(res)
	checkError(err)

	version_data := version_info["data"].(map[string]interface{})

	version_struct := Version{
		manifestID:        version_data["manifestId"].(string),
		branch:            version_data["branch"].(string),
		version:           version_data["version"].(string),
		buildVersion:      version_data["buildVersion"].(string),
		engineVersion:     version_data["engineVersion"].(string),
		riotClientVersion: version_data["riotClientVersion"].(string),
		riotClientBuild:   version_data["riotClientBuild"].(string),
	}

	defer res.Body.Close()

	return PlayerInfo{
		puuid:                 entitlement.subject,
		acct:                  ACCTInfoData,
		ban:                   BansData,
		country:               player_info["country"].(string),
		country_at:            player_info["country_at"].(float64),
		email_verified:        player_info["email_verified"].(bool),
		jti:                   player_info["jti"].(string),
		original_platform_id:  original_platform_id.(string),
		phone_number_verified: player_info["phone_number_verified"].(bool),
		player_locale:         player_info["player_locale"].(string),
		preferred_username:    player_info["preferred_username"].(string),
		pw:                    PWData,
		sub:                   player_info["sub"].(string),
		username:              player_info["username"].(string),
		client_platform:       client_platform_base64,
		version:               version_struct,
	}
}
