package main

import (
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

	socket.Load()

	services.Socket.SetPMHook(func(message *services.NewMessage) {
		log.Printf("Received PM from %s: %s", message.MessageFrom, message.RawMessage)
		socket.CMDHandler.HandleCommand(services.Socket, message)
	})

	services.Socket.SetOrderHook(func(order *services.SubscriptionsNewOrder) {
		log.Printf("New order: %s | %s | %d * %s @ %d platinum", order.User.GameName, order.OrderType, order.Quantity, order.Item.EN.Name, order.Price)

		services.DB.InsertOrder(order)
	})

	// Create a new Discord session using the provided bot token.
	s, err = discordgo.New("Bot " + os.Getenv("TOKEN"))

	if err != nil {
		log.Fatalf("Error creating Discord session: %s", err)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		s.UpdateStatusComplex(discordgo.UpdateStatusData{
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

	defer s.Close()

	// Sync the item list every day, but check if it's been
	// more than 24 hours since last sync every 5 minutes
	go func() {
		for {

			if services.Syncing {
				time.Sleep(time.Second * 5)
				continue
			}

			ls, ds, err := services.DB.GetLastSynced()

			if err != nil && err != gorm.ErrRecordNotFound {
				log.Fatalf("Error getting last synced time: %s", err)
			}

			if err == gorm.ErrRecordNotFound {
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

			// if is has been more than 24 hours since last sync
			if ls.Add(24 * time.Hour).Before(time.Now()) {

				var deep bool = false

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
