package embeddable

import "context"

type authPrincipalContextKey struct{}

func WithAuthPrincipal(ctx context.Context, principal AuthPrincipal) context.Context {
	return context.WithValue(ctx, authPrincipalContextKey{}, principal)
}

func GetAuthPrincipal(ctx context.Context) (AuthPrincipal, bool) {
	if ctx == nil {
		return AuthPrincipal{}, false
	}
	principal, ok := ctx.Value(authPrincipalContextKey{}).(AuthPrincipal)
	return principal, ok
}
