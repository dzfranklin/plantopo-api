package authn

import (
	"log/slog"
	"strings"
)

type DevAuthenticator struct {
	*WorkOS
}

const devImpersonatePrefix = "dev-impersonate:"

func (d *DevAuthenticator) Verify(token string) (string, error) {
	if strings.HasPrefix(token, devImpersonatePrefix) {
		slog.Warn("dev impersonation", "token", token)
		return token[len(devImpersonatePrefix):], nil
	}
	return d.WorkOS.Verify(token)
}
