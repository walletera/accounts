package publicapi

import (
    "context"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "strings"
    "time"

    "github.com/walletera/accounts/internal/app"

    "github.com/cucumber/godog"
    slogwatcher "github.com/walletera/logs-watcher/slog"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
    "go.uber.org/zap/zapcore"
)

const (
    appKey                    = "app"
    appCtxCancelFuncKey       = "appCtxCancelFuncKey"
    logsWatcherKey            = "logsWatcher"
    logsWatcherWaitForTimeout = 5 * time.Second
    publicApiHttpServerPort   = 8484
    mongodbURL                = "mongodb://localhost:27017/?retryWrites=true&w=majority"
)

var mongodbClient *mongo.Client

func beforeScenarioHook(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
    handler, err := newZapHandler()
    if err != nil {
        return ctx, err
    }
    logsWatcher := slogwatcher.NewWatcher(handler)
    ctx = context.WithValue(ctx, logsWatcherKey, logsWatcher)

    client, err := getMongodbClient()
    if err != nil {
        return ctx, err
    }

    // cleanup database before each scenario
    err = client.Database("accounts").Collection("accounts").Drop(ctx)
    if err != nil {
        return nil, err
    }

    return ctx, nil
}

func afterScenarioHook(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
    logsWatcher := logsWatcherFromCtx(ctx)

    appFromCtx(ctx).Stop(ctx)
    foundLogEntry := logsWatcher.WaitFor("accounts stopped", logsWatcherWaitForTimeout)
    if !foundLogEntry {
        return ctx, fmt.Errorf("app termination failed (didn't find expected log entry)")
    }

    err = logsWatcher.Stop()
    if err != nil {
        return ctx, fmt.Errorf("failed stopping the logsWatcher: %w", err)
    }

    return ctx, nil
}

func aRunningAccountsService(ctx context.Context) (context.Context, error) {
    logHandler := logsWatcherFromCtx(ctx).DecoratedHandler()

    appCtx, appCtxCancelFunc := context.WithCancel(ctx)

    paymentsRMApp, err := app.NewApp(
        app.WithPublicAPIConfig(app.PublicAPIConfig{
            PublicAPIHttpServerPort: publicApiHttpServerPort,
        }),
        app.WithMongoDBURL(mongodbURL),
        app.WithLogHandler(logHandler),
    )
    if err != nil {
        appCtxCancelFunc()
        return ctx, fmt.Errorf("failed initializing paymentsRMApp: %s", err.Error())
    }

    err = paymentsRMApp.Run(appCtx)
    if err != nil {
        appCtxCancelFunc()
        return ctx, fmt.Errorf("failed running accounts app: %s", err.Error())
    }

    ctx = context.WithValue(ctx, appKey, paymentsRMApp)
    ctx = context.WithValue(ctx, appCtxCancelFuncKey, appCtxCancelFunc)

    foundLogEntry := logsWatcherFromCtx(ctx).WaitFor("accounts started", logsWatcherWaitForTimeout)
    if !foundLogEntry {
        return ctx, fmt.Errorf("accounts app startup failed (didn't find expected log entry)")
    }

    return ctx, nil
}

func anAuthorizedWalleteraCustomer(ctx context.Context) (context.Context, error) {
    // TODO
    return ctx, nil
}

func createAccount(ctx context.Context, requestBody string) (context.Context, error) {
    url := fmt.Sprintf("http://127.0.0.1:%d/accounts", publicApiHttpServerPort)
    bodyReader := strings.NewReader(requestBody)
    request, err := http.NewRequest(http.MethodPost, url, bodyReader)
    if err != nil {
        return ctx, fmt.Errorf("failed to create request: %w", err)
    }

    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Authorization", "Bearer ajsonwebtoken")
    resp, err := http.DefaultClient.Do(request)
    if err != nil {
        return ctx, fmt.Errorf("failed to send request: %w", err)
    }

    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            panic(err)
        }
    }(resp.Body)

    ctx = context.WithValue(ctx, responseStatusCodeContextKey, resp.StatusCode)
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return ctx, fmt.Errorf("failed reading response body: %w", err)
    }
    ctx = context.WithValue(ctx, createAccountResponseBodyKey, body)
    return ctx, nil
}

func responseStatusCodeFromCtx(ctx context.Context) int {
    value := ctx.Value(responseStatusCodeContextKey)
    if value == nil {
        panic("responseStatusCodeContextKey not found in context")
    }
    createAccountResponseStatusCode, ok := value.(int)
    if !ok {
        panic("responseStatusCodeContextKey has invalid type")
    }
    return createAccountResponseStatusCode
}

func logsWatcherFromCtx(ctx context.Context) *slogwatcher.Watcher {
    value := ctx.Value(logsWatcherKey)
    if value == nil {
        panic("logs watcher not found in context")
    }
    watcher, ok := value.(*slogwatcher.Watcher)
    if !ok {
        panic("logs watcher has invalid type")
    }
    return watcher
}

func appFromCtx(ctx context.Context) *app.App {
    value := ctx.Value(appKey)
    if value == nil {
        panic("paymentsRMApp not found in context")
    }
    paymentsRMApp, ok := value.(*app.App)
    if !ok {
        panic("paymentsRMApp has invalid type")
    }
    return paymentsRMApp
}

func newZapHandler() (slog.Handler, error) {
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
    zapLogger, err := zapConfig.Build()
    if err != nil {
        return nil, err
    }
    if zapLogger.Core() == nil {
        return nil, fmt.Errorf("zapLogger.Core() is nil")
    }
    return zapslog.NewHandler(zapLogger.Core()), nil
}

func getMongodbClient() (*mongo.Client, error) {
    if mongodbClient != nil {
        return mongodbClient, nil
    }

    MongodbUri := "mongodb://localhost:27017/?retryWrites=true&w=majority"

    // Use the SetServerAPIOptions() method to set the Stable API version to 1
    serverAPI := options.ServerAPI(options.ServerAPIVersion1)
    opts := options.Client().ApplyURI(MongodbUri).SetServerAPIOptions(serverAPI)

    // Create a new client and connect to the server
    mongodbClient, err := mongo.Connect(opts)
    if err != nil {
        return nil, err
    }

    return mongodbClient, nil
}
