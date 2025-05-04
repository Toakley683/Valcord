package types

import (
	"errors"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

type Lockfile_type struct {
	Name       string
	Process_id int
	Port       string
	Password   string
	Protocol   string
}

// Gets the local Lockfile from "%AppDataLocal%/Riot Games/Riot Client/Config/lockfile" -> Required for access token

func GetLockfile(doError bool) (lock Lockfile_type) {

	userCacheDir, err := os.UserCacheDir()
	checkError(err)

	dir := userCacheDir + "/Riot Games/Riot Client/Config/lockfile"

	_, err = os.Stat(dir)

	if errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist

		if doError {
			checkError(errors.New("no lockfile detected, game is closed"))
		}

		return Lockfile_type{}
	}
	checkError(err)

	file, err := os.ReadFile(dir)
	checkError(err)

	lockfileContents := (string(file))

	split := strings.Split(lockfileContents, ":")

	convertedPort, err := strconv.Atoi(split[1])
	checkError(err)

	return Lockfile_type{
		Name:       split[0],
		Process_id: convertedPort,
		Port:       split[2],
		Password:   split[3],
		Protocol:   split[4],
	}

}
