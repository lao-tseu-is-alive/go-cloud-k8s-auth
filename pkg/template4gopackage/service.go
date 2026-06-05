package template4gopackage

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
	P                   // change Permissions of one template_4_your_project_name
	O                   // change Owner of one Template4ServiceName
	A                   // Audit log of changes of one template_4_your_project_name and read only special _fields like _created_by
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

// List sends a list of template_4_your_project_names in the store based on the given parameters filters
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name?limit=3&ofset=0' |jq
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name?limit=3&type=112' |jq
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
			list = make([]*Template4ServiceNameList, 0)
			return ctx.JSON(http.StatusOK, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// Create allows to insert a new template_4_your_project_name
// curl -s -XPOST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "3999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 3,"name": "Gil-Parcelle","description": "just a nice parcelle test","external_id": 345678912,"inactivated": false,"managed_by": 999, "more_data": NULL,"pos_x":2537603.0 ,"pos_y":1152613.0   }' 'http://localhost:9090/goapi/v1/template_4_your_project_name'
func (s Service) Create(ctx echo.Context) error {
	handlerName := "Create"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToCreate(currentUserId, typeTemplate4ServiceName) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	newTemplate4ServiceName := &Template4ServiceName{
		CreatedBy: int32(currentUserId),
	}
	if err := ctx.Bind(newTemplate4ServiceName); err != nil {
		msg := fmt.Sprintf("Create has invalid format [%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Create Template4ServiceName Bind ok", "template_4_your_project_name", newTemplate4ServiceName.Name)
	if len(strings.Trim(newTemplate4ServiceName.Name, " ")) < 1 {

		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(newTemplate4ServiceName.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN, "name", MinNameLength)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, newTemplate4ServiceName.Id) {
		msg := fmt.Sprintf("This id (%v) already exist !", newTemplate4ServiceName.Id)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	template_4_your_project_nameCreated, err := s.Store.Create(reqCtx, *newTemplate4ServiceName)
	if err != nil {
		msg := fmt.Sprintf("Create had an error saving template_4_your_project_name:%#v, err:%#v", *newTemplate4ServiceName, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Create success", "template_4_your_project_nameId", template_4_your_project_nameCreated.Id)
	return ctx.JSON(http.StatusCreated, template_4_your_project_nameCreated)
}

// Count returns the number of template_4_your_project_names found after filtering data with any given CountParams
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name/count' |jq
func (s Service) Count(ctx echo.Context, params CountParams) error {
	handlerName := "Count"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	numTemplate4ServiceNames, err := s.Store.Count(reqCtx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem counting template_4_your_project_names :%v", err))
	}
	return ctx.JSON(http.StatusOK, numTemplate4ServiceNames)
}

// Delete will remove the given template_4_your_project_nameId entry from the store, and if not present will return 400 Bad Request
// curl -v -XDELETE -H "Content-Type: application/json" -H "Authorization: Bearer $token" 'http://localhost:8888/api/users/3' ->  204 No Content if present and delete it
// curl -v -XDELETE -H "Content-Type: application/json"  -H "Authorization: Bearer $token" 'http://localhost:8888/users/93333' -> 400 Bad Request
func (s Service) Delete(ctx echo.Context, template_4_your_project_nameId uuid.UUID) error {
	handlerName := "GeoJson"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, template_4_your_project_nameId) == false {
		msg := fmt.Sprintf("Delete(%v) cannot delete this id, it does not exist !", template_4_your_project_nameId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	if !s.Store.IsUserOwner(reqCtx, template_4_your_project_nameId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this template_4_your_project_name")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToDelete(currentUserId, typeTemplate4ServiceName) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	err := s.Store.Delete(reqCtx, template_4_your_project_nameId, currentUserId)
	if err != nil {
		msg := fmt.Sprintf("Delete(%v) got an error: %#v ", template_4_your_project_nameId, err)
		s.Log.Error(msg)
		return echo.NewHTTPError(http.StatusInternalServerError, msg)
	}
	return ctx.NoContent(http.StatusNoContent)

}

// Get will retrieve the Template4ServiceName with the given id in the store and return it
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name/9999971f-53d7-4eb6-8898-97f257ea5f27' |jq
func (s Service) Get(ctx echo.Context, template_4_your_project_nameId uuid.UUID) error {
	handlerName := "Get"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, template_4_your_project_nameId) == false {
		msg := fmt.Sprintf("Get(%v) cannot get this id, it does not exist !", template_4_your_project_nameId)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToGet(currentUserId, typeTemplate4ServiceName) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	template_4_your_project_name, err := s.Store.Get(reqCtx, template_4_your_project_nameId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving template_4_your_project_name :%v", err))
		} else {
			msg := fmt.Sprintf("Get(%v) no rows found in db", template_4_your_project_nameId)
			s.Log.Info(msg)
			return ctx.JSON(http.StatusNotFound, msg)
		}
	}
	return ctx.JSON(http.StatusOK, template_4_your_project_name)
}

// Update will change the attributes values for the template_4_your_project_name identified by the given template_4_your_project_nameId
// curl -s -XPUT -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "3999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 3,"name": "Gil-Parcelle","description": "just a nice parcelle test by GIL","external_id": 345678912,"inactivated": false,"managed_by": 999, "more_data": {"info_value": 3230 },"pos_x":2537603.0 ,"pos_y":1152613.0   }' 'http://localhost:9090/goapi/v1/template_4_your_project_name/3999971f-53d7-4eb6-8898-97f257ea5f27' |jq
func (s Service) Update(ctx echo.Context, template_4_your_project_nameId uuid.UUID) error {
	handlerName := "GeoJson"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "template_4_your_project_nameId", template_4_your_project_nameId, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	if s.Store.Exist(reqCtx, template_4_your_project_nameId) == false {
		msg := fmt.Sprintf("Update(%v) cannot update this id, it does not exist !", template_4_your_project_nameId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	if !s.Store.IsUserOwner(reqCtx, template_4_your_project_nameId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this template_4_your_project_name")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToUpdate(currentUserId, typeTemplate4ServiceName) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/

	updateTemplate4ServiceName := new(Template4ServiceName)
	if err := ctx.Bind(updateTemplate4ServiceName); err != nil {
		msg := fmt.Sprintf("Update has invalid format error:[%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(updateTemplate4ServiceName.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(updateTemplate4ServiceName.Name) < MinNameLength {

		msg := fmt.Sprintf(FieldMinLengthIsN+FoundNum, "name", MinNameLength, len(updateTemplate4ServiceName.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	updateTemplate4ServiceName.LastModifiedBy = &currentUserId
	//TODO handle update of validated field correctly by adding validated time & user
	// handle update of managed_by field correctly by checking if user is a valid active one
	template_4_your_project_nameUpdated, err := s.Store.Update(reqCtx, template_4_your_project_nameId, *updateTemplate4ServiceName)
	if err != nil {
		msg := fmt.Sprintf("Update had an error saving template_4_your_project_name:%#v, err:%#v", *updateTemplate4ServiceName, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Update success", "template_4_your_project_nameId", template_4_your_project_nameUpdated.Id)
	return ctx.JSON(http.StatusOK, template_4_your_project_nameUpdated)
}

// ListByExternalId sends a list of template_4_your_project_names in the store as json based of the given filters
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name/by-external-id/345678912?limit=3&ofset=0' |jq
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
			list = make([]*Template4ServiceNameList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// Search returns a list of template_4_your_project_names in the store as json based of the given search criteria
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name/search?limit=3&ofset=0' |jq
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/template_4_your_project_name/search?limit=3&type=112' |jq
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
			list = make([]*Template4ServiceNameList, 0)
			return ctx.JSON(http.StatusOK, list)
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.Search :%v", err))
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// TypeTemplate4ServiceNameList sends a list of TypeTemplate4ServiceName based on the given TypeTemplate4ServiceNameListParams parameters filters
func (s Service) TypeTemplate4ServiceNameList(ctx echo.Context, params TypeTemplate4ServiceNameListParams) error {
	handlerName := "TypeTemplate4ServiceNameList"
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
	list, err := s.Store.ListTypeTemplate4ServiceName(reqCtx, offset, limit, params)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.ListTypeTemplate4ServiceName :%v", err))
		} else {
			list = make([]*TypeTemplate4ServiceNameList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// TypeTemplate4ServiceNameCreate will insert a new TypeTemplate4ServiceName in the store
func (s Service) TypeTemplate4ServiceNameCreate(ctx echo.Context) error {
	handlerName := "TypeTemplate4ServiceNameCreate"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeTemplate4ServiceNames)
	}
	newTypeTemplate4ServiceName := &TypeTemplate4ServiceName{
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
	if err := ctx.Bind(newTypeTemplate4ServiceName); err != nil {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameCreate has invalid format [%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(newTypeTemplate4ServiceName.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(newTypeTemplate4ServiceName.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN+", found %d", "name", MinNameLength, len(newTypeTemplate4ServiceName.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	//s.Log.Info("# Create() before Store.TypeTemplate4ServiceNameCreate newTemplate4ServiceName : %#v\n", newTemplate4ServiceName)
	typeTemplate4ServiceNameCreated, err := s.Store.CreateTypeTemplate4ServiceName(reqCtx, *newTypeTemplate4ServiceName)
	if err != nil {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameCreate had an error saving template_4_your_project_name:%#v, err:%#v", *newTypeTemplate4ServiceName, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("TypeTemplate4ServiceNameCreate success", "typeTemplate4ServiceNameId", typeTemplate4ServiceNameCreated.Id)
	return ctx.JSON(http.StatusCreated, typeTemplate4ServiceNameCreated)
}

func (s Service) TypeTemplate4ServiceNameCount(ctx echo.Context, params TypeTemplate4ServiceNameCountParams) error {
	handlerName := "TypeTemplate4ServiceNameCount"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// Use request context for cancellation and tracing support
	reqCtx := ctx.Request().Context()
	numTemplate4ServiceNames, err := s.Store.CountTypeTemplate4ServiceName(reqCtx, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem counting template_4_your_project_names :%v", err))
	}
	return ctx.JSON(http.StatusOK, numTemplate4ServiceNames)
}

// TypeTemplate4ServiceNameDelete will remove the given TypeTemplate4ServiceName entry from the store, and if not present will return 400 Bad Request
func (s Service) TypeTemplate4ServiceNameDelete(ctx echo.Context, typeTemplate4ServiceNameId int32) error {
	handlerName := "TypeTemplate4ServiceNameDelete"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeTemplate4ServiceNames)
	}
	reqCtx := ctx.Request().Context()
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(reqCtx, existTypeTemplate4ServiceName, typeTemplate4ServiceNameId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameDelete(%v) cannot delete this id, it does not exist !", typeTemplate4ServiceNameId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	} else {
		err := s.Store.DeleteTypeTemplate4ServiceName(reqCtx, typeTemplate4ServiceNameId, currentUserId)
		if err != nil {
			msg := fmt.Sprintf("TypeTemplate4ServiceNameDelete(%v) got an error: %#v ", typeTemplate4ServiceNameId, err)
			s.Log.Error(msg)
			return echo.NewHTTPError(http.StatusInternalServerError, msg)
		}
		return ctx.NoContent(http.StatusNoContent)
	}
}

// TypeTemplate4ServiceNameGet will retrieve the Template4ServiceName with the given id in the store and return it
func (s Service) TypeTemplate4ServiceNameGet(ctx echo.Context, typeTemplate4ServiceNameId int32) error {
	handlerName := "TypeTemplate4ServiceNameGet"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeTemplate4ServiceNames)
	}
	reqCtx := ctx.Request().Context()
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(reqCtx, existTypeTemplate4ServiceName, typeTemplate4ServiceNameId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameGet(%v) cannot retrieve this id, it does not exist !", typeTemplate4ServiceNameId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	typeTemplate4ServiceName, err := s.Store.GetTypeTemplate4ServiceName(reqCtx, typeTemplate4ServiceNameId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving TypeTemplate4ServiceName :%v", err))
	}
	return ctx.JSON(http.StatusOK, typeTemplate4ServiceName)
}

func (s Service) TypeTemplate4ServiceNameUpdate(ctx echo.Context, typeTemplate4ServiceNameId int32) error {
	handlerName := "TypeTemplate4ServiceNameUpdate"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), s.Log)
	// get the current user from JWT TOKEN
	claims := s.Server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	s.Log.Info("handler called", "handler", handlerName, "userId", currentUserId)
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	if !claims.User.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeTemplate4ServiceNames)
	}
	reqCtx := ctx.Request().Context()
	typeTemplate4ServiceNameCount, err := s.DbConn.GetQueryInt(reqCtx, existTypeTemplate4ServiceName, typeTemplate4ServiceNameId)
	if err != nil || typeTemplate4ServiceNameCount < 1 {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameUpdate(%v) cannot update this id, it does not exist !", typeTemplate4ServiceNameId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	uTypeTemplate4ServiceName := new(TypeTemplate4ServiceName)
	if err := ctx.Bind(uTypeTemplate4ServiceName); err != nil {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameUpdate has invalid format error:[%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(uTypeTemplate4ServiceName.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(uTypeTemplate4ServiceName.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN+", found %d", "name", MinNameLength, len(uTypeTemplate4ServiceName.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	uTypeTemplate4ServiceName.LastModifiedBy = &currentUserId
	template_4_your_project_nameUpdated, err := s.Store.UpdateTypeTemplate4ServiceName(reqCtx, typeTemplate4ServiceNameId, *uTypeTemplate4ServiceName)
	if err != nil {
		msg := fmt.Sprintf("TypeTemplate4ServiceNameUpdate had an error saving typeTemplate4ServiceName:%#v, err:%#v", *uTypeTemplate4ServiceName, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("TypeTemplate4ServiceNameUpdate success", "typeTemplate4ServiceNameId", template_4_your_project_nameUpdated.Id)
	return ctx.JSON(http.StatusOK, template_4_your_project_nameUpdated)
}
