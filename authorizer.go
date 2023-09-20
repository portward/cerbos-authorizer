package cerbosauthorizer

import (
	"context"
	"slices"
	"sort"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	effectv1 "github.com/cerbos/cerbos/api/genpb/cerbos/effect/v1"
	"github.com/portward/registry-auth/auth"
	"golang.org/x/exp/maps"
)

var _ auth.Authorizer = Authorizer{}

// Authorizer uses [Cerbos] to authorize resource requests.
//
// [Cerbos]: https://cerbos.dev
type Authorizer struct {
	client *cerbos.GRPCClient

	defaultRoles []string
}

// NewAuthorizer returns a new [Authorizer].
func NewAuthorizer(client *cerbos.GRPCClient, defaultRoles []string) Authorizer {
	return Authorizer{
		client:       client,
		defaultRoles: defaultRoles,
	}
}

// Authorize implements the [auth.Authorizer] interface.
func (a Authorizer) Authorize(ctx context.Context, subject auth.Subject, requestedScopes []auth.Scope) ([]auth.Scope, error) {
	principal := cerbos.NewPrincipal(subject.ID().String()).
		WithAttributes(subject.Attributes()) // TODO: allow limiting what attributes are attached to the principal

	if rolesAttr, ok := subject.Attribute("roles"); ok { // TODO: allow extracting roles from subject
		principal = principal.WithRoles(extractRoles(rolesAttr)...)
	}

	if len(principal.Roles()) == 0 {
		principal = principal.WithRoles(a.defaultRoles...)
	}

	resourceBatch := cerbos.NewResourceBatch()

	for _, scope := range requestedScopes {
		resource := cerbos.NewResource(scope.Type, scope.Name)

		if scope.Class != "" {
			resource = resource.WithAttr("class", scope.Class)
		}

		resourceBatch.Add(resource, scope.Actions...)
	}

	resp, err := a.client.CheckResources(ctx, principal, resourceBatch)
	if err != nil { // TODO: check if error means no authorization
		return nil, err
	}

	var scopes auth.Scopes

	for _, result := range resp.GetResults() {
		res := result.GetResource()

		scope := auth.Scope{
			Resource: auth.Resource{
				Type: res.GetKind(),
				Name: res.GetId(),
			},
		}

		actions := result.GetActions()
		actionKeys := maps.Keys(actions)

		sort.Strings(actionKeys)

		for _, action := range actionKeys {
			effect := actions[action]

			if effect != effectv1.Effect_EFFECT_ALLOW {
				continue
			}

			scope.Actions = append(scope.Actions, action)
		}

		if len(scope.Actions) == 0 {
			continue
		}

		scopes = append(scopes, scope)
	}

	return scopes, nil
}

func extractRoles(attr any) []string {
	var roles []string

	switch attr := attr.(type) {
	case []any:
		for _, v := range attr {
			role, ok := v.(string)
			if !ok {
				continue
			}

			roles = append(roles, role)
		}
	case []string:
		return slices.Clone(attr)
	}

	return roles
}
