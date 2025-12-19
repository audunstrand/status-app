package domain

import (
	"strings"
	"testing"
)

func TestNewTeamID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid team ID", "team-123", false},
		{"empty string", "", true},
		{"too long", strings.Repeat("a", 101), true},
		{"max length", strings.Repeat("a", 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamID, err := NewTeamID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTeamID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && teamID.String() != tt.input {
				t.Errorf("NewTeamID().String() = %v, want %v", teamID.String(), tt.input)
			}
		})
	}
}

func TestNewTeamName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid name", "Engineering", "Engineering", false},
		{"with whitespace", "  Engineering  ", "Engineering", false},
		{"empty string", "", "", true},
		{"only whitespace", "   ", "", true},
		{"too long", strings.Repeat("a", 101), "", true},
		{"max length", strings.Repeat("a", 100), strings.Repeat("a", 100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamName, err := NewTeamName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTeamName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && teamName.String() != tt.want {
				t.Errorf("NewTeamName().String() = %v, want %v", teamName.String(), tt.want)
			}
		})
	}
}

func TestNewUpdateContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid content", "Working on feature X", false},
		{"empty string", "", true},
		{"exactly 500 chars", strings.Repeat("a", 500), false},
		{"501 chars", strings.Repeat("a", 501), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewUpdateContent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUpdateContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && content.String() != tt.input {
				t.Errorf("NewUpdateContent().String() = %v, want %v", content.String(), tt.input)
			}
		})
	}
}

func TestNewAuthor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid author", "john.doe", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			author, err := NewAuthor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && author.String() != tt.input {
				t.Errorf("NewAuthor().String() = %v, want %v", author.String(), tt.input)
			}
		})
	}
}

func TestNewSlackChannel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid channel", "C12345", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := NewSlackChannel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlackChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && channel.String() != tt.input {
				t.Errorf("NewSlackChannel().String() = %v, want %v", channel.String(), tt.input)
			}
		})
	}
}

func TestNewSlackUserID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid user ID", "U12345", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := NewSlackUserID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlackUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && userID.String() != tt.input {
				t.Errorf("NewSlackUserID().String() = %v, want %v", userID.String(), tt.input)
			}
		})
	}
}
