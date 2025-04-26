package types

import "github.com/bwmarrin/discordgo"

var (
	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
)
