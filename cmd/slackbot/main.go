package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/yourusername/status-app/internal/config"
)

type SlackBot struct {
	cfg       *config.Config
	client    *http.Client
	slackAPI  *slack.Client
}

func NewSlackBot(cfg *config.Config, slackAPI *slack.Client) *SlackBot {
	return &SlackBot{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		slackAPI: slackAPI,
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	api := slack.New(
		cfg.SlackBotToken,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(cfg.SlackSigningKey),
	)

	bot := NewSlackBot(cfg, api)

	client := socketmode.New(api)

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					continue
				}

				client.Ack(*evt.Request)
				bot.handleEvent(eventsAPIEvent)

			case socketmode.EventTypeSlashCommand:
				// Handle slash commands
				client.Ack(*evt.Request)

			case socketmode.EventTypeInteractive:
				// Handle interactive components
				client.Ack(*evt.Request)
			}
		}
	}()

	go func() {
		if err := client.Run(); err != nil {
			log.Fatalf("Slack client error: %v", err)
		}
	}()

	log.Println("Slack bot running")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}

func (bot *SlackBot) handleEvent(event slackevents.EventsAPIEvent) {
	ctx := context.Background()
	
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			log.Printf("Bot mentioned by user %s in channel %s: %s", ev.User, ev.Channel, ev.Text)
			
			channelID := ev.Channel
			channelName := bot.getChannelName(channelID)
			
			if err := bot.sendStatusUpdate(ctx, channelID, channelName, ev.Text, ev.User); err != nil {
				log.Printf("Failed to send status update: %v", err)
				bot.sendSlackMessage(ev.Channel, "❌ Failed to record your status update. Please try again.")
				return
			}
			
			log.Printf("Successfully submitted status update for team %s", channelID)
			bot.sendSlackMessage(ev.Channel, "✅ Status update recorded!")
			
		case *slackevents.MessageEvent:
			// Ignore bot messages and message subtypes (edits, deletes, etc)
			if ev.BotID != "" || ev.SubType != "" {
				return
			}
			
			log.Printf("Received message from user %s in channel %s: %s", ev.User, ev.Channel, ev.Text)
			
			channelID := ev.Channel
			channelName := bot.getChannelName(channelID)
			
			// Send status update to Commands service
			if err := bot.sendStatusUpdate(ctx, channelID, channelName, ev.Text, ev.User); err != nil {
				log.Printf("Failed to send status update: %v", err)
				bot.sendSlackMessage(ev.Channel, "❌ Failed to record your status update. Please try again.")
				return
			}
			
			log.Printf("Successfully submitted status update for team %s", channelID)
			bot.sendSlackMessage(ev.Channel, "✅ Status update recorded!")
		}
	}
}

func (bot *SlackBot) sendStatusUpdate(ctx context.Context, channelID, channelName, content, author string) error {
	payload := map[string]string{
		"content":      content,
		"author":       author,
		"channel_name": channelName,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	url := bot.cfg.CommandsURL + "/teams/" + channelID + "/updates"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bot.cfg.APISecret)
	
	resp, err := bot.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		log.Printf("Failed to submit update: status %d", resp.StatusCode)
	}
	
	return nil
}

func (bot *SlackBot) getChannelName(channelID string) string {
	info, err := bot.slackAPI.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		log.Printf("Failed to get channel info for %s: %v", channelID, err)
		return channelID
	}
	return info.Name
}

func (bot *SlackBot) sendSlackMessage(channel, message string) {
	_, _, err := bot.slackAPI.PostMessage(
		channel,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		log.Printf("Failed to send Slack message: %v", err)
	}
}
