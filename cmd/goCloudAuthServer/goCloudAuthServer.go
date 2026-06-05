package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/metadata"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/tools"
	"github.com/prometheus/client_golang/prometheus"
	"https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/gen/auth/v1/authv1connect"
	"https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/pkg/auth"
	"https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/pkg/version"
)

const (
	defaultPort                = 9090
	defaultLogName             = "stderr"
	defaultDBPort              = 5432
	defaultDBIp                = "127.0.0.1"
	defaultDBSslMode           = "prefer"
	defaultJwtStatusUrl        = "/status"
	defaultJwtCookieName       = "goJWT_token"
	defaultAppInfoUrl          = "/goAppInfo"
	defaultWebRootDir          = "goCloudAuthFront/dist/"
	defaultSqlDbMigrationsPath = "db/migrations"
	defaultSecuredApi          = "/goapi/v1"
	defaultAdminUser           = "goadmin"
	defaultAdminEmail          = "goadmin@yourdomain.org"
	defaultAdminId             = 960901
	charsetUTF8                = "charset=UTF-8"
	MIMEAppJSON                = "application/json"
	MIMEHtml                   = "text/html"
	MIMEHtmlCharsetUTF8        = MIMEHtml + "; " + charsetUTF8
	MIMEAppJSONCharsetUTF8     = MIMEAppJSON + "; " + charsetUTF8
)

// content holds our static web server content.
//
//go:embed goCloudAuthFront/dist/*
var content embed.FS

// sqlMigrations holds our db migrations sql files using https://github.com/golang-migrate/migrate
// in the line above you SHOULD have the same path  as const defaultSqlDbMigrationsPath
//
//go:embed db/migrations/*.sql
var sqlMigrations embed.FS

// UserLogin defines model for UserLogin.
type UserLogin struct {
	PasswordHash string `json:"password_hash"`
	Username     string `json:"username"`
}

type Service struct {
	Logger        *slog.Logger
	dbConn        database.DB
	server        *goHttpEcho.Server
	jwtCookieName string
}

// login is just a trivial example to test this server
// you should use the jwt token returned from LoginUser  in github.com/lao-tseu-is-alive/go-cloud-k8s-user-group'
// and share the same secret with the above component
func (s *Service) login(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("login", ctx.Request(), s.Logger)
	uLogin := new(UserLogin)
	login := ctx.FormValue("login")
	passwordHash := ctx.FormValue("hashed")
	s.Logger.Debug("login: %s, hash: %s ", login, passwordHash)
	// maybe it was not a form but a fetch data post
	if len(strings.Trim(login, " ")) < 1 {
		if err := ctx.Bind(uLogin); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user login or json format in request body")
		}
	} else {
		uLogin.Username = login
		uLogin.PasswordHash = passwordHash
	}
	s.Logger.Debug("About to check username: %s , password: %s", uLogin.Username, uLogin.PasswordHash)

	reqCtx := ctx.Request().Context()
	if s.server.Authenticator.AuthenticateUser(reqCtx, uLogin.Username, uLogin.PasswordHash) {
		userInfo, err := s.server.Authenticator.GetUserInfoFromLogin(reqCtx, login)
		if err != nil {
			myErrMsg := fmt.Sprintf("Error getting user info from login: %v", err)
			s.Logger.Error(myErrMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
		}
		token, err := s.server.JwtCheck.GetTokenFromUserInfo(userInfo)
		if err != nil {
			myErrMsg := fmt.Sprintf("Error getting jwt token from user info: %v", err)
			s.Logger.Error(myErrMsg)
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
		}
		// Prepare the response
		response := map[string]string{
			"jwtStatus": "success",
			"token":     token.String(),
		}
		s.Logger.Info("LoginUser() successful", "login", login)
		return ctx.JSON(http.StatusOK, response)
	} else {
		myErrMsg := "username not found or password invalid"
		s.Logger.Warn(myErrMsg)
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"jwtStatus": myErrMsg, "token": ""})
	}
}

func (s *Service) GetStatus(ctx echo.Context) error {
	goHttpEcho.TraceHttpRequest("GetStatus", ctx.Request(), s.Logger)
	// get the current user from JWT TOKEN
	claims := s.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	s.Logger.Info("in restricted : ", "currentUserId", currentUserId)
	// you can check if the user is not active anymore and RETURN 401 Unauthorized
	//if !s.Store.IsUserActive(currentUserId) {
	//	return echo.NewHTTPError(http.StatusUnauthorized, "current calling user is not active anymore")
	//}
	return ctx.JSON(http.StatusOK, claims)
}

func (s *Service) IsDBAlive() bool {
	dbVer, err := s.dbConn.GetVersion(context.Background())
	if err != nil {
		return false
	}
	if len(dbVer) < 2 {
		return false
	}
	return true
}

func (s *Service) checkReady(string) bool {
	// we decide what makes us ready, is a valid  connection to the database
	if !s.IsDBAlive() {
		return false
	}
	return true
}

