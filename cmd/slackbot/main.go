package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					continue
				}

				client.Ack(*evt.Request)
				bot.handleSlashCommand(cmd)

			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					continue
				}

				client.Ack(*evt.Request)
				bot.handleInteractive(callback)
			}
		}
	}()

	go func() {
		if err := client.Run(); err != nil {
			log.Fatalf("Slack client error: %v", err)
		}
	}()

	// Start HTTP server for metrics
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "slackbot",
		})
	})
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8081", // Different port from backend
		Handler: mux,
	}

	go func() {
		log.Println("Metrics server running on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	log.Println("Slack bot running")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	
	// Shutdown metrics server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func (bot *SlackBot) handleEvent(event slackevents.EventsAPIEvent) {
	ctx := context.Background()
	
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			slackMessagesReceivedTotal.WithLabelValues("mention").Inc()
			log.Printf("Bot mentioned by user %s in channel %s: %s", ev.User, ev.Channel, ev.Text)
			
			channelID := ev.Channel
			channelName := bot.getChannelName(channelID)
			
			if err := bot.sendStatusUpdate(ctx, channelID, channelName, ev.Text, ev.User); err != nil {
				slackbotErrorsTotal.WithLabelValues("backend_error").Inc()
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
			
			slackMessagesReceivedTotal.WithLabelValues("direct_message").Inc()
			log.Printf("Received message from user %s in channel %s: %s", ev.User, ev.Channel, ev.Text)
			
			channelID := ev.Channel
			channelName := bot.getChannelName(channelID)
			
			// Send status update to Commands service
			if err := bot.sendStatusUpdate(ctx, channelID, channelName, ev.Text, ev.User); err != nil {
				slackbotErrorsTotal.WithLabelValues("backend_error").Inc()
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
		backendAPICallsTotal.WithLabelValues("submit_update", "error").Inc()
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		backendAPICallsTotal.WithLabelValues("submit_update", "error").Inc()
		log.Printf("Failed to submit update: status %d", resp.StatusCode)
		return fmt.Errorf("backend returned status %d", resp.StatusCode)
	}
	
	backendAPICallsTotal.WithLabelValues("submit_update", "success").Inc()
	return nil
}

func (bot *SlackBot) getChannelName(channelID string) string {
	info, err := bot.slackAPI.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		slackAPICallsTotal.WithLabelValues("get_conversation_info", "error").Inc()
		log.Printf("Failed to get channel info for %s: %v", channelID, err)
		return channelID
	}
	slackAPICallsTotal.WithLabelValues("get_conversation_info", "success").Inc()
	return info.Name
}

func (bot *SlackBot) sendSlackMessage(channel, message string) {
	_, _, err := bot.slackAPI.PostMessage(
		channel,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		slackAPICallsTotal.WithLabelValues("post_message", "error").Inc()
		log.Printf("Failed to send Slack message: %v", err)
		return
	}
	slackAPICallsTotal.WithLabelValues("post_message", "success").Inc()
	slackMessagesSentTotal.Inc()
}

func (bot *SlackBot) handleSlashCommand(cmd slack.SlashCommand) {
	slackMessagesReceivedTotal.WithLabelValues("slash_command").Inc()
	slackCommandsHandledTotal.WithLabelValues(cmd.Command).Inc()
	log.Printf("Received slash command: %s from user %s in channel %s", cmd.Command, cmd.UserID, cmd.ChannelID)

	switch cmd.Command {
	case "/set-team-name":
		bot.openTeamNameModal(cmd)
	case "/updates":
		bot.showTeamUpdates(cmd)
	default:
		bot.slackAPI.PostEphemeral(
			cmd.ChannelID,
			cmd.UserID,
			slack.MsgOptionText("Unknown command", false),
		)
	}
}

func (bot *SlackBot) handleInteractive(callback slack.InteractionCallback) {
	log.Printf("Received interactive callback: %s", callback.Type)

	switch callback.Type {
	case slack.InteractionTypeViewSubmission:
		if callback.View.CallbackID == "set_team_name" {
			bot.handleTeamNameSubmission(callback)
		}
	}
}

func (bot *SlackBot) openTeamNameModal(cmd slack.SlashCommand) {
	modalRequest := slack.ModalViewRequest{
		Type: slack.VTModal,
		Title: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Set Team Name",
		},
		Close: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Cancel",
		},
		Submit: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Save",
		},
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewInputBlock(
					"team_name_block",
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Team Name",
					},
					nil,
					&slack.PlainTextInputBlockElement{
						Type:        slack.METPlainTextInput,
						ActionID:    "team_name_input",
						Placeholder: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Enter team name"},
					},
				),
			},
		},
		CallbackID: "set_team_name",
		PrivateMetadata: cmd.ChannelID,
	}

	_, err := bot.slackAPI.OpenView(cmd.TriggerID, modalRequest)
	if err != nil {
		log.Printf("Failed to open modal: %v", err)
	}
}

