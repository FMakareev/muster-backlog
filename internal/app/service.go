// Package app exposes the application-level service bound into the frontend.
package app

// Version is the application version. It is overridden at build time via
// -ldflags "-X github.com/FMakareev/muster-backlog/internal/app.Version=..."
// so that the binary reports the tag it was released under.
var Version = "0.0.0-dev"

// Service is the root service bound into the frontend. Feature services are
// added alongside it as they arrive; this one carries only what the shell needs
// before any project is loaded.
type Service struct{}

// NewService returns the root application service.
func NewService() *Service {
	return &Service{}
}

// AppVersion reports the version this binary was built from.
func (s *Service) AppVersion() string {
	return Version
}
