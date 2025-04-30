package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lynn9388/supsub"
)

var (
	SeperationBar = "||                                                                            ||"
)

type MatchPlayerIdentity struct {
	PlayerCardID           string
	PlayerTitleID          string
	AccountLevel           float64
	PreferredLevelBorderID string
	Incognito              bool
	HideAccountLevel       bool
}

type SeasonalBadgeInfo struct {
	SeasonID        string
	NumberOfWins    float64
	Rank            float64
	LeaderboardRank float64
}

type CurrentMatchPlayer struct {
	Subject           string
	TeamID            string
	CharacterID       string
	PlayerIdentity    MatchPlayerIdentity
	SeasonalBadgeInfo SeasonalBadgeInfo
	IsCoach           bool
	IsAssociated      bool
	PlatformType      string
	GameName          string
	TagLine           string
}

type ConnectionDetails struct {
	GameServerHosts        []string
	GameServerHost         string
	GameServerPort         float64
	GameServerObfuscatedIP float64
	GameClientHash         float64
	PlayerKey              string
}

type MatchmakingData struct {
	QueueID  string
	IsRanked bool
}

type CurrentGameMatch struct {
	MatchID           string
	Version           float64
	State             string
	MapID             string
	ModeID            string
	ProvisioningFlow  string
	GamePodID         string
	AllMUCName        string
	TeamMUCName       string
	TeamVoiceID       string
	TeamMatchToken    string
	IsReconnectable   bool
	ConnectionDetails ConnectionDetails
	Players           []CurrentMatchPlayer
	MatchmakingData   MatchmakingData
}

type CurrentAgentSelectPlayer struct {
	Subject                 string
	CharacterID             string
	CharacterSelectionState string
	CompetitiveTier         float64
	PlayerIdentity          MatchPlayerIdentity
	SeasonalBadgeInfo       SeasonalBadgeInfo
	IsCaptain               bool
	PlatformType            string
	GameName                string
	TagLine                 string
}

type CurrentAgentSelectTeam struct {
	TeamID  string
	Players []CurrentAgentSelectPlayer
}

type CurrentAgentSelect struct {
	ID                   string
	Version              float64
	AllyTeam             CurrentAgentSelectTeam
	EnemyTeam            CurrentAgentSelectTeam
	EnemyTeamSize        float64
	EnemyTeamLockCount   float64
	PregameState         string
	LastUpdated          string
	MapID                string
	MapSelectStep        float64
	GamePodID            string
	Mode                 string
	VoiceSessionID       string
	MUCName              string
	TeamMatchToken       string
	QueueID              string
	ProvisioningFlow     string
	IsRanked             bool
	PhaseTimeRemainingNS float64
	StepTimeRemainingNS  float64
}

type EmbedInput struct {
	Subject        string
	CharacterID    string
	PlayerIdentity MatchPlayerIdentity
	GameName       string
	TagLine        string
}

type ProfileEmbedInput struct {
	Subject        string
	CharacterID    string
	PlayerIdentity MatchPlayerIdentity
	GameName       string
	TagLine        string
	matchHistory   MatchHistory
	loadout        *Loadout
}

func StringLengther(main string, num int) string {

	return main + strings.Repeat(" ", num-len(main))

}

// Uses the region obtained from ShooterGame.log, along with match ID of current game

func GetCurrentAgentSelectID(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) string {

	// Get MatchID of agent select

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/pregame/v1/players/"+player.puuid, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var agent_select_player map[string]interface{}

	agent_select_player, err = GetJSON(res)
	checkError(err)

	// Use MatchID of current game to get rest of game stats

	if agent_select_player["MatchID"] == nil {
		return ""
	}

	return agent_select_player["MatchID"].(string)

}

func GetCurrentMatchID(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) string {

	// Get MatchID of current game

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/core-game/v1/players/"+player.puuid, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var match_game_player map[string]interface{}

	match_game_player, err = GetJSON(res)
	checkError(err)

	// Use MatchID of current game to get rest of game stats

	if match_game_player["MatchID"] == nil {
		return ""
	}

	return match_game_player["MatchID"].(string)

}

func CheckForMatch(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional, client http.Client, lastMatchID string, discord *discordgo.Session) string {

	matchID := GetCurrentMatchID(player, entitlement, regions)

	if matchID == lastMatchID {
		return matchID
	}

	if matchID == "" {

		// Incase we just went back to menu
		return matchID

	}

	// Should call whenever we go into a new match

	fmt.Println("New match found: '" + matchID + "'")

	Request_match(player, entitlement, regions, Settings["current_session_channel"], discord)

	return matchID

}

