package template4gopackage

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
	var numberOfTypeTemplate4ServiceNames int
	errTypeTemplate4ServiceNameTable := pgConn.QueryRow(ctx, typeTemplate4ServiceNameCount).Scan(&numberOfTypeTemplate4ServiceNames)
	if errTypeTemplate4ServiceNameTable != nil {
		log.Error("Unable to retrieve the number of typeTemplate4ServiceName", "error", err)
		return nil, errTypeTemplate4ServiceNameTable
	}

	if numberOfTypeTemplate4ServiceNames > 0 {
		log.Info("database contains records in template_4_your_project_name_db_schema.type_template_4_your_project_name", "count", numberOfTypeTemplate4ServiceNames)
	} else {
		log.Warn("template_4_your_project_name_db_schema.type_template_4_your_project_name is empty - it should contain at least one row")
		return nil, fmt.Errorf("«template_4_your_project_name_db_schema.type_template_4_your_project_name» contains %w should not be empty", numberOfTypeTemplate4ServiceNames)
	}

	return &psql, err
}

func (db *PGX) GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error) {
	db.log.Debug("trace: entering GeoJson", "offset", offset, "limit", limit)
	if params.Type != nil {
		db.log.Info("param type", "type", *params.Type)
	}
	if params.CreatedBy != nil {
		db.log.Info("params.CreatedBy", "createdBy", *params.CreatedBy)
	}
	var (
		mayBeResultIsNull *string
		err               error
	)
	isInactive := false
	if params.Inactivated != nil {
		isInactive = *params.Inactivated
	}
	listTemplate4ServiceNames := baseGeoJsonTemplate4ServiceNameSearch + listTemplate4ServiceNamesConditions
	if params.Validated != nil {
		db.log.Debug("params.Validated is not nil ")
		isValidated := *params.Validated
		listTemplate4ServiceNames += " AND validated = coalesce($6, validated) " + geoJsonListEndOfQuery
		err = pgxscan.Select(ctx, db.Conn, &mayBeResultIsNull, listTemplate4ServiceNames,
			limit, offset, &params.Type, &params.CreatedBy, isInactive, isValidated)
	} else {
		listTemplate4ServiceNames += geoJsonListEndOfQuery
		err = pgxscan.Select(ctx, db.Conn, &mayBeResultIsNull, listTemplate4ServiceNames,
			limit, offset, &params.Type, &params.CreatedBy, isInactive)
	}
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "List", err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Info("List returned no results")
		return "", pgx.ErrNoRows
	}
	return *mayBeResultIsNull, nil
}