func checkHealthy(string) bool {
	// you decide what makes you ready, may be it is the connection to the database
	//if !IsDBAlive() {
	//	return false
	//}
	return true
}

func initMetadataOrFail(db database.DB, l *slog.Logger) {
	// checking metadata information
	metadataService := metadata.Service{Log: l, Db: db}
	metaDataCtx, metaDataCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer metaDataCancel()
	metadataService.CreateMetadataTableOrFail(metaDataCtx)
	found, ver := metadataService.GetServiceVersionOrFail(metaDataCtx, version.AppName)
	if found {
		l.Info("retrieved service", "app", version.AppName, "version", ver, "status", "found")
	} else {
		l.Info("impossible to retrieved service", "app", version.AppName, "version", ver, "status", "not found")
	}
	metadataService.SetServiceVersionOrFail(metaDataCtx, version.AppName, version.Version)
}

func runMigrationsOrFail(dbDsn string) {
	// begin section go-migrate db migration with embed files in go program
	// https://github.com/golang-migrate/migrate
	d, err := iofs.New(sqlMigrations, defaultSqlDbMigrationsPath)
	if err != nil {
		log.Fatalf("💥💥 error doing iofs.New for db migrations  error: %v\n", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, strings.Replace(dbDsn, "postgres", "pgx5", 1))
	if err != nil {
		log.Fatalf("💥💥 error doing migrate.NewWithSourceInstance(iofs, dbURL:%s)  error: %v\n", dbDsn, err)
	}

	err = m.Up()
	if err != nil {
		//if err == m.
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("💥💥 error doing migrate.Up error: %v\n", err)
		}
	}
	// end section go-migrate db migration with embed files in go program
}

