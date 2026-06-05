// Package template_4_your_project_name provides mappers between domain types and Proto types.
// This bridges the gap between the database layer (Domain) and the API layer (Proto).
package template4gopackage

import (
	"time"

	"github.com/google/uuid"
	template_4_your_project_namev1 "github.com/your-github-account/template-4-your-project-name/gen/template_4_your_project_name/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Helper Functions
// =============================================================================

// timeToTimestamp converts a *time.Time to *timestamppb.Timestamp
func timeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// timestampToTime converts a *timestamppb.Timestamp to *time.Time
func timestampToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// stringPtr returns a pointer to the string, or nil if empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefString safely dereferences a string pointer, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// int32Ptr returns a pointer to the int32, or nil if zero
func int32Ptr(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

// derefInt32 safely dereferences an int32 pointer, returning 0 if nil
func derefInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// boolPtr returns a pointer to the bool
func boolPtr(b bool) *bool {
	return &b
}

// derefBool safely dereferences a bool pointer, returning false if nil
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// mapToStruct converts a map[string]interface{} to *structpb.Struct
func mapToStruct(m *map[string]interface{}) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(*m)
	if err != nil {
		return nil
	}
	return s
}

// structToMap converts a *structpb.Struct to *map[string]interface{}
func structToMap(s *structpb.Struct) *map[string]interface{} {
	if s == nil {
		return nil
	}
	m := s.AsMap()
	return &m
}

// statusToString converts a *Template4ServiceNameStatus to string
func statusToString(s *Template4ServiceNameStatus) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

// stringToStatus converts a string to *Template4ServiceNameStatus
func stringToStatus(s string) *Template4ServiceNameStatus {
	if s == "" {
		return nil
	}
	status := Template4ServiceNameStatus(s)
	return &status
}

// =============================================================================
// Template4ServiceName Mappers
// =============================================================================

// DomainTemplate4ServiceNameToProto converts a domain Template4ServiceName to a Proto Template4ServiceName
func DomainTemplate4ServiceNameToProto(t *Template4ServiceName) *template_4_your_project_namev1.Template4ServiceName {
	if t == nil {
		return nil
	}
	return &template_4_your_project_namev1.Template4ServiceName{
		Id:                t.Id.String(),
		TypeId:            t.TypeId,
		Name:              t.Name,
		Description:       derefString(t.Description),
		Comment:           derefString(t.Comment),
		ExternalId:        derefInt32(t.ExternalId),
		ExternalRef:       derefString(t.ExternalRef),
		BuildAt:           timeToTimestamp(t.BuildAt),
		Status:            statusToString(t.Status),
		ContainedBy:       derefString(t.ContainedBy),
		ContainedByOld:    derefInt32(t.ContainedByOld),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timeToTimestamp(t.InactivatedTime),
		InactivatedBy:     derefInt32(t.InactivatedBy),
		InactivatedReason: derefString(t.InactivatedReason),
		Validated:         derefBool(t.Validated),
		ValidatedTime:     timeToTimestamp(t.ValidatedTime),
		ValidatedBy:       derefInt32(t.ValidatedBy),
		ManagedBy:         derefInt32(t.ManagedBy),
		CreatedAt:         timeToTimestamp(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timeToTimestamp(t.LastModifiedAt),
		LastModifiedBy:    derefInt32(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timeToTimestamp(t.DeletedAt),
		DeletedBy:         derefInt32(t.DeletedBy),
		MoreData:          mapToStruct(t.MoreData),
		PosX:              t.PosX,
		PosY:              t.PosY,
	}
}

// ProtoTemplate4ServiceNameToDomain converts a Proto Template4ServiceName to a domain Template4ServiceName.
// Returns an error if UUID parsing fails.
func ProtoTemplate4ServiceNameToDomain(t *template_4_your_project_namev1.Template4ServiceName) (*Template4ServiceName, error) {
	if t == nil {
		return nil, nil
	}

	var id uuid.UUID
	var err error
	if t.Id != "" {
		id, err = uuid.Parse(t.Id)
		if err != nil {
			return nil, err
		}
	}

	return &Template4ServiceName{
		Id:                id,
		TypeId:            t.TypeId,
		Name:              t.Name,
		Description:       stringPtr(t.Description),
		Comment:           stringPtr(t.Comment),
		ExternalId:        int32Ptr(t.ExternalId),
		ExternalRef:       stringPtr(t.ExternalRef),
		BuildAt:           timestampToTime(t.BuildAt),
		Status:            stringToStatus(t.Status),
		ContainedBy:       stringPtr(t.ContainedBy),
		ContainedByOld:    int32Ptr(t.ContainedByOld),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timestampToTime(t.InactivatedTime),
		InactivatedBy:     int32Ptr(t.InactivatedBy),
		InactivatedReason: stringPtr(t.InactivatedReason),
		Validated:         boolPtr(t.Validated),
		ValidatedTime:     timestampToTime(t.ValidatedTime),
		ValidatedBy:       int32Ptr(t.ValidatedBy),
		ManagedBy:         int32Ptr(t.ManagedBy),
		CreatedAt:         timestampToTime(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timestampToTime(t.LastModifiedAt),
		LastModifiedBy:    int32Ptr(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timestampToTime(t.DeletedAt),
		DeletedBy:         int32Ptr(t.DeletedBy),
		MoreData:          structToMap(t.MoreData),
		PosX:              t.PosX,
		PosY:              t.PosY,
	}, nil
}

// DomainTemplate4ServiceNameListToProto converts a domain Template4ServiceNameList to a Proto Template4ServiceNameList
func DomainTemplate4ServiceNameListToProto(t *Template4ServiceNameList) *template_4_your_project_namev1.Template4ServiceNameList {
	if t == nil {
		return nil
	}
	return &template_4_your_project_namev1.Template4ServiceNameList{
		Id:          t.Id.String(),
		TypeId:      t.TypeId,
		Name:        t.Name,
		Description: derefString(t.Description),
		ExternalId:  derefInt32(t.ExternalId),
		Inactivated: t.Inactivated,
		Validated:   derefBool(t.Validated),
		Status:      statusToString(t.Status),
		CreatedBy:   t.CreatedBy,
		CreatedAt:   timeToTimestamp(t.CreatedAt),
		PosX:        t.PosX,
		PosY:        t.PosY,
	}
}

// DomainTemplate4ServiceNameListSliceToProto converts a slice of domain Template4ServiceNameList to Proto Template4ServiceNameList
func DomainTemplate4ServiceNameListSliceToProto(items []*Template4ServiceNameList) []*template_4_your_project_namev1.Template4ServiceNameList {
	if items == nil {
		return nil
	}
	result := make([]*template_4_your_project_namev1.Template4ServiceNameList, len(items))
	for i, item := range items {
		result[i] = DomainTemplate4ServiceNameListToProto(item)
	}
	return result
}

// =============================================================================
// TypeTemplate4ServiceName Mappers
// =============================================================================

// DomainTypeTemplate4ServiceNameToProto converts a domain TypeTemplate4ServiceName to a Proto TypeTemplate4ServiceName
func DomainTypeTemplate4ServiceNameToProto(t *TypeTemplate4ServiceName) *template_4_your_project_namev1.TypeTemplate4ServiceName {
	if t == nil {
		return nil
	}
	return &template_4_your_project_namev1.TypeTemplate4ServiceName{
		Id:                t.Id,
		Name:              t.Name,
		Description:       derefString(t.Description),
		Comment:           derefString(t.Comment),
		ExternalId:        derefInt32(t.ExternalId),
		TableName:         derefString(t.TableName),
		GeometryType:      derefString(t.GeometryType),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timeToTimestamp(t.InactivatedTime),
		InactivatedBy:     derefInt32(t.InactivatedBy),
		InactivatedReason: derefString(t.InactivatedReason),
		ManagedBy:         derefInt32(t.ManagedBy),
		IconPath:          t.IconPath,
		CreatedAt:         timeToTimestamp(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timeToTimestamp(t.LastModifiedAt),
		LastModifiedBy:    derefInt32(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timeToTimestamp(t.DeletedAt),
		DeletedBy:         derefInt32(t.DeletedBy),
		MoreDataSchema:    mapToStruct(t.MoreDataSchema),
	}
}

// ProtoTypeTemplate4ServiceNameToDomain converts a Proto TypeTemplate4ServiceName to a domain TypeTemplate4ServiceName
func ProtoTypeTemplate4ServiceNameToDomain(t *template_4_your_project_namev1.TypeTemplate4ServiceName) *TypeTemplate4ServiceName {
	if t == nil {
		return nil
	}
	return &TypeTemplate4ServiceName{
		Id:                t.Id,
		Name:              t.Name,
		Description:       stringPtr(t.Description),
		Comment:           stringPtr(t.Comment),
		ExternalId:        int32Ptr(t.ExternalId),
		TableName:         stringPtr(t.TableName),
		GeometryType:      stringPtr(t.GeometryType),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timestampToTime(t.InactivatedTime),
		InactivatedBy:     int32Ptr(t.InactivatedBy),
		InactivatedReason: stringPtr(t.InactivatedReason),
		ManagedBy:         int32Ptr(t.ManagedBy),
		IconPath:          t.IconPath,
		CreatedAt:         timestampToTime(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timestampToTime(t.LastModifiedAt),
		LastModifiedBy:    int32Ptr(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timestampToTime(t.DeletedAt),
		DeletedBy:         int32Ptr(t.DeletedBy),
		MoreDataSchema:    structToMap(t.MoreDataSchema),
	}
}

// DomainTypeTemplate4ServiceNameListToProto converts a domain TypeTemplate4ServiceNameList to a Proto TypeTemplate4ServiceNameList
func DomainTypeTemplate4ServiceNameListToProto(t *TypeTemplate4ServiceNameList) *template_4_your_project_namev1.TypeTemplate4ServiceNameList {
	if t == nil {
		return nil
	}
	return &template_4_your_project_namev1.TypeTemplate4ServiceNameList{
		Id:           t.Id,
		Name:         t.Name,
		ExternalId:   derefInt32(t.ExternalId),
		IconPath:     t.IconPath,
		CreatedAt:    timeToTimestamp(&t.CreatedAt),
		TableName:    derefString(t.TableName),
		GeometryType: derefString(t.GeometryType),
		Inactivated:  t.Inactivated,
	}
}

// DomainTypeTemplate4ServiceNameListSliceToProto converts a slice of domain TypeTemplate4ServiceNameList to Proto
func DomainTypeTemplate4ServiceNameListSliceToProto(items []*TypeTemplate4ServiceNameList) []*template_4_your_project_namev1.TypeTemplate4ServiceNameList {
	if items == nil {
		return nil
	}
	result := make([]*template_4_your_project_namev1.TypeTemplate4ServiceNameList, len(items))
	for i, item := range items {
		result[i] = DomainTypeTemplate4ServiceNameListToProto(item)
	}
	return result
}
