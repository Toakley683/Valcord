package main

import (
	"math"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/MasterDimmy/go-cls"
	"github.com/bwmarrin/discordgo"
	"github.com/lynn9388/supsub"
	"github.com/ncruces/zenity"

	Types "valcord/types"
)

type DISCORD_BOT_DATA struct {
	token string
}

var (
	ItemTypes = map[string]string{
		"Agents": "01bb38e1-da47-4e6a-9b3d-945fe4655707",
		//"Contracts": "f85cb6f7-33e5-4dc8-b609-ec7212301948",
		"Sprays":      "d5f120f8-ff8c-4aac-92ea-f2b5acbe9475",
		"Gun Buddies": "dd3bf334-87f3-40bd-b043-682a57a8dc3a",
		"Cards":       "3f296c07-64c3-494c-923b-fe692a4fa1bd",
		"Skins":       "e7c63390-eda7-46e0-bb7a-a6abdacd2433",
		"Titles":      "de7caa6b-adf7-4588-bbd1-143831e786c6",
	}

	ItemTypeOptions = func() []*discordgo.ApplicationCommandOptionChoice {

		R := make([]*discordgo.ApplicationCommandOptionChoice, len(ItemTypes))

		I := 0

		for N, V := range ItemTypes {

			R[I] = &discordgo.ApplicationCommandOptionChoice{
				Name:  N,
				Value: V,
			}

			I++

		}

		return R

	}

	ItemTypeSelect = func() []discordgo.SelectMenuOption {

		R := make([]discordgo.SelectMenuOption, len(ItemTypes))

		I := 0

		for N, V := range ItemTypes {

			R[I] = discordgo.SelectMenuOption{
				Label: N,
				Value: V,
			}

			I++

		}

		return R

	}

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
		{
			Name:        "show_owned",
			Description: "Prints own items of X type",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "What type of item",
					Required:    true,
					Choices:     ItemTypeOptions(),
				},
			},
		},
	}
)

func clampIntegar(V int, Min int, Max int) int {

	if V > Max {
		return Max
	}

	if V < Min {
		return Min
	}

	return V

}

type showOwnedItemsCallback struct {
	sendMessage   func(*discordgo.WebhookParams) (*discordgo.Message, error)
	updateMessage func(*discordgo.Message, *discordgo.WebhookEdit)
}

