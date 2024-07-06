package commands

import (
	"fmt"
	"time"
	"vaportrader/src/constants"
	"vaportrader/src/services"

	"github.com/bwmarrin/discordgo"
)

func LinkCommand() Command {
	return Command{
		Name:        "link",
		Description: "Used to link your Warframe Market account to your Discord account.",
		Usage:       "link username: VaporTrader",
		Category:    "Utility",
		Cooldown:    5 * time.Second,
		Handler:     LinkHandler,
		Permissions: LinkPermissions,
		Modal:       LinkModal,
		Action:      LinkAction,
		Options:     []*discordgo.ApplicationCommandOption{
			// {
			// 	Name:        "username",
			// 	Description: "The username you of your Warframe Market account.",
			// 	Type:        discordgo.ApplicationCommandOptionString,
			// 	Required:    true,
			// },
		},
	}
}

type AccountLinkStatus struct {
	ID          string
	Phase       uint8
	Code        string
	Username    *string
	Profile     *services.ApiProfile
	Interaction *discordgo.Interaction
}

type LinkModalResponse struct {
	Username string `json:"username_field"`
}

func LinkAction(s *discordgo.Session, m *discordgo.InteractionCreate, ctx ActionContext) (bool, error) {
	if ctx.Action.CustomID == "link_account_wfm_"+ctx.User.ID+"_accept" {

		rawEntry := services.KV.Get("link_account_wfm_" + ctx.User.ID)

		if rawEntry != nil {
			entry := rawEntry.Value.(AccountLinkStatus)

			if entry.Phase == 1 {
				entry.Phase = 2

				s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredMessageUpdate,
				})

				_, err := s.InteractionResponseEdit(entry.Interaction, &discordgo.WebhookEdit{
					Embeds: &[]*discordgo.MessageEmbed{
						{
							Title:       "Here's your code!",
							Description: "Send this code to the [Vapor Trader](https://warframe.market/profile/VaporTrader) account as a private message. This will automatically link your Discord, and Warframe Market accounts",
							Color:       constants.ThemeColor,
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:   "Code",
									Value:  entry.Code,
									Inline: true,
								},
								{
									Name: "Expires",
									// Show the expiry time as a relative time
									Value:  rawEntry.Expiry.Format("in 2 minutes"),
									Inline: true,
								},
								{
									Name:  "Usage",
									Value: "```\nlink " + entry.Code + "\n```",
								},
							},
						},
					},
					Components: &[]discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Take me there",
									Style:    discordgo.LinkButton,
									URL:      "https://warframe.market/profile/VaporTrader",
									Disabled: false,
								},
							},
						},
					},
				})

				entry.Interaction = m.Interaction

				if err != nil {
					println(err.Error())
					return false, err
				}

				services.KV.Set("link_account_wfm_"+ctx.User.ID, entry, time.Minute*15)

			} else {
				services.KV.Delete("link_account_wfm_" + ctx.User.ID)
				return false, fmt.Errorf("Link request in invalid state - Clearing")
			}
		}
	} else if ctx.Action.CustomID == "link_account_wfm_"+ctx.User.ID+"_reject" {
		rawEntry := services.KV.Get("link_account_wfm_" + ctx.User.ID)

		if rawEntry != nil {
			entry := rawEntry.Value.(AccountLinkStatus)

			s.InteractionResponseDelete(m.Interaction)
			err := s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "modals_link_account_wfm_" + ctx.User.ID,
					Title:    "Link your Warframe Market account",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "username_field",
									Label:       "What is your Warframe Market username?",
									Style:       discordgo.TextInputShort,
									Placeholder: "(Case sensitive) VaporTrader",
									Value:       *entry.Username,
									Required:    true,
									MaxLength:   60,
									MinLength:   1,
								},
							},
						},
					},
				},
			})

			if err != nil {
				println(err.Error())
				return false, err
			}

			entry.Phase = 0
			entry.Profile = nil
			entry.Username = nil
			services.KV.Set("link_account_wfm_"+ctx.User.ID, entry, time.Minute*15)
		} else {
			return false, fmt.Errorf("No Link entry found for user %s - Somehow you broke it?\nPlease report this incident to the developer.", ctx.User.ID)
		}

	}

	return true, nil
}

