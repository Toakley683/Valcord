package main

import (
	"log"
	"os"
	"os/signal"

	Types "valcord/types"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

var (
	settings = check_settings()

	general_valorant_information ValorantInformation
)

type ValorantInformation struct {
	lock_file     Types.Lockfile_type
	entitlements  Types.EntitlementsTokenResponse
	player_info   Types.PlayerInfo
	regional_data Types.Regional
}

func main() {

	Types.Init_val_details()
	lockfile := Types.GetLockfile()
	entitlements := Types.GetEntitlementsToken(lockfile)
	player_info := Types.GetPlayerInfo(entitlements)
	region_data := Types.GetRegionData()

	general_valorant_information = ValorantInformation{
		lock_file:     lockfile,
		entitlements:  entitlements,
		player_info:   player_info,
		regional_data: region_data,
	}

	discord_setup()
	defer discord.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Shutting down..")

}
