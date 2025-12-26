package app

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/ogen-go/ogen/middleware"
    "github.com/walletera/accounts/internal/adapters/input/http/public"
    "github.com/walletera/accounts/internal/adapters/mongodb"
    "github.com/walletera/accounts/pkg/logattr"
    "github.com/walletera/accounts/publicapi"
    "go.mongodb.org/mongo-driver/v2/mongo/options"

    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
)

type App struct {
    mongodbURL        string
    mongoClient       *mongo.Client
    publicAPIConfig   Optional[PublicAPIConfig]
    logHandler        slog.Handler
    logger            *slog.Logger
    httpServersToStop []*http.Server
}

func NewApp(opts ...Option) (*App, error) {
    app := &App{}
    err := setDefaultOpts(app)
    if err != nil {
        return nil, fmt.Errorf("failed setting default options: %w", err)
    }
    for _, opt := range opts {
        opt(app)
    }
    return app, nil
}

func (app *App) Run(ctx context.Context) error {
    app.logger = slog.
        New(app.logHandler).
        With(logattr.ServiceName("accounts"))

    app.logger.Info("accounts started")

    var httpServersToStop []*http.Server

    if app.publicAPIConfig.Set {
        publicApiHttpServer, err := app.startPublicAPIHTTPServer(app.logger)
        if err != nil {
            return fmt.Errorf("failed starting public api http server: %w", err)
        }

        httpServersToStop = append(httpServersToStop, publicApiHttpServer)
    }
    app.httpServersToStop = httpServersToStop

    return nil
}

func (app *App) Stop(ctx context.Context) {
    err := app.mongoClient.Disconnect(context.TODO())
    if err != nil {
        app.logger.Error("error disconnecting from mongo", logattr.Error(err.Error()))
    }
    for _, httpServer := range app.httpServersToStop {
        err := httpServer.Shutdown(ctx)
        if err != nil {
            app.logger.Error("error stopping http server", logattr.Error(err.Error()))
        }
    }
    app.logger.Info("accounts stopped")
}

func setDefaultOpts(app *App) error {
    zapLogger, err := newZapLogger()
    if err != nil {
        return err
    }
    app.logHandler = zapslog.NewHandler(zapLogger.Core())
    return nil
}

func newZapLogger() (*zap.Logger, error) {
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
    zapConfig := zap.Config{
        Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
        Development:       false,
        DisableStacktrace: true,
        Sampling: &zap.SamplingConfig{
            Initial:    100,
            Thereafter: 100,
        },
        Encoding:         "json",
        EncoderConfig:    encoderConfig,
        OutputPaths:      []string{"stderr"},
        ErrorOutputPaths: []string{"stderr"},
    }
    return zapConfig.Build()
}

func (app *App) startPublicAPIHTTPServer(appLogger *slog.Logger) (*http.Server, error) {
    serverAPI := options.ServerAPI(options.ServerAPIVersion1)
    bsonOpts := &options.BSONOptions{
        UseJSONStructTags: true,
    }
    opts := options.Client().
        ApplyURI(app.mongodbURL).
        SetServerAPIOptions(serverAPI).
        SetBSONOptions(bsonOpts)

    // Create a new client and connect to the server
    client, err := mongo.Connect(opts)
    if err != nil {
        return nil, fmt.Errorf("error connecting to mongodb: %w", err)
    }
    app.mongoClient = client

    repository := mongodb.NewAccountsRepository(app.mongoClient, "accounts", "accounts")

    reqLoggingMiddleware := func(req middleware.Request, next middleware.Next) (middleware.Response, error) {
        app.logger.
            With("http-method", req.Raw.Method).
            With("http-path", req.Raw.URL.String()).
            Debug("handling request")
        return next(req)
    }

    server, err := publicapi.NewServer(
        public.NewHandler(
            repository,
            appLogger.With(logattr.Component("http.PublicAPIHandler")),
        ),
        &public.SecurityHandler{},
        publicapi.WithMiddleware(reqLoggingMiddleware),
    )
    if err != nil {
        panic(err)
    }
    httpServer := &http.Server{
        Addr:    fmt.Sprintf("0.0.0.0:%d", app.publicAPIConfig.Value.PublicAPIHttpServerPort),
        Handler: server,
    }

    go func() {
        defer appLogger.Info("http server stopped")
        if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            appLogger.Error("http server error", logattr.Error(err.Error()))
        }
    }()

    appLogger.Info("http server started")

    return httpServer, nil
}
