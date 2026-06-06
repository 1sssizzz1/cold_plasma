package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type BonusService struct {
	users repository.UserRepository
	bonus repository.BonusRepository
	tx    repository.TxManager
}

func NewBonusService(users repository.UserRepository, bonus repository.BonusRepository, tx repository.TxManager) *BonusService {
	return &BonusService{users: users, bonus: bonus, tx: tx}
}

func (s *BonusService) Balance(ctx context.Context, userID int64) (int, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return 0, fmt.Errorf("пользователь не найден: %w", ErrNotFound)
		}
		return 0, err
	}
	return u.BonusPoints, nil
}

func (s *BonusService) Logs(ctx context.Context, userID int64, limit int) ([]models.BonusLog, error) {
	return s.bonus.ListLogsByUserID(ctx, userID, limit)
}

func (s *BonusService) FindByPhone(ctx context.Context, phone string) (models.User, error) {
	phone, err := NormalizeRussianPhone(phone)
	if err != nil {
		return models.User{}, err
	}

	u, err := s.users.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.User{}, fmt.Errorf("пользователь с таким телефоном не найден: %w", ErrNotFound)
		}
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *BonusService) Spend(ctx context.Context, userID int64, amount int, comment string) (int, error) {
	comment = strings.TrimSpace(comment)
	if amount <= 0 {
		return 0, fmt.Errorf("amount должен быть > 0: %w", ErrValidation)
	}
	if comment == "" {
		comment = "Списание бонусов"
	}

	var newBalance int
	err := s.tx.WithTx(ctx, func(ctx context.Context, repos repository.TxRepos) error {
		u, err := repos.User().GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return fmt.Errorf("пользователь не найден: %w", ErrNotFound)
			}
			return err
		}
		if u.IsBlocked {
			return fmt.Errorf("пользователь заблокирован: %w", ErrForbidden)
		}
		if u.BonusPoints < amount {
			return fmt.Errorf("недостаточно бонусов: %w", ErrValidation)
		}

		newBalance = u.BonusPoints - amount
		if err := repos.User().UpdateBonusPoints(ctx, userID, newBalance); err != nil {
			return err
		}
		if err := repos.Bonus().AddLog(ctx, userID, "spend", amount, comment); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newBalance, nil
}

func (s *BonusService) SpendByPhone(ctx context.Context, phone string, amount int, comment string) (models.User, error) {
	phone, err := NormalizeRussianPhone(phone)
	if err != nil {
		return models.User{}, err
	}
	comment = strings.TrimSpace(comment)
	if amount <= 0 {
		return models.User{}, fmt.Errorf("amount должен быть > 0: %w", ErrValidation)
	}
	if comment == "" {
		comment = "Списание бонусов администратором"
	}

	var out models.User
	err = s.tx.WithTx(ctx, func(ctx context.Context, repos repository.TxRepos) error {
		u, err := repos.User().GetByPhone(ctx, phone)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return fmt.Errorf("пользователь с таким телефоном не найден: %w", ErrNotFound)
			}
			return err
		}
		if u.IsBlocked {
			return fmt.Errorf("пользователь заблокирован: %w", ErrForbidden)
		}
		if u.BonusPoints < amount {
			return fmt.Errorf("недостаточно бонусов: %w", ErrValidation)
		}

		newBalance := u.BonusPoints - amount
		if err := repos.User().UpdateBonusPoints(ctx, u.ID, newBalance); err != nil {
			return err
		}
		if err := repos.Bonus().AddLog(ctx, u.ID, "spend", amount, comment); err != nil {
			return err
		}
		u.BonusPoints = newBalance
		u.PasswordHash = ""
		out = u
		return nil
	})
	if err != nil {
		return models.User{}, err
	}
	return out, nil
}

func (s *BonusService) AwardByPhone(ctx context.Context, phone string, amount int, comment string) (models.User, error) {
	phone, err := NormalizeRussianPhone(phone)
	if err != nil {
		return models.User{}, err
	}
	comment = strings.TrimSpace(comment)
	if amount <= 0 {
		return models.User{}, fmt.Errorf("amount должен быть > 0: %w", ErrValidation)
	}
	if comment == "" {
		comment = "Начисление бонусов администратором"
	}

	var out models.User
	err = s.tx.WithTx(ctx, func(ctx context.Context, repos repository.TxRepos) error {
		u, err := repos.User().GetByPhone(ctx, phone)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return fmt.Errorf("пользователь с таким телефоном не найден: %w", ErrNotFound)
			}
			return err
		}
		if u.IsBlocked {
			return fmt.Errorf("пользователь заблокирован: %w", ErrForbidden)
		}
		newBalance := u.BonusPoints + amount
		if err := repos.User().UpdateBonusPoints(ctx, u.ID, newBalance); err != nil {
			return err
		}
		if err := repos.Bonus().AddLog(ctx, u.ID, "earn", amount, comment); err != nil {
			return err
		}
		u.BonusPoints = newBalance
		u.PasswordHash = ""
		out = u
		return nil
	})
	if err != nil {
		return models.User{}, err
	}
	return out, nil
}
