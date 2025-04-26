package main

import (
	"fmt"
	"reflect"

	"github.com/bwmarrin/discordgo"

	Types "valcord/types"
)

type DISCORD_BOT_DATA struct {
	token string
}

func ToMap(in interface{}, tag string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(tag); tagv != "" {
			// set key of map to value in struct field
			out[tagv] = v.Field(i).Interface()
		}
	}
	return out, nil
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

	important_channels = make(map[string]*discordgo.Channel)
)

func setupComponents() {

	Types.CommandHandlers["setup_channel"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		value := i.ApplicationCommandData().Options[0].Value

		settings["current_session_channel"] = i.ChannelID

		check_settings_data(settings)

		fmt.Println("Channel (" + i.ChannelID + ") has now been designated '" + value.(string) + "'")

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Successfully set channel to type " + value.(string),
			},
		})

	}

	Types.CommandHandlers["agent_select_request"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		fmt.Println("Requested Agent Select")

		AgentSelect := Types.GetAgentSelectInfo(general_valorant_information.player_info, general_valorant_information.entitlements, general_valorant_information.regional_data)

		_, err := s.ChannelMessageSendEmbed(i.ChannelID, Types.NewAgentSelectEmbed(AgentSelect))
		checkError(err)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)
	}

	Types.CommandHandlers["match_request"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		Types.Request_match(general_valorant_information.player_info, general_valorant_information.entitlements, general_valorant_information.regional_data, i.ChannelID, s)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		s.InteractionResponseDelete(i.Interaction)
	}

	Types.CommandHandlers["request_shop"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		fmt.Println("Requested Shop")

		Type := i.ApplicationCommandData().Options[0].Value.(string)

		fmt.Println(Type)

		Messages := Types.RequestShopEmbed(Type, general_valorant_information.player_info, general_valorant_information.entitlements, general_valorant_information.regional_data)

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

func discord_setup() {

	discord_bot_data := DISCORD_BOT_DATA{
		token: settings["discord_api_token"],
	}

	var err error

	discord, err = discordgo.New("Bot " + discord_bot_data.token)
	checkError(err)

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {

		fmt.Println("Ready")

		important_channels["current_session_channel"], err = discord.Channel(settings["current_session_channel"])

		if important_channels["current_session_channel"] == nil {
			fmt.Println("No 'current-session' server selected! Use '/setup_channel channel_type:current_session' in the channel to set!")
		}

	})

	err = discord.Open()
	checkError(err)

	setupComponents()

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

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

	})

	allCommands, err := discord.ApplicationCommands(discord.State.User.ID, settings["server_id"])
	checkError(err)

	//command_cleanup()

	for _, v := range commands {

		var test discordgo.ApplicationCommand

		for _, y := range allCommands {

			if y.Name == v.Name {

				test = *y
				break

			}

		}

		if test.Name == "" {

			fmt.Println("Trying to init '" + v.Name + "'")

			cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, settings["server_id"], v)
			checkError(err)
			fmt.Println("Initialized '" + cmd.Name + "'")

		}

	}

	fmt.Println(discord.State.User.ID)

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
