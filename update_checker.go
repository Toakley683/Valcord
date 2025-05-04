package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	Types "valcord/types"

	"github.com/MasterDimmy/go-cls"
)

var versionSHA string
var version string

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

	cls.CLS()

	if update_data["tag_name"] == versionSHA {

		// Version is up to date!

		fmt.Println("Successful, current version is most up to date")

		time.Sleep(time.Second)
		return

	}

	// Version is not updated

	fmt.Println("Current Valcord version is not updated")
	fmt.Println("\nCurrent:")
	fmt.Println("\tValcord " + version + "\n")
	fmt.Println("Latest:")
	fmt.Println("\t" + update_data["name"].(string))

	WaitDelay := time.Second * 10
	Duration := time.Duration(WaitDelay).Seconds()

	fmt.Println("\nWill attempt to run normally in (" + strconv.Itoa(int(Duration)) + ") seconds..")

	fmt.Println("\nYou can find the newest version at 'https://github.com/Toakley683/Valcord/releases/latest'\n")

	time.Sleep(WaitDelay)

}
