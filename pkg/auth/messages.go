package auth

const (
	FieldCannotBeEmpty        = "field %s cannot be empty or contain only spaces"
	FieldMinLengthIsN         = "field %s minimum length is %d"
	OnlyAdminCanManageUsers   = "only admin users can manage other users"
	SelectFailedInNWithErrorE = "pgxscan.Select unexpectedly failed in %s, error : %v"
)
