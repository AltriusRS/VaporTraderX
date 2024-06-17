package constants

import "github.com/bwmarrin/discordgo"

const Version = "0.1"
const Author = "Altrius"
const AuthorID = "180639594017062912"
const AuthorString = "<@" + AuthorID + ">" + " - " + Author
const BotName = "VaporTrader"
const ThemeColor = 0xae6eb4

var Footer = &discordgo.MessageEmbedFooter{
	Text: BotName + " " + Version + " | Made with ❤️ by " + Author,
}

const WFMAuthor = "FatalCenturion"

const WFMFooter = BotName + " " + Version + " | Made with ❤️ by " + WFMAuthor