func CheckForAgentSelect(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional, client http.Client, lastAgentSelectID string, discord *discordgo.Session) string {

	agentSelectID := GetCurrentAgentSelectID(player, entitlement, regions)

	if agentSelectID == lastAgentSelectID {
		return agentSelectID
	}

	if agentSelectID == "" {

		// Incase we just went back to menu
		return agentSelectID

	}

	// Should call whenever we go into a new match

	fmt.Println("New agent select found: '" + agentSelectID + "'")

	Request_agentSelect(player, entitlement, regions, Settings["current_session_channel"], discord)

	return agentSelectID

}

func ListenForMatch(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional, client http.Client, checkSecondDelta time.Duration, discord *discordgo.Session) {

	go func() {

		// Create new thread for listening so we don't block the rest of the program

		lastAgentSelectID := ""
		lastMatchID := ""

		for {

			fmt.Println("Checking match status..")

			lastAgentSelectID = CheckForAgentSelect(player, entitlement, regions, client, lastAgentSelectID, discord)
			lastMatchID = CheckForMatch(player, entitlement, regions, client, lastMatchID, discord)

			time.Sleep(checkSecondDelta)

		}

	}()

}

func GetAgentSelectInfo(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) CurrentAgentSelect {

	// Get MatchID of agent select

	var MatchID string = GetCurrentAgentSelectID(player, entitlement, regions)

	if MatchID == "" {
		// Not in agent select
		fmt.Println("Not currently in agent select")
		return CurrentAgentSelect{}
	}

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/pregame/v1/matches/"+MatchID, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var agent_select_information map[string]interface{}

	agent_select_information, err = GetJSON(res)

	teams := [2]string{"AllyTeam", "EnemyTeam"}
	team_data := map[string]CurrentAgentSelectTeam{}

	for _, team_name := range teams {

		if agent_select_information[team_name] == nil {
			// 'X' team does not exist (Likely custom game with only 1 team)
			// Create a team of empty players for data structure
			fmt.Println("No team '" + team_name + "'")

			team_data[team_name] = CurrentAgentSelectTeam{
				TeamID:  "",
				Players: []CurrentAgentSelectPlayer{},
			}
			continue
		}

		var team_details = agent_select_information[team_name].(map[string]interface{})

		var Players_raw = team_details["Players"].([]interface{})
		var Players = make([]CurrentAgentSelectPlayer, len(Players_raw))
		checkError(err)

		for Index := range Players_raw {

			ply := Players_raw[Index].(map[string]interface{})
			ply_id := ply["PlayerIdentity"].(map[string]interface{})
			badge_info := ply["SeasonalBadgeInfo"].(map[string]interface{})

			var Played_Identity = MatchPlayerIdentity{
				PlayerCardID:           ply_id["PlayerCardID"].(string),
				PlayerTitleID:          ply_id["PlayerTitleID"].(string),
				AccountLevel:           ply_id["AccountLevel"].(float64),
				PreferredLevelBorderID: ply_id["PreferredLevelBorderID"].(string),
				Incognito:              ply_id["Incognito"].(bool),
				HideAccountLevel:       ply_id["HideAccountLevel"].(bool),
			}

			var SeasonalBadgeInfo = SeasonalBadgeInfo{
				SeasonID:        badge_info["SeasonID"].(string),
				NumberOfWins:    badge_info["NumberOfWins"].(float64),
				Rank:            badge_info["Rank"].(float64),
				LeaderboardRank: badge_info["LeaderboardRank"].(float64),
			}

			var Player_Info = CurrentAgentSelectPlayer{
				Subject:                 ply["Subject"].(string),
				CharacterID:             ply["CharacterID"].(string),
				CharacterSelectionState: ply["CharacterSelectionState"].(string),
				CompetitiveTier:         ply["CompetitiveTier"].(float64),
				PlayerIdentity:          Played_Identity,
				SeasonalBadgeInfo:       SeasonalBadgeInfo,
				IsCaptain:               ply["IsCaptain"].(bool),
				PlatformType:            ply["PlatformType"].(string),
			}

			Players[Index] = Player_Info
		}

		team_data[team_name] = CurrentAgentSelectTeam{
			TeamID:  team_details["TeamID"].(string),
			Players: Players,
		}

	}

	var ProvisioningFlow interface{}
	if agent_select_information["ProvisioningFlow"] == nil {
		ProvisioningFlow = ""
	} else {
		ProvisioningFlow = agent_select_information["ProvisioningFlow"].(string)
	}
	var agent_select_struct = CurrentAgentSelect{
		ID:                   agent_select_information["ID"].(string),
		Version:              agent_select_information["Version"].(float64),
		AllyTeam:             team_data["AllyTeam"],
		EnemyTeam:            team_data["EnemyTeam"],
		EnemyTeamSize:        agent_select_information["EnemyTeamSize"].(float64),
		EnemyTeamLockCount:   agent_select_information["EnemyTeamLockCount"].(float64),
		PregameState:         agent_select_information["PregameState"].(string),
		LastUpdated:          agent_select_information["LastUpdated"].(string),
		MapID:                agent_select_information["MapID"].(string),
		MapSelectStep:        agent_select_information["MapSelectStep"].(float64),
		GamePodID:            agent_select_information["GamePodID"].(string),
		Mode:                 agent_select_information["Mode"].(string),
		VoiceSessionID:       agent_select_information["VoiceSessionID"].(string),
		ProvisioningFlow:     ProvisioningFlow.(string),
		MUCName:              agent_select_information["MUCName"].(string),
		TeamMatchToken:       agent_select_information["TeamMatchToken"].(string),
		QueueID:              agent_select_information["QueueID"].(string),
		IsRanked:             agent_select_information["IsRanked"].(bool),
		PhaseTimeRemainingNS: agent_select_information["PhaseTimeRemainingNS"].(float64),
		StepTimeRemainingNS:  agent_select_information["StepTimeRemainingNS"].(float64),
	}

	return agent_select_struct

}

