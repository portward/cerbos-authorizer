//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	_ "github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

const (
	goImageRepo = "golang"
	goVersion   = "1.21.1"
)

func Test(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	var cerbos *dagger.Container

	var test *dagger.Container

	// Prepare
	{
		client := client.Pipeline("Prepare")

		{
			client := client.Pipeline("Cerbos")

			cerbos = cerbosContainer(client)
		}
	}

	// Build
	{
		client := client.Pipeline("Build")

		{
			client := client.Pipeline("Test container")

			test = testContainer(client).
				WithServiceBinding("cerbos", cerbos).
				WithEnvVariable("CERBOS_ADDRESS", "cerbos:3592")
		}
	}

	testContainerID, err := test.ID(ctx)
	if err != nil {
		return err
	}

	dir := client.Host().
		Directory(".", dagger.HostDirectoryOpts{
			Exclude: []string{
				".devenv/",
				".direnv/",
				".github/",
				"bin/",
				"build/",
				"ci/",
				"Dockerfile",
				"var/",
			},
		})

	_, err = client.Pipeline("Test").
		Container(dagger.ContainerOpts{
			ID: testContainerID,
		}).
		WithMountedDirectory("/src", dir).
		WithFocus().
		WithExec([]string{"go", "test", "-race", "./..."}).
		Sync(ctx)
	if err != nil {
		return err
	}

	return err
}

// TODO: add go cache
func testContainer(client *dagger.Client) *dagger.Container {
	return client.Container().
		From(fmt.Sprintf("%s:%s", goImageRepo, goVersion)).
		WithEntrypoint(nil).
		WithWorkdir("/src")
}

func cerbosContainer(client *dagger.Client) *dagger.Container {
	config := client.Host().Directory("./etc/cerbos/policies")

	return client.Container().From("ghcr.io/cerbos/cerbos:0.30.0").
		WithExposedPort(3592, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithExposedPort(3593, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithMountedDirectory("/policies", config).
		WithExec(nil)
}
