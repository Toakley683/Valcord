package types

import (
	"net/http"
)

type Title struct {
	displayName        string
	titleText          string
	isHiddenIfNotOwned bool
}

type Buddy struct {
	displayName        string
	displayIcon        string
	charmLevel         int
	isHiddenIfNotOwned bool
}

type PlayerCard struct {
	displayName        string
	isHiddenIfNotOwned bool
	displayIcon        string
	wideArt            string
	largeArt           string
}

type LevelBorder struct {
	displayName   string
	startingLevel float64
	iconURL       string
}

type Ornaments struct {
	Title       Title
	PlayerCard  PlayerCard
	LevelBorder LevelBorder
}

type Spray struct {
	displayName     string
	displayIcon     string
	fullIcon        string
	fullTransparent string
	animationPNG    string
	animationGif    string
}

// Get's spray data

func SprayData(sprayuuid string) Spray {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/sprays/"+sprayuuid, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var spray_data map[string]interface{}

	data, err := GetJSON(res)
	checkError(err)

	spray_data = data["data"].(map[string]interface{})

	if spray_data["animationGif"] == nil {
		spray_data["animationGif"] = "https://media.valorant-api.com/sprays/" + sprayuuid + "/fulltransparenticon.png"
	}

	return Spray{
		displayName:     spray_data["displayName"].(string),
		displayIcon:     "https://media.valorant-api.com/sprays/" + sprayuuid + "/displayIcon.png",
		fullIcon:        "https://media.valorant-api.com/sprays/" + sprayuuid + "/fullicon.png",
		fullTransparent: "https://media.valorant-api.com/sprays/" + sprayuuid + "/fulltransparenticon.png",
		animationPNG:    "https://media.valorant-api.com/sprays/" + sprayuuid + "/animation.png",
		animationGif:    spray_data["animationGif"].(string),
	}

}

// Get's Card data

func CardData(carduuid string) PlayerCard {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/playercards/"+carduuid, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var card_data map[string]interface{}

	data, err := GetJSON(res)
	checkError(err)

	card_data = data["data"].(map[string]interface{})

	return PlayerCard{
		displayName:        card_data["displayName"].(string),
		isHiddenIfNotOwned: card_data["isHiddenIfNotOwned"].(bool),
		displayIcon:        "https://media.valorant-api.com/playercards/" + carduuid + "/displayicon.png",
		wideArt:            "https://media.valorant-api.com/playercards/" + carduuid + "/wideart.png",
		largeArt:           "https://media.valorant-api.com/playercards/" + carduuid + "/largeart.png",
	}

}

// Get's Title data

func TitleData(titleuuid string) Title {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/playertitles/"+titleuuid, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var title_data map[string]interface{}

	data, err := GetJSON(res)
	checkError(err)

	title_data = data["data"].(map[string]interface{})

	return Title{
		displayName:        title_data["displayName"].(string),
		titleText:          title_data["titleText"].(string),
		isHiddenIfNotOwned: title_data["isHiddenIfNotOwned"].(bool),
	}

}

// Get's Buddy data

func BuddyData(buddyuuid string) Buddy {

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/buddies/levels/"+buddyuuid, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var buddy_data map[string]interface{}

	data, err := GetJSON(res)
	checkError(err)

	buddy_data = data["data"].(map[string]interface{})

	return Buddy{
		displayName:        buddy_data["displayName"].(string),
		displayIcon:        "https://media.valorant-api.com/buddylevels/" + buddyuuid + "/displayicon.png",
		charmLevel:         int(buddy_data["charmLevel"].(float64)),
		isHiddenIfNotOwned: buddy_data["hideIfNotOwned"].(bool),
	}

}

// Uses match data to get playercards, playertitles and level borders

func GetOrnamentsFromPlayer(plr_data MatchPlayerIdentity) Ornaments {
	//https://valorant-api.com/v1/playertitles/UUID
	//https://valorant-api.com/v1/playercards/UUID
	//https://valorant-api.com/v1/levelborders/UUID

	req, err := http.NewRequest("GET", "https://valorant-api.com/v1/playertitles/"+plr_data.PlayerTitleID, nil)
	checkError(err)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var title_info map[string]interface{}

	title_info, err = GetJSON(res)
	checkError(err)

	title_data := title_info["data"].(map[string]interface{})

	req, err = http.NewRequest("GET", "https://valorant-api.com/v1/playercards/"+plr_data.PlayerCardID, nil)
	checkError(err)

	res, err = Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var card_info map[string]interface{}

	card_info, err = GetJSON(res)
	checkError(err)

	card_data := card_info["data"].(map[string]interface{})

	req, err = http.NewRequest("GET", "https://valorant-api.com/v1/levelborders/"+plr_data.PreferredLevelBorderID, nil)
	checkError(err)

	res, err = Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var border_info map[string]interface{}

	border_info, err = GetJSON(res)
	checkError(err)

	border_data := border_info["data"].(map[string]interface{})

	ornament := Ornaments{
		Title: Title{
			displayName:        title_data["displayName"].(string),
			titleText:          title_data["titleText"].(string),
			isHiddenIfNotOwned: title_data["isHiddenIfNotOwned"].(bool),
		},
		PlayerCard: PlayerCard{
			displayName:        card_data["displayName"].(string),
			isHiddenIfNotOwned: card_data["isHiddenIfNotOwned"].(bool),
			displayIcon:        "https://media.valorant-api.com/playercards/" + plr_data.PlayerCardID + "/displayicon.png",
			wideArt:            "https://media.valorant-api.com/playercards/" + plr_data.PlayerCardID + "/wideart.png",
			largeArt:           "https://media.valorant-api.com/playercards/" + plr_data.PlayerCardID + "/largeart.png",
		},
		LevelBorder: LevelBorder{
			displayName:   border_data["displayName"].(string),
			startingLevel: border_data["startingLevel"].(float64),
			iconURL:       "https://media.valorant-api.com/levelborders/" + plr_data.PreferredLevelBorderID + "/levelnumberappearance.png",
		},
	}

	return ornament

}
