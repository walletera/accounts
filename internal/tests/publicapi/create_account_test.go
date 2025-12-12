package publicapi

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "strings"
    "testing"

    "github.com/cucumber/godog"
    "github.com/walletera/accounts/internal/app"
)

const (
    createAccountResponseStatusCodeKey = "createAccountResponseStatusCodeKey"
    createAccountResponseBodyKey       = "createAccountResponseBodyKey"
)

func TestCreateAccount(t *testing.T) {

    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features/create_account.feature"},
            TestingT: t, // Testing instance that will run subtests.
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
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

func theCustomerSendsTheFollowingRequestToTheEndpointCustomerIdAccounts(ctx context.Context, requestBody *godog.DocString) (context.Context, error) {
    if requestBody.Content == "" {
        return ctx, fmt.Errorf("request body is empty")
    }
    url := fmt.Sprintf("http://127.0.0.1:%d/accounts", publicApiHttpServerPort)
    bodyReader := strings.NewReader(requestBody.Content)
    request, err := http.NewRequest(http.MethodPost, url, bodyReader)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    request.Header.Set("Content-Type", "application/json")
    request.Header.Set("Authorization", "Bearer ajsonwebtoken")
    resp, err := http.DefaultClient.Do(request)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }

    defer func(Body io.ReadCloser) {
        err := Body.Close()
        if err != nil {
            panic(err)
        }
    }(resp.Body)

    ctx = context.WithValue(ctx, createAccountResponseStatusCodeKey, resp.StatusCode)
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed reading response body: %w", err)
    }
    ctx = context.WithValue(ctx, createAccountResponseBodyKey, body)

    return ctx, nil
}

func theAccountsServiceProducesTheFollowingLog(ctx context.Context, logMsg string) (context.Context, error) {
    logsWatcher := logsWatcherFromCtx(ctx)
    foundLogEntry := logsWatcher.WaitFor(logMsg, logsWatcherWaitForTimeout)
    if !foundLogEntry {
        return ctx, fmt.Errorf("didn't find expected log entry")
    }
    return ctx, nil
}

func theEndpointReturnsTheHttpStatusCode(ctx context.Context, expectedStatusCode int) (context.Context, error) {
    statusCode := createAccountResponseStatusCodeFromCtx(ctx)
    if statusCode != expectedStatusCode {
        body := createAccountResponseBodyFromCtx(ctx)
        return ctx, fmt.Errorf("expected http status code %d but got %d -- response body is %s", expectedStatusCode, statusCode, body)
    }
    return ctx, nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Before(beforeScenarioHook)
    ctx.Step(`^a running accounts service$`, aRunningAccountsService)
    ctx.Step(`^an authorized walletera customer$`, anAuthorizedWalleteraCustomer)
    ctx.Step(`^the accounts service receives the following request on the endpoint /accounts:$`, theCustomerSendsTheFollowingRequestToTheEndpointCustomerIdAccounts)
    ctx.Step(`^the endpoint returns the http status code (\d+)$`, theEndpointReturnsTheHttpStatusCode)
    ctx.Step(`^the accounts service produces the following log$`, theAccountsServiceProducesTheFollowingLog)
    ctx.After(afterScenarioHook)
}

func createAccountResponseStatusCodeFromCtx(ctx context.Context) int {
    value := ctx.Value(createAccountResponseStatusCodeKey)
    if value == nil {
        panic("createAccountResponseStatusCodeKey not found in context")
    }
    createAccountResponseStatusCode, ok := value.(int)
    if !ok {
        panic("createAccountResponseStatusCodeKey has invalid type")
    }
    return createAccountResponseStatusCode
}

func createAccountResponseBodyFromCtx(ctx context.Context) []byte {
    value := ctx.Value(createAccountResponseBodyKey)
    if value == nil {
        panic("createAccountResponseStatusCodeKey not found in context")
    }
    createAccountResponseStatusCode, ok := value.([]byte)
    if !ok {
        panic("createAccountResponseStatusCodeKey has invalid type")
    }
    return createAccountResponseStatusCode
}
