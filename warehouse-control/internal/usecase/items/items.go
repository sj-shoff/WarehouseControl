package items_usecase

import (
	"context"
	"errors"
	"fmt"

	"warehouse-control/internal/domain"

	customErr "warehouse-control/internal/domain/errors"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type ItemsUsecase struct {
	repo     itemsRepository
	logger   *zlog.Zerolog
	validate *validator.Validate
}

func NewService(repo itemsRepository, logger *zlog.Zerolog) *ItemsUsecase {
	return &ItemsUsecase{
		repo:     repo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (s *ItemsUsecase) CreateItem(ctx context.Context, item *domain.Item, username string) (int64, error) {
	if err := s.validate.Struct(item); err != nil {
		s.logger.Error().Err(err).Msg("Validation failed")
		return 0, fmt.Errorf("%w: %v", customErr.ErrInvalidInput, err)
	}
	s.logger.Info().Str("user", username).Msg("Creating item")
	id, err := s.repo.CreateItem(ctx, item, username)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create item")
		if errors.Is(err, customErr.ErrDatabase) {
			return 0, customErr.ErrDatabase
		}
		return 0, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Str("user", username).Msg("Item created")
	return id, nil
}

func (s *ItemsUsecase) GetItems(ctx context.Context, limit, offset int, search string) ([]*domain.Item, int, error) {
	if limit <= 0 {
		limit = 100
	}
	s.logger.Info().Msg("Getting items")
	items, total, err := s.repo.GetItems(ctx, limit, offset, search)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get items")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, 0, customErr.ErrDatabase
		}
		return nil, 0, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int("count", len(items)).Msg("Items retrieved")
	return items, total, nil
}

func (s *ItemsUsecase) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	if id <= 0 {
		return nil, customErr.ErrInvalidInput
	}
	s.logger.Info().Int64("id", id).Msg("Getting item")
	item, err := s.repo.GetItemByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("Failed to get item")
		if errors.Is(err, customErr.ErrItemNotFound) {
			return nil, customErr.ErrItemNotFound
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Msg("Item retrieved")
	return item, nil
}

func (s *ItemsUsecase) UpdateItem(ctx context.Context, id int64, item *domain.Item, username string) error {
	if id <= 0 {
		return customErr.ErrInvalidInput
	}
	if err := s.validate.Struct(item); err != nil {
		s.logger.Error().Err(err).Msg("Validation failed")
		return fmt.Errorf("%w: %v", customErr.ErrInvalidInput, err)
	}
	s.logger.Info().Int64("id", id).Str("user", username).Msg("Updating item")
	err := s.repo.UpdateItem(ctx, id, item, username)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("Failed to update item")
		if errors.Is(err, customErr.ErrItemNotFound) {
			return customErr.ErrItemNotFound
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return customErr.ErrDatabase
		}
		return fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Str("user", username).Msg("Item updated")
	return nil
}

func (s *ItemsUsecase) DeleteItem(ctx context.Context, id int64, username string) error {
	if id <= 0 {
		return customErr.ErrInvalidInput
	}
	s.logger.Info().Int64("id", id).Str("user", username).Msg("Deleting item")
	err := s.repo.DeleteItem(ctx, id, username)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("Failed to delete item")
		if errors.Is(err, customErr.ErrItemNotFound) {
			return customErr.ErrItemNotFound
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return customErr.ErrDatabase
		}
		return fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Str("user", username).Msg("Item deleted")
	return nil
}

func (s *ItemsUsecase) BulkDeleteItems(ctx context.Context, ids []int64, username string) error {
	s.logger.Info().Str("user", username).Msg("Bulk deleting items")
	err := s.repo.BulkDeleteItems(ctx, ids, username)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to bulk delete items")
		if errors.Is(err, customErr.ErrDatabase) {
			return customErr.ErrDatabase
		}
		return fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Str("user", username).Msg("Items bulk deleted")
	return nil
}
