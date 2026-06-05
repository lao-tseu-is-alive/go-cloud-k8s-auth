package auth

const (
	FieldCannotBeEmpty                          = "field %s cannot be empty or contain only spaces"
	FieldMinLengthIsN                           = "field %s minimum length is %d"
	FoundNum                                    = ", found %d"
	FunctionNReturnedNoResults                  = "%s returned no results "
	OnlyAdminCanManageTypeAuths = "only admin user can manage type go_cloud_auth"
	SelectFailedInNWithErrorE                   = "pgxscan.Select unexpectedly failed in %s, error : %v"
)