func LinkModal(s *discordgo.Session, m *discordgo.InteractionCreate, ctx ModalContext) (bool, error) {
	username := ctx.Options["username_field"]

	rawEntry := services.KV.Get("link_account_wfm_" + ctx.User.ID)

	if rawEntry != nil {
		entry := rawEntry.Value.(AccountLinkStatus)

		if entry.Phase == 0 {
			entry.Phase = 1

			if entry.Username == nil {
				entry.Username = &username
			}

			if entry.Profile == nil {
				apiProfile, err := services.API.GetUser(*entry.Username)

				if err != nil {
					return false, err
				}

				entry.Profile = apiProfile
			}

			var thumbnail *discordgo.MessageEmbedThumbnail = nil

			if entry.Profile.Avatar != nil {
				thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: "https://warframe.market/static/assets/" + *entry.Profile.Avatar,
				}
			}

			entry.Interaction = m.Interaction

			var banText string = "No"

			if entry.Profile.Banned {
				banText = "Yes"
			}

			services.KV.Set("link_account_wfm_"+ctx.User.ID, entry, time.Minute*15)
			err := s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title: "Is this your Warframe Market account?",
							Color: constants.ThemeColor,
							URL:   fmt.Sprintf("https://warframe.market/profile/%s", entry.Profile.IngameName),
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:   "Username",
									Value:  entry.Profile.IngameName,
									Inline: true,
								},
								{
									Name:   "Status",
									Value:  fmt.Sprintf("**%s**", entry.Profile.Status),
									Inline: true,
								},
								{
									Name:   "Banned",
									Value:  banText,
									Inline: true,
								},
							},
							Thumbnail: thumbnail,
						},
					},
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									CustomID: "link_account_wfm_" + ctx.User.ID + "_accept",
									Label:    "Yes",
									Style:    discordgo.SuccessButton,
									Disabled: false,
								},
								discordgo.Button{
									CustomID: "link_account_wfm_" + ctx.User.ID + "_reject",
									Label:    "No",
									Style:    discordgo.DangerButton,
									Disabled: false,
								},
							},
						},
					},
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})

			if err != nil {
				return false, err
			}
		}

	} else {
		return false, fmt.Errorf("No Link entry found for user %s - Somehow you broke it?\nPlease report this incident to the developer.", ctx.User.ID)
	}

	return true, nil
}

func LinkHandler(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, error) {

	rawEntry := services.KV.Get("link_account_wfm_" + ctx.User.ID)

	if rawEntry == nil {
		code, err := services.GenerateOTP(ctx.User.ID)

		println("Generated code: " + code)

		if err != nil {
			return false, err
		}

		services.KV.Set(code+":totp", "link_account_wfm_"+ctx.User.ID, time.Hour)

		entry := services.KV.Set("link_account_wfm_"+ctx.User.ID, AccountLinkStatus{
			ID:    ctx.User.ID,
			Phase: 0,
			Code:  code,
		}, time.Minute*15)

		err = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "modals_link_account_wfm_" + ctx.User.ID,
				Title:    "Link your Warframe Market account",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "username_field",
								Label:       "What is your Warframe Market username?",
								Style:       discordgo.TextInputShort,
								Placeholder: "(Case sensitive) VaporTrader",
								Required:    true,
								MaxLength:   60,
								MinLength:   1,
							},
						},
					},
				},
			},
		})

		if err != nil {
			println(err.Error())
			return false, err
		}

		println("Initialized new link request - " + entry.Value.(AccountLinkStatus).Code)
	} else {
		// user is already linking an account
		return false, fmt.Errorf("You aready have an active link attempt. Please complete that first (or wait for 15 minutes for it to expire).")
	}

	return true, nil
}

func LinkPermissions(s *discordgo.Session, m *discordgo.InteractionCreate, ctx CommandContext) (bool, string, error) {

	if !ctx.User.WfmID.Valid {
		return true, "", nil
	}

	if ctx.User.WfmID.Valid {
		return false, "You already have a Warframe Market account linked to this bot\nIf you wish to link another account, please use `/unlink` to invalidate the linked account.", nil
	}

	return false, "This error shouldn't happen, as it indicates a corrupted database.", nil
}
