Feature: Create account

  Background: the accounts service is up and running
    Given a running accounts service

  Scenario: an account is created successfully
    Given an authorized walletera customer
    When  the accounts service receives the following request on the endpoint /accounts:
    """json
    {
      "id": "bdf48329-d870-4fb4-882a-0fa0aef28a63",
      "customerId": "f423bd83-a401-4264-813b-83d7e4f057d6",
      "currency": "ARS",
      "institutionName": "dinopay",
      "institutionId": "dinopay",
      "accountDetails": {
        "accountType": "cvu",
        "cuit": "23679876453",
        "routingInfo": {
          "cvuRoutingInfoType": "cvu",
          "cvu": "1122334455667788554433"
        }
      }
    }
    """
    Then the endpoint returns the http status code 201
    And the accounts service produces the following log
    """
    account saved
    """
