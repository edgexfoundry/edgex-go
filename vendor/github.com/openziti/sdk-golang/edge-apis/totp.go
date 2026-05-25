package edge_apis

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TotpCodeResult represents the outcome of requesting a TOTP code from a user or provider,
// containing either the code string or an error if the request failed.
type TotpCodeResult struct {
	Code string
	Err  error
}

// TotpEnrollmentResult represents the outcome of a TOTP enrollment request. A non-empty Code
// means the user scanned the QR code and entered the resulting code to complete enrollment.
// A non-nil Err means the user cancelled or denied enrollment.
type TotpEnrollmentResult struct {
	Code string
	Err  error
}

// TotpEnrollmentProvider is called during OIDC authentication when an identity has not yet
// enrolled in TOTP but their auth policy requires it. Implementations show the provisioning
// URL (as a QR code or plain text) to the user, collect their initial TOTP code, and return
// it via the channel. Send a non-nil Err to cancel enrollment.
type TotpEnrollmentProvider interface {
	// GetTotpEnrollmentCode receives the provisioning URL for QR code display and returns
	// a channel that delivers the TOTP code entered by the user, or an error to cancel.
	GetTotpEnrollmentCode(provisioningUrl string) <-chan TotpEnrollmentResult
}

// TotpEnrollmentProviderFunc is a function adapter that implements TotpEnrollmentProvider.
type TotpEnrollmentProviderFunc func(provisioningUrl string) <-chan TotpEnrollmentResult

// GetTotpEnrollmentCode implements TotpEnrollmentProvider.
func (f TotpEnrollmentProviderFunc) GetTotpEnrollmentCode(provisioningUrl string) <-chan TotpEnrollmentResult {
	return f(provisioningUrl)
}

// TotpTokenResult represents the outcome of exchanging a TOTP code for a session token,
// including the token value, issuance timestamp, and any errors encountered.
type TotpTokenResult struct {
	Token    string
	IssuedAt time.Time
	Err      error
}

// TotpCodeProvider supplies TOTP codes for multi-factor authentication. Implementations typically
// prompt users to enter codes from authenticator apps.
type TotpCodeProvider interface {
	// GetTotpCode returns a channel that delivers the TOTP code result.
	GetTotpCode() <-chan TotpCodeResult
}

// TotpCodeProviderFunc is a function adapter that implements TotpCodeProvider.
type TotpCodeProviderFunc func() <-chan TotpCodeResult

// NewTotpCodeProviderFromChStringFunc adapts legacy func(chan string) callbacks to the TotpCodeProvider interface.
// This enables backward compatibility while allowing a smoother migration path to the new interface.
func NewTotpCodeProviderFromChStringFunc(stringFunc func(ch chan string)) TotpCodeProvider {
	return TotpCodeProviderFunc(func() <-chan TotpCodeResult {
		resultCh := make(chan TotpCodeResult)

		go func() {
			stringCh := make(chan string)
			go stringFunc(stringCh)

			code := <-stringCh

			resultCh <- TotpCodeResult{
				Code: code,
			}
		}()

		return resultCh
	})
}

func (f TotpCodeProviderFunc) GetTotpCode() <-chan TotpCodeResult {
	return f()
}

// TotpTokenRequestor exchanges TOTP codes with the authentication service for session tokens.
type TotpTokenRequestor interface {
	// RequestTotpToken exchanges a TOTP code for a session token.
	RequestTotpToken(code string) <-chan TotpTokenResult
}

// TotpTokenProvider coordinates the complete TOTP authentication flow, obtaining codes and exchanging them for tokens.
type TotpTokenProvider interface {
	// Request initiates a TOTP token request, returning a channel with the result.
	Request() <-chan TotpTokenResult
}

// TotpTokenProviderFunc is a function adapter that implements TotpTokenProvider.
type TotpTokenProviderFunc func() <-chan TotpTokenResult

// Request implements TotpTokenProvider.
func (f TotpTokenProviderFunc) Request() <-chan TotpTokenResult {
	return f()
}

// SingularTokenRequestor serializes TOTP token requests, ensuring only one is active at a time.
// This prevents duplicate authentication attempts when multiple operations require TOTP.
type SingularTokenRequestor struct {
	isRequesting   sync.Mutex
	codeProvider   TotpCodeProvider
	tokenRequestor TotpTokenRequestor
}

const totpCodeProviderTimeout = 5 * time.Minute
const totpTokenRequestorTimeout = 30 * time.Second

// NewSingularTokenRequestor creates a token requestor that coordinates code collection and token exchange.
// Only one request can be active at a time; subsequent requests return nil if one is already in progress.
func NewSingularTokenRequestor(codeProvider TotpCodeProvider, tokenRequestor TotpTokenRequestor) *SingularTokenRequestor {
	return &SingularTokenRequestor{
		codeProvider:   codeProvider,
		tokenRequestor: tokenRequestor,
	}
}

// Request initiates a TOTP token request, returning nil if a request is already in progress.
// The returned channel delivers the token result once the code is collected and exchanged.
func (r *SingularTokenRequestor) Request() <-chan TotpTokenResult {
	if lockObtained := r.isRequesting.TryLock(); !lockObtained {
		//outstanding request don't do anything
		return nil
	}

	tokenCh := make(chan TotpTokenResult)
	codeCh := r.codeProvider.GetTotpCode()

	go func() {
		defer r.isRequesting.Unlock()

		select {
		case codeResult := <-codeCh:
			if codeResult.Err != nil {
				tokenCh <- TotpTokenResult{
					Token: "",
					Err:   fmt.Errorf("error getting totp code: %v", codeResult.Err),
				}
				return
			}
			code := strings.TrimSpace(codeResult.Code)

			if code == "" {
				tokenCh <- TotpTokenResult{
					Token: "",
					Err:   errors.New("empty totp code entered"),
				}
				return
			}

			select {
			case tokenResult := <-r.tokenRequestor.RequestTotpToken(code):
				tokenCh <- tokenResult
			case <-time.After(totpTokenRequestorTimeout):
				tokenCh <- TotpTokenResult{
					Token: "",
					Err:   errors.New("timed out waiting for totp token"),
				}
				return
			}

			return
		case <-time.After(totpCodeProviderTimeout):
			tokenCh <- TotpTokenResult{
				Token: "",
				Err:   errors.New("timed out waiting for totp code"),
			}
			return
		}
	}()

	return tokenCh
}
