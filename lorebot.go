package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/nlopes/slack"
)

type Lorebot struct {
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
				l.Pg.UpvoteLore(message.User, message.Text)
				return
			}
			fmt.Println("User: " + message.User + " + lore id: " + l.LorebotID)
			// Can't lore the lorebot
			if message.User == "" {
				fmt.Println("Ingoring self lore")
				return
			}
			l.Pg.InsertLore(message.User, message.Text)
			msg := Message{ChannelID: channelId, Content: "Lore added: <@" + message.User + ">: " + message.Text}
			l.MessageQueue <- msg
			return
		}
	}
}

func (l *Lorebot) HandleMessage(ev *slack.MessageEvent) {
	spl := strings.Split(ev.Text, " ")
	if len(spl) < 2 {
		return
	}
	userID := parseUserID(spl[0])
	if userID == l.LorebotID {
		cmd := spl[1]
		var lores []Lore = nil
		switch cmd {
		case "help":
			out := "Usage: @lorebot <help | random | recent | search <query> | top | user <username>> | highscores"
			msg := Message{ChannelID: ev.Channel, Content: out}
			fmt.Println("Trying to write message: " + out)
			l.MessageQueue <- msg
			return
		case "highscores":
			// scores := l.Pg.Highscores()
			out := "```"
			out += "Most lored messages: \n"
			out += "Most points: "
			/*
				for _, score := range scores {
					// TODO: Build scores

				}
			*/
			out += "```"
			msg := Message{ChannelID: ev.Channel, Content: out}
			l.MessageQueue <- msg
			return
		case "random":
			lores = l.Pg.RandomLore()
		case "recent":
			lores = l.Pg.RecentLore()
		case "user":
			if len(spl) != 3 {
				return
			}
			parsedUser := parseUserID(spl[2])
			lores = l.Pg.LoreForUser(parsedUser)
		case "search":
			if len(spl) < 3 {
				return
			}
			query := strings.Join(spl[2:], " ")
			lores = l.Pg.SearchLore(query)
		case "top":
			lores = l.Pg.TopLore()
		}

		// If we have some lores to share, send them to slack
		if lores != nil {
			out := ""
			for _, lore := range lores {
				out += "<@" + lore.userID + ">" + ": " + lore.Message + " (" + strconv.Itoa(lore.Score) + ")" + "\n"
			}
			msg := Message{ChannelID: ev.Channel, Content: out}
			l.MessageQueue <- msg
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
	go l.MessageWorker()
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

func parseUserID(unparsed string) string {
	userID := strings.Replace(unparsed, "<", "", 1)
	userID = strings.Replace(userID, ">", "", 1)
	userID = strings.Replace(userID, "@", "", 1)
	return userID
}

func NewLorebot(conf *Configuration) *Lorebot {
	bot := Lorebot{
		Pg:           NewPostgresClient(conf),
		SlackAPI:     slack.New(conf.Token),
		LorebotID:    conf.BotID,
		MessageQueue: make(chan Message, 1000),
	}
	bot.SlackAPI.SetDebug(true)

	return &bot
}
