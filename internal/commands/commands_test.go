package commands

import (
	"testing"
	"time"
)

func TestSubmitStatusUpdate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     SubmitStatusUpdate
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: SubmitStatusUpdate{
				TeamID:    "team-123",
				Content:   "Working on feature X",
				Author:    "Alice",
				SlackUser: "alice",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing team_id",
			cmd: SubmitStatusUpdate{
				Content:   "Working on feature X",
				Author:    "Alice",
				SlackUser: "alice",
			},
			wantErr: true,
			errMsg:  "team_id is required",
		},
		{
			name: "missing content",
			cmd: SubmitStatusUpdate{
				TeamID:    "team-123",
				Author:    "Alice",
				SlackUser: "alice",
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "missing author",
			cmd: SubmitStatusUpdate{
				TeamID:    "team-123",
				Content:   "Working on feature X",
				SlackUser: "alice",
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "missing slack_user",
			cmd: SubmitStatusUpdate{
				TeamID:  "team-123",
				Content: "Working on feature X",
				Author:  "Alice",
			},
			wantErr: true,
			errMsg:  "slack_user is required",
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
	tests := []struct {
		name    string
		cmd     RegisterTeam
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: RegisterTeam{
				Name:         "Engineering",
				SlackChannel: "#engineering",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cmd: RegisterTeam{
				SlackChannel: "#engineering",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing slack_channel",
			cmd: RegisterTeam{
				Name: "Engineering",
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
	tests := []struct {
		name    string
		cmd     UpdateTeam
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid command",
			cmd: UpdateTeam{
				TeamID:       "team-123",
				Name:         "Engineering",
				SlackChannel: "#engineering",
			},
			wantErr: false,
		},
		{
			name: "missing team_id",
			cmd: UpdateTeam{
				Name:         "Engineering",
				SlackChannel: "#engineering",
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
