// Package auth provides mappers between domain types and Proto types.
package auth

import (
	"github.com/google/uuid"
	authv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
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

// =============================================================================
// User Mappers: Domain <-> Proto
// =============================================================================

// DomainUserToProto converts a domain User to a Proto User.
func DomainUserToProto(u *User) *authv1.User {
	if u == nil {
		return nil
	}
	return &authv1.User{
		Id:          u.ID.String(),
		ExternalId:  u.AlternateAppID,
		Email:       u.Email,
		Name:        u.Name,
		AvatarUrl:   u.AvatarURL,
		Provider:    u.Provider,
		Disabled:    !u.IsActive, // is_active -> disabled (inverted)
		Roles:       u.Roles,
		CreatedAt:   timeToTimestamp(u.CreatedAt),
		LastLoginAt: timeToTimestamp(u.LastLoginAt),
	}
}

// ProtoUserToDomain converts a Proto User to a domain User.
// Returns an error if UUID parsing fails.
func ProtoUserToDomain(p *authv1.User) (*User, error) {
	if p == nil {
		return nil, nil
	}

	var id uuid.UUID
	var err error
	if p.Id != "" {
		id, err = uuid.Parse(p.Id)
		if err != nil {
			return nil, err
		}
	}

	return &User{
		ID:             id,
		AlternateAppID: p.ExternalId,
		Email:          p.Email,
		Name:           p.Name,
		AvatarURL:      p.AvatarUrl,
		Provider:       p.Provider,
		Roles:          p.Roles,
		IsActive:       !p.Disabled, // disabled -> is_active (inverted)
	}, nil
}

// =============================================================================
// UserList Mappers
// =============================================================================

// DomainUserListToProto converts a domain UserList to a Proto UserList.
func DomainUserListToProto(u *UserList) *authv1.UserList {
	if u == nil {
		return nil
	}
	return &authv1.UserList{
		Id:          u.ID.String(),
		ExternalId:  u.AlternateAppID,
		Email:       u.Email,
		Name:        u.Name,
		Disabled:    !u.IsActive,
		CreatedAt:   timeToTimestamp(u.CreatedAt),
		LastLoginAt: timeToTimestamp(u.LastLoginAt),
	}
}

// DomainUserListSliceToProto converts a slice of domain UserList to Proto.
func DomainUserListSliceToProto(items []*UserList) []*authv1.UserList {
	if items == nil {
		return nil
	}
	result := make([]*authv1.UserList, len(items))
	for i, item := range items {
		result[i] = DomainUserListToProto(item)
	}
	return result
}

// =============================================================================
// JWT Mappers: Domain -> common-libs UserInfo for JWT claims
// =============================================================================

// DomainUserToJwtUserInfo converts a domain User to a goHttpEcho.UserInfo for JWT.
// Uses AlternateAppID as the integer UserId for JWT compatibility.
// groupIDs should be pre-fetched from the user_groups table.
func DomainUserToJwtUserInfo(u *User, groupIDs []int) *goHttpEcho.UserInfo {
	if u == nil {
		return nil
	}

	isAdmin := false
	for _, role := range u.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	return &goHttpEcho.UserInfo{
		UserId:     int(u.AlternateAppID),
		ExternalId: int(u.AlternateAppID),
		Name:       u.Name,
		Email:      u.Email,
		Login:      u.Email, // use email as login for OAuth users
		IsAdmin:    isAdmin,
		Groups:     groupIDs,
	}
}
