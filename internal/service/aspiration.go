package service

import (
	"log"
	"time"
	"bytes"
	"errors"
	"context"

	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/mapper"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/repository"

	"gorm.io/gorm"
)

type DiscussionBroadcaster interface {
	BroadcastDiscussion(aspiration *model.Aspiration, discussion *model.Discussion)
}

type AspirationService interface {
	// CRUD Operations
	Create(ctx context.Context, sessionID uint, request *dto.AspirationRequest) (*dto.AspirationResponse, error)
	Delete(ctx context.Context, aspirationID uint) error
	GetAll(ctx context.Context, sessionID uint, sessionRole model.RoleEnum, request *dto.AspirationFilterRequest) ([]dto.AspirationResponse, *dto.Pagination, error)
	GetByID(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum) (*dto.AspirationResponse, error)
	Close(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum) error

	// Count Functions
	Count(ctx context.Context, request *dto.AspirationFilterRequest) (int64, error)
	CountDaily(ctx context.Context, request *dto.AspirationFilterRequest) (map[string]int64, error)

	// Vote and Discussion Functions
	Vote(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum, request *dto.VoteRequest) (*dto.VoteResponse, error)
	Discussion(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum, request *dto.DiscussionRequest) (*dto.DiscussionResponse, error)

	// Export Aspirations to Excel
	Export(ctx context.Context, request *dto.AspirationFilterRequest) (*bytes.Buffer, error)
}

type aspirationService struct {
	repository  repository.AspirationRepository
	user        UserService
	broadcaster DiscussionBroadcaster
}

func NewAspirationService(repository repository.AspirationRepository, user UserService, broadcaster DiscussionBroadcaster) AspirationService {
	service := &aspirationService{repository, user, broadcaster}
	service.autoClose()
	return service
}

// Attachment Validation Rules
var attachmentValidate = utils.Validate{
	AllowedTypes: []string{"image/jpeg", "image/png", "application/pdf"},
	MaxSize:      10 << 20,
	MaxFiles:     5,
	MaxTotalSize: 25 << 20,
}

// Create a New Aspiration
func (s *aspirationService) Create(ctx context.Context, sessionID uint, request *dto.AspirationRequest) (*dto.AspirationResponse, error) {
	err := utils.ValidateFiles(request.Files, attachmentValidate);
	if err != nil { return nil, apperror.ErrBadRequest(err.Error()) }

	attachments, err := utils.SaveFiles(request.Files, "aspiration")
	if err != nil { return nil, apperror.ErrInternal("Failed to save attachments") }

	aspiration := mapper.ToAspirationModel(sessionID, request, attachments)
	err = s.repository.Create(ctx, &aspiration)
	if err != nil {
		cleanupErr := cleanupAttachments(attachments);
		if cleanupErr != nil { log.Printf("failed to clean up attachments after aspiration create failure: %v", cleanupErr) }

		if errors.Is(err, gorm.ErrForeignKeyViolated) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to create aspiration")
	}

	response := mapper.ToAspirationResponse(&aspiration, sessionID, model.RoleUser)
	return &response, nil
}

// Delete a Aspiration
func (s *aspirationService) Delete(ctx context.Context, aspirationID uint) error {
	aspiration, err := s.repository.GetByID(ctx, aspirationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return apperror.ErrNotFound("Aspiration not found") }
		return apperror.ErrInternal("Failed to retrieve aspiration")
	}

	err = s.repository.Delete(ctx, aspirationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return apperror.ErrNotFound("Aspiration not found") }
		return apperror.ErrInternal("Failed to delete aspiration")
	}

	attachmentPaths := aspirationAttachmentPaths(aspiration)
	err = utils.DeleteFiles(attachmentPaths);
	if err != nil { log.Printf("failed to delete attachments for aspiration %d: %v", aspirationID, err) }

	return nil
}

// Get All Aspirations
func (s *aspirationService) GetAll(ctx context.Context, sessionID uint, sessionRole model.RoleEnum, request *dto.AspirationFilterRequest) ([]dto.AspirationResponse, *dto.Pagination, error) {
	filter := mapper.ToAspirationFilter(request)
	aspirations, count, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, nil, apperror.ErrInternal("Failed to retrieve aspirations") }

	response := mapper.ToAspirationsResponse(aspirations, sessionID, sessionRole)
	pagination := mapper.ToPaginationResponse(count, filter.Page, filter.Limit)
	return response, pagination, nil
}

