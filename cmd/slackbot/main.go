package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/yourusername/status-app/internal/config"
)

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
				handleEvent(eventsAPIEvent)

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

func handleEvent(event slackevents.EventsAPIEvent) {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			log.Printf("Message: %v", ev.Text)
			// TODO: Parse message and send command to command service
		}
	}
}
