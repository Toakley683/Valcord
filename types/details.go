package types

import (
	"net/http"
	"strings"
)

// Initializing Data
type CurrencyImagesStruct struct {
	ValorantPoints string
	KingdomCredits string
	FreeAgents     string
	Radianite      string
}

type ValorantRank struct {
	Tier         int
	TierName     string
	DivisionName string
	RankColor    string
	RankIcon     string
}

var (
	AgentDetails  map[string]PlayableAgent
	DefaultAgents map[string]PlayableAgent
	MapDetails    map[string]PlayableMap
	RankDetails   map[int]ValorantRank
	Regions       map[string]string

	CurrencyImages = CurrencyImagesStruct{
		ValorantPoints: "https://media.valorant-api.com/currencies/85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741/largeicon.png",
		KingdomCredits: "https://media.valorant-api.com/currencies/85ca954a-41f2-ce94-9b45-8ca3dd39a00d/largeicon.png",
		FreeAgents:     "https://media.valorant-api.com/currencies/f08d4ae3-939c-4576-ab26-09ce1f23bb37/largeicon.png",
		Radianite:      "https://media.valorant-api.com/currencies/e59aa87c-4cbf-517a-5983-6e81511be9b7/largeicon.png",
	}

	CurrencyIDToImage map[string]string = map[string]string{
		"85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741": "https://media.valorant-api.com/currencies/85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741/largeicon.png",
		"85ca954a-41f2-ce94-9b45-8ca3dd39a00d": "https://media.valorant-api.com/currencies/85ca954a-41f2-ce94-9b45-8ca3dd39a00d/largeicon.png",
		"f08d4ae3-939c-4576-ab26-09ce1f23bb37": "https://media.valorant-api.com/currencies/f08d4ae3-939c-4576-ab26-09ce1f23bb37/largeicon.png",
		"e59aa87c-4cbf-517a-5983-6e81511be9b7": "https://media.valorant-api.com/currencies/e59aa87c-4cbf-517a-5983-6e81511be9b7/largeicon.png",
	}
)

func ItemIDWTypeToStruct(Type string, ItemID string, Quantity int) Item {

	if ItemID == "" {
		return Item{}
	}

	switch Type {
	case "01bb38e1-da47-4e6a-9b3d-945fe4655707":

		// Agents

		Data := AgentDetails[ItemID]

		return Item{
			ItemTypeID:    Type,
			ItemID:        ItemID,
			Amount:        Quantity,
			Description:   Data.description,
			Name:          Data.displayName,
			DisplayIcon:   Data.displayIcon,
			StreamedVideo: Data.RoleIcon,
			Color:         Data.MainColor,
		}

	case "d5f120f8-ff8c-4aac-92ea-f2b5acbe9475":

		// Sprays

		Data := SprayData(ItemID)

		return Item{
			ItemTypeID:    Type,
			ItemID:        ItemID,
			Amount:        Quantity,
			Name:          Data.displayName,
			Description:   "",
			DisplayIcon:   Data.animationGif,
			StreamedVideo: Data.animationGif,
		}

	case "dd3bf334-87f3-40bd-b043-682a57a8dc3a":

		// Gun Buddies

		Data := BuddyData(ItemID)

		return Item{
			ItemTypeID:    Type,
			ItemID:        ItemID,
			Amount:        Quantity,
			Description:   "",
			Name:          Data.displayName,
			DisplayIcon:   Data.displayIcon,
			StreamedVideo: Data.displayIcon,
		}

	case "3f296c07-64c3-494c-923b-fe692a4fa1bd":

		// Cards

		Data := CardData(ItemID)

		return Item{
			ItemTypeID:    Type,
			ItemID:        ItemID,
			Amount:        Quantity,
			Description:   "",
			Name:          Data.displayName,
			DisplayIcon:   Data.displayIcon,
			StreamedVideo: Data.wideArt,
		}

	case "de7caa6b-adf7-4588-bbd1-143831e786c6":

		// Titles

		Data := TitleData(ItemID)

		return Item{
			ItemTypeID:    Type,
			ItemID:        ItemID,
			Amount:        Quantity,
			Description:   "",
			Name:          Data.titleText,
			DisplayIcon:   "",
			StreamedVideo: "",
		}

	case "e7c63390-eda7-46e0-bb7a-a6abdacd2433":

		// Skins

		Data := GetWeaponData(ItemID)

		Video := ""

		if Data["streamedVideo"] != nil {
			Video = Data["streamedVideo"].(string)
		}

		ItemLevel := ""

		if Data["levelItem"] != nil {
			ItemLevel = Data["levelItem"].(string)
		}

		return Item{
			ItemTypeID:    Type,
			ItemID:        ItemID,
			Amount:        Quantity,
			Description:   "",
			Name:          Data["displayName"].(string),
			DisplayIcon:   "https://media.valorant-api.com/weaponskinlevels/" + ItemID + "/displayicon.png",
			StreamedVideo: Video,
			LevelItem:     ItemLevel,
		}

	}

	return Item{}

}

