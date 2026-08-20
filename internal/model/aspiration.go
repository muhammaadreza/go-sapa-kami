package model

import "time"

// Type Enums
type TypeEnum string
type StatusEnum string
type VoteEnum string

// Constants for enums
const (
	TypeCriticism  TypeEnum = "criticism"
	TypeInnovation TypeEnum = "innovation"

	StatusOpen     StatusEnum = "open"
	StatusClose    StatusEnum = "close"
	StatusAccepted StatusEnum = "accepted"
	StatusRejected StatusEnum = "rejected"

	VoteAccept 	   VoteEnum = "accept"
	VoteReject 	   VoteEnum = "reject"
)

// Models
// Aspiration Model
type Aspiration struct {
	ID          	uint         `gorm:"primaryKey"`
	SenderID    	uint         `gorm:"<-:create;not null"`
	Subject     	string       `gorm:"<-:create;not null"`
	Body        	string       `gorm:"<-:create;not null"`
	Attachments 	Attachments  `gorm:"type:jsonb;default:'[]'"`
	Type        	TypeEnum     `gorm:"<-:create;type:type_enum;not null"`
	Status      	StatusEnum   `gorm:"<-:update;type:status_enum;default:'open';not null"`
	CreatedAt   	time.Time    `gorm:"<-:create"`
	ExpiredAt       *time.Time   `gorm:"<-:create;default:null"`

	AcceptVoteCount int64 		 `gorm:"->;-:migration"`
	RejectVoteCount int64 		 `gorm:"->;-:migration"`
	DiscussionCount int64 		 `gorm:"->;-:migration"`

	Sender     	 	User         `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	Votes       	[]Vote       `gorm:"foreignKey:AspirationID"`
	Discussions 	[]Discussion `gorm:"foreignKey:AspirationID"`
}

// Vote Model
type Vote struct {
	ID           uint      `gorm:"primaryKey"`
	AspirationID uint      `gorm:"<-:create;not null;uniqueIndex:idx_aspiration_voter"`
	VoterID      uint      `gorm:"<-:create;not null;uniqueIndex:idx_aspiration_voter"`
	Value        VoteEnum  `gorm:"type:vote_enum;not null"`

	CreatedAt 	time.Time  `gorm:"<-:create"`
	UpdatedAt 	time.Time  `gorm:"autoUpdateTime"`

	Voter       User       `gorm:"foreignKey:VoterID;constraint:OnDelete:CASCADE"`
	Aspiration 	Aspiration `gorm:"foreignKey:AspirationID;constraint:OnDelete:CASCADE"`
}

// Discussion Model
type Discussion struct {
	ID           uint        `gorm:"primaryKey"`
	AspirationID uint        `gorm:"<-:create;not null"`
	SenderID     uint        `gorm:"<-:create;not null"`
	Message      string      `gorm:"<-:create;not null"`
	Attachments  Attachments `gorm:"<-:create;type:jsonb;default:'[]'"`
	CreatedAt    time.Time   `gorm:"<-:create"`

	Sender     	 User        `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	Aspiration 	 Aspiration  `gorm:"foreignKey:AspirationID;constraint:OnDelete:CASCADE"`
}

// Filter Structs
type AspirationFilter struct {
	Query     *string
	Type      *TypeEnum
	Status    *StatusEnum
	StartDate *string
	EndDate   *string
	SenderID  *uint
	Limit     int
	Page      int

	SortKey   *string
	SortOrder *string
}
