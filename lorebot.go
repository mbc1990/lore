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
			// Replace C2147483705 with your Channel ID
			// rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C2147483705"))

		case *slack.MessageEvent:
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
