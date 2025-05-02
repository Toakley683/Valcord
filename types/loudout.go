package types

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type LoadoutItem struct {
	DisplayIcon string
	TypeID      string
	Buddy       Buddy
	VideoStream string
	weaponInfo  WeaponInfo
}

type Expression struct {
	AssetID string
	TypeID  string
	Name    string
	IconURL string
}

type Loadout struct {
	Expressions []Expression
	Items       map[string]LoadoutItem
}

type WeaponInfo struct {
	displayName      string
	contentTierName  string
	contentTierIcon  string
	contentTierRank  int
	contentTierColor string
}

var (
	WeaponIDToName = map[string]string{}
)

func Init_Weapons() {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/weapons", nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var weapon_data map[string]interface{}

	weapon_data, err = GetJSON(res)
	checkError(err)

	weapons := weapon_data["data"].([]interface{})

	for _, weapon := range weapons {

		weapon := weapon.(map[string]interface{})

		WeaponIDToName[weapon["uuid"].(string)] = weapon["displayName"].(string)

	}

}

func GetWeaponFromID(ID string) WeaponInfo {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/weapons/skins/"+ID, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var weapon_data map[string]interface{}

	weapon_data, err = GetJSON(res)
	checkError(err)

	weapon := weapon_data["data"].(map[string]interface{})

	contentTierName := ""
	contentTierIcon := ""
	contentTierRank := 0
	contentTierColor := ""

	if weapon["contentTierUuid"] != nil {

		req, err = http.NewRequest("GET", "https://valorant-api.com/v1/contenttiers/"+weapon["contentTierUuid"].(string), nil)
		checkError(err)

		res, err = Client.Do(req)
		checkError(err)

		defer res.Body.Close()

		var tier_data map[string]interface{}

		tier_data, err = GetJSON(res)
		checkError(err)

		tier := tier_data["data"].(map[string]interface{})

		contentTierName = tier["devName"].(string)
		contentTierIcon = "https://media.valorant-api.com/contenttiers/" + weapon["contentTierUuid"].(string) + "/displayicon.png"
		contentTierRank = int(tier["rank"].(float64))
		contentTierColor = tier["highlightColor"].(string)

	}

	return WeaponInfo{
		displayName:      weapon["displayName"].(string),
		contentTierName:  contentTierName,
		contentTierIcon:  contentTierIcon,
		contentTierRank:  contentTierRank,
		contentTierColor: contentTierColor,
	}

}

func GetUnlockedAgents(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) []PlayableAgent {

	//pd.{Shard}.a.pvp.net/store/v1/entitlements/{PlayerID}/01bb38e1-da47-4e6a-9b3d-945fe4655707 < AgentTypeID

	req, err := http.NewRequest("GET", "https://pd."+regions.shard+".a.pvp.net/store/v1/entitlements/"+player.sub+"/01bb38e1-da47-4e6a-9b3d-945fe4655707", nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.riotClientVersion)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var tier_data map[string]interface{}

	tier_data, err = GetJSON(res)
	checkError(err)

	if tier_data["Entitlements"] == nil {
		return []PlayableAgent{} // Return Empty agent list
	}

	agents := tier_data["Entitlements"].([]interface{})

	agentArray := make([]PlayableAgent, len(agents))

	for I, AgentData := range agents {

		AgentData := AgentData.(map[string]interface{})

		AgentID := AgentData["ItemID"].(string)

		agentArray[I] = AgentDetails[AgentID]

	}

	return agentArray

}

func GetItem(itemD interface{}, loadoutItem chan LoadoutItem) {

	item := itemD.(map[string]interface{})["Sockets"].(map[string]interface{})

	// Check skin chroma ( ID = "3ad1b2b2-acdb-4524-852f-954a76ddae0a" )
	SkinChroma := item["3ad1b2b2-acdb-4524-852f-954a76ddae0a"].(map[string]interface{})["Item"].(map[string]interface{})

	// Check weaponID ( ID = "bcef87d6-209b-46c6-8b19-fbe40bd95abc" )
	WeaponID := item["bcef87d6-209b-46c6-8b19-fbe40bd95abc"].(map[string]interface{})["Item"].(map[string]interface{})["ID"].(string)
	WeaponInfo := GetWeaponFromID(WeaponID)

	// Check buddyID ( ID = "dd3bf334-87f3-40bd-b043-682a57a8dc3a" )

	BuddyInfo := Buddy{}

	if item["dd3bf334-87f3-40bd-b043-682a57a8dc3a"] != nil {

		BuddyID := item["dd3bf334-87f3-40bd-b043-682a57a8dc3a"].(map[string]interface{})["Item"].(map[string]interface{})["ID"].(string)
		BuddyInfo = BuddyData(BuddyID)
	}

	video_link := ""

	if SkinChroma["streamedVideo"] != nil {
		video_link = SkinChroma["streamedVideo"].(string)
	}

	DisplayIcon := ""

	if SkinChroma["ID"] != nil {

		DisplayIcon = "https://media.valorant-api.com/weaponskinchromas/" + SkinChroma["ID"].(string) + "/fullrender.png"

	}

	loadoutItem <- LoadoutItem{
		DisplayIcon: DisplayIcon,
		TypeID:      itemD.(map[string]interface{})["ID"].(string),
		Buddy:       BuddyInfo,
		VideoStream: video_link,
		weaponInfo:  WeaponInfo,
	}

}

