package main

import (
	"os"
	"time"

	"github.com/MasterDimmy/go-cls"
	"github.com/bwmarrin/discordgo"

	Types "valcord/types"
)

type DISCORD_BOT_DATA struct {
	token string
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "setup_channel",
			Description: "Sets channel up to become a type",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel_type",
					Description: "Sets the type of channel",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "current_session",
							Value: "current-session",
						},
					},
				},
			},
		},
		{
			Name:        "agent_select_request",
			Description: "Prints current agentSelect info",
		},
		{
			Name:        "match_request",
			Description: "Prints current match info",
		},
		{
			Name:        "request_shop",
			Description: "Prints current shop",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "banner",
					Description: "Do you want banner only or only rotating shop",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Banner",
							Value: "banner",
						},
						{
							Name:  "Rotation",
							Value: "rotation",
						},
						{
							Name:  "Accessory",
							Value: "accessory",
						},
						{
							Name:  "Night Market",
							Value: "night_market",
						},
					},
				},
			},
		},
	}
)

func setupComponents() {

	Types.CommandHandlers["setup_channel"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		value := i.ApplicationCommandData().Options[0].Value

		settings["current_session_channel"] = i.ChannelID

		Types.CheckSettingsData(settings)

		Types.NewLog("Channel (" + i.ChannelID + ") has now been designated '" + value.(string) + "'")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Successfully set channel to type " + value.(string),
			},
		})

	}

	Types.CommandHandlers["agent_select_request"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)

		Types.Request_agentSelect(general_valorant_information.player_info, general_valorant_information.regional_data, i.ChannelID, s)

	}

	Types.CommandHandlers["match_request"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)

		Types.Request_match(general_valorant_information.player_info, general_valorant_information.regional_data, i.ChannelID, s)

	}

	Types.CommandHandlers["request_shop"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		Types.NewLog("Shop has been requested")

		Type := i.ApplicationCommandData().Options[0].Value.(string)

		Messages := Types.RequestShopEmbed(Type, general_valorant_information.player_info, general_valorant_information.regional_data)

		for _, Message := range Messages {
			_, err := s.ChannelMessageSendComplex(i.ChannelID, &Message)
			checkError(err)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)
	}
}

var (
	discord *discordgo.Session
)

func NoChannelWithID() {

	cls.CLS()

	Types.NewLog("No channel with id '" + settings["current_session_channel"] + "' found\n")

	Types.NewLog("Ensure channel exists")
	Types.NewLog("Ensure ChannelID is correct (Will be automatically reset for this purpose)\n ")

	settings["current_session_channel"] = ""
	Types.CheckSettingsData(settings)

	cleanup()
	os.Exit(1)

}

func checkChannelID() {

	_, err := discord.Channel(settings["current_session_channel"])

	if err != nil {

		if err.Error() == `HTTP 404 Not Found, {"message": "Unknown Channel", "code": 10003}` {

			NoChannelWithID()

		}

		if err.Error() == `HTTP 404 Not Found, {"message": "404: Not Found", "code": 0}` {

			NoChannelWithID()

		}

		settings["current_session_channel"] = ""
		Types.CheckSettingsData(settings)

		checkError(err)

	}

}

func serverInaccessable(inviteLink string) {

	cls.CLS()

	Types.NewLog("Bot is not in server with id '" + settings["server_id"] + "'")
	Types.NewLog("Make sure to invite the bot into a server!\n ")

	Types.NewLog("Bot Invite Link: '" + inviteLink + "'\n")

	Types.NewLog("Saved ServerID will be reset incase of error\n ")

	settings["server_id"] = ""
	Types.CheckSettingsData(settings)

	cleanup()
	os.Exit(1)

}

func checkServerID(inviteLink string) {

	_, err := discord.Guild(settings["server_id"])

	if err != nil {

		Types.NewLog("`" + err.Error() + "`")

		if err.Error() == `HTTP 404 Not Found, {"message": "Unknown Guild", "code": 10004}` {

			serverInaccessable(inviteLink)

		}

		if err.Error() == `HTTP 404 Not Found, {"message": "404: Not Found", "code": 0}` {

			serverInaccessable(inviteLink)

		}

		settings["server_id"] = ""
		Types.CheckSettingsData(settings)

		checkError(err)

	}

}

func discord_setup() {

	discord_bot_data := DISCORD_BOT_DATA{
		token: settings["discord_api_token"],
	}

	var err error

	discord, err = discordgo.New("Bot " + discord_bot_data.token)
	checkError(err)

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {

		Types.NewLog("Discord bot: Ready")

		// Listen for matches to auto-send match data

		Types.ListenForMatch(general_valorant_information.player_info, general_valorant_information.regional_data, Types.Client, time.Second*20, discord)

	})

	err = discord.Open()

	InviteLink := "https://discord.com/oauth2/authorize?client_id=" + discord.State.User.ID + "&permissions=93184&integration_type=0&scope=bot'"

	checkServerID(InviteLink)
	checkChannelID()

	if err != nil {

		if err.Error() == "websocket: close 4004: Authentication failed." {

			// Bot token invalid

			Types.NewLog("Bot token provided was invalid..")
			Types.NewLog("Reseting saved bot token")

			settings["discord_api_token"] = ""
			Types.CheckSettingsData(settings)

			cleanup()
			os.Exit(1)

		}

		checkError(err)

	}

	if Flags["Reset"] {
		Types.NewLog("Cleaning up commands..")
		command_cleanup()
		cleanup()
		os.Exit(1)
	}

	if Flags["Link"] {
		Types.NewLog("Bot Invite Link: '" + "https://discord.com/oauth2/authorize?client_id=" + discord.State.User.ID + "&permissions=93184&integration_type=0&scope=bot'")
		cleanup()
		os.Exit(1)
	}

	setupComponents()

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		go func() {

			switch i.Type {
			case discordgo.InteractionApplicationCommand:
				if h, ok := Types.CommandHandlers[i.ApplicationCommandData().Name]; ok {
					h(s, i)
				}
			case discordgo.InteractionMessageComponent:
				if h, ok := Types.CommandHandlers[i.MessageComponentData().CustomID]; ok {
					h(s, i)
				}
			}

		}()

	})

	allCommands, err := discord.ApplicationCommands(discord.State.User.ID, settings["server_id"])
	checkError(err)

	for _, v := range commands {

		var command discordgo.ApplicationCommand

		for _, y := range allCommands {

			if y.Name == v.Name {

				command = *y
				break

			}

		}

		if command.Name == "" {

			Types.NewLog("Trying to init '" + v.Name + "'")

			cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, settings["server_id"], v)
			checkError(err)
			Types.NewLog("Initialized '" + cmd.Name + "'")

		}

	}

	Types.NewLog("Discord bot UserID: " + discord.State.User.ID)

}

func command_cleanup() {

	commands, err := discord.ApplicationCommands(discord.State.User.ID, settings["server_id"])
	checkError(err)

	for _, v := range commands {

		Types.NewLog("Cleaning command: ", v)

		if v == nil {
			continue
		}
		err := discord.ApplicationCommandDelete(discord.State.User.ID, settings["server_id"], v.ID)
		checkError(err)
	}

}
