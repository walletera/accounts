package publicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/cucumber/godog"
	"github.com/walletera/accounts/publicapi"
)

const (
	getAccountOkKey = "getPayment"
)

func TestGetAccount(t *testing.T) {

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeGetPaymentFeature,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/get_account.feature"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeGetPaymentFeature(ctx *godog.ScenarioContext) {
	ctx.Before(beforeScenarioHook)
	ctx.Given(`^a running accounts service$`, aRunningAccountsService)
	ctx.Given(`^a list of existing accounts:$`, aListOfExistingAccounts)
	ctx.When(`^the accounts service receives a GET request on endpoint \/accounts with filters (.+)$`, theAccountsServiceReceivesAGETRequestOnEndpointAccountsWithFilters)
	ctx.Step(`^the endpoint returns the http status code (\d+)$`, theEndpointReturnsTheHttpStatusCode)
	ctx.After(afterScenarioHook)
}

func theAccountsServiceReceivesAGETRequestOnEndpointAccountsWithFilters(ctx context.Context, filters string) (context.Context, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/accounts%s", publicApiHttpServerPort, filters)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
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

	ctx = context.WithValue(ctx, responseStatusCodeContextKey, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("expected http status code %d but got %d", http.StatusOK, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed reading response body: %w", err)
	}
	var accountList publicapi.ListAccountsOKApplicationJSON
	err = json.Unmarshal(respBody, &accountList)
	if err != nil {
		return ctx, fmt.Errorf("failed to decode response: %w -- %s", err, respBody)
	}

	if len(accountList) > 1 {
		return ctx, fmt.Errorf("expected response to contain no more than 1 account, but got %d", len(accountList))
	}

	if len(accountList) == 1 {
		ctx = context.WithValue(ctx, getAccountOkKey, accountList[0])
	}

	return ctx, nil
}
