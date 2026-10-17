//go:build go1.25

package security

import "net/http"

// [start csrf]

// ProtectForms rejects cross-origin, state-changing browser requests:
// the defense against cross-site request forgery that Go 1.25 added to
// net/http. Requests from our own site and from trusted origins pass.
func ProtectForms(next http.Handler) (http.Handler, error) {
	cop := http.NewCrossOriginProtection()
	if err := cop.AddTrustedOrigin("https://admin.example.com"); err != nil {
		return nil, err
	}
	return cop.Handler(next), nil
}

// [end csrf]
