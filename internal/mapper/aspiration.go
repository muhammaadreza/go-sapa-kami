package mapper

import (
	"time"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/model"
)

// Aspiration Request to Aspiration Model
func ToAspirationModel(senderID uint, request *dto.AspirationRequest, attachments model.Attachments) model.Aspiration {
	var expiredAt *time.Time
	if request.Type == model.TypeInnovation {
		expTime := time.Now().AddDate(0, 0, 14)
		expMidnight := time.Date(expTime.Year(), expTime.Month(), expTime.Day(), 0, 0, 0, 0, expTime.Location())
		expiredAt = &expMidnight
	}

	return model.Aspiration{
		SenderID:    senderID,
		Type:        request.Type,
		Subject:     request.Subject,
		Body:        request.Body,
		Attachments: attachments,
		ExpiredAt:   expiredAt,
	}
}

// Vote Request to Vote Model
func ToVoteModel(aspirationID, voterID uint, request *dto.VoteRequest) model.Vote {
	return model.Vote{
		AspirationID: aspirationID,
		VoterID:      voterID,
		Value:        request.Value,
	}
}

// Discussion Request to Discussion Model
func ToDiscussionModel(aspirationID, senderID uint, request *dto.DiscussionRequest, attachments model.Attachments) model.Discussion {
	return model.Discussion{
		AspirationID: aspirationID,
		SenderID:     senderID,
		Message:      request.Message,
		Attachments:  attachments,
	}
}

// Aspiration Model to Aspiration Response
func ToAspirationResponse(aspiration *model.Aspiration, sessionID uint, sessionRole model.RoleEnum) dto.AspirationResponse {
	var sender *dto.UserResponse
	if aspiration.Type == model.TypeInnovation || aspiration.SenderID == sessionID {
		user := ToUserNested(&aspiration.Sender)
		sender = &user
	}

	voteCount := dto.VoteCount{
		Accept: aspiration.AcceptVoteCount,
		Reject: aspiration.RejectVoteCount,
	}

	var discussions []dto.DiscussionResponse
	for _, discussion := range aspiration.Discussions {
		discussions = append(discussions, ToDiscussionResponse(aspiration, sessionID, sessionRole, &discussion))
	}

	var userVote *model.VoteEnum
	for _, vote := range aspiration.Votes {
		if vote.VoterID == sessionID {
			vVal := vote.Value
			userVote = &vVal
			break
		}
	}

	var expiredAt *string
	if aspiration.ExpiredAt != nil {
		strExpired := aspiration.ExpiredAt.Format("2006-01-02 15:04:05")
		expiredAt = &strExpired
	}

	return dto.AspirationResponse{
		ID:              aspiration.ID,
		Type:            aspiration.Type,
		Status:          aspiration.Status,
		Subject:         aspiration.Subject,
		Body:            aspiration.Body,
		Attachments:     aspiration.Attachments,
		Sender:          sender,
		VoteCount:       voteCount,
		Discussions:     discussions,
		DiscussionCount: aspiration.DiscussionCount,
		UserVote:        userVote,
		CreatedAt:       aspiration.CreatedAt.Format("2006-01-02 15:04:05"),
		ExpiredAt:       expiredAt,
	}
}

// Vote Model to Vote Response
func ToVoteResponse(vote *model.Vote) dto.VoteResponse {
	voter := ToUserNested(&vote.Voter)

	return dto.VoteResponse{
		ID:        vote.ID,
		Voter:     &voter,
		Value:     vote.Value,
		CreatedAt: vote.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Discussion Model to Discussion Response
func ToDiscussionResponse(aspiration *model.Aspiration, sessionID uint, sessionRole model.RoleEnum, discussion *model.Discussion) dto.DiscussionResponse {
	isYou := discussion.SenderID == sessionID
	isSender := discussion.SenderID == aspiration.SenderID

	var sender *dto.UserResponse
	if canViewDiscussionSender(aspiration, sessionID, sessionRole, discussion) {
		user := ToUserNested(&discussion.Sender)
		sender = &user
	} else {
		maskID := aspiration.Type == model.TypeCriticism && discussion.SenderID == aspiration.SenderID
		user := ToUserAnonymous(&discussion.Sender, maskID)
		sender = &user
	}

	return dto.DiscussionResponse{
		ID:          discussion.ID,
		Sender:      sender,
		Message:     discussion.Message,
		Attachments: discussion.Attachments,
		IsYou:       isYou,
		IsSender:    isSender,
		CreatedAt:   discussion.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Aspirations List Model to Aspirations List Response
func ToAspirationsResponse(aspirations []model.Aspiration, sessionID uint, sessionRole model.RoleEnum) []dto.AspirationResponse {
	response := make([]dto.AspirationResponse, len(aspirations))
	for i, aspiration := range aspirations {
		response[i] = ToAspirationResponse(&aspiration, sessionID, sessionRole)
	}

	return response
}

// Filter Request to Filter
func ToAspirationFilter(request *dto.AspirationFilterRequest) model.AspirationFilter {
	return model.AspirationFilter{
		Query:     request.Query,
		Type:      request.Type,
		Status:    request.Status,
		StartDate: request.StartDate,
		EndDate:   request.EndDate,
		SenderID:  request.SenderID,
		Limit:     request.Limit,
		Page:      request.Page,
		SortKey:   request.SortKey,
		SortOrder: request.SortOrder,
	}
}

// Helper function to determine if the discussion sender can be viewed by the session user
func canViewDiscussionSender(aspiration *model.Aspiration, sessionID uint, sessionRole model.RoleEnum, discussion *model.Discussion) bool {
	// 1. Pembuat komentar selalu bisa melihat namanya sendiri
	if discussion.SenderID == sessionID { return true }

	// 2. Jika yang sedang melihat adalah si pembuat aspirasi (Sender),
	// maka dia TIDAK BISA melihat nama Admin/pihak manajemen yang membalas diskusinya
	if sessionID == aspiration.SenderID { return false }

	// 3. Jika bukan Admin/SuperAdmin, tutup aksesnya
	if sessionRole != model.RoleAdmin && sessionRole != model.RoleSuperAdmin { return false }

	// 4. Admin tidak bisa melihat nama pelapor jika aspirasinya tipe Kritik
	if aspiration.Type == model.TypeCriticism && discussion.SenderID == aspiration.SenderID { return false }

	return true
}
