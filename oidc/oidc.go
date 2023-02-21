package oidcverify

import (
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/hadrienk/connect-go-interceptors/common"
	"net/http"
	"strings"
)
import "github.com/coreos/go-oidc/v3/oidc"

func NewOIDCInterceptor(verifier *oidc.IDTokenVerifier, handler func(token *oidc.IDToken) error) connect.Interceptor {
	return common.ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
		rawToken := header.Get("Authorization")
		rawToken = strings.TrimPrefix(rawToken, "Bearer")
		rawToken = strings.TrimSpace(rawToken)

		idToken, err := verifier.Verify(ctx, rawToken)
		if err != nil {
			return ctx, err
		}
		return ctx, handler(idToken)
	})
}
