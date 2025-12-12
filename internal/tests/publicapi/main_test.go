package publicapi

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

const (
    containersStartTimeout = 60 * time.Second
)

func TestMain(m *testing.M) {
    ctx, cancelCtx := context.WithTimeout(context.Background(), containersStartTimeout)
    defer cancelCtx()

    stopMongo, err := startMongoDBContainer(ctx)
    if err != nil {
        panic(err)
    }
    defer func() {
        err = stopMongo()
        if err != nil {
            panic(err)
        }
    }()

    m.Run()
}

const (
    startupTimeout               = 10 * time.Second
    containersTerminationTimeout = 20 * time.Second
)

func startMongoDBContainer(ctx context.Context) (func() error, error) {
    req := testcontainers.ContainerRequest{
        Image:        "mongodb/mongodb-community-server",
        Name:         "mongodb",
        ExposedPorts: []string{"27017:27017"},
        WaitingFor:   wait.NewExecStrategy([]string{"mongosh", "--eval", "show dbs"}).WithStartupTimeout(startupTimeout),
        LogConsumerCfg: &testcontainers.LogConsumerConfig{
            Consumers: []testcontainers.LogConsumer{NewContainerLogConsumer("mongodb")},
        },
    }
    mongodbC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("error creating mongodb container: %w", err)
    }

    return func() error {
        terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), containersTerminationTimeout)
        defer terminationCtxCancel()
        return mongodbC.Terminate(terminationCtx)
    }, nil
}
