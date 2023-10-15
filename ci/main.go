//go:build mage
// +build mage

package main

import (
	"context"
)

type Ci struct{}

func (m *Ci) MyFunction(ctx context.Context, stringArg string) (*Container, error) {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg}).Sync(ctx)
}
