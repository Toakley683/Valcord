package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/MasterDimmy/go-cls"

	"github.com/ncruces/zenity"

	Types "valcord/types"
)

var (
	settings map[string]string

	Retries = 0
)

func BeginChecks() {

	settings = Types.CheckSettings()

	PasteDiscordToken()
}

func VerifyBotToken(token string) bool {

	if token == "" {
		return false
	}

	matched, err := regexp.MatchString(`[\w-]{26}\.[\w-]{6}\.[\w-]{38}`, token)
	checkError(err)

	fmt.Println(matched)

	return matched

}

func VerifyServerID(serverID string) bool {

	if serverID == "" {
		return false
	}

	if len(serverID) > 19 {
		return false
	}

	matched, err := regexp.MatchString(`^\d{18,19}$`, serverID)
	checkError(err)

	return matched

}

func VerifyChannelID(channelID string) bool {

	if channelID == "" {
		return false
	}

	if len(channelID) > 19 {
		return false
	}

	matched, err := regexp.MatchString(`^\d{18,19}$`, channelID)
	checkError(err)

	return matched

}

func PasteDiscordToken() {

	// Request Discord Bot API Token

	Retries = Retries + 1

	if Retries > 25 {
		fmt.Println("Ran out of tries")
		os.Exit(0)
	}

	if !VerifyBotToken(settings["discord_api_token"]) {

		var DiscordToken string
		var err error

		if Retries <= 1 {

			DiscordToken, err = zenity.Entry("Enter Discord Bot Token:",
				zenity.Title("Valcord"))

		} else {

			DiscordToken, err = zenity.Entry("Enter Discord Bot Token:",
				zenity.Title("Invalid Token Given"))

		}

		if err != nil && err.Error() == "dialog canceled" {
			os.Exit(0)
		}

		checkError(err)

		Matched := VerifyBotToken(DiscordToken)

		if Matched {

			settings["discord_api_token"] = DiscordToken
			settings["server_id"] = ""
			settings["current_session_channel"] = ""
			Types.CheckSettingsData(settings)

			Retries = 0
			CheckForServerID()
			return

		}

		PasteDiscordToken()

	}

	Retries = 0
	CheckForServerID()

}

func CheckForServerID() {

	// Request Session Channel ID

	Retries = Retries + 1

	if !VerifyServerID(settings["server_id"]) {

		var ServerID string
		var err error

		if Retries <= 1 {

			ServerID, err = zenity.Entry("Enter ServerID:",
				zenity.Title("Valcord"))

		} else {

			ServerID, err = zenity.Entry("Enter ServerID:",
				zenity.Title("Invalid ServerID Given"))

		}

		if err != nil && err.Error() == "dialog canceled" {
			os.Exit(0)
		}

		checkError(err)

		Matched := VerifyServerID(ServerID)

		if Matched {

			settings["server_id"] = ServerID
			settings["current_session_channel"] = ""
			Types.CheckSettingsData(settings)

			Retries = 0
			CheckForChannelID()

			return

		}

		CheckForServerID()

	}

	Retries = 0
	CheckForChannelID()

}

func CheckForChannelID() {

	// Request Session Channel ID

	Retries = Retries + 1

	if !VerifyChannelID(settings["current_session_channel"]) {

		var SessionChannelID string
		var err error

		if Retries <= 1 {

			SessionChannelID, err = zenity.Entry("Enter Session ChannelID:",
				zenity.Title("Valcord"))

		} else {

			SessionChannelID, err = zenity.Entry("Enter Session ChannelID:",
				zenity.Title("Invalid Session ChannelID Given"))

		}

		if err != nil && err.Error() == "dialog canceled" {
			os.Exit(0)
		}

		checkError(err)

		Matched := VerifyChannelID(SessionChannelID)

		if Matched {

			settings["current_session_channel"] = SessionChannelID
			Types.CheckSettingsData(settings)

			CheckForOpenGame()

			return

		}

		CheckForChannelID()

	}

	CheckForOpenGame()

}

func CheckForOpenGame() {

	// Check for game being open

	RetryDelay := time.Millisecond * 7500

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	Client := http.Client{Transport: tr}

	var res *http.Response

	for {

		lockfile := Types.GetLockfile(false)

		if lockfile.Port == "" {
			Types.NewLog("Client: Closed")
			Types.NewLog("Game is not open, retrying..")

			time.Sleep(RetryDelay)
			continue
		}

		req, err := http.NewRequest("GET", "https://127.0.0.1:"+lockfile.Port+"/entitlements/v1/token", nil)
		checkError(err)

		req.Header.Add("Authorization", "Basic "+Types.BasicAuth("riot", lockfile.Password))

		Res, Err := Client.Do(req)
		res = Res

		if Err != nil {

			splitError := strings.Split(Err.Error(), " ")
			finalError := strings.Join(splitError[6:], " ")

			if finalError == "No connection could be made because the target machine actively refused it." {

				// Game is not open

				Types.NewLog("Client: Closed")
				Types.NewLog("Game is not open, retrying..")

				time.Sleep(RetryDelay)
				continue

			}

			checkError(Err)

		}

		Types.NewLog("Client: Open")

		break

	}

	time.Sleep(time.Second * 1)
	cls.CLS()

	res.Body.Close()

	AppInit()

}
