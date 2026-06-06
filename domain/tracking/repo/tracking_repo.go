package repo

import "context"

type TrackingRepo[T any] interface {
	Save(ctx context.Context, entity T) error
	FindByID(ctx context.Context, id int64) (T, error)
	FindAll(ctx context.Context, offset, limit int) ([]T, error)
	Delete(ctx context.Context, id int64) error
}
