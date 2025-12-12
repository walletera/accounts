package publicapi

import (
    "context"
    "fmt"
    "log/slog"
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
    err = client.Database("payments").Collection("payments").Drop(ctx)
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