func GetCurrentMatchInfo(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) CurrentGameMatch {

	// Get MatchID of current game

	var MatchID string = GetCurrentMatchID(player, entitlement, regions)

	if MatchID == "" {
		fmt.Println("Not currently in match")
		return CurrentGameMatch{}
	}

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/core-game/v1/matches/"+MatchID, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var match_information map[string]interface{}

	match_information, err = GetJSON(res)
	checkError(err)

	var connection_details = match_information["ConnectionDetails"].(map[string]interface{})

	var GameServerHosts_raw = connection_details["GameServerHosts"].([]interface{})
	var GameServerHosts = make([]string, len(GameServerHosts_raw))

	for Index, HostName := range GameServerHosts_raw {
		GameServerHosts[Index] = HostName.(string)
	}

	var Players_raw = match_information["Players"].([]interface{})
	var Players = make([]CurrentMatchPlayer, len(Players_raw))

	for Index := range Players_raw {

		ply := Players_raw[Index].(map[string]interface{})
		ply_id := ply["PlayerIdentity"].(map[string]interface{})
		badge_info := ply["SeasonalBadgeInfo"].(map[string]interface{})

		var Played_Identity = MatchPlayerIdentity{
			PlayerCardID:           ply_id["PlayerCardID"].(string),
			PlayerTitleID:          ply_id["PlayerTitleID"].(string),
			AccountLevel:           ply_id["AccountLevel"].(float64),
			PreferredLevelBorderID: ply_id["PreferredLevelBorderID"].(string),
			Incognito:              ply_id["Incognito"].(bool),
			HideAccountLevel:       ply_id["HideAccountLevel"].(bool),
		}

		var SeasonalBadgeInfo = SeasonalBadgeInfo{
			SeasonID:        badge_info["SeasonID"].(string),
			NumberOfWins:    badge_info["NumberOfWins"].(float64),
			Rank:            badge_info["Rank"].(float64),
			LeaderboardRank: badge_info["LeaderboardRank"].(float64),
		}

		var Player_Info = CurrentMatchPlayer{
			Subject:           ply["Subject"].(string),
			TeamID:            ply["TeamID"].(string),
			CharacterID:       ply["CharacterID"].(string),
			PlayerIdentity:    Played_Identity,
			SeasonalBadgeInfo: SeasonalBadgeInfo,
			IsCoach:           ply["IsCoach"].(bool),
			IsAssociated:      ply["IsAssociated"].(bool),
			PlatformType:      ply["PlatformType"].(string),
		}

		Players[Index] = Player_Info
	}

	var matchmakingStruct MatchmakingData = MatchmakingData{}

	if match_information["MatchmakingData"] != nil {

		var Matchmaking_raw = match_information["MatchmakingData"].(map[string]interface{})

		matchmakingStruct = MatchmakingData{
			QueueID:  Matchmaking_raw["QueueID"].(string),
			IsRanked: Matchmaking_raw["IsRanked"].(bool),
		}

	}

	fmt.Println(matchmakingStruct.QueueID)

	connection_details_struct := ConnectionDetails{
		GameServerHosts:        GameServerHosts,
		GameServerHost:         connection_details["GameServerHost"].(string),
		GameServerPort:         connection_details["GameServerPort"].(float64),
		GameServerObfuscatedIP: connection_details["GameServerObfuscatedIP"].(float64),
		GameClientHash:         connection_details["GameClientHash"].(float64),
		PlayerKey:              connection_details["PlayerKey"].(string),
	}

	var match_struct = CurrentGameMatch{
		MatchID:           match_information["MatchID"].(string),
		Version:           match_information["Version"].(float64),
		State:             match_information["State"].(string),
		MapID:             match_information["MapID"].(string),
		ModeID:            match_information["ModeID"].(string),
		ProvisioningFlow:  match_information["ProvisioningFlow"].(string),
		GamePodID:         match_information["GamePodID"].(string),
		AllMUCName:        match_information["AllMUCName"].(string),
		TeamMUCName:       match_information["TeamMUCName"].(string),
		TeamVoiceID:       match_information["TeamVoiceID"].(string),
		TeamMatchToken:    match_information["TeamMatchToken"].(string),
		IsReconnectable:   match_information["IsReconnectable"].(bool),
		ConnectionDetails: connection_details_struct,
		Players:           Players,
		MatchmakingData:   matchmakingStruct,
	}

	return match_struct

}