func GetLoadout(LoudoutInfo map[string]interface{}, PUUID string) map[string]Loadout {

	// map[string]Loadout ( Index == Player UUID )

	if LoudoutInfo["Loadouts"] == nil {

		return map[string]Loadout{}

	}

	loadouts := LoudoutInfo["Loadouts"].([]interface{})

	fmt.Println("Loadout count: " + strconv.Itoa(len(loadouts)))

	Loadouts := map[string]Loadout{}

	for _, playerLoadout := range loadouts {

		pLoudout := playerLoadout.(map[string]interface{})

		var playerLoadout map[string]interface{}

		if pLoudout["Loadout"] == nil {
			playerLoadout = pLoudout
		} else {
			playerLoadout = pLoudout["Loadout"].(map[string]interface{})
		}

		if PUUID != "" {

			if playerLoadout["Subject"].(string) != PUUID {
				continue
			}

		}

		// Expressions

		expressions := playerLoadout["Expressions"].(map[string]interface{})["AESSelections"].([]interface{})
		expressionsMap := make([]Expression, len(expressions))

		for I, expression := range expressions {

			expression := expression.(map[string]interface{})
			Type := ""

			switch expression["TypeID"].(string) {
			case "03a572de-4234-31ed-d344-ababa488f981":
				Type = "Flex"
			case "d5f120f8-ff8c-4aac-92ea-f2b5acbe9475":
				Type = "Spray"
			}

			Name := ""
			IconURL := ""

			if Type == "Spray" {

				sprayData := SprayData(expression["AssetID"].(string))

				Name = sprayData.displayName
				IconURL = sprayData.fullTransparent

			}

			if Type == "Flex" {

				flexData := FlexData(expression["AssetID"].(string))

				Name = flexData.displayName
				IconURL = flexData.displayIcon

			}

			if Type == "" {
				checkError(errors.New("Player loadout, expression type unknown: (" + expression["TypeID"].(string) + ")"))
			}

			expressionsMap[I] = Expression{
				AssetID: expression["AssetID"].(string),
				TypeID:  expression["TypeID"].(string),
				Name:    Name,
				IconURL: IconURL,
			}

		}

		// Items

		items := playerLoadout["Items"].(map[string]interface{})
		itemMap := map[string]LoadoutItem{}

		for Type, itemD := range items {

			loadout := make(chan LoadoutItem)

			go GetItem(itemD, loadout)

			itemMap[Type] = <-loadout

		}

		loadout := Loadout{
			Expressions: expressionsMap,
			Items:       itemMap,
		}

		Loadouts[playerLoadout["Subject"].(string)] = loadout

	}

	return Loadouts

}

func GetMatchLoudout(matchUUID string, PUUID string, player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) map[string]Loadout {

	//"https://pd." + regions.shard + ".a.pvp.net/mmr/v1/players/" + PlayerUUID

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/core-game/v1/matches/"+matchUUID+"/loadouts", nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.riotClientVersion)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var loadout_information map[string]interface{}

	loadout_information, err = GetJSON(res)
	checkError(err)

	return GetLoadout(loadout_information, PUUID)

}

func GetAgentSelectLoudout(matchUUID string, PUUID string, player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) map[string]Loadout {

	//"https://pd." + regions.shard + ".a.pvp.net/mmr/v1/players/" + PlayerUUID

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/pregame/v1/matches/"+matchUUID+"/loadouts", nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.riotClientVersion)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var loadout_information map[string]interface{}

	loadout_information, err = GetJSON(res)
	checkError(err)

	return GetLoadout(loadout_information, PUUID)
}
