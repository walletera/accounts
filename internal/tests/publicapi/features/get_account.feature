Feature: Get Account by Id

  Background: the accounts service is up and running
    Given a running accounts service
    And a list of existing accounts:
    """
    data/existing_accounts.json
    """

  Scenario Outline: an account can be retrieved using different query parameters
    When the accounts service receives a GET request on endpoint /accounts with filters <filters>
    Then the endpoint returns the http status code <statusCode>


    Examples:
      | accountId                                | statusCode |
      | ?id=bdf48329-d870-4fb4-882a-0fa0aef28a63 | 200        |
      | ?cvu=6677889900112233445566              | 200        |
      | ?dinopayAccountNumber=DP-123456789       | 200        |