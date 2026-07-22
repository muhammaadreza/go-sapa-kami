package model

import "time"

// Type Enums
type TypeEnum string

// Constants for Type enums
const (
	TypeCriticism  TypeEnum = "criticism"
	TypeInnovation TypeEnum = "innovation"
)

// Messages Model
type Message struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint		`gorm:"<-:create;not null"`
	Subject   string	`gorm:"<-:create;not null"`
	Body	  string    `gorm:"<-:create;not null"`
	Type      TypeEnum  `gorm:"<-:create;type:type_enum;not null"`
	CreatedAt time.Time `gorm:"<-:create"`

	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// DTOs
// Message Request
type MessageRequest struct {
	Type    TypeEnum `json:"type" binding:"required"`
	Subject string   `json:"subject" binding:"required"`
	Body    string   `json:"body" binding:"required"`
}

// Message Response
type MessageResponse struct {
	ID        uint 	        `json:"id"`
	Type      TypeEnum      `json:"type"`
	Subject   string        `json:"subject"`
	Body      string        `json:"body"`
	User      *UserResponse `json:"user,omitempty"`

	CreatedAt string        `json:"created_at"`
}

// Message Filter Request
type MessageFilterRequest struct {
	Query     *string   `form:"query" binding:"omitempty"`
	Type      *TypeEnum `form:"type" binding:"omitempty,oneof=criticism innovation"`
	StartDate *string   `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
    EndDate   *string   `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
	UserID    *uint     `form:"user_id"`
	Limit     int       `form:"limit" binding:"omitempty,min=1,max=100"`
	Page      int       `form:"page" binding:"omitempty,min=1"`

	SortKey   *string   `form:"sort_key" binding:"omitempty,oneof=id type user_name subject created_at"`
	SortOrder *string   `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// Mapping Functions
// Message Request to Message Model
func (request *MessageRequest) ToModel(userID uint) Message {
	return Message{
		UserID:  userID,
		Type: 	 request.Type,
		Subject: request.Subject,
		Body:    request.Body,
	}
}

// Message Model to Message Response
func (message *Message) ToResponse() MessageResponse {
	user := message.User.ToNested()
	if message.Type == TypeCriticism { user.Name = "anonymous" }
	return MessageResponse{
		ID: 	   message.ID,
		User: 	   &user,
		Type: 	   message.Type,
		Subject:   message.Subject,
		Body: 	   message.Body,
		CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Message Model to Message Nested
func (message *Message) ToNested() MessageResponse {
	return MessageResponse{
		ID: 	   message.ID,
		Type: 	   message.Type,
		Subject:   message.Subject,
		Body: 	   message.Body,
		CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List Mapping Functions
// Message Model List to Message Response List
func ToMessagesResponseList(messages []Message) []MessageResponse {
	response := make([]MessageResponse, len(messages))
	for i, message := range messages {
		response[i] = message.ToResponse()
	}

	return response
}
