package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
)

type Permission int8 // enum
const (
	R Permission = iota // Read implies List (SELECT in DB, or GET in API)
	W                   // implies INSERT,UPDATE, DELETE
	M                   // Update or Put only
	D                   // Delete only
	C                   // Create only (Insert, Post)
	P                   // change Permissions of one go_cloud_auth
	O                   // change Owner of one Auth
	A                   // Audit log of changes of one go_cloud_auth and read only special _fields like _created_by
)

func (s Permission) String() string {
	switch s {
	case R:
		return "R"
	case W:
		return "W"
	case M:
		return "M"
	case D:
		return "D"
	case C:
		return "C"
	case P:
		return "P"
	case O:
		return "O"
	case A:
		return "A"
	}
	return "ErrorPermissionUnknown"
}

type Service struct {
	Log              *slog.Logger
	DbConn           database.DB
	Store            Storage
	Server           *goHttpEcho.Server
	ListDefaultLimit int
}

func (s Service) GeoJson(ctx echo.Context, params GeoJsonParams) error {
	handlerName := "GeoJson"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	jsonResult, err := s.Store.GeoJson(reqCtx, offset, limit, params)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.List :%v", err))
		} else {
			jsonResult = "empty"
			return ctx.JSONBlob(http.StatusOK, []byte(jsonResult))
		}
	}
	return ctx.JSONBlob(http.StatusOK, []byte(jsonResult))
}