func NewAgentSelectEmbed(agentSelect CurrentAgentSelect, player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) []*discordgo.MessageSend {

	PlayerIDs := make([]string, len(agentSelect.AllyTeam.Players)+len(agentSelect.EnemyTeam.Players))

	for I, Player := range agentSelect.AllyTeam.Players {
		PlayerIDs[I] = Player.Subject
	}

	for I, Player := range agentSelect.EnemyTeam.Players {
		PlayerIDs[I+len(agentSelect.AllyTeam.Players)] = Player.Subject
	}

	json_data, err := json.MarshalIndent(PlayerIDs, "", "	")
	checkError(err)

	body := bytes.NewBuffer(json_data)
	checkError(err)

	req, err := http.NewRequest("PUT", "https://pd."+regions.shard+".a.pvp.net/name-service/v2/players", body)
	checkError(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	checkError(err)

	names := []any{}

	json.Unmarshal(data, &names)
	checkError(err)

	PlayerNameLookup := make(map[string]map[string]string)

	for _, NameData := range names {

		NameData := NameData.(map[string]interface{})

		PlayerNameLookup[NameData["Subject"].(string)] = map[string]string{
			"name":    NameData["GameName"].(string),
			"tagLine": NameData["TagLine"].(string),
		}

	}

	var Title string = "Valorant Match"

	if agentSelect.ID == "" {

		return []*discordgo.MessageSend{
			{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       Title,
						Color:       10628401, // Red Error Color
						Description: "Not currently in Agent Select",
					},
				},
			},
		}
	}

	message_size := 0

	if len(agentSelect.AllyTeam.Players) > 10 {
		// For deathmatches

		return []*discordgo.MessageSend{}
	}

	if len(agentSelect.EnemyTeam.Players) > 10 {
		// For deathmatches

		return []*discordgo.MessageSend{}

	}

	if len(agentSelect.AllyTeam.Players) > 0 {
		message_size = message_size + 1
	}

	if len(agentSelect.EnemyTeam.Players) > 0 {
		message_size = message_size + 1
	}

	final_list := make([]*discordgo.MessageSend, message_size)

	embeds0 := make([]*discordgo.MessageEmbed, len(agentSelect.AllyTeam.Players))
	embeds1 := make([]*discordgo.MessageEmbed, len(agentSelect.EnemyTeam.Players))

	for I, P := range agentSelect.AllyTeam.Players {

		input := EmbedInput{
			Subject:        P.Subject,
			CharacterID:    P.CharacterID,
			PlayerIdentity: P.PlayerIdentity,
			GameName:       PlayerNameLookup[P.Subject]["name"],
			TagLine:        PlayerNameLookup[P.Subject]["tagLine"],
		}

		embeds0[I] = matchEmbedFromPlayer(&input, 3124052, &regions, &entitlement, &player, nil)

	}

	if len(agentSelect.AllyTeam.Players) > 0 {
		final_list[0] = &discordgo.MessageSend{
			Embeds: embeds0,
		}
	}

	for I, P := range agentSelect.EnemyTeam.Players {

		input := EmbedInput{
			Subject:        P.Subject,
			CharacterID:    P.CharacterID,
			PlayerIdentity: P.PlayerIdentity,
			GameName:       PlayerNameLookup[P.Subject]["name"],
			TagLine:        PlayerNameLookup[P.Subject]["tagLine"],
		}

		embeds1[I] = matchEmbedFromPlayer(&input, 11348780, &regions, &entitlement, &player, nil)

	}

	if len(agentSelect.EnemyTeam.Players) > 0 {
		final_list[1] = &discordgo.MessageSend{
			Embeds: embeds1,
		}
	}

	OptionList := make([]discordgo.SelectMenuOption, len(agentSelect.AllyTeam.Players)+len(agentSelect.EnemyTeam.Players))

	for Index, Player := range agentSelect.AllyTeam.Players {

		OptionList[Index] = discordgo.SelectMenuOption{
			Label:       PlayerNameLookup[Player.Subject]["name"] + ":" + PlayerNameLookup[Player.Subject]["tagLine"],
			Value:       Player.Subject,
			Description: "Selects this player",
		}

	}

	for Index, Player := range agentSelect.EnemyTeam.Players {

		OptionList[Index+len(agentSelect.AllyTeam.Players)] = discordgo.SelectMenuOption{
			Label:       PlayerNameLookup[Player.Subject]["name"] + " " + supsub.ToSup(PlayerNameLookup[Player.Subject]["tagLine"]),
			Value:       Player.Subject,
			Description: "Selects this player",
		}

	}

	final_list[len(final_list)-1].Components = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select_player_agent",
					Placeholder: "Select player",
					Options:     OptionList,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{

				discordgo.Button{
					CustomID: "exit_agent_select",
					Label:    "Exit",
					Style:    discordgo.DangerButton,
				},
			},
		},
	}

	final_list[0].Content = SeperationBar

	return final_list
}

