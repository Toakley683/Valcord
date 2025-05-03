package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	Types "valcord/types"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

var (
	general_valorant_information ValorantInformation
)

type ValorantInformation struct {
	lock_file     Types.Lockfile_type
	entitlements  Types.EntitlementsTokenResponse
	player_info   Types.PlayerInfo
	regional_data Types.Regional
}

func cleanup() {

	fmt.Println("Cleaning up data for exit..")

	fmt.Println("Closing discord bot..")
	err := discord.Close()

	if err != nil {

		fmt.Println("Could not disable discord bot: Error(" + err.Error() + ")")

	} else {

		fmt.Println("Discord bot: Closed")

	}

}

func AppStartup() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	defer cleanup()

	// Add listener for when Valorant is (Calling 127.0.0.1:{LockfilePort} )

	Types.Init_val_details()
	lockfile := Types.GetLockfile()
	entitlements := Types.GetEntitlementsToken(lockfile)
	player_info := Types.GetPlayerInfo()
	region_data := Types.GetRegionData()

	general_valorant_information = ValorantInformation{
		lock_file:     lockfile,
		entitlements:  entitlements,
		player_info:   player_info,
		regional_data: region_data,
	}

	discord_setup()

	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Shutting down..")
}

func main() {

	BeginChecks()

}
