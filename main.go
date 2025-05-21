package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/MasterDimmy/go-cls"
	"github.com/getlantern/systray"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/ncruces/zenity"

	Types "valcord/types"
)

func checkError(err error) {
	if err != nil {
		Types.NewLog(err)
		zenity.Error(err.Error(),
			zenity.Title("Vencord: Error"))
		log.Fatal(err)
	}
}

var (
	general_valorant_information ValorantInformation

	Flags = map[string]bool{}

	menuStatus = &systray.MenuItem{}
	menuUpdate = &systray.MenuItem{}

	ValcordLockfile *os.File

	menuListenForMatch *bool = Pointer(true)
)

func Pointer[T any](d T) *T {
	return &d
}

type ValorantInformation struct {
	lock_file     Types.Lockfile_type
	entitlements  Types.EntitlementsTokenResponse
	player_info   Types.PlayerInfo
	regional_data Types.Regional
}

func cleanDiscordBot() {
	Types.NewLog("\t - Cleaning up discord bot..")

	if discord != nil {

		Types.NewLog("\t - Discord Client: Closing")
		go discord.Close()
		Types.NewLog("\t - Discord Client: Closed")
	}
}

func cleanLockFile() {

	Types.NewLog("\t - Cleaning up lock file..")

	_, err := os.Stat(Types.LockFileDir)

	if err == nil {

		// File exists

		Types.NewLog("\t - Lock File: Found")

		Data, _ := os.ReadFile(Types.LockFileDir)

		if string(Data) == strconv.Itoa(os.Getpid()) {

			// Check if contents of lockfile are == to PID (Process ID) ensuring only 1 process runs at any time

			Types.NewLog("\t - Lock File: PID-" + string(Data))

			go ValcordLockfile.Close()

			err := os.Remove(Types.LockFileDir)

			if err != nil {
				Types.NewLog(err)
			} else {
				Types.NewLog("\t - Lock File: Closed")
			}

		}

	}
}

func cleanLogFile() {

	Types.NewLog(" - Cleaning up Log File..")

	if Types.LogFile != nil {

		Types.NewLog(" - Log File: Closing")
		Types.NewLog(" - Log File: Closed")
		go Types.LogFile.Close()

	}
}

func cleanup() {
	Types.NewLog("Cleanup: Pending\n")

	Types.NewLog("cleanUpSequence: [")

	sequence := make([]func(), 2)

	sequence[0] = cleanLockFile
	sequence[1] = cleanDiscordBot

	for Index, f := range sequence {
		f()

		if Index == len(sequence)-1 {
			continue
		}

		Types.NewLog("")
	}

	defer cleanLogFile()

	Types.NewLog("]\n")
	Types.NewLog("Cleanup: Complete\n")

}

