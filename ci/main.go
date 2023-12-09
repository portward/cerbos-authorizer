package main

import (
	"fmt"
	"path/filepath"
)

const (
	goVersion           = "1.21.5"
	golangciLintVersion = "v1.55.2"
	cerbosVersion       = "0.32.0"
)

type Ci struct{}

func (m *Ci) Test() *Container {
	return dag.Go().
		FromContainer(
			dag.Go().
				FromVersion(goVersion).
				WithSource(dag.Host().Directory(root())).
				Container().
				WithServiceBinding("cerbos", cerbos()).
				WithEnvVariable("CERBOS_ADDRESS", "cerbos:3593"),
		).
		Exec([]string{"go", "test", "-race", "-v", "./..."})
}

func cerbos() *Service {
	config := dag.Host().Directory(filepath.Join(root(), "etc/cerbos/policies"))

	return dag.Container().From(fmt.Sprintf("ghcr.io/cerbos/cerbos:%s", cerbosVersion)).
		WithExposedPort(3592).
		WithExposedPort(3593).
		WithMountedDirectory("/policies", config).
		AsService()
}

func (m *Ci) Lint() *Container {
	return dag.GolangciLint().
		Run(GolangciLintRunOpts{
			Version:   golangciLintVersion,
			GoVersion: goVersion,
			Source:    dag.Host().Directory(root()),
			Verbose:   true,
		})
}