func getPlayerNames(PlayerIDS []string, player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) map[string]map[string]string {

	json_data, err := json.MarshalIndent(PlayerIDS, "", "	")
	checkError(err)

	body := bytes.NewBuffer(json_data)
	checkError(err)

	req, err := http.NewRequest("PUT", "https://pd."+regions.shard+".a.pvp.net/name-service/v2/players", body)
	checkError(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := Client.Do(req)
	checkError(err)

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	checkError(err)

	names := []any{}

	json.Unmarshal(data, &names)
	checkError(err)

	returnedNames := make(map[string]map[string]string, len(names))

	for _, name := range names {

		nameData := name.(map[string]interface{})

		returnedNames[nameData["Subject"].(string)] = map[string]string{
			"name":    nameData["GameName"].(string),
			"tagLine": nameData["TagLine"].(string),
		}

	}

	return returnedNames

}

func newMatchEmbed(match CurrentGameMatch, player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) []*discordgo.MessageSend {

	PlayerIDs := make([]string, len(match.Players))

	var MainPlayer CurrentMatchPlayer

	for I, Player := range match.Players {
		if Player.Subject == player.sub {
			MainPlayer = Player
		}
		PlayerIDs[I] = Player.Subject
	}

	PlayerNameLookup := getPlayerNames(PlayerIDs, player, entitlement, regions)

	var Title string = "Valorant Match"

	if match.MatchID == "" {

		return []*discordgo.MessageSend{
			{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       Title,
						Color:       10628401, // Red Error Color
						Description: "Not currently in Match",
					},
				},
			},
		}
	}

	AllyTeam := map[int]CurrentMatchPlayer{}
	EnemyTeam := map[int]CurrentMatchPlayer{}

	AllyTeamID := MainPlayer.TeamID
	for Index, P := range match.Players {

		var isEnemy bool = false

		if len(match.Players) > 10 {

			if Index >= len(match.Players)/2 {
				isEnemy = true
			}

		}

		if P.TeamID == AllyTeamID && !isEnemy {

			// Friendly Player
			P.GameName = PlayerNameLookup[P.Subject]["name"]
			P.TagLine = PlayerNameLookup[P.Subject]["tagLine"]

			AllyTeam[len(AllyTeam)] = P

			//fmt.Println(len(AllyTeamNames))
		} else {
			isEnemy = true
		}

		if isEnemy {
			// Enemy Player
			P.GameName = PlayerNameLookup[P.Subject]["name"]
			P.TagLine = PlayerNameLookup[P.Subject]["tagLine"]

			EnemyTeam[len(EnemyTeam)] = P

			//fmt.Println(len(EnemyTeamNames))
		}
	}

	message_size := 0

	if len(AllyTeam) > 0 {
		message_size = message_size + 1
	}

	if len(EnemyTeam) > 0 {
		message_size = message_size + 1
	}

	final_list := make([]*discordgo.MessageSend, message_size)

	embeds0 := make([]*discordgo.MessageEmbed, len(AllyTeam))
	embeds1 := make([]*discordgo.MessageEmbed, len(EnemyTeam))

	for I, P := range AllyTeam {

		input := EmbedInput{
			Subject:        P.Subject,
			CharacterID:    P.CharacterID,
			PlayerIdentity: P.PlayerIdentity,
			GameName:       P.GameName,
			TagLine:        P.TagLine,
		}

		embeds0[I] = matchEmbedFromPlayer(&input, 3124052, &regions, &entitlement, &player, &match.MatchmakingData)

	}

	if len(AllyTeam) > 0 {
		final_list[0] = &discordgo.MessageSend{
			Embeds: embeds0,
		}
	}

	if len(EnemyTeam) > 0 {

		final_list[1] = &discordgo.MessageSend{
			Embeds: embeds1,
		}
	}

	for I, P := range EnemyTeam {

		input := EmbedInput{
			Subject:        P.Subject,
			CharacterID:    P.CharacterID,
			PlayerIdentity: P.PlayerIdentity,
			GameName:       P.GameName,
			TagLine:        P.TagLine,
		}

		embeds1[I] = matchEmbedFromPlayer(&input, 11348780, &regions, &entitlement, &player, &match.MatchmakingData)

	}

	OptionList := make([]discordgo.SelectMenuOption, len(AllyTeam)+len(EnemyTeam))

	for Index, Player := range match.Players {

		OptionList[Index] = discordgo.SelectMenuOption{
			Label:       PlayerNameLookup[Player.Subject]["name"] + " " + supsub.ToSup(PlayerNameLookup[Player.Subject]["tagLine"]),
			Value:       Player.Subject,
			Description: "Selects this player",
		}

	}

	final_list[len(final_list)-1].Components = []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select_player",
					Placeholder: "Select player",
					Options:     OptionList,
				},
			},
		},
	}

	final_list[0].Content = SeperationBar

	return final_list
}

