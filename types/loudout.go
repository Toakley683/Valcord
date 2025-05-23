package types

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"
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

type OwnedEntitlement struct {
	TypeID string
	ItemID string
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

func GetOwnedItems(player PlayerInfo, regions Regional, itemType string) []OwnedEntitlement {

	entitlement := GetEntitlementsToken(GetLockfile(true))

	req, err := http.NewRequest("GET", "https://pd."+regions.shard+".a.pvp.net/store/v1/entitlements/"+player.sub+"/"+itemType, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.riotClientVersion)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var item_data map[string]interface{}

	item_data, err = GetJSON(res)
	checkError(err)

	if item_data["Entitlements"] == nil {
		return []OwnedEntitlement{} // Return Empty agent list
	}

	data := item_data["Entitlements"].([]interface{})

	var finalData []interface{}

	if itemType == "e7c63390-eda7-46e0-bb7a-a6abdacd2433" {

		type IndexedItem struct {
			Index int
			Value interface{}
		}

		finalDataChan := make(chan IndexedItem, len(data))

		var wg1 sync.WaitGroup

		var finalIndex int32 = 0

		for _, Val := range data {

			wg1.Add(1)

			go func(Val interface{}) {

				defer wg1.Done()

				V := Val.(map[string]interface{})

				ItemData := ItemIDWTypeToStruct(itemType, V["ItemID"].(string), 0)

				if ItemData.LevelItem == "" {

					index := int(atomic.AddInt32(&finalIndex, 1)) - 1

					finalDataChan <- IndexedItem{
						Index: index,
						Value: Val,
					}

				}

			}(Val)

		}

		wg1.Wait()
		close(finalDataChan)

		finalData = make([]interface{}, len(finalDataChan))

		for i := range finalDataChan {

			finalData[i.Index] = i.Value

			NewLog(i.Index)

		}

		NewLog(len(finalDataChan))

	} else {
		finalData = data
	}

	DataLength := len(finalData)

	if itemType == "01bb38e1-da47-4e6a-9b3d-945fe4655707" {

		// Is Agent

		// Add default agents because they aren't included

		DataLength = len(finalData) + len(DefaultAgents)

	}

	returnedData := make([]OwnedEntitlement, DataLength)

	for Index, Data := range finalData {

		ItemData := Data.(map[string]interface{})

		returnedData[Index] = OwnedEntitlement{
			TypeID: ItemData["TypeID"].(string),
			ItemID: ItemData["ItemID"].(string),
		}

	}

	if itemType == "01bb38e1-da47-4e6a-9b3d-945fe4655707" {

		// Is Agent

		// Add default agents because they aren't included

		I := 0

		for _, Value := range DefaultAgents {

			returnedData[len(data)+I] = OwnedEntitlement{
				TypeID: "4e60e748-bce6-4faa-9327-ebbe6089d5fe",
				ItemID: Value.UUID,
			}

			I++

		}

	}

	return returnedData

}

func GetUnlockedAgents(player PlayerInfo, regions Regional) []PlayableAgent {

	//pd.{Shard}.a.pvp.net/store/v1/entitlements/{PlayerID}/01bb38e1-da47-4e6a-9b3d-945fe4655707 < AgentTypeID

	entitlement := GetEntitlementsToken(GetLockfile(true))

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

func GetItem(itemD interface{}) LoadoutItem {

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

	return LoadoutItem{
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

	type ChanItem struct {
		Index string
		Value Loadout
	}

	Loadouts := map[string]Loadout{}
	loadoutOutput := make(chan ChanItem, len(loadouts))

	var wg sync.WaitGroup

	for _, playerLoadout := range loadouts {

		wg.Add(1)

		go func(plrLoadout interface{}) {

			defer wg.Done()

			pLoudout := plrLoadout.(map[string]interface{})

			var playerLoadout map[string]interface{}

			if pLoudout["Loadout"] == nil {
				playerLoadout = pLoudout
			} else {
				playerLoadout = pLoudout["Loadout"].(map[string]interface{})
			}

			if PUUID != "" {

				if playerLoadout["Subject"].(string) != PUUID {
					return
				}

			}

			// Expressions

			expressions := playerLoadout["Expressions"].(map[string]interface{})["AESSelections"].([]interface{})
			expressionsMap := make([]Expression, len(expressions))

			for I, expression := range expressions {

				expression := expression.(map[string]interface{})

				if expression["AssetID"] == nil {
					continue
				}

				Data := ItemIDWTypeToStruct(expression["TypeID"].(string), expression["AssetID"].(string), 1)

				expressionsMap[I] = Expression{
					AssetID: Data.ItemID,
					TypeID:  Data.ItemTypeID,
					Name:    Data.Name,
					IconURL: Data.DisplayIcon,
				}

			}

			// Items

			items := playerLoadout["Items"].(map[string]interface{})
			itemMap := map[string]LoadoutItem{}

			T := time.Now()

			type ChanLItem struct {
				Index string
				Value LoadoutItem
			}

			itemOutput := make(chan ChanLItem, len(items))

			var wg2 sync.WaitGroup

			for Type, itemD := range items {

				wg2.Add(1)

				go func(Type string, itemD interface{}) {

					defer wg2.Done()

					lItem := GetItem(itemD)

					itemOutput <- ChanLItem{
						Index: Type,
						Value: lItem,
					}

				}(Type, itemD)

			}

			wg2.Wait()
			close(itemOutput)

			for Info := range itemOutput {
				itemMap[Info.Index] = Info.Value
			}

			NewLog("Loadout took:", time.Since(T).Seconds())

			loadoutOutput <- ChanItem{
				Index: playerLoadout["Subject"].(string),
				Value: Loadout{
					Expressions: expressionsMap,
					Items:       itemMap,
				},
			}

		}(playerLoadout)

	}

	wg.Wait()
	close(loadoutOutput)

	for Info := range loadoutOutput {

		Loadouts[Info.Index] = Info.Value

	}

	return Loadouts

}

func GetMatchLoudout(matchUUID string, PUUID string, player PlayerInfo, regions Regional) map[string]Loadout {

	//"https://pd." + regions.shard + ".a.pvp.net/mmr/v1/players/" + PlayerUUID

	entitlement := GetEntitlementsToken(GetLockfile(true))

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

func GetAgentSelectLoudout(matchUUID string, PUUID string, player PlayerInfo, regions Regional) map[string]Loadout {

	//"https://pd." + regions.shard + ".a.pvp.net/mmr/v1/players/" + PlayerUUID

	entitlement := GetEntitlementsToken(GetLockfile(true))

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
