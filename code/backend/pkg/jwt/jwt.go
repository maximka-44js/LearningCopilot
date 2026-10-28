package jwt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	providerInitTimeout = 10 * time.Second
	verifyTimeout       = 5 * time.Second
)

var (
	// ErrUnexpectedSigningMethod is returned when the JWT signing method is not expected.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	// ErrTokenGenerationUnsupported is returned when token generation is not configured.
	ErrTokenGenerationUnsupported = errors.New("token generation unsupported")
	// ErrTokenVerificationUnavailable is returned when token verification is not configured.
	ErrTokenVerificationUnavailable = errors.New("token verification unavailable")
	// ErrEmptyToken is returned when the token string is empty.
	ErrEmptyToken = errors.New("empty token")
	// ErrEmptySubject is returned when the token has no subject.
	ErrEmptySubject = errors.New("empty subject")
	// ErrInvalidAudience is returned when the token audience does not match the client.
	ErrInvalidAudience = errors.New("invalid audience")
	// ErrInvalidIssuer is returned when the token issuer does not match the configured issuer.
	ErrInvalidIssuer = errors.New("invalid issuer")
)

// Manager handles JWT token verification (Keycloak) and optional HMAC token generation for tests.
type Manager struct {
	verifier *oidc.IDTokenVerifier
	clientID string
	issuer   string
	secret   []byte
	duration time.Duration
}

// New creates a Keycloak-backed JWT manager using OIDC discovery.
func New(issuer, clientID string) (*Manager, error) {
	return NewWithDiscovery(issuer, issuer, clientID)
}

// NewWithDiscovery creates a verifier using separate discovery and token issuers.
func NewWithDiscovery(discoveryIssuer, issuer, clientID string) (*Manager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), providerInitTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, discoveryIssuer)
	if err != nil {
		return nil, fmt.Errorf("jwt - New - oidc.NewProvider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		SkipIssuerCheck:   discoveryIssuer != issuer,
		SkipClientIDCheck: true,
	})

	return &Manager{verifier: verifier, clientID: clientID, issuer: issuer}, nil
}

// NewHMAC creates a HMAC-backed JWT manager for local testing.
func NewHMAC(secret string, duration time.Duration) *Manager {
	return &Manager{
		secret:   []byte(secret),
		duration: duration,
	}
}

// GenerateToken creates a new JWT token for the given user ID.
func (m *Manager) GenerateToken(userID string) (string, error) {
	if len(m.secret) == 0 {
		return "", ErrTokenGenerationUnsupported
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(m.duration)),
	})

	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("jwt - GenerateToken - token.SignedString: %w", err)
	}

	return tokenString, nil
}

// ParseToken validates a JWT token and returns the user ID.
func (m *Manager) ParseToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", ErrEmptyToken
	}

	if m.verifier != nil {
		ctx, cancel := context.WithTimeout(context.Background(), verifyTimeout)
		defer cancel()

		idToken, err := m.verifier.Verify(ctx, tokenString)
		if err != nil {
			return "", fmt.Errorf("jwt - ParseToken - verifier.Verify: %w", err)
		}

		var claims struct {
			Subject         string          `json:"sub"`
			Audience        json.RawMessage `json:"aud"`
			AuthorizedParty string          `json:"azp"`
		}
		if err := idToken.Claims(&claims); err != nil {
			return "", fmt.Errorf("jwt - ParseToken - Claims: %w", err)
		}
		if m.issuer != "" && idToken.Issuer != m.issuer {
			return "", fmt.Errorf("%w: expected %q got %q", ErrInvalidIssuer, m.issuer, idToken.Issuer)
		}
		if claims.Subject == "" {
			return "", ErrEmptySubject
		}
		if m.clientID != "" && !audienceMatches(claims.Audience, claims.AuthorizedParty, m.clientID) {
			return "", ErrInvalidAudience
		}

		return claims.Subject, nil
	}

	if len(m.secret) == 0 {
		return "", ErrTokenVerificationUnavailable
	}

	parser := jwtlib.NewParser(jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}))
	claims := &jwtlib.RegisteredClaims{}
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return m.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("jwt - ParseToken - jwtlib.ParseWithClaims: %w", err)
	}
	if !token.Valid {
		return "", errors.New("jwt - ParseToken - invalid token")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("jwt - ParseToken - GetSubject: %w", err)
	}
	if sub == "" {
		return "", ErrEmptySubject
	}

	return sub, nil
}

func audienceMatches(aud json.RawMessage, azp string, clientID string) bool {
	if clientID == "" {
		return true
	}
	if azp == clientID {
		return true
	}
	var single string
	if err := json.Unmarshal(aud, &single); err == nil {
		return single == clientID
	}
	var list []string
	if err := json.Unmarshal(aud, &list); err == nil {
		for _, item := range list {
			if item == clientID {
				return true
			}
		}
	}
	return false
}