func matchEmbedFromPlayer(P *EmbedInput, color int, regions *Regional, entitlement *EntitlementsTokenResponse, playerInfo *PlayerInfo, matchmakingData *MatchmakingData) *discordgo.MessageEmbed {

	fmt.Println("Loading: " + P.GameName)

	IncludedGamemodes := map[string]bool{
		"competitive": true,
	}

	if matchmakingData != nil {

		IncludedGamemodes[matchmakingData.QueueID] = true

	}

	mmr := GetPlayerMMR(regions, entitlement, playerInfo, P.Subject, IncludedGamemodes)

	agent := AgentDetails[strings.ToLower(P.CharacterID)]

	fmt.Println(mmr.PeakRank)

	CurrentGameMMR := mmr.Competitive

	if matchmakingData != nil {

		switch matchmakingData.QueueID {
		case "competitive":
			CurrentGameMMR = mmr.Competitive
		case "swiftplay":
			CurrentGameMMR = mmr.Swiftplay
		case "deathmatch":
			CurrentGameMMR = mmr.Deathmatch
		case "unrated":
			CurrentGameMMR = mmr.Unrated
		}
	}

	if CurrentGameMMR == nil {
		CurrentGameMMR = mmr.Competitive
	}

	WinPercentage := strconv.FormatFloat(100/CurrentGameMMR.TotalGames*CurrentGameMMR.TotalWins, 'f', 2, 64)

	var WinPercText string

	if matchmakingData != nil {

		if matchmakingData.QueueID != "competitive" && CurrentGameMMR == mmr.Competitive {

			WinPercText = "`Comp Win%: " + WinPercentage + "%`"

		} else {

			WinPercText = "`Win%: " + WinPercentage + "%`"

		}
	} else {

		WinPercText = "`Comp Win%: " + WinPercentage + "%`"

	}

	levelString := strconv.Itoa(int(P.PlayerIdentity.AccountLevel))

	if P.PlayerIdentity.HideAccountLevel {
		levelString = "(Hidden)"
	}

	info := "`Level: " + levelString + "` "
	info = info + "`Peak: " + RankDetails[mmr.PeakRank].TierName + "` "

	info = info + WinPercText + " \n"

	if P.GameName == "" {
		P.GameName = "undefined"
	}

	fmt.Println("Rank icon: " + RankDetails[mmr.CurrentRank].RankIcon)

	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    P.GameName + " " + supsub.ToSup(P.TagLine),
			IconURL: agent.displayIcon,
		},
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Value:  info,
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: RankDetails[mmr.CurrentRank].RankIcon,
		},
	}
}