func showOwnedItems(Type string, callbackInfo showOwnedItemsCallback) {

	Types.NewLog(Type)

	Items := Types.GetOwnedItems(general_valorant_information.player_info, general_valorant_information.regional_data, Type)

	DefaultLLength := 10
	MaxPages := int(math.Ceil(float64(len(Items))/float64(DefaultLLength))) - 1

	EmbedList := func(PageIndex int) *[]*discordgo.MessageEmbed {

		LLength := DefaultLLength
		S := LLength * PageIndex

		if len(Items)-S < 10 {
			LLength = len(Items) - S
		}

		Types.NewLog("Start:", S, "Items:", len(Items))

		if len(Items)-S <= 0 {

			return &[]*discordgo.MessageEmbed{
				{
					Title: "No content available",
				},
			}
		}

		Types.NewLog(LLength)

		ReturnedList := make([]*discordgo.MessageEmbed, LLength)

		for Index := range LLength {

			if Index+S >= len(Items) {
				break
			}

			Data := Items[Index+S]

			ItemData := Types.ItemIDWTypeToStruct(Type, Data.ItemID, 0)

			AuthorIcon := ""
			ColorHex := "0xffffff"
			Description := ""
			DisplayIcon := ItemData.StreamedVideo

			if Type == ItemTypes["Agents"] {
				AuthorIcon = ItemData.StreamedVideo
				DisplayIcon = ItemData.DisplayIcon
			}

			if Type == ItemTypes["Skins"] {
				AuthorIcon = ""
				DisplayIcon = ItemData.DisplayIcon
			}

			if ItemData.Color != "" {
				ColorHex = "0x" + ItemData.Color[:len(ItemData.Color)-2]
			}

			if ItemData.Description != "" {
				Description = ItemData.Description
			}

			Color, err := strconv.ParseInt(ColorHex, 0, 0)
			checkError(err)

			ReturnedList[Index] = &discordgo.MessageEmbed{
				Description: Description,
				Author: &discordgo.MessageEmbedAuthor{
					Name:    ItemData.Name + supsub.ToSup("("+strconv.Itoa(Index+S+1)+")"),
					IconURL: AuthorIcon,
				},
				Color: int(Color),
			}

			if DisplayIcon != "" {

				Types.NewLog("("+strconv.Itoa(Index)+") Icon:", DisplayIcon)

			}

			switch Type {
			default:
				ReturnedList[Index].Image = &discordgo.MessageEmbedImage{
					URL: DisplayIcon,
				}
			case "01bb38e1-da47-4e6a-9b3d-945fe4655707": // IsAgent
				ReturnedList[Index].Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: DisplayIcon,
				}
			}

			if Index == LLength-1 {
				ReturnedList[Index].Footer = &discordgo.MessageEmbedFooter{
					Text: "Entries ( " + strconv.Itoa(LLength-Index+S) + " to " + strconv.Itoa(LLength+S) + " ) ( Total " + strconv.Itoa(len(Items)) + " )" + " ( Page " + strconv.Itoa(PageIndex+1) + " )",
				}

			}

		}

		return &ReturnedList

	}

	// Add list button to change item type

	msgData, err := callbackInfo.sendMessage(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "Loading..",
			},
		},
	})
	checkError(err)

	UpdateMessage := func(Page int) {

		Types.NewLog("Page:", Page)

		callbackInfo.updateMessage(msgData, &discordgo.WebhookEdit{
			Content: func(A string) *string {
				return &A
			}(""),
			Embeds: EmbedList(Page),
			Components: &[]discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "content_reselect",
							Placeholder: "Reselect Type",
							Options:     ItemTypeSelect(),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style:    2,
							CustomID: "content_prev",
							Label:    "Prev",
						},
						discordgo.Button{
							Style:    2,
							CustomID: "content_next",
							Label:    "Next",
						},
					},
				},
			},
		})

	}

	UpdateMessage(0)

	PageIndex := 0

	Types.CommandHandlers["content_prev"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		go s.InteractionResponseDelete(i.Interaction)

		PageIndex = clampIntegar(PageIndex-1, 0, MaxPages)

		UpdateMessage(PageIndex)

	}

	Types.CommandHandlers["content_next"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		go s.InteractionResponseDelete(i.Interaction)

		PageIndex = clampIntegar(PageIndex+1, 0, MaxPages)

		UpdateMessage(PageIndex)

	}

	Types.CommandHandlers["content_reselect"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Requested",
			},
		})

		go s.InteractionResponseDelete(i.Interaction)

		Type = i.MessageComponentData().Values[0]

		Items = Types.GetOwnedItems(general_valorant_information.player_info, general_valorant_information.regional_data, Type)

		MaxPages = int(math.Ceil(float64(len(Items))/float64(DefaultLLength))) - 1
		PageIndex = 0

		UpdateMessage(0)

	}
}

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

	Types.CommandHandlers["show_owned"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		Response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		}

		Response.Data.Flags = discordgo.MessageFlagsEphemeral

		s.InteractionRespond(i.Interaction, Response)

		Types.NewLog("Items has been requested")

		Type := i.ApplicationCommandData().Options[0].Value.(string)

		cb := showOwnedItemsCallback{
			sendMessage: func(wp *discordgo.WebhookParams) (*discordgo.Message, error) {
				return s.FollowupMessageCreate(i.Interaction, true, wp)
			},
			updateMessage: func(m *discordgo.Message, we *discordgo.WebhookEdit) {
				s.FollowupMessageEdit(i.Interaction, m.ID, we)
			},
		}

		showOwnedItems(Type, cb)
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

	zenity.Error("Discord bot is not in server with a channel with the ID '"+settings["current_session_channel"]+"' \n\nWe will open the instruction guide for it now",
		zenity.Title("Valcord"))

	cmd := "cmd.exe"
	args := []string{"/c", "start", "https://github.com/Toakley683/Valcord/wiki/Retrieving-Session-Channel"}

	exec.Command(cmd, args...).Start()

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

	zenity.Error("Discord bot is not in server with ID '"+settings["server_id"]+"' \n\nWe will open the invite link for it now",
		zenity.Title("Valcord"))

	cmd := "cmd.exe"
	args := []string{"/c", "start", inviteLink}

	exec.Command(cmd, args...).Start()

	os.Exit(0)

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

		Types.ListenForMatch(general_valorant_information.player_info, general_valorant_information.regional_data, Types.Client, time.Second*20, discord, menuListenForMatch)

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

	commandInit()

	Types.NewLog("Discord bot UserID: " + discord.State.User.ID)

}

func commandInit() {

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
}

func command_cleanup() {

	if discord == nil {
		return
	}

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