// List returns the list of existing template_4_your_project_names with the given offset and limit.
func (db *PGX) List(ctx context.Context, offset, limit int, params ListParams) ([]*Template4ServiceNameList, error) {
	db.log.Debug("trace: entering List", "offset", offset, "limit", limit)
	if params.Type != nil {
		db.log.Info("param type", "type", *params.Type)
	}
	if params.CreatedBy != nil {
		db.log.Info("params.CreatedBy", "createdBy", *params.CreatedBy)
	}
	var (
		res []*Template4ServiceNameList
		err error
	)
	isInactive := false
	if params.Inactivated != nil {
		isInactive = *params.Inactivated
	}
	listTemplate4ServiceNames := baseTemplate4ServiceNameListQuery + listTemplate4ServiceNamesConditions
	if params.Validated != nil {
		db.log.Debug("params.Validated is not nil ")
		isValidated := *params.Validated
		listTemplate4ServiceNames += " AND validated = coalesce($6, validated) " + template_4_your_project_nameListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTemplate4ServiceNames,
			limit, offset, &params.Type, &params.CreatedBy, isInactive, isValidated)
	} else {
		listTemplate4ServiceNames += template_4_your_project_nameListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTemplate4ServiceNames,
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

// ListByExternalId returns the list of existing template_4_your_project_names having given externalId with the given offset and limit.
func (db *PGX) ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*Template4ServiceNameList, error) {
	db.log.Debug("trace: entering ListByExternalId", "externalId", externalId)
	var res []*Template4ServiceNameList
	listByExternalIdTemplate4ServiceNames := baseTemplate4ServiceNameListQuery + listByExternalIdTemplate4ServiceNamesCondition + template_4_your_project_nameListOrderBy
	err := pgxscan.Select(ctx, db.Conn, &res, listByExternalIdTemplate4ServiceNames, limit, offset, externalId)
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

func (db *PGX) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*Template4ServiceNameList, error) {
	db.log.Debug("trace: entering Search", "offset", offset, "limit", limit)
	var (
		res []*Template4ServiceNameList
		err error
	)
	searchTemplate4ServiceNames := baseTemplate4ServiceNameListQuery + listTemplate4ServiceNamesConditions
	if params.Keywords != nil {
		searchTemplate4ServiceNames += " AND text_search @@ plainto_tsquery('french', unaccent($6))"
		if params.Validated != nil {
			searchTemplate4ServiceNames += " AND validated = coalesce($7, validated) " + template_4_your_project_nameListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchTemplate4ServiceNames,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords, &params.Validated)
		} else {
			searchTemplate4ServiceNames += template_4_your_project_nameListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchTemplate4ServiceNames,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords)
		}
	} else {
		if params.Validated != nil {
			searchTemplate4ServiceNames += " AND validated = coalesce($6, validated) " + template_4_your_project_nameListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchTemplate4ServiceNames,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Validated)
		} else {
			searchTemplate4ServiceNames += template_4_your_project_nameListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchTemplate4ServiceNames,
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

// Get will retrieve the template_4_your_project_name with given id
func (db *PGX) Get(ctx context.Context, id uuid.UUID) (*Template4ServiceName, error) {
	db.log.Debug("trace: entering Get", "id", id)
	res := &Template4ServiceName{}
	err := pgxscan.Get(ctx, db.Conn, res, getTemplate4ServiceName, id)
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

// Exist returns true only if a template_4_your_project_name with the specified id exists in store.
func (db *PGX) Exist(ctx context.Context, id uuid.UUID) bool {
	db.log.Debug("trace: entering Exist", "id", id)
	count, err := db.dbi.GetQueryInt(ctx, existTemplate4ServiceName, id)
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

// Count returns the number of template_4_your_project_name stored in DB
func (db *PGX) Count(ctx context.Context, params CountParams) (int32, error) {
	db.log.Debug("trace : entering Count()")
	var (
		count int
		err   error
	)
	queryCount := countTemplate4ServiceName + " WHERE _deleted = false AND position IS NOT NULL "
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

// Create will store the new Template4ServiceName in the database
func (db *PGX) Create(ctx context.Context, t Template4ServiceName) (*Template4ServiceName, error) {
	db.log.Debug("trace: entering Create", "name", t.Name, "id", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, createTemplate4ServiceName,
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
	createdTemplate4ServiceName, err := db.Get(ctx, t.Id)
	if err != nil {
		return nil, fmt.Errorf("error %w: template_4_your_project_name was created, but can not be retrieved", err)
	}
	return createdTemplate4ServiceName, nil
}

// Update the template_4_your_project_name stored in DB with given id and other information in struct
func (db *PGX) Update(ctx context.Context, id uuid.UUID, t Template4ServiceName) (*Template4ServiceName, error) {
	db.log.Debug("trace: entering Update", "id", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, updateTemplate4ServiceName,
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
	updatedTemplate4ServiceName, err := db.Get(ctx, t.Id)
	if err != nil {
		return nil, fmt.Errorf("error %w: template_4_your_project_name was updated, but can not be retrieved", err)
	}
	return updatedTemplate4ServiceName, nil
}

// Delete the template_4_your_project_name stored in DB with given id
func (db *PGX) Delete(ctx context.Context, id uuid.UUID, userId int32) error {
	db.log.Debug("trace: entering Delete", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteTemplate4ServiceName, userId, id)
	if err != nil {
		db.log.Error("template_4_your_project_name could not be deleted", "id", id, "error", err)
		return fmt.Errorf("template_4_your_project_name could not be deleted: %w", err)
	}
	if rowsAffected < 1 {
		db.log.Error("template_4_your_project_name was not deleted", "id", id)
		return fmt.Errorf("template_4_your_project_name was not marked for deletetion")
	}
	return nil
}

// IsTemplate4ServiceNameActive returns true if the template_4_your_project_name with the specified id has the inactivated attribute set to false
func (db *PGX) IsTemplate4ServiceNameActive(ctx context.Context, id uuid.UUID) bool {
	db.log.Debug("trace: entering IsTemplate4ServiceNameActive", "id", id)
	count, err := db.dbi.GetQueryInt(ctx, isActiveTemplate4ServiceName, id)
	if err != nil {
		db.log.Error("IsTemplate4ServiceNameActive could not be retrieved from DB", "id", id, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("IsTemplate4ServiceNameActive is true", "id", id, "count", count)
		return true
	} else {
		db.log.Info("IsTemplate4ServiceNameActive is false", "id", id, "count", count)
		return false
	}
}

// IsUserOwner returns true only if userId is the creator of the record (owner) of this template_4_your_project_name in store.
func (db *PGX) IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool {
	db.log.Debug("trace: entering IsUserOwner", "id", id, "userId", userId)
	count, err := db.dbi.GetQueryInt(ctx, existTemplate4ServiceNameOwnedBy, id, userId)
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

// CreateTypeTemplate4ServiceName will store the new TypeTemplate4ServiceName in the database
func (db *PGX) CreateTypeTemplate4ServiceName(ctx context.Context, tt TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error) {
	db.log.Debug("trace: entering CreateTypeTemplate4ServiceName", "name", tt.Name, "createdBy", tt.CreatedBy)
	var lastInsertId int = 0
	err := db.Conn.QueryRow(ctx, createTypeTemplate4ServiceName,
		tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, &tt.GeometryType, //$6
		&tt.ManagedBy, tt.IconPath, tt.CreatedBy, &tt.MoreDataSchema).Scan(&lastInsertId)
	if err != nil {
		db.log.Error("CreateTypeTemplate4ServiceName unexpectedly failed", "name", tt.Name, "error", err)
		return nil, err
	}
	db.log.Info("CreateTypeTemplate4ServiceName success", "name", tt.Name, "id", lastInsertId)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdTypeTemplate4ServiceName, err := db.GetTypeTemplate4ServiceName(ctx, int32(lastInsertId))
	if err != nil {
		return nil, fmt.Errorf("error %w: typeTemplate4ServiceName was created, but can not be retrieved", err)
	}
	return createdTypeTemplate4ServiceName, nil
}

// UpdateTypeTemplate4ServiceName updates the TypeTemplate4ServiceName stored in DB with given id and other information in struct
func (db *PGX) UpdateTypeTemplate4ServiceName(ctx context.Context, id int32, tt TypeTemplate4ServiceName) (*TypeTemplate4ServiceName, error) {
	db.log.Debug("trace: entering UpdateTypeTemplate4ServiceName", "id", id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, updateTypeTing,
		id, tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, //$6
		&tt.GeometryType, tt.Inactivated, &tt.InactivatedTime, &tt.InactivatedBy, &tt.InactivatedReason, //$11
		&tt.ManagedBy, tt.IconPath, &tt.LastModifiedBy, &tt.MoreDataSchema) //$14
	if err != nil {

		db.log.Error("UpdateTypeTemplate4ServiceName unexpectedly failed", "id", id, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("UpdateTypeTemplate4ServiceName no row was updated", "id", id)
		return nil, err
	}

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedTypeTemplate4ServiceName, err := db.GetTypeTemplate4ServiceName(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error %w: template_4_your_project_name was updated, but can not be retrieved", err)
	}
	return updatedTypeTemplate4ServiceName, nil
}

// DeleteTypeTemplate4ServiceName deletes the TypeTemplate4ServiceName stored in DB with given id
func (db *PGX) DeleteTypeTemplate4ServiceName(ctx context.Context, id int32, userId int32) error {
	db.log.Debug("trace: entering DeleteTypeTemplate4ServiceName", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteTypeTemplate4ServiceName, userId, id)
	if err != nil {
		db.log.Error("typetemplate_4_your_project_name could not be deleted", "id", id, "error", err)
		return fmt.Errorf("typetemplate_4_your_project_name could not be deleted: %w", err)
	}
	if rowsAffected < 1 {
		db.log.Error("typetemplate_4_your_project_name was not deleted", "id", id)
		return fmt.Errorf("typetemplate_4_your_project_name was not marked for deletion")
	}
	return nil
}

// ListTypeTemplate4ServiceName returns the list of existing TypeTemplate4ServiceName with the given offset and limit.
func (db *PGX) ListTypeTemplate4ServiceName(ctx context.Context, offset, limit int, params TypeTemplate4ServiceNameListParams) ([]*TypeTemplate4ServiceNameList, error) {
	db.log.Debug("trace : entering ListTypeTemplate4ServiceName")
	var (
		res []*TypeTemplate4ServiceNameList
		err error
	)
	listTypeTemplate4ServiceNames := typeTemplate4ServiceNameListQuery
	if params.Keywords != nil {
		listTypeTemplate4ServiceNames += listTypeTemplate4ServiceNamesConditionsWithKeywords + typeTemplate4ServiceNameListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTypeTemplate4ServiceNames,
			limit, offset, &params.Keywords, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	} else {
		listTypeTemplate4ServiceNames += listTypeTemplate4ServiceNamesConditionsWithoutKeywords + typeTemplate4ServiceNameListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTypeTemplate4ServiceNames,
			limit, offset, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	}

	if err != nil {
		db.log.Error("ListTypeTemplate4ServiceName failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("ListTypeTemplate4ServiceName returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// GetTypeTemplate4ServiceName will retrieve the TypeTemplate4ServiceName with given id
func (db *PGX) GetTypeTemplate4ServiceName(ctx context.Context, id int32) (*TypeTemplate4ServiceName, error) {
	db.log.Debug("trace: entering GetTypeTemplate4ServiceName", "id", id)
	res := &TypeTemplate4ServiceName{}
	err := pgxscan.Get(ctx, db.Conn, res, getTypeTemplate4ServiceName, id)
	if err != nil {
		db.log.Error("GetTypeTemplate4ServiceName failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("GetTypeTemplate4ServiceName returned no results", "id", id)
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// CountTypeTemplate4ServiceName returns the number of TypeTemplate4ServiceName based on search criteria
func (db *PGX) CountTypeTemplate4ServiceName(ctx context.Context, params TypeTemplate4ServiceNameCountParams) (int32, error) {
	db.log.Debug("trace : entering CountTypeTemplate4ServiceName()")
	var (
		count int
		err   error
	)
	queryCount := countTypeTemplate4ServiceName + " WHERE 1 = 1 "
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
		db.log.Error("CountTypeTemplate4ServiceName failed", "error", err)
		return 0, err
	}
	return int32(count), nil
}

// GetTypeTemplate4ServiceNameMaxId will retrieve maximum value of TypeTemplate4ServiceName id existing in store.
func (db *PGX) GetTypeTemplate4ServiceNameMaxId(ctx context.Context) (int32, error) {
	db.log.Debug("trace : entering GetTypeTemplate4ServiceNameMaxId")
	existingMaxId, err := db.dbi.GetQueryInt(ctx, typeTemplate4ServiceNameMaxId)
	if err != nil {
		db.log.Error("GetTypeTemplate4ServiceNameMaxId() failed", "error", err)
		return 0, err
	}
	return int32(existingMaxId), nil
}
