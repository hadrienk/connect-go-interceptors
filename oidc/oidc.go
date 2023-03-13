package oidcverify

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/hadrienk/connect-go-interceptors/common"
	"net/http"
	"strings"
)
import "github.com/coreos/go-oidc/v3/oidc"

type option func(token *oidc.IDToken) error

type oidcTokenKey struct{}

func GetToken(ctx context.Context) (*oidc.IDToken, bool) {
	value := ctx.Value(oidcTokenKey{})
	if idToken, ok := value.(*oidc.IDToken); ok {
		return idToken, true
	}
	return nil, false
}

// Validate that the given role is present in the claims of the oidc.IDToken.
func WithRole(role string) option {
	return WithHandler(func(token *oidc.IDToken) error {
		var claims struct {
			Roles []string `json:"roles"`
		}
		if err := token.Claims(&claims); err != nil {
			return connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("no roles claim", err))
		}
		for _, r := range claims.Roles {
			if strings.TrimSpace(r) == strings.TrimSpace(role) {
				return nil
			}
		}
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("missing role %s", role))
	})
}

// Add a custom handler that gets called with the validated oidc.IDToken
func WithHandler(handler func(token *oidc.IDToken) error) option {
	return option(handler)
}

func NewOIDCInterceptor(verifier *oidc.IDTokenVerifier, options ...option) connect.Interceptor {
	return common.ContextHeaderInterceptor(func(ctx context.Context, header http.Header) (context.Context, error) {
		rawToken := header.Get("Authorization")
		rawToken = strings.TrimPrefix(rawToken, "Bearer")
		rawToken = strings.TrimSpace(rawToken)
		idToken, err := verifier.Verify(ctx, rawToken)
		if err != nil {
			return ctx, connect.NewError(connect.CodeUnauthenticated, err)
		}
		for _, option := range options {
			err := option(idToken)
			if err != nil {
				return ctx, err
			}
		}
		return context.WithValue(ctx, oidcTokenKey{}, idToken), nil
	})
}
