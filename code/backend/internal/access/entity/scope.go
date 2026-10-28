package entity

import (
	"fmt"
	"sort"
	"strings"

	d "github.com/maximka-44js/lc/internal/app/domain"
)

type ScopeDomain string

type ScopeAccess string

const (
	ScopeDomainStorage  ScopeDomain = "storage"
	ScopeDomainChat     ScopeDomain = "chat"
	ScopeDomainProjects ScopeDomain = "projects"
)

const (
	ScopeAccessRead  ScopeAccess = "read"
	ScopeAccessWrite ScopeAccess = "write"
)

type Scope struct {
	Domain ScopeDomain
	Access ScopeAccess
}

func ParseScope(raw string) (Scope, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return Scope{}, d.NewValidationError(CodeAPIKeyInvalid, "scope must be in domain:access format")
	}

	s := Scope{Domain: ScopeDomain(parts[0]), Access: ScopeAccess(parts[1])}
	if err := s.Validate(); err != nil {
		return Scope{}, err
	}

	return s, nil
}

func (s Scope) Validate() error {
	switch s.Domain {
	case ScopeDomainStorage, ScopeDomainChat, ScopeDomainProjects:
	default:
		return d.NewValidationError(CodeAPIKeyInvalid, fmt.Sprintf("invalid scope domain: %s", s.Domain))
	}

	switch s.Access {
	case ScopeAccessRead, ScopeAccessWrite:
	default:
		return d.NewValidationError(CodeAPIKeyInvalid, fmt.Sprintf("invalid scope access: %s", s.Access))
	}

	return nil
}

func (s Scope) String() string {
	return string(s.Domain) + ":" + string(s.Access)
}

func NormalizeScopes(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, d.NewValidationError(CodeAPIKeyInvalid, "at least one scope is required")
	}

	uniq := make(map[string]struct{}, len(input))
	for _, raw := range input {
		s, err := ParseScope(raw)
		if err != nil {
			return nil, err
		}
		uniq[s.String()] = struct{}{}
	}

	out := make([]string, 0, len(uniq))
	for v := range uniq {
		out = append(out, v)
	}
	sort.Strings(out)

	return out, nil
}

func ScopeAccessForMethod(method string) ScopeAccess {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		return ScopeAccessRead
	default:
		return ScopeAccessWrite
	}
}

func HasScope(scopes []string, domain ScopeDomain, access ScopeAccess) bool {
	needle := Scope{Domain: domain, Access: access}.String()
	for _, scope := range scopes {
		if scope == needle {
			return true
		}
	}
	return false
}
