package types

import (
	"log"
	"net/http"
	"strconv"
)

type SeasonExtra struct {
	displayName string
	title       string
	startTime   string
	endTime     string
}

type Season struct {
	PeakRank                   int
	SeasonID                   string
	ExtraData                  SeasonExtra
	NumberOfWins               int
	NumberOfWinsWithPlacements int
	NumberOfGames              int
	Rank                       int
	CapstoneWins               int
	LeaderboardRank            int
	CompetitiveTier            int
	RankedRating               int
	WinsByTier                 map[string]int
	GamesNeededForRating       int
	TotalWinsNeededForRank     int
}

type MMRGamemode struct {
	GamemodeType                      string
	TotalGamesNeededForRating         int
	TotalGamesNeededForLeaderboard    int
	CurrentSeasonGamesNeededForRating int
	Seasons                           []Season
}

type Career struct {
	PeakRank                      int
	CurrentRank                   int
	RankedRating                  int
	Competitive                   MMRGamemode
	Deathmatch                    MMRGamemode
	Ggteam                        MMRGamemode
	Hurm                          MMRGamemode
	Seeding                       MMRGamemode
	Spikerush                     MMRGamemode
	Swiftplay                     MMRGamemode
	Unrated                       MMRGamemode
	DerankProtectedGamesRemaining int
}

func GetSeasonExtraData(UUID string) SeasonExtra {

	//https://valorant-api.com/v1/seasons/

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/seasons/"+UUID, nil)
	checkError(err)

	res, err := client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var season_information map[string]interface{}

	season_information, err = GetJSON(res)
	checkError(err)

	season_information = season_information["data"].(map[string]interface{})

	var displayName string = ""
	var title string = ""
	var startTime string = ""
	var endTime string = ""

	if season_information["displayName"] != nil {
		displayName = season_information["displayName"].(string)
	}

	if season_information["title"] != nil {
		title = season_information["title"].(string)
	}

	if season_information["startTime"] != nil {
		startTime = season_information["startTime"].(string)
	}

	if season_information["endTime"] != nil {
		endTime = season_information["endTime"].(string)
	}

	return SeasonExtra{
		displayName: displayName,
		title:       title,
		startTime:   startTime,
		endTime:     endTime,
	}

}

func GetSeasons(gamemodeName string, gamemodeData map[string]interface{}) ([]Season, int) {

	SeasonalInfo := gamemodeData["SeasonalInfoBySeasonID"].(map[string]interface{})

	Seasons := make([]Season, len(SeasonalInfo))

	SeasonNumID := 0

	var PeakRank int = 0

	for _, SeasonData := range SeasonalInfo {

		SeasonData := SeasonData.(map[string]interface{})

		var WinsByTier map[string]int

		if SeasonData["WinsByTier"] != nil {

			winsByTier_data := SeasonData["WinsByTier"].(map[string]interface{})

			WinsByTier = make(map[string]int, len(winsByTier_data))

			for Index, Value := range winsByTier_data {

				if gamemodeName == "competitive" {

					R, err := strconv.Atoi(Index)
					checkError(err)

					if R > PeakRank {

						PeakRank = R

					}
				}

				WinsByTier[Index] = int(Value.(float64))

			}

		} else {
			WinsByTier = make(map[string]int, 0)
		}

		Seasons[SeasonNumID] = Season{
			SeasonID:                   SeasonData["SeasonID"].(string),
			ExtraData:                  GetSeasonExtraData(SeasonData["SeasonID"].(string)),
			NumberOfWins:               int(SeasonData["NumberOfWins"].(float64)),
			NumberOfWinsWithPlacements: int(SeasonData["NumberOfWinsWithPlacements"].(float64)),
			NumberOfGames:              int(SeasonData["NumberOfGames"].(float64)),
			Rank:                       int(SeasonData["Rank"].(float64)),
			CapstoneWins:               int(SeasonData["CapstoneWins"].(float64)),
			LeaderboardRank:            int(SeasonData["LeaderboardRank"].(float64)),
			CompetitiveTier:            int(SeasonData["CompetitiveTier"].(float64)),
			RankedRating:               int(SeasonData["RankedRating"].(float64)),
			WinsByTier:                 WinsByTier,
			GamesNeededForRating:       int(SeasonData["GamesNeededForRating"].(float64)),
			TotalWinsNeededForRank:     int(SeasonData["TotalWinsNeededForRank"].(float64)),
		}

		SeasonNumID = SeasonNumID + 1

	}

	return Seasons, PeakRank

}

func GetPlayerMMR(regions Regional, entitlement EntitlementsTokenResponse, player PlayerInfo, PlayerUUID string, ReturnedCareers map[string]bool) Career {

	//"https://pd." + regions.shard + ".a.pvp.net/mmr/v1/players/" + PlayerUUID

	req, err := http.NewRequest("GET", "https://pd."+regions.shard+".a.pvp.net/mmr/v1/players/"+PlayerUUID, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.riotClientVersion)

	res, err := client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var match_information map[string]interface{}

	match_information, err = GetJSON(res)
	checkError(err)

	queueSkills := match_information["QueueSkills"].(map[string]interface{})

	gamemodes := make(map[string]MMRGamemode, len(queueSkills))

	PeakRank := 0

	for gamemodeName, gamemodeData := range queueSkills {

		log.Println("Trying type: " + gamemodeName)

		if len(ReturnedCareers) != 0 {

			if !ReturnedCareers[gamemodeName] {

				continue

			}

		}

		gamemodeData := gamemodeData.(map[string]interface{})

		Season, Peak_Rank := GetSeasons(gamemodeName, gamemodeData)

		if Peak_Rank > PeakRank {

			PeakRank = Peak_Rank

		}

		gamemode := MMRGamemode{
			GamemodeType:                      gamemodeName,
			TotalGamesNeededForRating:         int(gamemodeData["TotalGamesNeededForRating"].(float64)),
			CurrentSeasonGamesNeededForRating: int(gamemodeData["CurrentSeasonGamesNeededForRating"].(float64)),
			Seasons:                           Season,
		}

		log.Println("Getting Career type: " + gamemodeName)

		gamemodes[gamemodeName] = gamemode

	}

	latestComp := match_information["LatestCompetitiveUpdate"].(map[string]interface{})

	career := Career{
		PeakRank:                      PeakRank,
		CurrentRank:                   int(latestComp["TierAfterUpdate"].(float64)),
		RankedRating:                  int(latestComp["RankedRatingAfterUpdate"].(float64)),
		Competitive:                   gamemodes["competitive"],
		Deathmatch:                    gamemodes["deathmatch"],
		Ggteam:                        gamemodes["ggteam"],
		Hurm:                          gamemodes["hurm"],
		Seeding:                       gamemodes["seeding"],
		Spikerush:                     gamemodes["spikerush"],
		Swiftplay:                     gamemodes["swiftplay"],
		Unrated:                       gamemodes["unrated"],
		DerankProtectedGamesRemaining: int(match_information["DerankProtectedGamesRemaining"].(float64)),
	}

	return career

}
