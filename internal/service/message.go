package service

import (
	"bytes"
	"errors"
	"context"

	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/utils"
	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/repository"

	"gorm.io/gorm"
)

type MessageService interface {
	Create(ctx context.Context, userID uint, request *model.MessageRequest) (*model.MessageResponse, error)
	Delete(ctx context.Context, messageID uint) error
	GetAll(ctx context.Context, filter model.MessageFilterRequest) ([]model.MessageResponse, *model.Pagination, error)
	GetByID(ctx context.Context, messageID uint) (*model.MessageResponse, error)
	Export(ctx context.Context, filter model.MessageFilterRequest) (*bytes.Buffer, error)
	Count(ctx context.Context, filter model.MessageFilterRequest) (int64, error)
	CountDaily(ctx context.Context, filter model.MessageFilterRequest) (map[string]int64, error)
}

type messageService struct {
	repository repository.MessageRepository
}

func NewMessageService(repository repository.MessageRepository) MessageService {
	return &messageService{ repository }
}

// Create a New Message
func (s *messageService) Create(ctx context.Context, userID uint, request *model.MessageRequest) (*model.MessageResponse, error) {
	message := request.ToModel(userID)
	err := s.repository.Create(ctx, &message)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) { return nil, apperror.ErrNotFound("User not found") }
		return nil, apperror.ErrInternal("Failed to create message")
	}

	response := message.ToResponse()
	return &response, nil
}

// Delete a Message
func (s *messageService) Delete(ctx context.Context, messageID uint) error {
	err := s.repository.Delete(ctx, messageID)
	if err != nil { return apperror.ErrInternal("Failed to delete message") }

	return nil
}

// Get All Messages
func (s *messageService) GetAll(ctx context.Context, filter model.MessageFilterRequest) ([]model.MessageResponse, *model.Pagination, error) {
	messages, count, err := s.repository.GetAll(ctx, filter)
	if err != nil { return nil, nil, apperror.ErrInternal("Failed to retrieve messages") }

	response := model.ToMessagesResponseList(messages)
	pagination := model.CreatePagination(count, filter.Page, filter.Limit)
	return response, pagination, nil
}

// Get Message by ID
func (s *messageService) GetByID(ctx context.Context, messageID uint) (*model.MessageResponse, error) {
	message, err := s.repository.GetByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.ErrNotFound("Message not found") }
		return nil, apperror.ErrInternal("Failed to retrieve message")
	}

	response := message.ToResponse()
	return &response, nil
}

// Export Messages to Excel
func (s *messageService) Export(ctx context.Context, filter model.MessageFilterRequest) (*bytes.Buffer, error) {
    messages, _, err := s.repository.GetAll(ctx, filter)
    if err != nil { return nil, apperror.ErrInternal("Failed to retrieve messages") }
    if len(messages) == 0 { return nil, apperror.ErrNotFound("No messages found") }

    excel, err := utils.GenerateMessagesExcel(messages)
	if err != nil { return nil, apperror.ErrInternal("Failed to export messages") }

	return excel, nil
}

// Count Messages
func (s *messageService) Count(ctx context.Context, filter model.MessageFilterRequest) (int64, error) {
	count, err := s.repository.Count(ctx, filter)
	if err != nil { return 0, apperror.ErrInternal("Failed to count messages") }

	return count, nil
}

// Count Daily Messages
func (s *messageService) CountDaily(ctx context.Context, filter model.MessageFilterRequest) (map[string]int64, error) {
	count, err := s.repository.CountDaily(ctx, filter)
	if err != nil { return nil, apperror.ErrInternal("Failed to count daily messages") }

	return count, nil
}