func (bot *SlackBot) handleTeamNameSubmission(callback slack.InteractionCallback) {
	channelID := callback.View.PrivateMetadata
	teamName := callback.View.State.Values["team_name_block"]["team_name_input"].Value

	ctx := context.Background()
	if err := bot.updateTeamName(ctx, channelID, teamName); err != nil {
		log.Printf("Failed to update team name: %v", err)
		bot.sendSlackMessage(channelID, "❌ Failed to update team name. Please try again.")
		return
	}

	log.Printf("Successfully updated team name for channel %s to '%s'", channelID, teamName)
	bot.sendSlackMessage(channelID, fmt.Sprintf("✅ Team name updated to '%s'", teamName))
}

func (bot *SlackBot) updateTeamName(ctx context.Context, channelID, teamName string) error {
	payload := map[string]string{
		"name": teamName,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := bot.cfg.CommandsURL + "/teams/" + channelID
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to update team name: status %d", resp.StatusCode)
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

func (bot *SlackBot) showTeamUpdates(cmd slack.SlashCommand) {
teamID := cmd.ChannelID

// Fetch updates from backend API
url := fmt.Sprintf("%s/teams/%s/updates?limit=10", bot.cfg.CommandsURL, teamID)
req, err := http.NewRequest("GET", url, nil)
if err != nil {
log.Printf("Failed to create request: %v", err)
bot.slackAPI.PostEphemeral(cmd.ChannelID, cmd.UserID,
slack.MsgOptionText("❌ Failed to fetch updates", false))
return
}

req.Header.Set("Authorization", "Bearer "+bot.cfg.APISecret)

resp, err := bot.client.Do(req)
if err != nil {
log.Printf("Failed to fetch updates: %v", err)
bot.slackAPI.PostEphemeral(cmd.ChannelID, cmd.UserID,
slack.MsgOptionText("❌ Failed to fetch updates", false))
return
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
log.Printf("Backend returned status %d", resp.StatusCode)
bot.slackAPI.PostEphemeral(cmd.ChannelID, cmd.UserID,
slack.MsgOptionText("❌ Failed to fetch updates", false))
return
}

var updates []struct {
UpdateID  string    `json:"update_id"`
Content   string    `json:"content"`
Author    string    `json:"author"`
CreatedAt time.Time `json:"created_at"`
}

if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
log.Printf("Failed to decode response: %v", err)
bot.slackAPI.PostEphemeral(cmd.ChannelID, cmd.UserID,
slack.MsgOptionText("❌ Failed to parse updates", false))
return
}

// Format and send response
if len(updates) == 0 {
bot.slackAPI.PostEphemeral(cmd.ChannelID, cmd.UserID,
slack.MsgOptionText("📝 No updates yet for this team", false))
return
}

message := "📝 *Recent Updates*\n\n"
for _, update := range updates {
message += fmt.Sprintf("• %s - _%s_\n", update.Content, update.CreatedAt.Format("Jan 02, 15:04"))
}

bot.slackAPI.PostEphemeral(cmd.ChannelID, cmd.UserID,
slack.MsgOptionText(message, false))
}