func CreatePlayerProfile(P *ProfileEmbedInput, player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional, discord *discordgo.Session) *discordgo.WebhookParams {

	PlayerID := P.Subject

	if PlayerID == "" {
		fmt.Println("No player ID given for profile")
		return &discordgo.WebhookParams{}
	}

	mmr := GetPlayerMMR(&regions, &entitlement, &player, P.Subject, map[string]bool{})
	ornament := GetOrnamentsFromPlayer(P.PlayerIdentity)

	fmt.Println("Selceted: " + P.GameName)

	embedCount := 3

	Embeds := make([]*discordgo.MessageEmbed, embedCount)

	RankHex := "0x" + RankDetails[mmr.CurrentRank].RankColor[:len(RankDetails[mmr.CurrentRank].RankColor)-2]

	Color, err := strconv.ParseInt(RankHex, 0, 0)
	checkError(err)

	fmt.Println(int(Color))

	// Title Embed

	TitleDescription := ""

	RR := 1.0 / 100.0 * float64(mmr.RankedRating)
	RRProgressCharacters := float64(25)

	RRS := int(math.Floor(RRProgressCharacters * RR))

	RRStart := strings.Repeat("▓", RRS)
	RREnd := strings.Repeat("░", int(RRProgressCharacters)-RRS)

	RRProgressText := RRStart + RREnd + " `- MMR: ( " + strconv.Itoa(mmr.RankedRating) + "/100 )`"

	TitleDescription = TitleDescription + RRProgressText

	Embeds[0] = &discordgo.MessageEmbed{
		Color: int(Color),
		Author: &discordgo.MessageEmbedAuthor{
			Name:    P.GameName + " " + supsub.ToSup(P.TagLine),
			IconURL: RankDetails[mmr.CurrentRank].RankIcon,
		},
		Description: TitleDescription,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: ornament.PlayerCard.displayIcon,
		},
	}

	completedEmbeds := 1

	// Match History Embed

	if len(P.matchHistory.History) > 0 {

		Description := ""

		for _, MatchInfo := range P.matchHistory.History {

			Description = Description + "`Match Type: " + strings.ToUpper(MatchInfo.QueueID) + "` "
			Description = Description + "`Score: " + strconv.Itoa(MatchInfo.previousMatchPlayerDetails.Stats.score) + "` "
			Description = Description + "`K: " + strconv.Itoa(MatchInfo.previousMatchPlayerDetails.Stats.kills) + "` "
			Description = Description + "`D: " + strconv.Itoa(MatchInfo.previousMatchPlayerDetails.Stats.deaths) + "` "
			Description = Description + "`A: " + strconv.Itoa(MatchInfo.previousMatchPlayerDetails.Stats.assists) + "` "
			Description = Description + "`Rounds: " + strconv.Itoa(MatchInfo.previousMatchPlayerDetails.Stats.roundsPlayed) + "` \n"

		}

		Embeds[completedEmbeds] = &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name: "Match History " + supsub.ToSup("Entries ( "+strconv.Itoa(P.matchHistory.BeginIndex)+" to "+strconv.Itoa(P.matchHistory.EndIndex)+" ) (Total "+strconv.Itoa(P.matchHistory.TotalEntries)+")"),
			},
			Description: Description,
			Color:       int(Color),
		}

		completedEmbeds = completedEmbeds + 1

	} else {

		Embeds[completedEmbeds] = &discordgo.MessageEmbed{
			Color:       int(Color),
			Title:       "Match History",
			Description: "No match history found",
		}

		completedEmbeds = completedEmbeds + 1

	}

	LoadoutDescription := ""
	LoadoutIndex := 0

	LongestName := 0

	for _, WeaponName := range P.loadout.Items {

		name := WeaponName.weaponInfo.displayName

		SkinNameSplit := strings.Split(name, " ")

		WeaponType := WeaponIDToName[WeaponName.TypeID]

		if SkinNameSplit[len(SkinNameSplit)-1] == WeaponType {

			SkinNameSplit[len(SkinNameSplit)-1] = ""

		}

		name = strings.Join(SkinNameSplit, " ")
		name = strings.TrimRight(name, " ")

		if len(name) > LongestName {
			LongestName = len(name)
		}

	}

	for ID, WeaponType := range WeaponIDToName {

		LoadoutIndex = LoadoutIndex + 1

		SkinName := P.loadout.Items[ID].weaponInfo.displayName

		SkinNameSplit := strings.Split(SkinName, " ")

		if SkinNameSplit[len(SkinNameSplit)-1] == WeaponType {

			SkinNameSplit[len(SkinNameSplit)-1] = ""

		}

		SkinName = strings.Join(SkinNameSplit, " ")
		SkinName = strings.TrimRight(SkinName, " ")

		//BuddyName := strings.ReplaceAll(P.loadout.Items[ID].Buddy.displayName, "_", " ")

		LoadoutDescription = LoadoutDescription + "`" + StringLengther(WeaponType+":", 10) + StringLengther(SkinName, LongestName+1) + "`"

		if P.loadout.Items[ID].Buddy.displayName != "" {

			//LoadoutDescription = LoadoutDescription + " `Buddy: " + BuddyName + "`"

		}

		LoadoutDescription = LoadoutDescription + "\n"

	}

	Embeds[completedEmbeds] = &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Loadout",
		},
		Description: LoadoutDescription,
		Color:       int(Color),
	}

	fmt.Println("Got profile!")

	return &discordgo.WebhookParams{
		Embeds: Embeds,
	}

}

