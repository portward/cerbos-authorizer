package cerbosauthorizer

import (
	"fmt"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"github.com/portward/registry-auth/auth"
)

// Config implements the [AuthorizerFactory] interface defined by Portward.
//
// [AuthorizerFactory]: https://pkg.go.dev/github.com/portward/portward/config#AuthorizerFactory
type Config struct {
	Address      string   `mapstructure:"address"`
	DefaultRoles []string `mapstructure:"defaultRoles"`
}

// New returns a new [Authorizer] from the configuration.
func (c Config) New() (auth.Authorizer, error) {
	client, err := cerbos.New(c.Address)
	if err != nil {
		return nil, err
	}

	return NewAuthorizer(client, c.DefaultRoles), nil
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("cerbos: address is required")
	}

	return nil
}