func AppShutdown() {

	Types.NewLog("Shutting down..\n")

	cleanup()

	cls.CLS()

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

func makeLink(src, dst string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()
	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", dst)
	if err != nil {
		return err
	}
	idispatch := cs.ToIDispatch()
	oleutil.PutProperty(idispatch, "TargetPath", src)
	oleutil.CallMethod(idispatch, "Save")
	return nil
}

func setStartMenu(onStartMenu bool) bool {

	AppdataDir, err := os.UserConfigDir()
	checkError(err)

	AppName := "Valcord.LNK"

	Dir := AppdataDir + `\Microsoft\Windows\Start Menu\Programs\` + AppName

	curDir, err := os.Executable()
	checkError(err)

	_, err = os.Stat(curDir)

	if err != nil {

		if errors.Is(err, fs.ErrNotExist) {
			return false
		}

	}

	switch onStartMenu {
	case true:
		E := makeLink(curDir, Dir)
		checkError(E)
	case false:
		E := os.Remove(Dir)
		checkError(E)
	}

	return true

}

func createInsLockfile() {

	Data := []byte(strconv.Itoa(os.Getpid()))

	err := os.WriteFile(Types.LockFileDir, Data, 0700)
	checkError(err)

	ValcordLockfile, err = os.OpenFile(Types.LockFileDir, os.O_RDWR, 0777)
	checkError(err)

}

func SystraySetup() {

	settings = Types.CheckSettings()

	defer systray.Quit()

	// Check to see if lockfile is there, otherwise close

	_, err := os.Stat(Types.LockFileDir)

	if errors.Is(err, fs.ErrNotExist) {

		// File does not exist

		createInsLockfile()

	} else {
		checkError(err)
	}

	if err == nil {

		err := os.Remove(Types.LockFileDir)

		if err != nil {

			Types.NewLog("This process is already running")

			zenity.Error("Process is already running",
				zenity.Title("Valcord"))

			log.Panicln("Process is already running")

		}

		createInsLockfile()

	}

	// Load system tray

	Icons := LoadIcons()

	systray.SetTitle("Valcord")
	systray.SetTooltip("Valcord")

	systray.SetIcon(Icons["logo"])

	Title := systray.AddMenuItem("Valcord", "")
	Title.Disable()

	systray.AddSeparator()

	menuStatus = systray.AddMenuItemCheckbox("Deactivated", "", false)
	menuStatus.SetTooltip("Application is not active")
	menuStatus.Disable()

	menuUpdate = systray.AddMenuItemCheckbox("Checking for updates..", "", false)
	menuUpdate.SetTooltip("Application is checking for update..")
	menuUpdate.Disable()

	systray.AddSeparator()

	Title = systray.AddMenuItem("Settings", "")
	Title.Disable()

	saved_lfm, err := strconv.ParseBool(settings["listen_for_matches"])
	if err != nil {
		saved_lfm = false
		Types.NewLog(err)
	}

	menuMatchListen := systray.AddMenuItemCheckbox("Listen for Matches", "Do you want match information to auto-post", saved_lfm)

	saved_sm, err := strconv.ParseBool(settings["in_startmenu"])
	if err != nil {
		saved_sm = false
		Types.NewLog(err)
	}

	setStartMenu(saved_sm)

	menuStartMenu := systray.AddMenuItemCheckbox("In Start Menu", "Would you like app to be able to open from Start Menu?", saved_sm)

	systray.AddSeparator()

	Title = systray.AddMenuItem("Information", "")
	Title.Disable()

	menuConfig := systray.AddMenuItem("Config", "Opens the config directory")
	menuDiscordBotInvite := systray.AddMenuItem("Discord Bot Invite", "Invites the bot to X server")

	systray.AddSeparator()

	menuCommandReload := systray.AddMenuItem("Reload Commands", "Clears and reloads the commands for the discord bot")
	menuCommandReloadConfirm := menuCommandReload.AddSubMenuItem("Confirm", "THIS MAY CAUSE PROBLEMS / DO NOT USE OFTEN")

	var listenForMatch bool = menuMatchListen.Checked()

	menuListenForMatch = &listenForMatch

	go func() {

		for {

			select {
			case <-menuMatchListen.ClickedCh:

				// Change whether program will listen for a new match automatically

				switch menuMatchListen.Checked() {
				case true:
					menuMatchListen.Uncheck()
				case false:
					menuMatchListen.Check()
				}

				Types.NewLog("Match listening set to:", menuMatchListen.Checked())
				*menuListenForMatch = menuMatchListen.Checked()

				settings["listen_for_matches"] = strconv.FormatBool(menuMatchListen.Checked())
				Types.CheckSettingsData(settings)

			case <-menuUpdate.ClickedCh:
				Types.NewLog("Clicked")

				link := "https://github.com/Toakley683/Valcord/wiki/Updating-Valcord"

				cmd := "cmd.exe"
				args := []string{"/c", "start", link}

				exec.Command(cmd, args...).Start()

			case <-menuStartMenu.ClickedCh:

				// Set flag to start this program on program start

				if !setStartMenu(!menuStartMenu.Checked()) {
					continue
				}

				switch menuStartMenu.Checked() {
				case true:
					menuStartMenu.Uncheck()
				case false:
					menuStartMenu.Check()
				}

				Types.NewLog("Run on start menu set to:", menuStartMenu.Checked())

				settings["in_startmenu"] = strconv.FormatBool(menuStartMenu.Checked())
				Types.CheckSettingsData(settings)

			case <-menuDiscordBotInvite.ClickedCh:

				if discord == nil {
					zenity.Error("Discord bot is not online, could not get invite link")
					continue
				}

				InviteLink := "https://discord.com/oauth2/authorize?client_id=" + discord.State.User.ID + "&permissions=93184&integration_type=0&scope=bot'"

				cmd := "cmd.exe"
				args := []string{"/c", "start", InviteLink}

				exec.Command(cmd, args...).Start()

			case <-menuConfig.ClickedCh:

				// Open the config directory for ease of access

				cmd := exec.Command(`explorer`, Types.Settings_directory)
				cmd.Run()

			case <-menuCommandReloadConfirm.ClickedCh:

				// Reload discord bot commands

				command_cleanup()
				commandInit()

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
