package version

var (
	// EDIT THESE VALUES AFTER CREATING THE REPO FROM TEMPLATE
	// -----------------------------------------------------

	// AppName  is the CamelCase name of your app (e.g., "User", "Product")
	AppName = "goCloudAuth"

	// GoPackage  is the name of your main service go package (e.g., "user", "product")
	// should be: all lowercase, short no hyphens, no underscores, no camelCase, usually one word
	GoPackage = "auth"

	// ServiceName is the name of your main entity/service first letter Capital (e.g., "User", "Product")
	ServiceName = "Auth"

	// DbSchemaName is the name of your main entity/service database schema can be the same as go package
	DbSchemaName = "go_auth_db_schema"

	// AppNameKebab is the kebab-case version for your github repository (e.g., "user", "product")
	AppNameKebab = "go-cloud-auth"

	// AppNameSnake is the snake-case version for database or directory (e.g., "user", "product")
	AppNameSnake = "go_cloud_auth"

	// Repository is the full GitHub repo URL
	Repository = "https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth"

	// Version starting point
	Version = "0.0.2"

	// Revision is auto-filled by build (do not edit manually)
	Revision = "unknown"
	// BuildStamp is auto-filled by build (do not edit manually)
	BuildStamp = "unknown"
)
