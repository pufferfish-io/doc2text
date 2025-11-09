package auth

import (
	"context"
	"strings"

	config "doc2text/internal/presentation/config"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const (
	CtxTokenKey ctxKey = "oidc_token_raw"
	CtxAzpKey   ctxKey = "oidc_azp"
	CtxSubKey   ctxKey = "oidc_sub"
	CtxClientID ctxKey = "oidc_client_id"
)

func NewUnaryAuthInterceptor(cfg config.OIDC) (grpc.UnaryServerInterceptor, error) {
	if cfg.Issuer == "" || cfg.JWKSURL == "" || cfg.Audience == "" {
		return nil, nil
	}

	keySet := oidc.NewRemoteKeySet(context.Background(), cfg.JWKSURL)
	verifier := oidc.NewVerifier(cfg.Issuer, keySet, &oidc.Config{ClientID: cfg.Audience})

	type tokenClaims struct {
		AZP      string      `json:"azp"`
		ClientID string      `json:"client_id"`
		Sub      string      `json:"sub"`
		Scope    string      `json:"scope"`
		Aud      interface{} `json:"aud"`
		Iss      string      `json:"iss"`
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		authz := ""
		if vals := md.Get("authorization"); len(vals) > 0 {
			authz = vals[0]
		} else if vals := md.Get("Authorization"); len(vals) > 0 {
			authz = vals[0]
		}
		if authz == "" {
			return nil, status.Error(codes.Unauthenticated, "authorization header not provided")
		}
		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}
		raw := parts[1]

		idToken, err := verifier.Verify(ctx, raw)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token verification failed: %v", err)
		}

		var claims tokenClaims
		if err := idToken.Claims(&claims); err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token claims: %v", err)
		}

		if cfg.ExpectedAzp != "" && claims.AZP != cfg.ExpectedAzp {
			return nil, status.Error(codes.PermissionDenied, "invalid azp")
		}

		ctx = context.WithValue(ctx, CtxTokenKey, raw)
		if claims.Sub != "" {
			ctx = context.WithValue(ctx, CtxSubKey, claims.Sub)
		}
		if claims.AZP != "" {
			ctx = context.WithValue(ctx, CtxAzpKey, claims.AZP)
		}
		if claims.ClientID != "" {
			ctx = context.WithValue(ctx, CtxClientID, claims.ClientID)
		}

		return handler(ctx, req)
	}, nil
}
