package publicapi

import (
    "testing"

    "github.com/cucumber/godog"
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

func aRunningAccountsService() error {
    return godog.ErrPending
}

func anAuthorizedWalleteraCustomer() error {
    return godog.ErrPending
}

func theCustomerSendsAPOSTRequestToTheEndpointFBdABDEFDAccounts(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10 int, arg11 *godog.DocString) error {
    return godog.ErrPending
}

func theEndpointReturnsTheHttpStatusCode(arg1 int) error {
    return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^a running accounts service$`, aRunningAccountsService)
    ctx.Step(`^an authorized walletera customer$`, anAuthorizedWalleteraCustomer)
    ctx.Step(`^the customer sends a POST request to the endpoint \/f(\d+)bd(\d+)-a(\d+)-(\d+)-(\d+)b-(\d+)d(\d+)e(\d+)f(\d+)d(\d+)\/accounts:$`, theCustomerSendsAPOSTRequestToTheEndpointFBdABDEFDAccounts)
    ctx.Step(`^the endpoint returns the http status code (\d+)$`, theEndpointReturnsTheHttpStatusCode)
}
