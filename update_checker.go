package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	Types "valcord/types"

	"github.com/MasterDimmy/go-cls"
)

var versionSHA string
var version string

func getShortSHA(SHA string) string {

	return strings.Join(strings.Split(SHA, "")[:7], "")

}

func GetLatestInfo() map[string]interface{} {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	Client := http.Client{Transport: tr}

	// Latest release https://github.com/Toakley683/Valcord/releases/latest

	req, err := http.NewRequest("GET", "https://api.github.com/repos/Toakley683/Valcord/releases/latest", nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var update_data map[string]interface{}

	update_data, err = Types.GetJSON(res)
	checkError(err)

	if update_data["status"] != nil {

		Types.NewLog("Couldn't find any versions")

		time.Sleep(time.Second)
		return map[string]interface{}{}

	}

	return update_data

}

func GetTagInfo(TagName string) map[string]interface{} {

	if TagName == "" {

		checkError(errors.New("no tag given for github tag api"))

	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	Client := http.Client{Transport: tr}

	// Tag API: https://api.github.com/repos/Toakley683/Valcord/git/ref/tags/{TagName}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/Toakley683/Valcord/git/ref/tags/"+TagName, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var tag_data map[string]interface{}

	tag_data, err = Types.GetJSON(res)
	checkError(err)

	if tag_data["status"] != nil {

		Types.NewLog("Couldn't find any tag")

		time.Sleep(time.Second)
		return map[string]interface{}{}

	}

	tag_data = tag_data["object"].(map[string]interface{})

	return tag_data

}

func checkUpdates() {

	if versionSHA == "" {

		// Test versions have no version

		menuUpdate.SetTitle("Update: Most Recent")
		menuBranch.SetTitle("Branch: Test")

		Types.NewLog("SHA: " + versionSHA)
		Types.NewLog("Updates will not be checked due to test mode")
		Types.NewLog("Running in test mode, requires version for release mode..")
		time.Sleep(time.Second * 2)

		cls.CLS()

		return

	}

	Types.NewLog("Running in release mode.")
	Types.NewLog("Checking for updates..")

	Types.NewLog("Current Version: " + version)
	Types.NewLog("Current SHA: " + versionSHA)

	update_data := GetLatestInfo()
	tag_data := GetTagInfo(update_data["tag_name"].(string))

	cls.CLS()

	updatedSHA := getShortSHA(tag_data["sha"].(string))

	if updatedSHA == versionSHA {

		// Version is up to date!

		Types.NewLog("VersionSHA: " + versionSHA)
		Types.NewLog("Successful, current version is most up to date")

		menuUpdate.SetTitle("Update: Most Recent")
		menuBranch.SetTitle("Branch: Release")

		time.Sleep(time.Second * 2)

		cls.CLS()
		return

	}

	// Version is not updated

	menuUpdate.SetTitle("Update available")
	menuUpdate.Enable()

	Types.NewLog("Current Valcord version is not updated")
	Types.NewLog("\nCurrent:")
	Types.NewLog("\tValcord " + version + " (" + versionSHA + ")\n")
	Types.NewLog("Latest:")
	Types.NewLog("\t" + update_data["name"].(string) + " (" + updatedSHA + ")")

	WaitDelay := time.Second * 2
	Duration := time.Duration(WaitDelay).Seconds()

	Types.NewLog("\nWill attempt to run normally in (" + strconv.Itoa(int(Duration)) + ") seconds..")

	Types.NewLog("\nYou can find the newest version at 'https://github.com/Toakley683/Valcord/releases/latest'")

	Types.NewLog("\n")

	time.Sleep(WaitDelay)

}
