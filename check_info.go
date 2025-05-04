package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MasterDimmy/go-cls"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	Types "valcord/types"
)

var (
	settings map[string]string
)

func BeginChecks() {

	settings = Types.CheckSettings()

	PasteDiscordToken()
}

func PasteDiscordToken() {

	// Request Discord Bot API Token

	if settings["discord_api_token"] == "" {

		app := tview.NewApplication()

		text := tview.NewTextView().
			SetText("Instructions on bot token: 'https://github.com/Toakley683/Valcord/wiki/Obtaining-a-Discord-Bot-Token'").
			SetTextAlign(tview.AlignLeft).
			SetDynamicColors(true).
			SetWordWrap(true)

		input := tview.NewInputField().
			SetLabel("Enter Discord Bot Token: ").
			SetPlaceholder(" [PASTE BOT TOKEN HERE]").
			SetFieldWidth(125).
			SetAcceptanceFunc(tview.InputFieldMaxLength(100))

		input.SetDoneFunc(func(key tcell.Key) {

			if key == tcell.KeyEnter {

				app.Stop()

				settings["discord_api_token"] = input.GetText()
				settings["server_id"] = ""
				settings["current_session_channel"] = ""
				Types.CheckSettingsData(settings)

				CheckForServerID()

			}

		})
		input.SetBorder(true)

		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(text, 2, 0, false).
			AddItem(input, 3, 1, true).
			AddItem(nil, 0, 1, false)

		frame := tview.NewFrame(flex).
			SetBorders(1, 1, 1, 1, 0, 0).
			AddText("╡ Valcord ╞", true, tview.AlignCenter, tview.Styles.PrimaryTextColor).
			SetBorders(1, 1, 1, 1, 0, 0)

		if err := app.SetRoot(frame, true).Run(); err != nil {
			panic(err)
		}

	} else {
		CheckForServerID()
	}

}
func CheckForServerID() {

	// Request Session Channel ID

	if settings["server_id"] == "" {

		app := tview.NewApplication()

		text := tview.NewTextView().
			SetText("Instructions on ServerID: 'https://github.com/Toakley683/Valcord/wiki/Setting-Session-Channel'").
			SetTextAlign(tview.AlignLeft).
			SetDynamicColors(true).
			SetWordWrap(true)

		input := tview.NewInputField().
			SetLabel("ServerID: ").
			SetPlaceholder(" [PASTE SERVERID HERE]").
			SetFieldWidth(125).
			SetAcceptanceFunc(tview.InputFieldMaxLength(100))

		input.SetDoneFunc(func(key tcell.Key) {

			if key == tcell.KeyEnter {

				app.Stop()

				settings["server_id"] = input.GetText()
				settings["current_session_channel"] = ""

				Types.CheckSettingsData(settings)

				CheckForChannelID()

			}

		})
		input.
			SetBorder(true).
			SetTitleAlign(tview.AlignCenter)

		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(text, 2, 0, false).
			AddItem(input, 3, 1, true).
			AddItem(nil, 0, 1, false)

		frame := tview.NewFrame(flex).
			SetBorders(1, 1, 1, 1, 0, 0).
			AddText("╡ Valcord ╞", true, tview.AlignCenter, tview.Styles.PrimaryTextColor).
			SetBorders(1, 1, 1, 1, 0, 0)

		if err := app.SetRoot(frame, true).Run(); err != nil {
			panic(err)
		}

	} else {
		CheckForChannelID()
	}

}

func CheckForChannelID() {

	// Request Session Channel ID

	if settings["current_session_channel"] == "" {

		app := tview.NewApplication()

		text := tview.NewTextView().
			SetText("Instructions on Session channelID: 'https://github.com/Toakley683/Valcord/wiki/Setting-Session-Channel'").
			SetTextAlign(tview.AlignLeft).
			SetDynamicColors(true).
			SetWordWrap(true)

		input := tview.NewInputField().
			SetLabel("Session Channel ID: ").
			SetPlaceholder(" [PASTE CHANNELID HERE]").
			SetFieldWidth(125).
			SetAcceptanceFunc(tview.InputFieldMaxLength(100))

		input.SetDoneFunc(func(key tcell.Key) {

			if key == tcell.KeyEnter {

				app.Stop()

				settings["current_session_channel"] = input.GetText()

				Types.CheckSettingsData(settings)

				CheckForOpenGame()

			}

		})
		input.
			SetBorder(true).
			SetTitleAlign(tview.AlignCenter)

		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(text, 2, 0, false).
			AddItem(input, 3, 1, true).
			AddItem(nil, 0, 1, false)

		frame := tview.NewFrame(flex).
			SetBorders(1, 1, 1, 1, 0, 0).
			AddText("╡ Valcord ╞", true, tview.AlignCenter, tview.Styles.PrimaryTextColor).
			SetBorders(1, 1, 1, 1, 0, 0)

		if err := app.SetRoot(frame, true).Run(); err != nil {
			panic(err)
		}

	}

	CheckForOpenGame()

}

type check_lockfile struct {
	name       string
	process_id int
	port       string
	password   string
	protocol   string
}

func getcheck_lockfile() (lock *check_lockfile) {

	userCacheDir, err := os.UserCacheDir()
	checkError(err)

	dir := userCacheDir + "/Riot Games/Riot Client/Config/lockfile"

	_, err = os.Stat(dir)

	if errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist
		return nil
	}
	checkError(err)

	file, err := os.ReadFile(dir)
	checkError(err)

	lockfileContents := (string(file))

	split := strings.Split(lockfileContents, ":")

	convertedPort, err := strconv.Atoi(split[1])
	checkError(err)

	return &check_lockfile{
		name:       split[0],
		process_id: convertedPort,
		port:       split[2],
		password:   split[3],
		protocol:   split[4],
	}

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

		lockfile := getcheck_lockfile()

		if lockfile == nil {
			fmt.Println("Client: Closed")
			fmt.Println("Game is not open, retrying..")

			time.Sleep(RetryDelay)
			continue
		}

		req, err := http.NewRequest("GET", "https://127.0.0.1:"+lockfile.port+"/entitlements/v1/token", nil)
		checkError(err)

		req.Header.Add("Authorization", "Basic "+Types.BasicAuth("riot", lockfile.password))

		Res, Err := Client.Do(req)
		res = Res

		if Err != nil {

			splitError := strings.Split(Err.Error(), " ")
			finalError := strings.Join(splitError[6:], " ")

			if finalError == "No connection could be made because the target machine actively refused it." {

				// Game is not open

				fmt.Println("Client: Closed")
				fmt.Println("Game is not open, retrying..")

				time.Sleep(RetryDelay)
				continue

			}

			checkError(Err)

		}

		fmt.Println("Client: Open")

		break

	}

	time.Sleep(time.Second * 1)
	cls.CLS()

	res.Body.Close()

	AppStartup()

}