func InitAgents() {

	// Initialize Agent details

	AgentDetails = make(map[string]PlayableAgent)
	DefaultAgents = make(map[string]PlayableAgent)

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/agents", nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var agents map[string]interface{}

	agents, err = GetJSON(res)
	checkError(err)

	data := agents["data"].([]interface{})

	for _, Agent := range data {

		agent_info := Agent.(map[string]interface{})

		if agent_info["role"] == nil {
			continue
		}

		roleData := agent_info["role"].(map[string]interface{})
		colorData := agent_info["backgroundGradientColors"].([]interface{})

		agent := PlayableAgent{
			UUID:             agent_info["uuid"].(string),
			displayName:      agent_info["displayName"].(string),
			description:      agent_info["description"].(string),
			developerName:    agent_info["developerName"].(string),
			backgroundURL:    "https://media.valorant-api.com/agents/" + agent_info["uuid"].(string) + "/background.png",
			fullportraitURL:  "https://media.valorant-api.com/agents/" + agent_info["uuid"].(string) + "/fullportrait.png",
			displayIcon:      "https://media.valorant-api.com/agents/" + agent_info["uuid"].(string) + "/displayicon.png",
			killfeedPortrait: "https://media.valorant-api.com/agents/" + agent_info["uuid"].(string) + "/killfeedportrait.png",
			Role:             roleData["displayName"].(string),
			RoleIcon:         roleData["displayIcon"].(string),
			MainColor:        colorData[0].(string),
		}

		AgentDetails[strings.ToLower(agent_info["uuid"].(string))] = agent

		if agent_info["isBaseContent"].(bool) {
			DefaultAgents[strings.ToLower(agent_info["uuid"].(string))] = agent
		}

	}
}

func InitMaps() {

	MapDetails = make(map[string]PlayableMap)

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/maps", nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var maps map[string]interface{}

	maps, err = GetJSON(res)
	checkError(err)

	map_data := maps["data"].([]interface{})

	for _, Map := range map_data {

		map_info := Map.(map[string]interface{})
		map_uuid := strings.ToLower(map_info["uuid"].(string))

		mapStruct := PlayableMap{
			uuid:           map_uuid,
			displayName:    map_info["displayName"].(string),
			ingameMapImage: "https://media.valorant-api.com/maps/" + map_uuid + "/displayicon.png",
			wideImage:      "https://media.valorant-api.com/maps/" + map_uuid + "/listviewicon.png",
			tallImage:      "https://media.valorant-api.com/maps/" + map_uuid + "/listviewicontall.png",
			splashImage:    "https://media.valorant-api.com/maps/" + map_uuid + "/splash.png",
			stylizedImage:  "https://media.valorant-api.com/maps/" + map_uuid + "/stylizedbackgroundimage.png",
			premierImage:   "https://media.valorant-api.com/maps/" + map_uuid + "/premierbackgroundimage.png",
			mapUrl:         map_info["mapUrl"].(string),
		}

		MapDetails[map_info["mapUrl"].(string)] = mapStruct

	}

}

func InitRanks() {

	// Initialize Rank details

	RankDetails = make(map[int]ValorantRank)

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/competitivetiers", nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var ranks map[string]interface{}

	ranks, err = GetJSON(res)
	checkError(err)

	data := ranks["data"].([]interface{})

	for _, RankData := range data {

		RankData := RankData.(map[string]interface{})

		TierData := RankData["tiers"].([]interface{})

		for _, Tier := range TierData {

			Tier := Tier.(map[string]interface{})

			var largeIcon string = ""

			if Tier["largeIcon"] != nil {
				largeIcon = Tier["largeIcon"].(string)
			}

			Rank := ValorantRank{
				Tier:         int(Tier["tier"].(float64)),
				TierName:     Tier["tierName"].(string),
				DivisionName: Tier["divisionName"].(string),
				RankColor:    Tier["color"].(string),
				RankIcon:     largeIcon,
			}

			RankDetails[int(Tier["tier"].(float64))] = Rank

		}

	}

}

func InitRegions() {

	Regions = make(map[string]string)

	Regions["NA"] = "North America"
	Regions["LATAM"] = "Latin America"
	Regions["BR"] = "Brazil"
	Regions["EU"] = "Europe"
	Regions["KR"] = "Korea"
	Regions["AP"] = "Asia Pacific"

}

func Init_val_details() {

	setup_networking()
	InitAgents()
	InitMaps()
	InitRanks()
	InitRegions()
	Init_Weapons()

	// Initialize Map details

}

// Data structures for organization

type PlayableAgent struct {
	UUID             string
	displayName      string
	description      string
	developerName    string
	backgroundURL    string
	fullportraitURL  string
	displayIcon      string
	killfeedPortrait string
	Role             string
	RoleIcon         string
	MainColor        string
}

type PlayableMap struct {
	uuid           string
	displayName    string
	ingameMapImage string
	wideImage      string
	tallImage      string
	splashImage    string
	stylizedImage  string
	premierImage   string
	mapUrl         string
}
