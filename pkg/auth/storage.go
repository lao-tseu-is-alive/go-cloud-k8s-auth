package auth

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

// Storage is an interface to different implementation of persistence for Auths/TypeAuth
type Storage interface {
	// List returns the list of existing go_cloud_auths with the given offset and limit.
	List(ctx context.Context, offset, limit int, params ListParams) ([]*UseAuthList, error)
	// GetByExternalId returns the list of existing go_cloud_auths having the given externalId with the given offset and limit.
	GetByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*AuthList, error)
	// Search returns the list of existing go_cloud_auths filtered by search params with the given offset and limit.
	Search(ctx context.Context, offset, limit int, params SearchParams) ([]*AuthList, error)
	// Get returns the go_cloud_auth with the specified go_cloud_auths ID.
	Get(ctx context.Context, id uuid.UUID) (*Auth, error)
	// Exist returns true only if a go_cloud_auths with the specified id exists in store.
	Exist(ctx context.Context, id uuid.UUID) bool
	// Count returns the total number of go_cloud_auths.
	Count(ctx context.Context, params CountParams) (int32, error)
	// Create saves a new go_cloud_auths in the storage.
	Create(ctx context.Context, go_cloud_auth Auth) (*Auth, error)
	// Update updates the go_cloud_auths with given ID in the storage.
	Update(ctx context.Context, id uuid.UUID, go_cloud_auth Auth) (*Auth, error)
	// Delete removes the go_cloud_auths with given ID from the storage.
	Delete(ctx context.Context, id uuid.UUID, userId int32) error
	// IsAuthActive returns true if the go_cloud_auth with the specified id has the inactivated attribute set to false
	IsAuthActive(ctx context.Context, id uuid.UUID) bool
}

func GetStorageInstanceOrPanic(ctx context.Context, dbDriver string, db database.DB, l *slog.Logger) Storage {
	var store Storage
	var err error
	switch dbDriver {
	case "pgx":
		store, err = NewPgxDB(ctx, db, l)
		if err != nil {
			l.Error("error doing NewPgxDB", "error", err)
			panic(err)
		}

	default:
		panic("unsupported DB driver type")
	}
	return store
}
