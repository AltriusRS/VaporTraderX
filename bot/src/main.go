package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"
	"vaportrader/src/commands"
	"vaportrader/src/services"
	"vaportrader/src/socket"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var s *discordgo.Session

func main() {
	// Find .env file
	err := godotenv.Load(".dev.env")

	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	services.InitDatabase()
	services.InitSocket(s)
	services.InitI18n()

	lm, err := json.Marshal(services.LanguageManager)

	print(string(lm))

	println(services.LanguageManager.Get(nil, "constants.wfm.footer", &map[string]interface{}{
		"maintainer": "Altrius",
	}))

	os.Exit(0)

	socket.Load()

	services.Socket.SetPMHook(func(message *services.NewMessage) {
		log.Printf("Received PM from %s: %s", message.MessageFrom, message.RawMessage)
		socket.CMDHandler.HandleCommand(services.Socket, message)
	})

	services.Socket.SetOrderHook(func(order *services.SubscriptionsNewOrder) {
		log.Printf("New order: %s | %s | %d * %s @ %d platinum", order.User.GameName, order.OrderType, order.Quantity, order.Item.EN.Name, order.Price)

		err := services.DB.InsertOrder(order)
		if err != nil {
			log.Printf("Error inserting order: %s", err)
			return
		}
	})

	// Create a new Discord session using the provided bot token.
	s, err = discordgo.New("Bot " + os.Getenv("TOKEN"))

	if err != nil {
		log.Fatalf("Error creating Discord session: %s", err)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		_ = s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name:  "Warframe",
					Type:  discordgo.ActivityTypeGame,
					State: "Watching the markets",
				},
			},
			AFK:    false,
			Status: "online",
		})
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.PresenceUpdate) {
		log.Printf("Presence update: %v - %v - %v", r.User.Username, r.Status, r.Presence.Activities[0])
	})

	s.Identify.Intents = discordgo.IntentsGuildPresences

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	commands.Load(s)

	// Add command handlers
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commands.CMDHandler.HandleCommand(s, i)
	})

	defer func(s *discordgo.Session) {
		err := s.Close()
		if err != nil {
			log.Fatalf("Error closing Discord session: %s", err)
		}
	}(s)

	// Sync the item list every day, but check if it's been
	// more than 24 hours since last sync every 5 minutes
	go func() {
		for {

			if services.Syncing {
				time.Sleep(time.Second * 5)
				continue
			}

			ls, ds, err := services.DB.GetLastSynced()

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Fatalf("Error getting last synced time: %s", err)
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("No last synced time found, syncing...")
				stateInfo := services.StateInfo{
					LastSynced: time.Now(),
				}

				err = services.DB.Create(&stateInfo)

				if err != nil {
					log.Fatalf("Error creating state info: %s", err)
				}
			}

			log.Printf("Time to next sync: %s", time.Until(ls.Add(24*time.Hour)))

			// if it has been more than 24 hours since last sync
			if ls.Add(24 * time.Hour).Before(time.Now()) {

				var deep = false

				if ds.Add(7 * 24 * time.Hour).Before(time.Now()) {
					deep = true
				}

				if deep {
					log.Println("Deep syncing...")
				} else {
					log.Println("Syncing...")
				}

				err = services.Sync(deep)
				if err != nil {
					log.Fatalf("Error syncing: %s", err)
				}
			}
			time.Sleep(time.Minute * 1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
