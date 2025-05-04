package main

import (
	"crypto/tls"
	"errors"
	"fmt"
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

		fmt.Println("Couldn't find any versions")

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

		fmt.Println("Couldn't find any tag")

		time.Sleep(time.Second)
		return map[string]interface{}{}

	}

	tag_data = tag_data["object"].(map[string]interface{})

	return tag_data

}

func checkUpdates() {

	if versionSHA == "" {

		// Test versions have no version

		fmt.Println("Running in test mode, requires version for release mode..")
		time.Sleep(time.Second)

		return

	}

	fmt.Println("Running in release mode.")
	fmt.Println("Checking for updates..")

	fmt.Println("Current Version: " + version)
	fmt.Println("Current SHA: " + versionSHA)

	update_data := GetLatestInfo()
	tag_data := GetTagInfo(update_data["tag_name"].(string))

	cls.CLS()

	updatedSHA := getShortSHA(tag_data["sha"].(string))

	if updatedSHA == versionSHA {

		// Version is up to date!

		fmt.Println("Successful, current version is most up to date")

		time.Sleep(time.Second)
		return

	}

	// Version is not updated

	fmt.Println("Current Valcord version is not updated")
	fmt.Println("\nCurrent:")
	fmt.Println("\tValcord " + version + " (" + versionSHA + ")\n")
	fmt.Println("Latest:")
	fmt.Println("\t" + update_data["name"].(string) + " (" + updatedSHA + ")")

	WaitDelay := time.Second * 10
	Duration := time.Duration(WaitDelay).Seconds()

	fmt.Println("\nWill attempt to run normally in (" + strconv.Itoa(int(Duration)) + ") seconds..")

	fmt.Println("\nYou can find the newest version at 'https://github.com/Toakley683/Valcord/releases/latest'")

	fmt.Print("\n")

	time.Sleep(WaitDelay)

}
