package cerbosauthorizer

import (
	"fmt"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"github.com/portward/registry-auth-config/config"
	"github.com/portward/registry-auth/auth"
)

func init() {
	config.RegisterAuthorizerFactory("cerbos", func() config.AuthorizerFactory { return Config{} })
}

// Config implements the [config.AuthorizerFactory] interface.
type Config struct {
	Address      string   `mapstructure:"address"`
	DefaultRoles []string `mapstructure:"defaultRoles"`
}

// New implements the [config.AuthorizerFactory] interface.
func (c Config) New() (auth.Authorizer, error) {
	client, err := cerbos.New(c.Address)
	if err != nil {
		return nil, err
	}

	return NewAuthorizer(client, c.DefaultRoles), nil
}

// Validate implements the [config.AuthorizerFactory] interface.
func (c Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("cerbos: address is required")
	}

	return nil
}
