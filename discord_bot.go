package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

		fmt.Println("Channel (" + i.ChannelID + ") has now been designated '" + value.(string) + "'")

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

		fmt.Println("Shop has been requested")

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

	settings["current_session_channel"] = ""
	Types.CheckSettingsData(settings)

	log.Println("No channel with id '" + settings["current_session_channel"] + "' found")

	cleanup()
	os.Exit(0)

}

func NoServerWithID() {

	settings["server_id"] = ""
	Types.CheckSettingsData(settings)

	log.Println("Server with id '" + settings["server_id"] + "' (Make sure bot is in server)")

	cleanup()
	os.Exit(0)

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

		checkError(err)

	}

}

func checkServerID() {

	_, err := discord.Guild(settings["server_id"])

	if err != nil {

		fmt.Println("`" + err.Error() + "`")

		if err.Error() == `HTTP 404 Not Found, {"message": "Unknown Guild", "code": 10004}` {

			NoServerWithID()

		}

		if err.Error() == `HTTP 404 Not Found, {"message": "404: Not Found", "code": 0}` {

			NoServerWithID()

		}

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

		checkServerID()
		checkChannelID()

		fmt.Println("Discord bot: Ready")

		// Listen for matches to auto-send match data

		Types.ListenForMatch(general_valorant_information.player_info, general_valorant_information.regional_data, Types.Client, time.Second*20, discord)

	})

	err = discord.Open()

	if err != nil {

		if err.Error() == "websocket: close 4004: Authentication failed." {

			// Bot token invalid

			fmt.Println("Bot token provided was invalid..")
			fmt.Println("Reseting saved bot token")

			settings["discord_api_token"] = ""
			Types.CheckSettingsData(settings)

			cleanup()
			os.Exit(0)

		}

		checkError(err)

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

	//command_cleanup()

	for _, v := range commands {

		var command discordgo.ApplicationCommand

		for _, y := range allCommands {

			if y.Name == v.Name {

				command = *y
				break

			}

		}

		if command.Name == "" {

			fmt.Println("Trying to init '" + v.Name + "'")

			cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, settings["server_id"], v)
			checkError(err)
			fmt.Println("Initialized '" + cmd.Name + "'")

		}

	}

	fmt.Println("Discord bot UserID: " + discord.State.User.ID)

}

/*func command_cleanup() {

	commands, err := discord.ApplicationCommands(discord.State.User.ID, settings["server_id"])
	checkError(err)

	for _, v := range commands {
		if v == nil {
			continue
		}
		err := discord.ApplicationCommandDelete(discord.State.User.ID, settings["server_id"], v.ID)
		checkError(err)
	}

}*/
