//go:build wireinject
// +build wireinject

package wired

import (
	"github.com/google/wire"
	s "github.com/uilki/lgc/api/server"
)

func InitializeServer(pass string) (*s.Server, error) {
	wire.Build(s.NewServer, s.NewBackLogger)
	return &s.Server{}, nil
}
