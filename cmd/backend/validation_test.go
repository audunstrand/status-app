package main

import (
	"testing"
)

func TestSubmitStatusUpdateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     SubmitStatusUpdateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: SubmitStatusUpdateRequest{
				Content: "Fixed the bug",
				Author:  "John Doe",
			},
			wantErr: false,
		},
		{
			name: "missing content",
			req: SubmitStatusUpdateRequest{
				Author: "John Doe",
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "content too long",
			req: SubmitStatusUpdateRequest{
				Content: string(make([]byte, 501)),
				Author:  "John Doe",
			},
			wantErr: true,
			errMsg:  "content must be 500 characters or less",
		},
		{
			name: "missing author",
			req: SubmitStatusUpdateRequest{
				Content: "Fixed the bug",
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "exactly 500 characters - valid",
			req: SubmitStatusUpdateRequest{
				Content: string(make([]byte, 500)),
				Author:  "John Doe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
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

func TestRegisterTeamRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterTeamRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: RegisterTeamRequest{
				Name:         "Engineering",
				SlackChannel: "#engineering",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			req: RegisterTeamRequest{
				SlackChannel: "#engineering",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing slack channel",
			req: RegisterTeamRequest{
				Name:         "Engineering",
			},
			wantErr: true,
			errMsg:  "slack_channel is required",
		},
		{
			name: "poll schedule optional",
			req: RegisterTeamRequest{
				Name:         "Engineering",
				SlackChannel: "#engineering",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
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
