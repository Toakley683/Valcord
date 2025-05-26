package types

import (
	"os"
	"regexp"
)

type Regional struct {
	Region string
	Shard  string
}

func GetRegionData() Regional {

	userCacheDir, err := os.UserCacheDir()
	checkError(err)

	dir := userCacheDir + "/VALORANT/Saved/Logs/ShooterGame.log"

	file, err := os.ReadFile(dir)
	checkError(err)

	shooterContents := (string(file))

	reg, err := regexp.Compile("https://glz-(.+?)-1.(.+?).a.pvp.net")
	checkError(err)

	region := reg.FindStringSubmatch(shooterContents)[1]
	shard := reg.FindStringSubmatch(shooterContents)[2]

	return Regional{
		Region: region,
		Shard:  shard,
	}

}
