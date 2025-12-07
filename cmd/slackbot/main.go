package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/yourusername/status-app/internal/config"
)

type SlackBot struct {
	cfg    *config.Config
	client *http.Client
}

func NewSlackBot(cfg *config.Config) *SlackBot {
	return &SlackBot{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bot := NewSlackBot(cfg)

	api := slack.New(
		cfg.SlackBotToken,
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(cfg.SlackSigningKey),
	)

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
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			log.Printf("Message: %v", ev.Text)
			// TODO: Parse message and send command to command service
			// bot.sendStatusUpdate(ev.Channel, ev.Text, ev.User)
		}
	}
}

func (bot *SlackBot) sendStatusUpdate(teamID, content, author string) error {
	payload := map[string]string{
		"team_id": teamID,
		"content": content,
		"author":  author,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", bot.cfg.CommandsURL+"/commands/submit-update", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if bot.cfg.APISecret != "" {
		req.Header.Set("Authorization", "Bearer "+bot.cfg.APISecret)
	}
	
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
