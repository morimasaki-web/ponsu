package request

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreatedPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload CreatedPayload
		wantErr bool
	}{
		{
			name:    "valid payload",
			payload: CreatedPayload{Title: "Test Request"},
			wantErr: false,
		},
		{
			name:    "empty title",
			payload: CreatedPayload{Title: ""},
			wantErr: true,
		},
		{
			name:    "title too long",
			payload: CreatedPayload{Title: string(make([]byte, 201))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRejectedPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload RejectedPayload
		wantErr bool
	}{
		{
			name:    "valid payload",
			payload: RejectedPayload{Reason: "Not approved"},
			wantErr: false,
		},
		{
			name:    "empty reason",
			payload: RejectedPayload{Reason: ""},
			wantErr: true,
		},
		{
			name:    "reason too long",
			payload: RejectedPayload{Reason: string(make([]byte, 501))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReturnedPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload ReturnedPayload
		wantErr bool
	}{
		{
			name:    "valid payload",
			payload: ReturnedPayload{Reason: "Needs revision"},
			wantErr: false,
		},
		{
			name:    "empty reason",
			payload: ReturnedPayload{Reason: ""},
			wantErr: true,
		},
		{
			name:    "reason too long",
			payload: ReturnedPayload{Reason: string(make([]byte, 501))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequestCommentedPayload_Validate(t *testing.T) {
	validCommentID := uuid.New().String()
	validUserID := uuid.New().String()
	validTime := time.Now()

	tests := []struct {
		name    string
		payload RequestCommentedPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: RequestCommentedPayload{
				CommentID:   validCommentID,
				UserID:      validUserID,
				Content:     "This is a comment",
				CommentedAt: validTime,
			},
			wantErr: false,
		},
		{
			name: "empty comment_id",
			payload: RequestCommentedPayload{
				CommentID:   "",
				UserID:      validUserID,
				Content:     "Comment",
				CommentedAt: validTime,
			},
			wantErr: true,
		},
		{
			name: "invalid comment_id UUID",
			payload: RequestCommentedPayload{
				CommentID:   "not-a-uuid",
				UserID:      validUserID,
				Content:     "Comment",
				CommentedAt: validTime,
			},
			wantErr: true,
		},
		{
			name: "empty user_id",
			payload: RequestCommentedPayload{
				CommentID:   validCommentID,
				UserID:      "",
				Content:     "Comment",
				CommentedAt: validTime,
			},
			wantErr: true,
		},
		{
			name: "invalid user_id UUID",
			payload: RequestCommentedPayload{
				CommentID:   validCommentID,
				UserID:      "not-a-uuid",
				Content:     "Comment",
				CommentedAt: validTime,
			},
			wantErr: true,
		},
		{
			name: "empty content",
			payload: RequestCommentedPayload{
				CommentID:   validCommentID,
				UserID:      validUserID,
				Content:     "",
				CommentedAt: validTime,
			},
			wantErr: true,
		},
		{
			name: "content too long",
			payload: RequestCommentedPayload{
				CommentID:   validCommentID,
				UserID:      validUserID,
				Content:     string(make([]byte, 1001)),
				CommentedAt: validTime,
			},
			wantErr: true,
		},
		{
			name: "zero commented_at",
			payload: RequestCommentedPayload{
				CommentID:   validCommentID,
				UserID:      validUserID,
				Content:     "Comment",
				CommentedAt: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
