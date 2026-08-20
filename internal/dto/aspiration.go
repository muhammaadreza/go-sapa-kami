package dto

import (
	"mime/multipart"
	"time"

	"go-sapa-kami/internal/model"
)

// Aspiration Request
type AspirationRequest struct {
	Type    model.TypeEnum          `form:"type" binding:"required,oneof=criticism innovation"`
	Subject string                  `form:"subject" binding:"required"`
	Body    string                  `form:"body" binding:"required"`
	Files   []*multipart.FileHeader `form:"files" binding:"omitempty"`
}

// Vote Request
type VoteRequest struct {
	Value model.VoteEnum `json:"value" binding:"required,oneof=accept reject"`
}

// Discussion Request
type DiscussionRequest struct {
	Message string                  `form:"message" binding:"required"`
	Files   []*multipart.FileHeader `form:"files" binding:"omitempty"`
}

// Aspiration Response
type AspirationResponse struct {
	ID              uint                 `json:"id"`
	Type            model.TypeEnum       `json:"type"`
	Status          model.StatusEnum     `json:"status"`
	Subject         string               `json:"subject"`
	Body            string               `json:"body"`
	Attachments     model.Attachments    `json:"attachments"`
	Sender          *UserResponse        `json:"sender,omitempty"`
	VoteCount       VoteCount            `json:"vote_count"`
	DiscussionCount int64                `json:"discussion_count"`
	Discussions     []DiscussionResponse `json:"discussions,omitempty"`
	UserVote        *model.VoteEnum      `json:"user_vote,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	ExpiredAt       *time.Time           `json:"expired_at,omitempty"`
}

// Vote Response
type VoteResponse struct {
	ID        uint           `json:"id"`
	Voter     *UserResponse  `json:"voter,omitempty"`
	Value     model.VoteEnum `json:"value"`
	CreatedAt time.Time      `json:"created_at"`
}

// Discussion Response
type DiscussionResponse struct {
	ID          uint              `json:"id"`
	Sender      *UserResponse     `json:"sender,omitempty"`
	Message     string            `json:"message"`
	Attachments model.Attachments `json:"attachments,omitempty"`
	IsYou       bool              `json:"is_you"`
	IsSender    bool              `json:"is_sender"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Vote Count Response
type VoteCount struct {
	Accept      int64 `json:"accept"`
	Reject      int64 `json:"reject"`
	TotalAdmins int64 `json:"total_admins"`
	Threshold   int64 `json:"threshold"`
}

// Aspiration Filter Request
type AspirationFilterRequest struct {
	Query     *string           `form:"query" binding:"omitempty"`
	Type      *model.TypeEnum   `form:"type" binding:"omitempty,oneof=criticism innovation"`
	Status    *model.StatusEnum `form:"status" binding:"omitempty,oneof=open close accepted rejected"`
	StartDate *string           `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate   *string           `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
	SenderID  *uint             `form:"sender_id"`
	Limit     int               `form:"limit" binding:"omitempty,min=1,max=100"`
	Page      int               `form:"page" binding:"omitempty,min=1"`

	SortKey   *string           `form:"sort_key" binding:"omitempty,oneof=id type sender_name subject created_at"`
	SortOrder *string           `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