func main() {
	logWriter, err := config.GetLogWriter(defaultLogName)
	if err != nil {
		log.Fatalf("💥💥 error getting log writer: %v'\n", err)
	}
	logLevel, err := config.GetLogLevel(golog.InfoLevel)
	if err != nil {
		log.Fatalf("💥💥 error getting log level: %v'\n", err)
	}
	l := golog.NewLogger("simple", logWriter, logLevel, version.AppName)
	l.Info("🚀 Starting", "app", version.AppName, "version", version.Version, "revision", version.Revision, "build", version.BuildStamp, "repository", version.REPOSITORY)

	dbDsn, err := config.GetPgDbDsnUrl(defaultDBIp, defaultDBPort, tools.ToSnakeCase(version.AppName), version.AppNameSnake, defaultDBSslMode)
	if err != nil {
		l.Error("💥💥 error getting database DSN", "error", err)
		os.Exit(1)
	}
	dbConnCtx, dbConnCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbConnCancel()
	db, err := database.GetInstance(dbConnCtx, "pgx", dbDsn, runtime.NumCPU(), l)
	if err != nil {
		l.Error("💥💥 error doing database.GetInstance", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	dbVersion, err := db.GetVersion(context.Background())
	if err != nil {
		l.Error("💥💥 error doing dbConn.GetVersion", "error", err)
		os.Exit(1)
	}
	l.Info("connected to db", "version", dbVersion)

	initMetadataOrFail(db, l)
	runMigrationsOrFail(dbDsn)

	// Get the ENV JWT_AUTH_URL value
	jwtAuthUrl, err := config.GetJwtAuthUrl()
	if err != nil {
		l.Error("💥💥 error getting JWT auth URL", "error", err)
		os.Exit(1)
	}
	jwtStatusUrl := config.GetJwtStatusUrl(defaultJwtStatusUrl)

	myVersionReader := goHttpEcho.NewSimpleVersionReader(
		version.AppName,
		version.Version,
		version.Repository,
		version.Revision,
		version.BuildStamp,
		jwtAuthUrl,
		jwtStatusUrl,
	)
	// Create a new JWT checker
	myJwt, err := goHttpEcho.GetNewJwtCheckerFromConfig(version.AppName, 60, l)
	if err != nil {
		l.Error("💥💥 error creating JWT checker", "error", err)
		os.Exit(1)
	}
	// Create a new Authenticator using factory function
	myAuthenticator, err := goHttpEcho.GetSimpleAdminAuthenticatorFromConfig(
		goHttpEcho.AdminDefaults{
			UserId:     defaultAdminId,
			ExternalId: 9999999,
			Login:      defaultAdminUser,
			Email:      defaultAdminEmail,
		},
		myJwt,
	)
	if err != nil {
		l.Error("💥💥 error creating authenticator", "error", err)
		os.Exit(1)
	}

	server, err := goHttpEcho.CreateNewServerFromEnv(
		defaultPort,
		"0.0.0.0", // defaultServerIp,
		&goHttpEcho.Config{
			ListenAddress: "",
			Authenticator: myAuthenticator,
			JwtCheck:      myJwt,
			VersionReader: myVersionReader,
			Logger:        l,
			WebRootDir:    defaultWebRootDir,
			Content:       content,
			RestrictedUrl: defaultSecuredApi,
		},
	)
	if err != nil {
		l.Error("💥💥 error creating server", "error", err)
		os.Exit(1)
	}

	cookieNameForJWT := config.GetJwtCookieName(defaultJwtCookieName)
	yourService := Service{
		Logger:        l,
		dbConn:        db,
		server:        server,
		jwtCookieName: cookieNameForJWT,
	}

	e := server.GetEcho()
	//e.Use(goHttpEcho.CookieToHeaderMiddleware(yourService.jwtCookieName, l))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://golux.lausanne.ch", "http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// begin prometheus stuff to create a custom counter metric
	customCounter := prometheus.NewCounter( // create new counter metric. This is replacement for `prometheus.Metric` struct
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_custom_requests_total", version.AppName),
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
	)
	if err := prometheus.Register(customCounter); err != nil { // register your new counter metric with default metrics registry
		l.Error("💥💥 error calling prometheus register", "error", err)
		os.Exit(1)
	}
	// https://echo.labstack.com/docs/middleware/prometheus
	mwConfig := echoprometheus.MiddlewareConfig{
		AfterNext: func(c echo.Context, err error) {
			customCounter.Inc() // use our custom metric in middleware. after every request increment the counter
		},
		// does not gather metrics on routes starting with `/health`
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/health")
		},
		Subsystem: version.AppName,
	}
	e.Use(echoprometheus.NewMiddlewareWithConfig(mwConfig)) // adds middleware to gather metrics
	// end prometheus stuff to create a custom counter metric

	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics
	e.GET("/readiness", server.GetReadinessHandler(yourService.checkReady, "Connection to DB"))
	e.GET("/health", server.GetHealthHandler(checkHealthy, "Connection to DB"))
	e.GET(defaultAppInfoUrl, server.GetAppInfoHandler())
	// Find a way to allow Login route to be available only in dev environment
	e.POST(jwtAuthUrl, yourService.login)
	// Call the DevRoutes function conditionally
	// This line will only compile if the 'dev' build tag is active.
	// Conditional compilation of dev routes

	if IsDevBuild {
		l.Info("Attempting to register dev routes...")
		DevRoutes(e, &yourService, jwtAuthUrl)
	}
	r := server.GetRestrictedGroup()
	r.GET(jwtStatusUrl, yourService.GetStatus)

	dbStorageCtx, dbStorageCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbStorageCancel()
	authStore := auth.GetStorageInstanceOrPanic(dbStorageCtx, "pgx", db, l)

	// Create business service (transport-agnostic)
	authBusinessService := auth.NewBusinessService(authStore, db, l, 50)

	// ---------------------------------------------------------
	// Connect + Vanguard: REST/gRPC/Connect transcoding
	// ---------------------------------------------------------
	// Create auth interceptor for JWT validation
	authInterceptor := auth.NewAuthInterceptor(myJwt, l)
	interceptors := connect.WithInterceptors(authInterceptor)

	// Create Connect servers (auth is handled by interceptor, not servers)
	authConnectServer := auth.NewAuthConnectServer(authBusinessService, l)
	typegoCloudAuthConnectServer := auth.NewTypeAuthConnectServer(authBusinessService, l)

	// Create service handlers with auth interceptor
	_, authHandler := authv1connect.NewAuthServiceHandler(authConnectServer, interceptors)
	_, typegoCloudAuthHandler := authv1connect.NewTypeAuthServiceHandler(typegoCloudAuthConnectServer, interceptors)

	// Create Vanguard services for HTTP transcoding
	authService := vanguard.NewService(
		authv1connect.goCloudAuthServiceName,
		authHandler,
	)
	typegoCloudAuthService := vanguard.NewService(
		authv1connect.TypegoCloudAuthServiceName,
		typegoCloudAuthHandler,
	)

	// Create transcoder for REST + gRPC + Connect
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{
		authService,
		typegoCloudAuthService,
	})
	if err != nil {
		l.Error("💥💥 error failed to create vanguard transcoder", "error", err)
		os.Exit(1)
	}

	// Mount transcoder into Echo with /goapi/v1 prefix
	// The transcoder handles REST endpoints defined in proto:
	// - GET /auth, POST /auth, etc. (defined in proto HTTP annotations)
	// - Connect endpoints: /auth.v1.goCloudAuthService/*
	//
	// We strip the /goapi/v1 prefix before passing to transcoder
	transcoderWithPrefix := http.StripPrefix(defaultSecuredApi, transcoder)

	e.Any(defaultSecuredApi+"/auth*", echo.WrapHandler(transcoderWithPrefix))
	e.Any(defaultSecuredApi+"/types*", echo.WrapHandler(transcoderWithPrefix))

	// Connect RPC endpoints (no prefix stripping needed)
	e.Any("/auth.v1.*", echo.WrapHandler(transcoder))

	l.Info("🚀 Connect+Vanguard handlers mounted for REST/gRPC transcoding", "securedUrl", defaultSecuredApi)

	err = server.StartServer()
	if err != nil {
		l.Error("💥💥 error starting server", "error", err)
		os.Exit(1)
	}

}
