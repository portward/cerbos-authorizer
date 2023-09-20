package cerbosauthorizer

import (
	"context"
	"maps"
	"os"
	"testing"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"github.com/portward/registry-auth/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type subjectStub struct {
	id    auth.SubjectID
	attrs map[string]any
}

// ID implements auth.Subject.
func (s subjectStub) ID() auth.SubjectID {
	return s.id
}

// Attribute implements auth.Subject.
func (s subjectStub) Attribute(key string) (any, bool) {
	v, ok := s.attrs[key]

	return v, ok
}

// Attributes implements auth.Subject.
func (s subjectStub) Attributes() map[string]any {
	return maps.Clone(s.attrs)
}

func TestAuthorizer(t *testing.T) {
	cerbosAddr := os.Getenv("CERBOS_ADDRESS")

	if cerbosAddr == "" {
		t.Skip("cerbos is not configured")
	}

	client, err := cerbos.New(cerbosAddr, cerbos.WithPlaintext())
	require.NoError(t, err)

	authorizer := NewAuthorizer(client, []string{"user"})

	t.Run("OK", func(t *testing.T) {
		subject := subjectStub{
			id: auth.SubjectIDFromString("user"),
			attrs: map[string]any{
				"roles": []string{"user"},
			},
		}

		requestedScopes := auth.Scopes{
			{ // This should be partially allowed
				Resource: auth.Resource{
					Type: "repository",
					Name: "image-in-root",
				},
				Actions: []string{"pull", "push"},
			},
			{ // This should be denied
				Resource: auth.Resource{
					Type: "repository",
					Name: "image/in/different/namespace",
				},
				Actions: []string{"pull"},
			},
			{ // This should be allowed
				Resource: auth.Resource{
					Type: "repository",
					Name: "user/image",
				},
				Actions: []string{"pull", "push"},
			},
		}

		grantedScopes, err := authorizer.Authorize(context.Background(), subject, requestedScopes)
		require.NoError(t, err)

		expectedScopes := []auth.Scope{
			{
				Resource: auth.Resource{
					Type: "repository",
					Name: "image-in-root",
				},
				Actions: []string{"pull"},
			},
			{
				Resource: auth.Resource{
					Type: "repository",
					Name: "user/image",
				},
				Actions: []string{"pull", "push"},
			},
		}

		assert.Equal(t, expectedScopes, grantedScopes)
	})

	t.Run("default roles and admin", func(t *testing.T) {
		subject := subjectStub{
			id: auth.SubjectIDFromString("admin"),
		}

		requestedScopes := auth.Scopes{
			{ // This should be allowed
				Resource: auth.Resource{
					Type: "repository",
					Name: "image-in-root",
				},
				Actions: []string{"pull", "push"},
			},
			{ // This should be denied
				Resource: auth.Resource{
					Type: "repository",
					Name: "image/in/different/namespace",
				},
				Actions: []string{"pull"},
			},
			{ // This should be allowed
				Resource: auth.Resource{
					Type: "repository",
					Name: "admin/image",
				},
				Actions: []string{"pull", "push"},
			},
		}

		grantedScopes, err := authorizer.Authorize(context.Background(), subject, requestedScopes)
		require.NoError(t, err)

		expectedScopes := []auth.Scope{
			{
				Resource: auth.Resource{
					Type: "repository",
					Name: "image-in-root",
				},
				Actions: []string{"pull", "push"},
			},
			{
				Resource: auth.Resource{
					Type: "repository",
					Name: "admin/image",
				},
				Actions: []string{"pull", "push"},
			},
		}

		assert.Equal(t, expectedScopes, grantedScopes)
	})
}
