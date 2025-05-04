package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

func checkUpdates() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	Client := http.Client{Transport: tr}

	// Latest release https://github.com/Toakley683/Valcord/releases/latest

	req, err := http.NewRequest("GET", "https://api.github.com/repos/Toakley683/Valcord/releases/latest", nil)
	checkError(err)

	res, err := Client.Do(req)

	defer res.Body.Close()

	fmt.Println(res)

	panic("Version incorrect")

}
