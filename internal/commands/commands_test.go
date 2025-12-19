package commands

import (
	"testing"
	"time"

	"github.com/yourusername/status-app/internal/domain"
)

func TestSubmitStatusUpdate_Validate(t *testing.T) {
	validTeamID, _ := domain.NewTeamID("team-123")
	validContent, _ := domain.NewUpdateContent("Working on feature X")
	validAuthor, _ := domain.NewAuthor("Alice")
	validSlackUser, _ := domain.NewSlackUserID("alice")

	tests := []struct {
		name    string
		cmd     SubmitStatusUpdate
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: SubmitStatusUpdate{
				TeamID:      validTeamID,
				ChannelName: "engineering",
				Content:     validContent,
				Author:      validAuthor,
				SlackUser:   validSlackUser,
				Timestamp:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing team_id",
			cmd: SubmitStatusUpdate{
				ChannelName: "engineering",
				Content:     validContent,
				Author:      validAuthor,
				SlackUser:   validSlackUser,
			},
			wantErr: true,
			errMsg:  "team_id is required",
		},
		{
			name: "missing content",
			cmd: SubmitStatusUpdate{
				TeamID:      validTeamID,
				ChannelName: "engineering",
				Author:      validAuthor,
				SlackUser:   validSlackUser,
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "missing author",
			cmd: SubmitStatusUpdate{
				TeamID:      validTeamID,
				ChannelName: "engineering",
				Content:     validContent,
				SlackUser:   validSlackUser,
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "missing slack_user",
			cmd: SubmitStatusUpdate{
				TeamID:      validTeamID,
				ChannelName: "engineering",
				Content:     validContent,
				Author:      validAuthor,
			},
			wantErr: true,
			errMsg:  "slack_user is required",
		},
		{
			name: "missing channel_name",
			cmd: SubmitStatusUpdate{
				TeamID:    validTeamID,
				Content:   validContent,
				Author:    validAuthor,
				SlackUser: validSlackUser,
			},
			wantErr: true,
			errMsg:  "channel_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestRegisterTeam_Validate(t *testing.T) {
	validName, _ := domain.NewTeamName("Engineering")
	validChannel, _ := domain.NewSlackChannel("#engineering")

	tests := []struct {
		name    string
		cmd     RegisterTeam
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: RegisterTeam{
				Name:         validName,
				SlackChannel: validChannel,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cmd: RegisterTeam{
				SlackChannel: validChannel,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing slack_channel",
			cmd: RegisterTeam{
				Name: validName,
			},
			wantErr: true,
			errMsg:  "slack_channel is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestUpdateTeam_Validate(t *testing.T) {
	validTeamID, _ := domain.NewTeamID("team-123")
	validName, _ := domain.NewTeamName("Engineering")
	validChannel, _ := domain.NewSlackChannel("#engineering")

	tests := []struct {
		name    string
		cmd     UpdateTeam
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: UpdateTeam{
				TeamID:       validTeamID,
				Name:         validName,
				SlackChannel: validChannel,
			},
			wantErr: false,
		},
		{
			name: "missing team_id",
			cmd: UpdateTeam{
				Name:         validName,
				SlackChannel: validChannel,
			},
			wantErr: true,
			errMsg:  "team_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
