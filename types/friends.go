package types

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

func CheckIfOnFriendsList(uuid string, lockfile Lockfile_type) bool {

	if uuid == GetPlayerInfo().sub {

		NewLog("Is owner, skipping over friend check")

		return true // Commented for debugging

	}

	req, err := http.NewRequest("GET", "https://127.0.0.1:"+lockfile.Port+"/chat/v4/friends", nil)
	checkError(err)

	req.Header.Add("Authorization", "Basic "+BasicAuth("riot", lockfile.Password))

	res, err := Client.Do(req)
	checkError(err)

	friends, err := GetJSON(res)
	checkError(err)

	if friends["friends"] == nil {
		return false
	}

	friends_array := friends["friends"].([]interface{})

	defer res.Body.Close()

	var wg sync.WaitGroup
	output := make(chan bool)

	wg.Add(1)

	go func() {

		defer wg.Done()

		for Index, Value := range friends_array {

			FinalVal := Value.(map[string]interface{})

			fmt.Println("Index:", Index, "friend username:", FinalVal["game_name"].(string)+":"+FinalVal["game_tag"].(string))

			if FinalVal["puuid"] != uuid {

				continue

			}

			output <- true
			break

		}

		output <- false

	}()

	if <-output {

		NewLog("Friend found")
		return true

	}

	wg.Wait()

	return false

}

func CheckIfRequestOutbound(uuid string, lockfile Lockfile_type) bool {

	req, err := http.NewRequest("GET", "https://127.0.0.1:"+lockfile.Port+"/chat/v4/friendrequests", nil)
	checkError(err)

	req.Header.Add("Authorization", "Basic "+BasicAuth("riot", lockfile.Password))

	res, err := Client.Do(req)
	checkError(err)

	requests, err := GetJSON(res)
	checkError(err)

	if requests["requests"] == nil {
		return false
	}

	requests_array := requests["requests"].([]interface{})

	defer res.Body.Close()

	var wg sync.WaitGroup
	output := make(chan bool)

	wg.Add(1)

	go func() {

		defer wg.Done()

		for _, Value := range requests_array {

			FinalVal := Value.(map[string]interface{})

			if FinalVal["puuid"] != uuid {
				continue
			}

			// Check if request is outbound

			if FinalVal["subscription"] != "pending_out" {
				continue
			}

			output <- true
			break

		}

		output <- false

	}()

	if <-output {

		NewLog("Friend found")
		return true

	}

	wg.Wait()

	return false

}

func sendFriendRequest(name string, tagline string, lockfile Lockfile_type) bool {

	textPayload := `{ "game_name":"` + name + `", ` + ` "game_tag":"` + tagline + `" }`

	payload := strings.NewReader(textPayload)

	req, err := http.NewRequest("POST", "https://127.0.0.1:"+lockfile.Port+"/chat/v4/friendrequests", payload)
	checkError(err)

	req.Header.Add("Authorization", "Basic "+BasicAuth("riot", lockfile.Password))

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var friend_request map[string]interface{}

	friend_request, err = GetJSON(res)
	checkError(err)

	if friend_request["errorCode"] != nil {

		NewLog("Friend Request Error:", friend_request["message"])
		return true

	}

	NewLog(friend_request)

	return true

}

func removeFriendRequest(uuid string, lockfile Lockfile_type) bool {

	textPayload := `{ "puuid":"` + uuid + `" }`

	NewLog(textPayload)

	payload := strings.NewReader(textPayload)

	req, err := http.NewRequest("DELETE", "https://127.0.0.1:"+lockfile.Port+"/chat/v4/friendrequests", payload)
	checkError(err)

	req.Header.Add("Authorization", "Basic "+BasicAuth("riot", lockfile.Password))

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	return true

}
