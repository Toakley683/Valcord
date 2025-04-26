package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
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

// Uses the region obtained from ShooterGame.log, along with match ID of current game

func GetCurrentAgentSelectID(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional) string {

	// Get MatchID of agent select

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/pregame/v1/players/"+player.puuid, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := client.Do(req)
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

func GetCurrentMatchID(player PlayerInfo, entitlement EntitlementsTokenResponse, regions Regional, client http.Client) string {

	// Get MatchID of current game

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/core-game/v1/players/"+player.puuid, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := client.Do(req)
	checkError(err)

	defer res.Body.Close()

	var match_game_player map[string]interface{}

	match_game_player, err = GetJSON(res)
	checkError(err)

	// Use MatchID of current game to get rest of game stats

	if match_game_player["MatchID"] == nil {
		fmt.Println("Not currently in match")
		return ""
	}

	return match_game_player["MatchID"].(string)

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

	res, err := client.Do(req)
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

	var MatchID string = GetCurrentMatchID(player, entitlement, regions, client)

	if MatchID == "" {
		return CurrentGameMatch{}
	}

	req, err := http.NewRequest("GET", "https://glz-"+regions.region+"-1."+regions.shard+".a.pvp.net/core-game/v1/matches/"+MatchID, nil)
	checkError(err)

	req.Header.Add("Authorization", "Bearer "+entitlement.accessToken)
	req.Header.Add("X-Riot-Entitlements-JWT", entitlement.token)
	req.Header.Add("X-Riot-ClientPlatform", player.client_platform)
	req.Header.Add("X-Riot-ClientVersion", player.version.version)

	res, err := client.Do(req)
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
	}

	return match_struct

}

func NewAgentSelectEmbed(agentSelect CurrentAgentSelect) *discordgo.MessageEmbed {

	/*var players map[int]CurrentAgentSelectPlayer

	if agentSelect.AllyTeam.Players != nil {

		for _, player := range agentSelect.AllyTeam.Players {
			players[len(players)] = player
		}

	}

	if agentSelect.EnemyTeam.Players != nil {

		for _, player := range agentSelect.AllyTeam.Players {
			players[len(players)] = player
		}

	}

	fmt.Println(players)

	var Title string = "Valorant Match"*/

	return &discordgo.MessageEmbed{
		Title: "Test",
		Color: 3124052, // Green Success Color
	}
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

	res, err := client.Do(req)
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
	for _, P := range match.Players {

		if P.TeamID == AllyTeamID {

			// Friendly Player
			P.GameName = PlayerNameLookup[P.Subject]["name"]
			P.TagLine = PlayerNameLookup[P.Subject]["tagLine"]

			AllyTeam[len(AllyTeam)] = P

			//fmt.Println(len(AllyTeamNames))
		} else {
			// Enemy Player
			P.GameName = PlayerNameLookup[P.Subject]["name"]
			P.TagLine = PlayerNameLookup[P.Subject]["tagLine"]

			EnemyTeam[len(EnemyTeam)] = P

			//fmt.Println(len(EnemyTeamNames))
		}
	}

	//AllyTeamNames

	/*AllyTeam = map[int]CurrentMatchPlayer{
		0: AllyTeam[0],
		1: AllyTeam[0],
		2: AllyTeam[0],
		3: AllyTeam[0],
		4: AllyTeam[0],
		5: AllyTeam[0],
	}

	EnemyTeam = map[int]CurrentMatchPlayer{
		0: AllyTeam[0],
		1: AllyTeam[0],
		2: AllyTeam[0],
		3: AllyTeam[0],
		4: AllyTeam[0],
		5: AllyTeam[0],
	}*/

	message_size := 0

	if len(AllyTeam) > 0 {
		message_size = message_size + 1
	}

	if len(EnemyTeam) > 0 {
		message_size = message_size + 1
	}

	final_list := make([]*discordgo.MessageSend, 2)

	embeds0 := make([]*discordgo.MessageEmbed, len(AllyTeam))
	embeds1 := make([]*discordgo.MessageEmbed, len(EnemyTeam))

	for I, P := range AllyTeam {

		embeds0[I] = matchEmbedFromPlayer(P, 3124052, regions, entitlement, player)

	}

	if len(AllyTeam) > 0 {
		final_list[0] = &discordgo.MessageSend{
			Embeds: embeds0,
		}
	}

	for I, P := range EnemyTeam {

		fmt.Println(P.GameName)

		embeds1[I] = matchEmbedFromPlayer(P, 11348780, regions, entitlement, player)

	}

	if len(EnemyTeam) > 0 {
		final_list[1] = &discordgo.MessageSend{
			Embeds: embeds1,
		}
	}

	return final_list
}

func matchEmbedFromPlayer(P CurrentMatchPlayer, color int, regions Regional, entitlement EntitlementsTokenResponse, playerInfo PlayerInfo) *discordgo.MessageEmbed {

	mmr := GetPlayerMMR(regions, entitlement, playerInfo, P.Subject, map[string]bool{
		"competitive": true,
	})

	agent := AgentDetails[strings.ToLower(P.CharacterID)]
	//ornament := GetOrnamentsFromPlayer(P.PlayerIdentity)

	fmt.Println(mmr.PeakRank)

	info := "`Level: " + strconv.Itoa(int(P.PlayerIdentity.AccountLevel)) + "` "
	info = info + "`Kills: " + strconv.Itoa(0) + "` "
	info = info + "`Deaths: " + strconv.Itoa(int(0)) + "` "
	info = info + "`Assists: " + strconv.Itoa(0) + "` \n"

	info = info + "`K/D: " + strconv.Itoa(int(0)) + "` "
	info = info + "`HS%: " + strconv.Itoa(int(0)) + "` "
	info = info + "`Peak: " + RankDetails[mmr.PeakRank].TierName + "` "
	info = info + "`First Bloods: " + strconv.Itoa(int(0)) + "`"

	if P.GameName == "" {
		P.GameName = "undefined"
	}

	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    P.GameName,
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
	CommandHandlers["switch_teams"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		/*isAlly := i.Message.Embeds[len(i.Message.Embeds)-1].Color == 3124052
		var team int

		if isAlly {
			team = 1
		} else {
			team = 0
		}

		edit := discordgo.NewMessageEdit(i.Message.ChannelID, i.Message.ID)

		edit.Embeds = &embeds[team]

		s.ChannelMessageEditComplex(edit)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		s.InteractionResponseDelete(i.Interaction)*/

	}
}
