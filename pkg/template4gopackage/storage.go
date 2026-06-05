package template4gopackage

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

// Storage is an interface to different implementation of persistence for Template4ServiceNames/TypeTemplate4ServiceName
type Storage interface {
	// GeoJson returns a geoJson of existing template_4_your_project_names with the given offset and limit.
	GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error)
	// List returns the list of existing template_4_your_project_names with the given offset and limit.
	List(ctx context.Context, offset, limit int, params ListParams) ([]*Template4ServiceNameList, error)
	// ListByExternalId returns the list of existing template_4_your_project_names having the given externalId with the given offset and limit.
	ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*Template4ServiceNameList, error)
	// Search returns the list of existing template_4_your_project_names filtered by search params with the given offset and limit.
	Search(ctx context.Context, offset, limit int, params SearchParams) ([]*Template4ServiceNameList, error)
	// Get returns the template_4_your_project_name with the specified template_4_your_project_names ID.
	Get(ctx context.Context, id uuid.UUID) (*Template4ServiceName, error)
	// Exist returns true only if a template_4_your_project_names with the specified id exists in store.
	Exist(ctx context.Context, id uuid.UUID) bool
	// Count returns the total number of template_4_your_project_names.
	Count(ctx context.Context, params CountParams) (int32, error)
	// Create saves a new template_4_your_project_names in the storage.
	Create(ctx context.Context, template_4_your_project_name Template4ServiceName) (*Template4ServiceName, error)
	// Update updates the template_4_your_project_names with given ID in the storage.
	Update(ctx context.Context, id uuid.UUID, template_4_your_project_name Template4ServiceName) (*Template4ServiceName, error)
	// Delete removes the template_4_your_project_names with given ID from the storage.
	Delete(ctx context.Context, id uuid.UUID, userId int32) error
	// IsTemplate4ServiceNameActive returns true if the template_4_your_project_name with the specified id has the inactivated attribute set to false
	IsTemplate4ServiceNameActive(ctx context.Context, id uuid.UUID) bool
	// IsUserOwner returns true only if userId is the creator of the record (owner) of this template_4_your_project_name in store.
	IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool
	// CreateTypeTemplate4ServiceName saves a new typeTemplate4ServiceName in the storage.
	CreateTypeTemplate4ServiceName(ctx context.Context, typeTemplate4ServiceName TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error)
	// UpdateTypeTemplate4ServiceName updates the typeTemplate4ServiceName with given ID in the storage.
	UpdateTypeTemplate4ServiceName(ctx context.Context, id int32, typeTemplate4ServiceName TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error)
	// DeleteTypeTemplate4ServiceName removes the typeTemplate4ServiceName with given ID from the storage.
	DeleteTypeTemplate4ServiceName(ctx context.Context, id int32, userId int32) error
	// ListTypeTemplate4ServiceName returns the list of active typeTemplate4ServiceNames with the given offset and limit.
	ListTypeTemplate4ServiceName(ctx context.Context, offset, limit int, params TypeTemplate4ServiceNameListParams) ([]*TypeTemplate4ServiceNameList, error)
	// GetTypeTemplate4ServiceName returns the typeTemplate4ServiceName with the specified template_4_your_project_names ID.
	GetTypeTemplate4ServiceName(ctx context.Context, id int32) (*TypeTemplate4ServiceName, error)
	// CountTypeTemplate4ServiceName returns the number of TypeTemplate4ServiceName based on search criteria
	CountTypeTemplate4ServiceName(ctx context.Context, params TypeTemplate4ServiceNameCountParams) (int32, error)
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
