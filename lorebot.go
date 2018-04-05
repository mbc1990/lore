package main

import "fmt"
import "log"
import "strings"
import "github.com/nlopes/slack"

type Lorebot struct {
	Conf         *Configuration
	Pg           *PostgresClient
	SlackAPI     *slack.Client
	LorebotID    string
	MessageQueue chan Message
}

type Message struct {
	ChannelID string
	Content   string
}

func (l *Lorebot) MessageWorker() {
	params := slack.PostMessageParameters{Username: "Lorebot", IconEmoji: ":lore:"}
	for msg := range l.MessageQueue {
		l.SlackAPI.PostMessage(msg.ChannelID, msg.Content, params)
	}
}

// channel + timestamp is a UUID for slack.
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
			if l.Pg.LoreExists(message.Text, message.User) {
				// TODO: Upvote lore
				return
			}
			fmt.Println("User: " + message.User + " + lore id: " + l.LorebotID)
			// Can't lore the lorebot
			if message.User == "" {
				fmt.Println("Ingoring self lore")
				return
			}
			l.Pg.InsertLore(message.User, message.Text)
			params := slack.PostMessageParameters{Username: "Lorebot", IconEmoji: ":lore:"}
			go l.SlackAPI.PostMessage(channelId, "Lore added: <@"+message.User+">: "+message.Text, params)
			return
		}
	}
}

func parseUserID(unparsed string) string {
	userId := strings.Replace(unparsed, "<", "", 1)
	userId = strings.Replace(userId, ">", "", 1)
	userId = strings.Replace(userId, "@", "", 1)
	return userId
}

func (l *Lorebot) HandleMessage(ev *slack.MessageEvent) {
	fmt.Println("event: ", ev)
	channel := ev.Channel
	spl := strings.Split(ev.Text, " ")
	userId := parseUserID(spl[0])
	cmd := spl[1]
	if userId == l.LorebotID {
		// TODO: Parse commands + arguments
		switch cmd {
		case "help":
			out := "Usage: @lorebot <help | recent | user <username>>"
			msg := &Message{ChannelID: channel, Content: out}
			fmt.Println("Trying to write message: " + out)
			l.MessageQueue <- *msg
		case "recent":
			out := ""
			recent := l.Pg.RecentLore()
			for _, lore := range recent {
				out += "<@" + lore.UserID + ">" + ": " + lore.Message + "\n"
			}
			msg := &Message{ChannelID: channel, Content: out}
			l.MessageQueue <- *msg
		case "user":
			// Make sure there are the right number of arguments
			if len(spl) != 3 {
				return
			}
			parsedUser := parseUserID(spl[2])
			out := ""
			lores := l.Pg.LoreForUser(parsedUser)
			for _, lore := range lores {
				out += "<@" + lore.UserID + ">" + ": " + lore.Message + "\n"
			}
			msg := &Message{ChannelID: channel, Content: out}
			l.MessageQueue <- *msg
		}
	}
}

func (l *Lorebot) HandleReaction(ev *slack.ReactionAddedEvent) {
	if ev.Reaction == "lore" {
		channel := ev.Item.Channel
		timestamp := ev.Item.Timestamp
		l.HandleLoreReact(channel, timestamp)
	}
}

func (l *Lorebot) Start() {
	fmt.Println("Starting...")

	go l.MessageWorker()

	// TODO: This is a race condition
	rtm := l.SlackAPI.NewRTM()
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			go l.HandleMessage(ev)
		case *slack.InvalidAuthEvent:
			log.Fatal("Invalid credentials")
		case *slack.ReactionAddedEvent:
			go l.HandleReaction(ev)
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

	lorebot.LorebotID = conf.BotID
	lorebot.MessageQueue = make(chan Message, 1000)
	return lorebot
}
