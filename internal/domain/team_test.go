package domain

import (
	"testing"
	"time"
)

func TestNewTeam(t *testing.T) {
	validID, _ := NewTeamID("team-123")
	validName, _ := NewTeamName("Engineering")
	validChannel, _ := NewSlackChannel("C12345")
	emptyID := TeamID{}
	emptyName := TeamName{}
	emptyChannel := SlackChannel{}

	tests := []struct {
		name         string
		id           TeamID
		teamName     TeamName
		slackChannel SlackChannel
		wantErr      bool
	}{
		{"valid team", validID, validName, validChannel, false},
		{"empty ID", emptyID, validName, validChannel, true},
		{"empty name", validID, emptyName, validChannel, true},
		{"empty channel", validID, validName, emptyChannel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := NewTeam(tt.id, tt.teamName, tt.slackChannel)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTeam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if team.ID().String() != tt.id.String() {
					t.Errorf("team.ID() = %v, want %v", team.ID(), tt.id)
				}
				if !team.IsRegistered() {
					t.Errorf("team.IsRegistered() = false, want true")
				}
			}
		})
	}
}

func TestTeam_Rename(t *testing.T) {
	tests := []struct {
		name    string
		newName TeamName
		wantErr bool
	}{
		{"valid rename", mustTeamName("Product"), false},
		{"same name", mustTeamName("Engineering"), true},
		{"empty name", TeamName{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamID, _ := NewTeamID("team-123")
			oldName, _ := NewTeamName("Engineering")
			channel, _ := NewSlackChannel("C12345")
			team, _ := NewTeam(teamID, oldName, channel)

			err := team.Rename(tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rename() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && team.Name().String() != tt.newName.String() {
				t.Errorf("team.Name() = %v, want %v", team.Name(), tt.newName)
			}
		})
	}
}

func mustTeamName(s string) TeamName {
	name, err := NewTeamName(s)
	if err != nil {
		panic(err)
	}
	return name
}

func TestNewUpdate(t *testing.T) {
	validID, _ := NewUpdateID("update-123")
	validTeamID, _ := NewTeamID("team-123")
	validContent, _ := NewUpdateContent("Working on feature X")
	validAuthor, _ := NewAuthor("john.doe")
	validSlackUser, _ := NewSlackUserID("U12345")
	validTime := time.Now()

	emptyID := UpdateID{}
	emptyTeamID := TeamID{}
	emptyContent := UpdateContent{}
	emptyAuthor := Author{}
	emptySlackUser := SlackUserID{}

	tests := []struct {
		name      string
		id        UpdateID
		teamID    TeamID
		content   UpdateContent
		author    Author
		slackUser SlackUserID
		timestamp time.Time
		wantErr   bool
	}{
		{"valid update", validID, validTeamID, validContent, validAuthor, validSlackUser, validTime, false},
		{"empty ID", emptyID, validTeamID, validContent, validAuthor, validSlackUser, validTime, true},
		{"empty team ID", validID, emptyTeamID, validContent, validAuthor, validSlackUser, validTime, true},
		{"empty content", validID, validTeamID, emptyContent, validAuthor, validSlackUser, validTime, true},
		{"empty author", validID, validTeamID, validContent, emptyAuthor, validSlackUser, validTime, true},
		{"empty slack user", validID, validTeamID, validContent, validAuthor, emptySlackUser, validTime, true},
		{"zero timestamp", validID, validTeamID, validContent, validAuthor, validSlackUser, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update, err := NewUpdate(tt.id, tt.teamID, tt.content, tt.author, tt.slackUser, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if update.ID().String() != tt.id.String() {
					t.Errorf("update.ID() = %v, want %v", update.ID(), tt.id)
				}
				if update.Content().String() != tt.content.String() {
					t.Errorf("update.Content() = %v, want %v", update.Content(), tt.content)
				}
			}
		})
	}
}
