package models

import "errors"

// ErrNotFound — доменная ошибка отсутствия записи.
// Используется во всех слоях, чтобы удобно делать errors.Is().
var ErrNotFound = errors.New("not found")

