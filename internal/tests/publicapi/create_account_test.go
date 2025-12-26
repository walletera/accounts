package publicapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

const (
	responseStatusCodeContextKey = "responseStatusCodeContextKey"
	createAccountResponseBodyKey = "createAccountResponseBodyKey"
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

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(beforeScenarioHook)
	ctx.Step(`^a running accounts service$`, aRunningAccountsService)
	ctx.Step(`^an authorized walletera customer$`, anAuthorizedWalleteraCustomer)
	ctx.Step(`^the accounts service receives the following request on the endpoint /accounts:$`, theAccountsServiceReceivesAPostRequestOnEndpointAccountWithTheFollowingBody)
	ctx.Step(`^the endpoint returns the http status code (\d+)$`, theEndpointReturnsTheHttpStatusCode)
	ctx.Step(`^the accounts service produces the following log$`, theAccountsServiceProducesTheFollowingLog)
	ctx.After(afterScenarioHook)
}

func theAccountsServiceReceivesAPostRequestOnEndpointAccountWithTheFollowingBody(ctx context.Context, requestBody *godog.DocString) (context.Context, error) {
	if requestBody.Content == "" {
		return ctx, fmt.Errorf("request body is empty")
	}

	return createAccount(ctx, requestBody.Content)
}

func theEndpointReturnsTheHttpStatusCode(ctx context.Context, expectedStatusCode int) (context.Context, error) {
	statusCode := responseStatusCodeFromCtx(ctx)
	if statusCode != expectedStatusCode {
		return ctx, fmt.Errorf("expected http status code %d but got %d", expectedStatusCode, statusCode)
	}
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
