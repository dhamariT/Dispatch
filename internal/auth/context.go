package auth

import (
	"context"

	"github.com/dhamariT/dispatch/internal/db"
)

type ctxKey int

const (
	userKey ctxKey = iota
	orgKey
	membershipKey
)

func WithUser(ctx context.Context, u db.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func UserFromCtx(ctx context.Context) (db.User, bool) {
	u, ok := ctx.Value(userKey).(db.User)
	return u, ok
}

func WithOrg(ctx context.Context, o db.Organization) context.Context {
	return context.WithValue(ctx, orgKey, o)
}

func OrgFromCtx(ctx context.Context) (db.Organization, bool) {
	o, ok := ctx.Value(orgKey).(db.Organization)
	return o, ok
}

func WithMembership(ctx context.Context, m db.OrganizationMember) context.Context {
	return context.WithValue(ctx, membershipKey, m)
}

func MembershipFromCtx(ctx context.Context) (db.OrganizationMember, bool) {
	m, ok := ctx.Value(membershipKey).(db.OrganizationMember)
	return m, ok
}