func Request_match(player_info PlayerInfo, entitlements EntitlementsTokenResponse, regional Regional, ChannelID string, discord *discordgo.Session) {

	fmt.Println("Requested Match")

	MatchInfo := GetCurrentMatchInfo(player_info, entitlements, regional)

	messages := newMatchEmbed(MatchInfo, player_info, entitlements, regional)

	for _, Message := range messages {

		if Message == nil {
			continue
		}

		_, err := discord.ChannelMessageSendComplex(ChannelID, Message)
		checkError(err)

	}

	CommandHandlers["select_player"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		Response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		}

		Response.Data.Flags = discordgo.MessageFlagsEphemeral

		s.InteractionRespond(i.Interaction, Response)

		PlayerID := i.MessageComponentData().Values[0]

		fmt.Println("ID: " + PlayerID)

		var P CurrentMatchPlayer

		for _, Ply := range MatchInfo.Players {

			if Ply.Subject != PlayerID {
				continue
			}

			P = Ply

		}

		loadout := GetMatchLoudout(MatchInfo.MatchID, P.Subject, player_info, entitlements, regional)[P.Subject]

		PlayerNames := getPlayerNames([]string{P.Subject}, player_info, entitlements, regional)[PlayerID]
		matchHistory, err := GetMatchHistoryOfUUID(P.Subject, 0, 10, &regional, &entitlements, player_info)
		checkError(err)

		input := &ProfileEmbedInput{
			Subject:        P.Subject,
			CharacterID:    P.CharacterID,
			PlayerIdentity: P.PlayerIdentity,
			GameName:       PlayerNames["name"],
			TagLine:        PlayerNames["tagLine"],
			matchHistory:   matchHistory,
			loadout:        &loadout,
		}

		FinalResponse := CreatePlayerProfile(input, player_info, entitlements, regional, discord)

		FinalResponse.Flags = discordgo.MessageFlagsEphemeral

		s.FollowupMessageCreate(i.Interaction, true, FinalResponse)

	}
}

func Request_agentSelect(player_info PlayerInfo, entitlements EntitlementsTokenResponse, regional Regional, ChannelID string, discord *discordgo.Session) {

	fmt.Println("Requested Match")

	AgentSelect := GetAgentSelectInfo(player_info, entitlements, regional)

	messages := NewAgentSelectEmbed(AgentSelect, player_info, entitlements, regional)

	for I, Message := range messages {

		if Message == nil {
			continue
		}

		fmt.Println("T : " + strconv.Itoa(I))

		_, err := discord.ChannelMessageSendComplex(ChannelID, Message)
		checkError(err)

	}

	CommandHandlers["select_player_agent"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)

		fmt.Println(i.ChannelID)

	}

	CommandHandlers["exit_agent_select"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)

		fmt.Println(i.ChannelID)

	}

}