// List sends a list of go_cloud_auths in the store based on the given parameters filters
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth?limit=3&ofset=0' |jq
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth?limit=3&type=112' |jq
func (s Service) List(ctx echo.Context, params ListParams) error {
	handlerName := "List"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	list, err := s.Store.List(reqCtx, offset, limit, params)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.List :%v", err))
		} else {
			list = make([]*AuthList, 0)
			return ctx.JSON(http.StatusOK, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// Create allows to insert a new go_cloud_auth
// curl -s -XPOST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "3999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 3,"name": "Gil-Parcelle","description": "just a nice parcelle test","external_id": 345678912,"inactivated": false,"managed_by": 999, "more_data": NULL,"pos_x":2537603.0 ,"pos_y":1152613.0   }' 'http://localhost:9090/goapi/v1/go_cloud_auth'
func (s Service) Create(ctx echo.Context) error {
	handlerName := "Create"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToCreate(currentUserId, typeAuth) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	newAuth := &Auth{
		CreatedBy: int32(currentUserId),
	}
	if err := ctx.Bind(newAuth); err != nil {
		msg := fmt.Sprintf("Create has invalid format [%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Create Auth Bind ok", "go_cloud_auth", newAuth.Name)
	if len(strings.Trim(newAuth.Name, " ")) < 1 {

		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(newAuth.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN, "name", MinNameLength)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, newAuth.Id) {
		msg := fmt.Sprintf("This id (%v) already exist !", newAuth.Id)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	go_cloud_authCreated, err := s.Store.Create(reqCtx, *newAuth)
	if err != nil {
		msg := fmt.Sprintf("Create had an error saving go_cloud_auth:%#v, err:%#v", *newAuth, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Create success", "go_cloud_authId", go_cloud_authCreated.Id)
	return ctx.JSON(http.StatusCreated, go_cloud_authCreated)
}

// Count returns the number of go_cloud_auths found after filtering data with any given CountParams
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth/count' |jq
func (s Service) Count(ctx echo.Context, params CountParams) error {
	handlerName := "Count"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	numAuths, err := s.Store.Count(reqCtx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem counting go_cloud_auths :%v", err))
	}
	return ctx.JSON(http.StatusOK, numAuths)
}

// Delete will remove the given go_cloud_authId entry from the store, and if not present will return 400 Bad Request
// curl -v -XDELETE -H "Content-Type: application/json" -H "Authorization: Bearer $token" 'http://localhost:8888/api/users/3' ->  204 No Content if present and delete it
// curl -v -XDELETE -H "Content-Type: application/json"  -H "Authorization: Bearer $token" 'http://localhost:8888/users/93333' -> 400 Bad Request
func (s Service) Delete(ctx echo.Context, go_cloud_authId uuid.UUID) error {
	handlerName := "GeoJson"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, go_cloud_authId) == false {
		msg := fmt.Sprintf("Delete(%v) cannot delete this id, it does not exist !", go_cloud_authId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	if !s.Store.IsUserOwner(reqCtx, go_cloud_authId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this go_cloud_auth")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToDelete(currentUserId, typeAuth) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	err := s.Store.Delete(reqCtx, go_cloud_authId, currentUserId)
	if err != nil {
		msg := fmt.Sprintf("Delete(%v) got an error: %#v ", go_cloud_authId, err)
		s.Log.Error(msg)
		return echo.NewHTTPError(http.StatusInternalServerError, msg)
	}
	return ctx.NoContent(http.StatusNoContent)

}

// Get will retrieve the Auth with the given id in the store and return it
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth/9999971f-53d7-4eb6-8898-97f257ea5f27' |jq
func (s Service) Get(ctx echo.Context, go_cloud_authId uuid.UUID) error {
	handlerName := "Get"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, go_cloud_authId) == false {
		msg := fmt.Sprintf("Get(%v) cannot get this id, it does not exist !", go_cloud_authId)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToGet(currentUserId, typeAuth) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	go_cloud_auth, err := s.Store.Get(reqCtx, go_cloud_authId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving go_cloud_auth :%v", err))
		} else {
			msg := fmt.Sprintf("Get(%v) no rows found in db", go_cloud_authId)
			s.Log.Info(msg)
			return ctx.JSON(http.StatusNotFound, msg)
		}
	}
	return ctx.JSON(http.StatusOK, go_cloud_auth)
}

// Update will change the attributes values for the go_cloud_auth identified by the given go_cloud_authId
// curl -s -XPUT -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "3999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 3,"name": "Gil-Parcelle","description": "just a nice parcelle test by GIL","external_id": 345678912,"inactivated": false,"managed_by": 999, "more_data": {"info_value": 3230 },"pos_x":2537603.0 ,"pos_y":1152613.0   }' 'http://localhost:9090/goapi/v1/go_cloud_auth/3999971f-53d7-4eb6-8898-97f257ea5f27' |jq
func (s Service) Update(ctx echo.Context, go_cloud_authId uuid.UUID) error {
	handlerName := "GeoJson"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "go_cloud_authId", go_cloud_authId, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, go_cloud_authId) == false {
		msg := fmt.Sprintf("Update(%v) cannot update this id, it does not exist !", go_cloud_authId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	if !s.Store.IsUserOwner(reqCtx, go_cloud_authId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this go_cloud_auth")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToUpdate(currentUserId, typeAuth) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/

	updateAuth := new(Auth)
	if err := ctx.Bind(updateAuth); err != nil {
		msg := fmt.Sprintf("Update has invalid format error:[%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(updateAuth.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(updateAuth.Name) < MinNameLength {

		msg := fmt.Sprintf(FieldMinLengthIsN+FoundNum, "name", MinNameLength, len(updateAuth.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	updateAuth.LastModifiedBy = &currentUserId
	//TODO handle update of validated field correctly by adding validated time & user
	// handle update of managed_by field correctly by checking if user is a valid active one
	go_cloud_authUpdated, err := s.Store.Update(reqCtx, go_cloud_authId, *updateAuth)
	if err != nil {
		msg := fmt.Sprintf("Update had an error saving go_cloud_auth:%#v, err:%#v", *updateAuth, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Update success", "go_cloud_authId", go_cloud_authUpdated.Id)
	return ctx.JSON(http.StatusOK, go_cloud_authUpdated)
}

// ListByExternalId sends a list of go_cloud_auths in the store as json based of the given filters
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth/by-external-id/345678912?limit=3&ofset=0' |jq
func (s Service) ListByExternalId(ctx echo.Context, externalId int32, params ListByExternalIdParams) error {
	handlerName := "ListByExternalId"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	list, err := s.Store.ListByExternalId(reqCtx, offset, limit, int(externalId))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.ListByExternalId :%v", err))
		} else {
			list = make([]*AuthList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// Search returns a list of go_cloud_auths in the store as json based of the given search criteria
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth/search?limit=3&ofset=0' |jq
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/go_cloud_auth/search?limit=3&type=112' |jq
func (s Service) Search(ctx echo.Context, params SearchParams) error {
	handlerName := "Search"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	list, err := s.Store.Search(reqCtx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			list = make([]*AuthList, 0)
			return ctx.JSON(http.StatusOK, list)
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.Search :%v", err))
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// TypeAuthList sends a list of TypeAuth based on the given TypeAuthListParams parameters filters
func (s Service) TypeAuthList(ctx echo.Context, params TypeAuthListParams) error {
	handlerName := "TypeAuthList"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	limit := 250
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	list, err := s.Store.ListTypeAuth(reqCtx, offset, limit, params)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.ListTypeAuth :%v", err))
		} else {
			list = make([]*TypeAuthList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// TypeAuthCreate will insert a new TypeAuth in the store
func (s Service) TypeAuthCreate(ctx echo.Context) error {
	handlerName := "TypeAuthCreate"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeAuths)
	}
	newTypeAuth := &TypeAuth{
		Comment:           nil,
		CreatedAt:         nil,
		CreatedBy:         int32(currentUserId),
		Deleted:           false,
		DeletedAt:         nil,
		DeletedBy:         nil,
		Description:       nil,
		ExternalId:        nil,
		GeometryType:      nil,
		Id:                0,
		Inactivated:       false,
		InactivatedBy:     nil,
		InactivatedReason: nil,
		InactivatedTime:   nil,
		LastModifiedAt:    nil,
		LastModifiedBy:    nil,
		ManagedBy:         nil,
		IconPath:          "",
		MoreDataSchema:    nil,
		Name:              "",
		TableName:         nil,
	}
	if err := ctx.Bind(newTypeAuth); err != nil {
		msg := fmt.Sprintf("TypeAuthCreate has invalid format [%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(newTypeAuth.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(newTypeAuth.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN+", found %d", "name", MinNameLength, len(newTypeAuth.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	//s.Log.Info("# Create() before Store.TypeAuthCreate newAuth : %#v\n", newAuth)
	typeAuthCreated, err := s.Store.CreateTypeAuth(reqCtx, *newTypeAuth)
	if err != nil {
		msg := fmt.Sprintf("TypeAuthCreate had an error saving go_cloud_auth:%#v, err:%#v", *newTypeAuth, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("TypeAuthCreate success", "typeAuthId", typeAuthCreated.Id)
	return ctx.JSON(http.StatusCreated, typeAuthCreated)
}

func (s Service) TypeAuthCount(ctx echo.Context, params TypeAuthCountParams) error {
	handlerName := "TypeAuthCount"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	numAuths, err := s.Store.CountTypeAuth(reqCtx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem counting go_cloud_auths :%v", err))
	}
	return ctx.JSON(http.StatusOK, numAuths)
}

// TypeAuthDelete will remove the given TypeAuth entry from the store, and if not present will return 400 Bad Request
func (s Service) TypeAuthDelete(ctx echo.Context, typeAuthId int32) error {
	handlerName := "TypeAuthDelete"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeAuths)
	}
	reqCtx := ctx.Request().Context()
	typeAuthCount, err := s.DbConn.GetQueryInt(reqCtx, existTypeAuth, typeAuthId)
	if err != nil || typeAuthCount < 1 {
		msg := fmt.Sprintf("TypeAuthDelete(%v) cannot delete this id, it does not exist !", typeAuthId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	} else {
		err := s.Store.DeleteTypeAuth(reqCtx, typeAuthId, currentUserId)
		if err != nil {
			msg := fmt.Sprintf("TypeAuthDelete(%v) got an error: %#v ", typeAuthId, err)
			s.Log.Error(msg)
			return echo.NewHTTPError(http.StatusInternalServerError, msg)
		}
		return ctx.NoContent(http.StatusNoContent)
	}
}

// TypeAuthGet will retrieve the Auth with the given id in the store and return it
func (s Service) TypeAuthGet(ctx echo.Context, typeAuthId int32) error {
	handlerName := "TypeAuthGet"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeAuths)
	}
	reqCtx := ctx.Request().Context()
	typeAuthCount, err := s.DbConn.GetQueryInt(reqCtx, existTypeAuth, typeAuthId)
	if err != nil || typeAuthCount < 1 {
		msg := fmt.Sprintf("TypeAuthGet(%v) cannot retrieve this id, it does not exist !", typeAuthId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	typeAuth, err := s.Store.GetTypeAuth(reqCtx, typeAuthId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving TypeAuth :%v", err))
	}
	return ctx.JSON(http.StatusOK, typeAuth)
}

func (s Service) TypeAuthUpdate(ctx echo.Context, typeAuthId int32) error {
	handlerName := "TypeAuthUpdate"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeAuths)
	}
	reqCtx := ctx.Request().Context()
	typeAuthCount, err := s.DbConn.GetQueryInt(reqCtx, existTypeAuth, typeAuthId)
	if err != nil || typeAuthCount < 1 {
		msg := fmt.Sprintf("TypeAuthUpdate(%v) cannot update this id, it does not exist !", typeAuthId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	uTypeAuth := new(TypeAuth)
	if err := ctx.Bind(uTypeAuth); err != nil {
		msg := fmt.Sprintf("TypeAuthUpdate has invalid format error:[%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(uTypeAuth.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(uTypeAuth.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN+", found %d", "name", MinNameLength, len(uTypeAuth.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	uTypeAuth.LastModifiedBy = &currentUserId
	go_cloud_authUpdated, err := s.Store.UpdateTypeAuth(reqCtx, typeAuthId, *uTypeAuth)
	if err != nil {
		msg := fmt.Sprintf("TypeAuthUpdate had an error saving typeAuth:%#v, err:%#v", *uTypeAuth, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("TypeAuthUpdate success", "typeAuthId", go_cloud_authUpdated.Id)
	return ctx.JSON(http.StatusOK, go_cloud_authUpdated)
}
