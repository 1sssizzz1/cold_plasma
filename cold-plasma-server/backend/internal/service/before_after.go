package service

import (
	"context"
	"errors"
	"strings"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type BeforeAfterService struct {
	results repository.BeforeAfterRepository
}

func NewBeforeAfterService(results repository.BeforeAfterRepository) *BeforeAfterService {
	return &BeforeAfterService{results: results}
}

func (s *BeforeAfterService) List(ctx context.Context, limit int) ([]models.BeforeAfterResult, error) {
	return s.results.ListPublic(ctx, limit)
}

func (s *BeforeAfterService) ListAll(ctx context.Context, limit int) ([]models.BeforeAfterResult, error) {
	return s.results.ListAll(ctx, limit)
}

func (s *BeforeAfterService) Save(ctx context.Context, id int64, p repository.BeforeAfterParams) (models.BeforeAfterResult, error) {
	p.Procedure = strings.TrimSpace(p.Procedure)
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	p.BeforeURL = strings.TrimSpace(p.BeforeURL)
	p.AfterURL = strings.TrimSpace(p.AfterURL)
	if p.BeforeURL == "" || p.AfterURL == "" {
		return models.BeforeAfterResult{}, se(ErrValidation, "Фото до и после обязательны")
	}
	if p.SortOrder < 0 {
		return models.BeforeAfterResult{}, se(ErrValidation, "Порядок не может быть отрицательным")
	}
	if id > 0 {
		out, err := s.results.Update(ctx, id, p)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return models.BeforeAfterResult{}, se(ErrNotFound, "Результат не найден")
			}
			return models.BeforeAfterResult{}, err
		}
		return out, nil
	}
	return s.results.Create(ctx, p)
}

func (s *BeforeAfterService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return se(ErrValidation, "Некорректный id")
	}
	if err := s.results.Delete(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrNotFound, "Результат не найден")
		}
		return err
	}
	return nil
}
