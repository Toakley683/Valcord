package types

import (
	"errors"
	"io/fs"
	"os"

	"github.com/goccy/go-yaml"
)

var (
	config_dir, _      = os.UserConfigDir()
	directory_name     = "Valcord"
	settings_directory = config_dir + "/" + directory_name
	settings_file      = settings_directory + "/settings.yaml"
)

func CheckSettingsData(settings map[string]string) map[string]string {

	default_settings := get_default_settings()

	for Index, Value := range default_settings {

		_, has := settings[Index]

		if !has {

			settings[Index] = Value

		}

	}

	// File doesn't exist

	data, err := yaml.Marshal(settings)
	checkError(err)

	err = os.WriteFile(settings_file, data, 0700)
	checkError(err)

	return settings

}

func check_settings_file() map[string]string {

	// Check if Directory exists, if not create it

	settings_data := get_default_settings()

	_, err := os.Stat(settings_file)

	if errors.Is(err, fs.ErrNotExist) {

		// File doesn't exist

		data, err := yaml.Marshal(settings_data)
		checkError(err)

		err = os.WriteFile(settings_file, data, 0700)
		checkError(err)

		check_settings_file()

		return nil

	}
	checkError(err)

	if err == nil {

		// File exists, read file and output settings

		data, err := os.ReadFile(settings_file)
		checkError(err)

		var settings map[string]string

		err = yaml.Unmarshal(data, &settings)
		checkError(err)

		return CheckSettingsData(settings)

	}

	return nil

}

func check_directory() map[string]string {

	// Check if Directory exists, if not create it

	_, err := os.Stat(settings_directory)

	if errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist

		err := os.Mkdir(settings_directory, 0700)
		checkError(err)

		check_directory()
		return nil
	}
	checkError(err)

	if err == nil {

		// Directory exists, continue steps
		return check_settings_file()

	}
	return nil

}

func get_default_settings() map[string]string {

	var settings map[string]string = map[string]string{
		"discord_api_token":       "DISCORD_TOKEN_HERE",
		"server_id":               "SERVER_ID_HERE",
		"owner_userid":            "USER_ID",
		"current_session_channel": "DO NOT SET MANUALLY",
	}

	return settings

}

func check_settings() map[string]string {

	return check_directory()

}

var (
	Settings map[string]string = check_settings()
)
