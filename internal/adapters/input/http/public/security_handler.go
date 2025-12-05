package public

import (
    "context"

    "github.com/walletera/accounts/publicapi"
)

type SecurityHandler struct {
    //pubKey *rsa.PublicKey
}

//func NewSecurityHandler(pubKey *rsa.PublicKey) *SecurityHandler {
//    return &SecurityHandler{
//        pubKey: pubKey,
//    }
//}

func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, operationName publicapi.OperationName, t publicapi.BearerAuth) (context.Context, error) {
    //wjwt, err := auth.ParseAndValidate(t.GetToken(), s.pubKey)
    //if err != nil {
    //    return nil, err
    //}
    //if len(wjwt.UID) == 0 {
    //    return nil, fmt.Errorf("uid is missing")
    //}
    //if wjwt.State != "active" {
    //    return nil, fmt.Errorf("customer is not active")
    //}
    return ctx, nil
}
