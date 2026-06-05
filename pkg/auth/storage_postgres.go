package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

type PGX struct {
	Conn *pgxpool.Pool
	dbi  database.DB
	log  *slog.Logger
}

// NewPgxDB will instantiate a new storage of type postgres and ensure schema exist
func NewPgxDB(ctx context.Context, db database.DB, log *slog.Logger) (Storage, error) {
	var psql PGX
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, err
	}
	psql.Conn = pgConn
	psql.dbi = db
	psql.log = log
	var numberOfTypeAuths int
	errTypeAuthTable := pgConn.QueryRow(ctx, typeAuthCount).Scan(&numberOfTypeAuths)
	if errTypeAuthTable != nil {
		log.Error("Unable to retrieve the number of typeAuth", "error", err)
		return nil, errTypeAuthTable
	}

	if numberOfTypeAuths > 0 {
		log.Info("database contains records in go_auth_db_schema.type_go_cloud_auth", "count", numberOfTypeAuths)
	} else {
		log.Warn("go_auth_db_schema.type_go_cloud_auth is empty - it should contain at least one row")
		return nil, fmt.Errorf("«go_auth_db_schema.type_go_cloud_auth» contains %w should not be empty", numberOfTypeAuths)
	}

	return &psql, err
}

// List returns the list of existing go_cloud_auths with the given offset and limit.
func (db *PGX) List(ctx context.Context, offset, limit int, params ListParams) ([]*AuthList, error) {
	db.log.Debug("trace: entering List", "offset", offset, "limit", limit)
	if params.Type != nil {
		db.log.Info("param type", "type", *params.Type)
	}
	if params.CreatedBy != nil {
		db.log.Info("params.CreatedBy", "createdBy", *params.CreatedBy)
	}
	var (
		res []*AuthList
		err error
	)
	isInactive := false
	if params.Inactivated != nil {
		isInactive = *params.Inactivated
	}
	listAuths := baseAuthListQuery + listAuthsConditions
	if params.Validated != nil {
		db.log.Debug("params.Validated is not nil ")
		isValidated := *params.Validated
		listAuths += " AND validated = coalesce($6, validated) " + go_cloud_authListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listAuths,
			limit, offset, &params.Type, &params.CreatedBy, isInactive, isValidated)
	} else {
		listAuths += go_cloud_authListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listAuths,
			limit, offset, &params.Type, &params.CreatedBy, isInactive)
	}
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "List", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("List returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// ListByExternalId returns the list of existing go_cloud_auths having given externalId with the given offset and limit.
func (db *PGX) ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*AuthList, error) {
	db.log.Debug("trace: entering ListByExternalId", "externalId", externalId)
	var res []*AuthList
	listByExternalIdAuths := baseAuthListQuery + listByExternalIdAuthsCondition + go_cloud_authListOrderBy
	err := pgxscan.Select(ctx, db.Conn, &res, listByExternalIdAuths, limit, offset, externalId)
	if err != nil {
		db.log.Error("ListByExternalId failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("ListByExternalId returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

func (db *PGX) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*AuthList, error) {
	db.log.Debug("trace: entering Search", "offset", offset, "limit", limit)
	var (
		res []*AuthList
		err error
	)
	searchAuths := baseAuthListQuery + listAuthsConditions
	if params.Keywords != nil {
		searchAuths += " AND text_search @@ plainto_tsquery('french', unaccent($6))"
		if params.Validated != nil {
			searchAuths += " AND validated = coalesce($7, validated) " + go_cloud_authListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchAuths,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords, &params.Validated)
		} else {
			searchAuths += go_cloud_authListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchAuths,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords)
		}
	} else {
		if params.Validated != nil {
			searchAuths += " AND validated = coalesce($6, validated) " + go_cloud_authListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchAuths,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Validated)
		} else {
			searchAuths += go_cloud_authListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchAuths,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated)
		}
	}

	if err != nil {
		db.log.Error("Search failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("Search returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// Get will retrieve the go_cloud_auth with given id
func (db *PGX) Get(ctx context.Context, id uuid.UUID) (*Auth, error) {
	db.log.Debug("trace: entering Get", "id", id)
	res := &Auth{}
	err := pgxscan.Get(ctx, db.Conn, res, getAuth, id)
	if err != nil {
		db.log.Error("Get failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("Get returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// Exist returns true only if a go_cloud_auth with the specified id exists in store.
func (db *PGX) Exist(ctx context.Context, id uuid.UUID) bool {
	db.log.Debug("trace: entering Exist", "id", id)
	count, err := db.dbi.GetQueryInt(ctx, existAuth, id)
	if err != nil {
		db.log.Error("Exist could not be retrieved from DB", "id", id, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("Exist: id does exist", "id", id, "count", count)
		return true
	} else {
		db.log.Info("Exist: id does not exist", "id", id, "count", count)
		return false
	}
}

// Count returns the number of go_cloud_auth stored in DB
func (db *PGX) Count(ctx context.Context, params CountParams) (int32, error) {
	db.log.Debug("trace : entering Count()")
	var (
		count int
		err   error
	)
	queryCount := countAuth + " WHERE _deleted = false AND position IS NOT NULL "
	withoutSearchParameters := true
	if params.Keywords != nil {
		withoutSearchParameters = false
		queryCount += `AND text_search @@ plainto_tsquery('french', unaccent($1))
		AND type_id = coalesce($2, type_id)
		AND _created_by = coalesce($3, _created_by)
		AND inactivated = coalesce($4, inactivated)
`
		if params.Validated != nil {
			db.log.Debug("params.Validated is not nil ")
			isValidated := *params.Validated
			queryCount += " AND validated = coalesce($4, validated) "
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Keywords, &params.Type, &params.CreatedBy, &params.Inactivated, isValidated)

		} else {
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Keywords, &params.Type, &params.CreatedBy, &params.Inactivated)
		}
	}
	if withoutSearchParameters {
		queryCount += `
		AND type_id = coalesce($1, type_id)
		AND _created_by = coalesce($2, _created_by)
		AND inactivated = coalesce($3, inactivated)
`
		if params.Validated != nil {
			db.log.Debug("params.Validated is not nil ")
			isValidated := *params.Validated
			queryCount += " AND validated = coalesce($4, validated) "
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Type, &params.CreatedBy, &params.Inactivated, isValidated)

		} else {
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Type, &params.CreatedBy, &params.Inactivated)
		}

	}

	if err != nil {
		db.log.Error("Count failed", "error", err)
		return 0, err
	}
	return int32(count), nil
}

// Create will store the new Auth in the database
func (db *PGX) Create(ctx context.Context, t Auth) (*Auth, error) {
	db.log.Debug("trace: entering Create", "name", t.Name, "id", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, createAuth,
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$14
		&t.ManagedBy, t.CreatedBy, &t.MoreData, t.PosX, t.PosY)
	if err != nil {
		db.log.Error("Create unexpectedly failed", "name", t.Name, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Create no row was created", "name", t.Name)
		return nil, err
	}
	db.log.Info("Create success", "name", t.Name, "id", t.Id)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdAuth, err := db.Get(ctx, t.Id)
	if err != nil {
		return nil, fmt.Errorf("error %w: go_cloud_auth was created, but can not be retrieved", err)
	}
	return createdAuth, nil
}

// Update the go_cloud_auth stored in DB with given id and other information in struct
func (db *PGX) Update(ctx context.Context, id uuid.UUID, t Auth) (*Auth, error) {
	db.log.Debug("trace: entering Update", "id", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, updateAuth,
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Inactivated, &t.InactivatedTime, &t.InactivatedBy, &t.InactivatedReason, //$15
		t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$18
		&t.ManagedBy, &t.LastModifiedBy, &t.MoreData, t.PosX, t.PosY) //$23
	if err != nil {

		db.log.Error("Update unexpectedly failed", "id", t.Id, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Update no row was updated", "id", t.Id)
		return nil, err
	}

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedAuth, err := db.Get(ctx, t.Id)
	if err != nil {
		return nil, fmt.Errorf("error %w: go_cloud_auth was updated, but can not be retrieved", err)
	}
	return updatedAuth, nil
}

// Delete the go_cloud_auth stored in DB with given id
func (db *PGX) Delete(ctx context.Context, id uuid.UUID, userId int32) error {
	db.log.Debug("trace: entering Delete", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteAuth, userId, id)
	if err != nil {
		db.log.Error("go_cloud_auth could not be deleted", "id", id, "error", err)
		return fmt.Errorf("go_cloud_auth could not be deleted: %w", err)
	}
	if rowsAffected < 1 {
		db.log.Error("go_cloud_auth was not deleted", "id", id)
		return fmt.Errorf("go_cloud_auth was not marked for deletetion")
	}
	return nil
}

// IsAuthActive returns true if the go_cloud_auth with the specified id has the inactivated attribute set to false
func (db *PGX) IsAuthActive(ctx context.Context, id uuid.UUID) bool {
	db.log.Debug("trace: entering IsAuthActive", "id", id)
	count, err := db.dbi.GetQueryInt(ctx, isActiveAuth, id)
	if err != nil {
		db.log.Error("IsAuthActive could not be retrieved from DB", "id", id, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("IsAuthActive is true", "id", id, "count", count)
		return true
	} else {
		db.log.Info("IsAuthActive is false", "id", id, "count", count)
		return false
	}
}

// IsUserOwner returns true only if userId is the creator of the record (owner) of this go_cloud_auth in store.
func (db *PGX) IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool {
	db.log.Debug("trace: entering IsUserOwner", "id", id, "userId", userId)
	count, err := db.dbi.GetQueryInt(ctx, existAuthOwnedBy, id, userId)
	if err != nil {
		db.log.Error("IsUserOwner could not be retrieved from DB", "id", id, "userId", userId, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("IsUserOwner is true", "id", id, "userId", userId, "count", count)
		return true
	} else {
		db.log.Info("IsUserOwner is false", "id", id, "userId", userId, "count", count)
		return false
	}
}

// CreateTypeAuth will store the new TypeAuth in the database
func (db *PGX) CreateTypeAuth(ctx context.Context, tt TypeAuth) (*TypeAuth, error) {
	db.log.Debug("trace: entering CreateTypeAuth", "name", tt.Name, "createdBy", tt.CreatedBy)
	var lastInsertId int = 0
	err := db.Conn.QueryRow(ctx, createTypeAuth,
		tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, &tt.GeometryType, //$6
		&tt.ManagedBy, tt.IconPath, tt.CreatedBy, &tt.MoreDataSchema).Scan(&lastInsertId)
	if err != nil {
		db.log.Error("CreateTypeAuth unexpectedly failed", "name", tt.Name, "error", err)
		return nil, err
	}
	db.log.Info("CreateTypeAuth success", "name", tt.Name, "id", lastInsertId)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdTypeAuth, err := db.GetTypeAuth(ctx, int32(lastInsertId))
	if err != nil {
		return nil, fmt.Errorf("error %w: typeAuth was created, but can not be retrieved", err)
	}
	return createdTypeAuth, nil
}

// UpdateTypeAuth updates the TypeAuth stored in DB with given id and other information in struct
func (db *PGX) UpdateTypeAuth(ctx context.Context, id int32, tt TypeAuth) (*TypeAuth, error) {
	db.log.Debug("trace: entering UpdateTypeAuth", "id", id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, updateTypeTing,
		id, tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, //$6
		&tt.GeometryType, tt.Inactivated, &tt.InactivatedTime, &tt.InactivatedBy, &tt.InactivatedReason, //$11
		&tt.ManagedBy, tt.IconPath, &tt.LastModifiedBy, &tt.MoreDataSchema) //$14
	if err != nil {

		db.log.Error("UpdateTypeAuth unexpectedly failed", "id", id, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("UpdateTypeAuth no row was updated", "id", id)
		return nil, err
	}

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedTypeAuth, err := db.GetTypeAuth(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error %w: go_cloud_auth was updated, but can not be retrieved", err)
	}
	return updatedTypeAuth, nil
}

// DeleteTypeAuth deletes the TypeAuth stored in DB with given id
func (db *PGX) DeleteTypeAuth(ctx context.Context, id int32, userId int32) error {
	db.log.Debug("trace: entering DeleteTypeAuth", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteTypeAuth, userId, id)
	if err != nil {
		db.log.Error("typego_cloud_auth could not be deleted", "id", id, "error", err)
		return fmt.Errorf("typego_cloud_auth could not be deleted: %w", err)
	}
	if rowsAffected < 1 {
		db.log.Error("typego_cloud_auth was not deleted", "id", id)
		return fmt.Errorf("typego_cloud_auth was not marked for deletion")
	}
	return nil
}

// ListTypeAuth returns the list of existing TypeAuth with the given offset and limit.
func (db *PGX) ListTypeAuth(ctx context.Context, offset, limit int, params TypeAuthListParams) ([]*TypeAuthList, error) {
	db.log.Debug("trace : entering ListTypeAuth")
	var (
		res []*TypeAuthList
		err error
	)
	listTypeAuths := typeAuthListQuery
	if params.Keywords != nil {
		listTypeAuths += listTypeAuthsConditionsWithKeywords + typeAuthListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTypeAuths,
			limit, offset, &params.Keywords, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	} else {
		listTypeAuths += listTypeAuthsConditionsWithoutKeywords + typeAuthListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTypeAuths,
			limit, offset, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	}

	if err != nil {
		db.log.Error("ListTypeAuth failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("ListTypeAuth returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// GetTypeAuth will retrieve the TypeAuth with given id
func (db *PGX) GetTypeAuth(ctx context.Context, id int32) (*TypeAuth, error) {
	db.log.Debug("trace: entering GetTypeAuth", "id", id)
	res := &TypeAuth{}
	err := pgxscan.Get(ctx, db.Conn, res, getTypeAuth, id)
	if err != nil {
		db.log.Error("GetTypeAuth failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("GetTypeAuth returned no results", "id", id)
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// CountTypeAuth returns the number of TypeAuth based on search criteria
func (db *PGX) CountTypeAuth(ctx context.Context, params TypeAuthCountParams) (int32, error) {
	db.log.Debug("trace : entering CountTypeAuth()")
	var (
		count int
		err   error
	)
	queryCount := countTypeAuth + " WHERE 1 = 1 "
	withoutSearchParameters := true
	if params.Keywords != nil {
		withoutSearchParameters = false
		queryCount += `AND text_search @@ plainto_tsquery('french', unaccent($1))
		AND _created_by = coalesce($2, _created_by)
		AND inactivated = coalesce($3, inactivated)
`
		count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Keywords, &params.CreatedBy, &params.Inactivated)
	}
	if withoutSearchParameters {
		queryCount += `
		AND _created_by = coalesce($1, _created_by)
		AND inactivated = coalesce($2, inactivated)
`
		count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.CreatedBy, &params.Inactivated)

	}
	if err != nil {
		db.log.Error("CountTypeAuth failed", "error", err)
		return 0, err
	}
	return int32(count), nil
}

// GetTypeAuthMaxId will retrieve maximum value of TypeAuth id existing in store.
func (db *PGX) GetTypeAuthMaxId(ctx context.Context) (int32, error) {
	db.log.Debug("trace : entering GetTypeAuthMaxId")
	existingMaxId, err := db.dbi.GetQueryInt(ctx, typeAuthMaxId)
	if err != nil {
		db.log.Error("GetTypeAuthMaxId() failed", "error", err)
		return 0, err
	}
	return int32(existingMaxId), nil
}