// Get Aspiration by ID
func (s *aspirationService) GetByID(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum) (*dto.AspirationResponse, error) {
	aspiration, err := s.repository.GetByID(ctx, aspirationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("Aspiration not found") }
		return nil, apperror.ErrInternal("Failed to retrieve aspiration")
	}

	if sessionRole == model.RoleUser && sessionID != aspiration.SenderID {
		return nil, apperror.ErrForbidden("You are not authorized to view this aspiration")
	}

	response := mapper.ToAspirationResponse(aspiration, sessionID, sessionRole)

	if aspiration.Type == model.TypeInnovation {
		adminRole := model.RoleAdmin
		filter := &dto.UserFilterRequest{Role: &adminRole}
		totalAdmins, err := s.user.Count(ctx, filter)
		if err == nil && totalAdmins > 0 {
			response.VoteCount.TotalAdmins = totalAdmins
			response.VoteCount.Threshold = (totalAdmins / 2) + 1
		}
	}

	return &response, nil
}

// Close an Aspiration
func (s *aspirationService) Close(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum) error {
	aspiration, err := s.repository.GetByID(ctx, aspirationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return apperror.ErrNotFound("Aspiration not found") }
		return apperror.ErrInternal("Failed to retrieve aspiration")
	}

	if sessionID != aspiration.SenderID && sessionRole != model.RoleSuperAdmin {
		return apperror.ErrForbidden("You are not authorized to close this aspiration")
	}

	if aspiration.Type != model.TypeCriticism { return apperror.ErrBadRequest("Only criticisms can be closed") }
	if aspiration.Status == model.StatusClose { return apperror.ErrBadRequest("Aspiration is already closed") }

	aspiration.Status = model.StatusClose
	err = s.repository.Update(ctx, aspiration)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return apperror.ErrNotFound("Aspiration not found") }
		return apperror.ErrInternal("Failed to close aspiration")
	}

	return nil
}

// Count Aspirations
func (s *aspirationService) Count(ctx context.Context, request *dto.AspirationFilterRequest) (int64, error) {
	filter := mapper.ToAspirationFilter(request)
	count, err := s.repository.Count(ctx, filter)
	if err != nil { return 0, apperror.ErrInternal("Failed to count aspirations") }

	return count, nil
}

// Count Daily Aspirations
func (s *aspirationService) CountDaily(ctx context.Context, request *dto.AspirationFilterRequest) (map[string]int64, error) {
	filter := mapper.ToAspirationFilter(request)
	count, err := s.repository.CountDaily(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to count daily aspirations") }

	return count, nil
}

// Vote and Discussion Functions
// Create a New Vote
func (s *aspirationService) Vote(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum, request *dto.VoteRequest) (*dto.VoteResponse, error) {
	aspiration, err := s.repository.GetByID(ctx, aspirationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("Aspiration not found") }
		return nil, apperror.ErrInternal("Failed to retrieve aspiration")
	}

	if sessionRole != model.RoleAdmin {
		return nil, apperror.ErrForbidden("You are not authorized to vote on this aspiration")
	}

	if aspiration.Status != model.StatusOpen {
		return nil, apperror.ErrForbidden("Voting is only allowed for open aspirations")
	}

	if sessionID == aspiration.SenderID {
		return nil, apperror.ErrForbidden("You cannot vote on your own aspiration")
	}

	vote := mapper.ToVoteModel(aspirationID, sessionID, request)
	err = s.repository.Vote(ctx, &vote)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) { return nil, apperror.ErrNotFound("Aspiration or User not found") }
		return nil, apperror.ErrInternal("Failed to create vote")
	}

	if aspiration.Type == model.TypeInnovation {
		admin := model.RoleAdmin
		filter := &dto.UserFilterRequest{ Role: &admin }
		adminCount, err := s.user.Count(ctx, filter)
		if err == nil && adminCount > 0 {
			acceptCount, _, totalVoters, err := s.repository.GetVoteStats(ctx, aspirationID)
			if err == nil {
				threshold := (adminCount / 2) + 1
				remainingVoters := adminCount - totalVoters

				var newStatus model.StatusEnum
				if acceptCount >= threshold {
					newStatus = model.StatusAccepted
				} else if acceptCount+remainingVoters < threshold {
					newStatus = model.StatusRejected
				}

				if newStatus != "" && newStatus != aspiration.Status {
					aspiration.Status = newStatus
					_ = s.repository.Update(ctx, aspiration)
				}
			}
		}
	}

	response := mapper.ToVoteResponse(&vote)
	return &response, nil
}

