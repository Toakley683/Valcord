package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/MasterDimmy/go-cls"
	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"

	Types "valcord/types"
)

func checkError(err error) {
	if err != nil {
		zenity.Error(err.Error(),
			zenity.Title("Vencord: Error"))
		log.Fatal(err)
	}
}

var (
	general_valorant_information ValorantInformation

	Flags = map[string]bool{}

	menuStatus = &systray.MenuItem{}
)

type ValorantInformation struct {
	lock_file     Types.Lockfile_type
	entitlements  Types.EntitlementsTokenResponse
	player_info   Types.PlayerInfo
	regional_data Types.Regional
}

func cleanup() {

	Types.NewLog("Cleaning up data for exit..")

	if discord != nil {

		Types.NewLog("Closing discord bot..")
		err := discord.Close()

		if err != nil {

			Types.NewLog("Could not disable discord bot: Error(" + err.Error() + ")")

		} else {

			Types.NewLog("Discord bot: Closed")

		}

	}

	Types.NewLog("Cleaning up Log File..")

	if Types.LogFile != nil {

		Types.NewLog("Log file: Closed")

		Types.LogFile.Close()

	}

}

func AppShutdown() {

	cleanup()

	cls.CLS()

	Types.NewLog("Shutting down..")

	os.Exit(0)

}

func AppInit() {

	Types.Init_val_details()
	lockfile := Types.GetLockfile(true)
	entitlements := Types.GetEntitlementsToken(lockfile)
	player_info := Types.GetPlayerInfo()
	region_data := Types.GetRegionData()

	general_valorant_information = ValorantInformation{
		lock_file:     lockfile,
		entitlements:  entitlements,
		player_info:   player_info,
		regional_data: region_data,
	}

	menuStatus.Check()
	menuStatus.SetTitle("Status: Activated")
	menuStatus.SetTooltip("Application is now active!")

	discord_setup()

	Types.NewLog("Press Ctrl+C to exit")

}

func ImmedieteFlags() {

	help := flag.Bool("help", false, "Get command info")
	reset := flag.Bool("clean-commands", false, "Clean Discord Commands")
	retrieve_link := flag.Bool("invite", false, "Get invite link for bot")

	flag.Parse()

	Flags["Reset"] = *reset
	Flags["Link"] = *retrieve_link

	if *help {

		HelpText := ""

		HelpText = HelpText + "Valcord commands:\n"
		HelpText = HelpText + "./valcord.exe [Command]\n"
		HelpText = HelpText + "\t--help = [ Prints this help text ]\n"
		HelpText = HelpText + "\t--clean-commands = [ Cleans all discord commands ]\n"
		HelpText = HelpText + "\t--invite = [ Generates invite link for bot ]\n"

		Types.NewLog(HelpText)

		os.Exit(1)

	}

}

func GetIconData(Client http.Client, url string) []byte {

	req, err := http.NewRequest("GET", url, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	checkError(err)

	return data

}

func LoadIcons() map[string][]byte {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	Client := http.Client{Transport: tr}

	export := map[string][]byte{}

	export["logo"] = GetIconData(Client, "https://raw.githubusercontent.com/Toakley683/Valcord/refs/heads/main/icons/valcord_logo.ico")

	return export

}

func AppStartup() {

	menuStatus.SetTitle("Status: Loading..")
	menuStatus.SetTooltip("Loading the application..")

	checkUpdates()

	ImmedieteFlags()

	BeginChecks()
}

func SystraySetup() {

	Icons := LoadIcons()

	systray.SetTitle("Valcord")
	systray.SetTooltip("Valcord")

	systray.SetIcon(Icons["logo"])

	menuStatus = systray.AddMenuItemCheckbox("Deactivated", "", false)
	menuStatus.SetTooltip("Application is not active")

	systray.AddSeparator()

	menuConfig := systray.AddMenuItem("Config", "Opens the config directory")

	go func() {

		for {

			select {
			case <-menuConfig.ClickedCh:
				cmd := exec.Command(`explorer`, Types.Settings_directory)
				cmd.Run()
			}
		}

	}()

	systray.AddSeparator()
	menuQuit := systray.AddMenuItem("Quit", "Quits application")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	cls.CLS()

	AppStartup()

	select {
	case <-menuQuit.ClickedCh:
		systray.Quit()
	case <-stop:
		systray.Quit()
	}
}

func main() {

	systray.Run(SystraySetup, AppShutdown)

}
