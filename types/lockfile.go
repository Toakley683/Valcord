package types

import (
	"os"
	"strconv"
	"strings"
)

type Lockfile_type struct {
	name       string
	process_id int
	port       string
	password   string
	protocol   string
}

// Gets the local Lockfile from "%AppDataLocal%/Riot Games/Riot Client/Config/lockfile" -> Required for access token

func GetLockfile() (lock Lockfile_type) {

	userCacheDir, err := os.UserCacheDir()
	checkError(err)

	dir := userCacheDir + "/Riot Games/Riot Client/Config/lockfile"

	file, err := os.ReadFile(dir)
	checkError(err)

	lockfileContents := (string(file))

	split := strings.Split(lockfileContents, ":")

	convertedPort, err := strconv.Atoi(split[1])
	checkError(err)

	return Lockfile_type{
		name:       split[0],
		process_id: convertedPort,
		port:       split[2],
		password:   split[3],
		protocol:   split[4],
	}

}
