Feature: Create account

  Background: the accounts service is up and running
    Given a running accounts service

  Scenario: a account is created successfully
    Given an authorized walletera customer
    When  the customer sends a POST request to the endpoint /f423bd83-a401-4264-813b-83d7e4f057d6/accounts:
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
