package types

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/ncruces/zenity"
)

var (
	config_dir, _      = os.UserConfigDir()
	directory_name     = "Valcord"
	Settings_directory = config_dir + "\\" + directory_name
	Logs_directory     = Settings_directory + "\\logs"
	Settings_file      = Settings_directory + "\\settings.yaml"
	AppFileDir         = Settings_directory + "\\application\\valcord.exe"
	LockFileDir        = Settings_directory + "\\lockfile.pid"

	Settings map[string]string
)

func CheckSettingsData(settings map[string]string) map[string]string {

	default_settings := get_default_settings()

	if settings == nil {
		return default_settings
	}

	for Index, Value := range default_settings {

		_, has := settings[Index]

		if !has {

			settings[Index] = Value

		}

	}

	// File doesn't exist

	data, err := yaml.Marshal(settings)
	checkError(err)

	err = os.WriteFile(Settings_file, data, 0700)
	checkError(err)

	return settings

}

func check_settings_file() map[string]string {

	// Check if File exists, if not create it

	settings_data := get_default_settings()

	_, err := os.Stat(Settings_file)

	if errors.Is(err, fs.ErrNotExist) {

		// File doesn't exist

		data, err := yaml.Marshal(settings_data)
		checkError(err)

		err = os.WriteFile(Settings_file, data, 0700)
		checkError(err)

		return check_settings_file()

	}
	checkError(err)

	if err == nil {

		// File exists, read file and output settings

		data, err := os.ReadFile(Settings_file)
		checkError(err)

		var settings map[string]string

		err = yaml.Unmarshal(data, &settings)
		checkError(err)

		return CheckSettingsData(settings)

	}
	return nil

}

func putAppToDir(NewDir string) {

	CurDir, err := os.Executable()
	checkError(err)

	data, _ := os.Stat(AppFileDir)

	if CurDir == AppFileDir {

		NewLog("Executable in correct spot, will continue..")
		return

	}

	if data != nil {
		NewLog("Found file in application folder..")
		NewLog("Removing old application")

		err := os.Remove(AppFileDir)

		if err != nil {

			zenity.Error("Could not remove old update file..\n\n" + err.Error())

		}
		checkError(err)

		NewLog("Removed old application")
	}

	renameErr := os.Rename(CurDir, AppFileDir)

	if renameErr != nil {

		// Is error

		Spaced := strings.Split(renameErr.Error(), " ")
		spaced := strings.Join(Spaced[len(Spaced)-3:], " ")

		switch spaced {
		case "different disk drive.":
			NewLog("Not on disk")

			data, err := os.ReadFile(CurDir)
			checkError(err)

			err = os.WriteFile(AppFileDir, data, 0700)
			checkError(err)

			zenity.Info("Please remove current executable, executable will be found in file that will open",
				zenity.Title("Valcord"),
			)

			cmd := exec.Command(`explorer`, NewDir)
			cmd.Run()

			renameErr = nil

			os.Exit(0)
			return

		case "Access is denied.":
			NewLog("Access was denied for setting file position")
			zenity.Error("No access to file, retry..",
				zenity.Title("Valcord"))
			return
		}

		NewLog(spaced)

	}
	checkError(renameErr)

	NewLog("File has been relocated")

}

func check_app_dir() {

	// Check if Directory exists, if not create it

	dir := Settings_directory + `\application`

	_, err := os.Stat(dir)

	if errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist

		err := os.Mkdir(dir, 0700)
		checkError(err)

		check_app_dir()
		return
	}
	checkError(err)

	if err == nil {

		// Directory exists, put current app there
		putAppToDir(dir)

	}

}

func check_directory() map[string]string {

	// Check if Directory exists, if not create it

	_, err := os.Stat(Settings_directory)

	if errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist

		err := os.Mkdir(Settings_directory, 0700)
		checkError(err)

		check_directory()
		return nil
	}
	checkError(err)

	if err == nil {

		// Directory exists, continue steps
		check_app_dir()

		return check_settings_file()

	}
	check_directory()
	return nil

}

func get_default_settings() map[string]string {

	var settings map[string]string = map[string]string{
		"discord_api_token":       "",
		"server_id":               "",
		"owner_userid":            "",
		"current_session_channel": "",
		"listen_for_matches":      "true",
		"in_startmenu":            "true",
	}

	return settings

}

func CheckSettings() map[string]string {

	Info := check_directory()

	Settings = Info

	return Info

}

var (
// Settings = CheckSettings()
)
