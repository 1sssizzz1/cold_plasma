package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type ProcedureService struct {
	procedures repository.ProcedureRepository
	reviews    repository.ProcedureReviewRepository
}

func NewProcedureService(procedures repository.ProcedureRepository, reviews repository.ProcedureReviewRepository) *ProcedureService {
	return &ProcedureService{procedures: procedures, reviews: reviews}
}

func (s *ProcedureService) List(ctx context.Context) ([]models.Procedure, error) {
	return s.procedures.ListActive(ctx)
}

func (s *ProcedureService) Get(ctx context.Context, id int64) (models.Procedure, error) {
	p, err := s.procedures.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.Procedure{}, fmt.Errorf("процедура не найдена: %w", ErrNotFound)
		}
		return models.Procedure{}, err
	}
	return p, nil
}

func (s *ProcedureService) ListAll(ctx context.Context) ([]models.Procedure, error) {
	return s.procedures.ListAll(ctx)
}

func (s *ProcedureService) Save(ctx context.Context, id int64, p repository.ProcedureParams) (models.Procedure, error) {
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)
	p.Category = strings.TrimSpace(p.Category)
	p.ImageURL = strings.TrimSpace(p.ImageURL)
	p.VideoURL = strings.TrimSpace(p.VideoURL)
	p.Services = strings.TrimSpace(p.Services)
	p.DurationStr = strings.TrimSpace(p.DurationStr)
	if p.Title == "" {
		return models.Procedure{}, se(ErrValidation, "Название процедуры обязательно")
	}
	if p.DurationMins < 0 || p.Price < 0 || p.BonusEarned < 0 {
		return models.Procedure{}, se(ErrValidation, "Длительность, цена и бонусы не могут быть отрицательными")
	}
	if id > 0 {
		out, err := s.procedures.Update(ctx, id, p)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return models.Procedure{}, se(ErrNotFound, "Процедура не найдена")
			}
			return models.Procedure{}, err
		}
		return out, nil
	}
	return s.procedures.Create(ctx, p)
}

func (s *ProcedureService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return se(ErrValidation, "Некорректный id")
	}
	if err := s.procedures.Delete(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrNotFound, "Процедура не найдена")
		}
		return err
	}
	return nil
}

func (s *ProcedureService) ListReviews(ctx context.Context, procedureID int64) ([]models.ProcedureReview, error) {
	return s.reviews.ListPublic(ctx, procedureID, 100)
}

func (s *ProcedureService) ListAllReviews(ctx context.Context) ([]models.ProcedureReview, error) {
	return s.reviews.ListAll(ctx, 300)
}

func (s *ProcedureService) AddReview(ctx context.Context, userID, procedureID int64, rating int, text string) (models.ProcedureReview, error) {
	if procedureID <= 0 {
		return models.ProcedureReview{}, se(ErrValidation, "Процедура обязательна")
	}
	if rating < 0 || rating > 5 {
		return models.ProcedureReview{}, se(ErrValidation, "Оценка должна быть от 0 до 5")
	}
	text = strings.TrimSpace(text)
	ok, err := s.reviews.CanUserReview(ctx, userID, procedureID)
	if err != nil {
		return models.ProcedureReview{}, err
	}
	if !ok {
		return models.ProcedureReview{}, se(ErrForbidden, "Отзыв можно оставить после завершённого сеанса")
	}
	return s.reviews.Upsert(ctx, userID, procedureID, rating, text)
}

func (s *ProcedureService) DeleteReview(ctx context.Context, id int64) error {
	if id <= 0 {
		return se(ErrValidation, "Некорректный id")
	}
	if err := s.reviews.Delete(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrNotFound, "Отзыв не найден")
		}
		return err
	}
	return nil
}