// Create a New Discussion
func (s *aspirationService) Discussion(ctx context.Context, aspirationID, sessionID uint, sessionRole model.RoleEnum, request *dto.DiscussionRequest) (*dto.DiscussionResponse, error) {
	aspiration, err := s.repository.GetByID(ctx, aspirationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("Aspiration not found") }
		return nil, apperror.ErrInternal("Failed to retrieve aspiration")
	}

	if aspiration.Status != model.StatusOpen {
		return nil, apperror.ErrForbidden("Discussion is only allowed for open aspirations")
	}

	if sessionRole != model.RoleAdmin && sessionRole != model.RoleSuperAdmin && sessionID != aspiration.SenderID {
		return nil, apperror.ErrForbidden("You are not authorized to discuss this aspiration")
	}

	err = utils.ValidateFiles(request.Files, attachmentValidate);
	if err != nil { return nil, apperror.ErrBadRequest(err.Error()) }

	attachments, err := utils.SaveFiles(request.Files, "discussion")
	if err != nil { return nil, apperror.ErrInternal("Failed to save attachments") }

	discussion := mapper.ToDiscussionModel(aspirationID, sessionID, request, attachments)
	err = s.repository.Discussion(ctx, &discussion)
	if err != nil {
		cleanupErr := cleanupAttachments(attachments);
		if cleanupErr != nil { log.Printf("failed to clean up attachments after discussion create failure: %v", cleanupErr) }

		if errors.Is(err, gorm.ErrForeignKeyViolated) { return nil, apperror.ErrNotFound("Aspiration or User not found") }
		return nil, apperror.ErrInternal("Failed to create discussion")
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastDiscussion(aspiration, &discussion)
	}

	response := mapper.ToDiscussionResponse(aspiration, sessionID, sessionRole, &discussion)

	return &response, nil
}

// Export Aspirations to Excel
func (s *aspirationService) Export(ctx context.Context, request *dto.AspirationFilterRequest) (*bytes.Buffer, error) {
	filter := mapper.ToAspirationFilter(request)
	filter.Limit = 0; filter.Page = 0

	aspirations, _, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to retrieve aspirations") }
	if len(aspirations) == 0 { return nil, apperror.ErrNotFound("No aspirations found") }

	excel, err := utils.GenerateAspirationsExcel(aspirations)
	if err != nil { return nil, apperror.ErrInternal("Failed to export aspirations") }

	return excel, nil
}

// Auto Close Aspirations
func (s *aspirationService) autoClose() {
	processExpired := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		admin := model.RoleAdmin
		filter := &dto.UserFilterRequest{Role: &admin}

		totalAdmins, err := s.user.Count(ctx, filter)
		if err != nil { log.Printf("[AutoCloseWorker] Error counting admins: %v", err); return }
		if totalAdmins <= 0 { log.Printf("[AutoCloseWorker] No admins found"); return }

		expired, err := s.repository.GetExpired(ctx)
		if err != nil { log.Printf("[AutoCloseWorker] Error finding expired innovations: %v", err); return }
		if len(expired) == 0 { return }

		threshold := (totalAdmins / 2) + 1
		for _, aspiration := range expired {
			aspirationToUpdate := aspiration

			acceptCount, _, _, err := s.repository.GetVoteStats(ctx, aspirationToUpdate.ID)
			if err != nil { continue }

			newStatus := model.StatusRejected
			if acceptCount >= threshold { newStatus = model.StatusAccepted }

			aspirationToUpdate.Status = newStatus
			err = s.repository.Update(ctx, &aspirationToUpdate)
			if err != nil { log.Printf("[AutoCloseWorker] Error updating aspiration %d: %v", aspirationToUpdate.ID, err) }
		}
	}

	go func() {
		processExpired()
		for {
			now := time.Now()
			nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := nextMidnight.Sub(now)

			<-time.After(duration)
			processExpired()
		}
	}()
}

// Helper Functions
// Get all attachment paths from an aspiration and its discussions
func aspirationAttachmentPaths(aspiration *model.Aspiration) []string {
	paths := make([]string, 0, len(aspiration.Attachments))
	for _, attachment := range aspiration.Attachments {
		if attachment.Path != "" { paths = append(paths, attachment.Path) }
	}

	for _, discussion := range aspiration.Discussions {
		for _, attachment := range discussion.Attachments {
			if attachment.Path != "" { paths = append(paths, attachment.Path) }
		}
	}

	return paths
}

// Cleanup attachments by deleting the files from the filesystem
func cleanupAttachments(attachments model.Attachments) error {
	paths := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		if attachment.Path != "" { paths = append(paths, attachment.Path) }
	}

	return utils.DeleteFiles(paths)
}
