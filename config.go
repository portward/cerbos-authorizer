package cerbos

import (
	"fmt"

	"github.com/cerbos/cerbos-sdk-go/cerbos"
	"github.com/portward/registry-auth/auth"
)

// Config implements the [AuthorizerFactory] interface defined by Portward.
//
// [AuthorizerFactory]: https://pkg.go.dev/github.com/portward/portward/config#AuthorizerFactory
type Config struct {
	Address      string        `mapstructure:"address"`
	Options      OptionsConfig `mapstructure:"options"`
	DefaultRoles []string      `mapstructure:"defaultRoles"`
}

// OptionsConfig implements options for the Cerbos client connection.
type OptionsConfig struct {
	Plaintext bool `mapstructure:"plaintext"`
}

// New returns a new [Authorizer] from the configuration.
func (c Config) New() (auth.Authorizer, error) {
	var options []cerbos.Opt

	if c.Options.Plaintext {
		options = append(options, cerbos.WithPlaintext())
	}

	client, err := cerbos.New(c.Address, options...)
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
