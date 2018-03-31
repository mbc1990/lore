package main

import "fmt"
import "strings"
import "github.com/nlopes/slack"

type Lorebot struct {
	Conf      *Configuration
	Pg        *PostgresClient
	SlackAPI  *slack.Client
	LorebotID string
}

// channel + timestamp is apparently a UUID for slack
// So when someone lore reacts, we look up the channel history,
// find the message with that timestamp, and store it
func (l *Lorebot) HandleLoreReact(channelId string, timestamp string) {
	params := slack.NewHistoryParameters()
	history, err := l.SlackAPI.GetChannelHistory(channelId, params)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	for _, message := range history.Messages {
		if message.Timestamp == timestamp {
			fmt.Println("adding lore for " + message.User + ": " + message.Text)
			spl := strings.Split(message.Text, " ")
			spl = spl[1:]
			cleaned := strings.Join(spl, " ")
			fmt.Println("Cleaned: " + cleaned)
			if l.Pg.LoreExists(cleaned, message.User) {
				// TODO: Upvote lore
				return
			}
			l.Pg.InsertLore(message.User, cleaned)
			break
		}
	}
}

func (l *Lorebot) Start() {
	rtm := l.SlackAPI.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: \n")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			// fmt.Println("Infos:", ev.Info)
			// fmt.Println("Connection counter:", ev.ConnectionCount)

		case *slack.MessageEvent:
			spl := strings.Split(ev.Text, " ")
			splMsg := spl[1:]
			cleaned := strings.Join(splMsg, " ")
			userId := strings.Replace(spl[0], "<", "", 1)
			userId = strings.Replace(userId, ">", "", 1)
			userId = strings.Replace(userId, "@", "", 1)
			channel := ev.Channel
			if userId == l.LorebotID {
				if cleaned == "recent" {
					out := ""
					recent := l.Pg.RecentLore()
					for _, lore := range recent {
						info, _ := l.SlackAPI.GetUserInfo(lore.UserID)
						fmt.Println("\nInfo", info)
						displayName := info.Profile.DisplayName
						out += displayName + ": " + lore.Message + "\n"
					}
					params := slack.PostMessageParameters{Username: "Lorebot", IconEmoji: ":lore:"}
					_, _, err := l.SlackAPI.PostMessage(channel, out, params)
					if err != nil {
						fmt.Printf("%s\n", err)
						return
					}
				}
			}

		case *slack.PresenceChangeEvent:
			// fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			// fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			// fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			// fmt.Printf("Invalid credentials")
			return

		case *slack.ReactionAddedEvent:
			if ev.Reaction == "lore" {
				channel := ev.Item.Channel
				timestamp := ev.Item.Timestamp
				go l.HandleLoreReact(channel, timestamp)
			}

		default:
			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func NewLorebot(conf *Configuration) *Lorebot {
	lorebot := new(Lorebot)
	lorebot.Conf = conf
	lorebot.Pg = NewPostgresClient(lorebot.Conf.PGHost, lorebot.Conf.PGPort,
		lorebot.Conf.PGUser, lorebot.Conf.PGPassword, lorebot.Conf.PGDbname)
	lorebot.SlackAPI = slack.New(lorebot.Conf.Token)
	lorebot.SlackAPI.SetDebug(true)

	// TODO: Move to conf file
	lorebot.LorebotID = "U9VEHRJHH"
	return lorebot
}
