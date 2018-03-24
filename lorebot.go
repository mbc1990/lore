package main

import "fmt"
import "log"
import "os"
import "github.com/nlopes/slack"

type Lorebot struct {
	Conf             *Configuration
	Pg               *PostgresClient
	AvailableClasses map[string]string
}

// channel + timestamp is apparently a UUID for slack
// So when someone lore reacts, we look up the channel history,
// find the message with that timestamp, and store it
func (l *Lorebot) HandleLoreReact(channel string, timestamp string) {
	// TODO: Fetch channel history
	// TODO: Search for matching timestamp
	// TODO: If found, add/update lore
	// TODO: If not found, send error message
}

func (l *Lorebot) Start() {
	api := slack.New(l.Conf.Token)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			// TODO: If lorebot is mentioned
			// TODO: Parse message
			// TODO: If parsed, perform command
			fmt.Printf("Message: %v\n", ev)

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		case *slack.ReactionAddedEvent:
			// TODO: Test for duplicate lore
			// TODO: If duplicate, increment score
			// TODO: If not duplicate, add lore
			// TODO: This doesn't give any information about the message that was reacted to?
			if ev.Reaction == "lore" {
				fmt.Printf("Lore detected")
				fmt.Printf("Reaction: %s\n", ev.Reaction)
				channel := ev.Item.Channel
				timestamp := ev.Item.Timestamp
				fmt.Printf("Channel: %s\n", channel)
				fmt.Printf("timestamp: %s\n", timestamp)
				go l.HandleLoreReact(channel, timestamp)

				// fmt.Printf("Item: %v\n", msg.Data)
				// fmt.Printf("Details: %v\n", msg)
			}

		default:
			// Ignore other events..
			fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func NewLorebot(conf *Configuration) *Lorebot {
	lorebot := new(Lorebot)
	lorebot.Conf = conf
	lorebot.Pg = NewPostgresClient(lorebot.Conf.PGHost, lorebot.Conf.PGPort,
		lorebot.Conf.PGUser, lorebot.Conf.PGPassword, lorebot.Conf.PGDbname)
	return lorebot
}
