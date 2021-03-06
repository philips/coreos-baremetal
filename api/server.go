package api

import (
	"net/http"

	"github.com/coreos/pkg/capnslog"
)

const (
	// APIVersion of the api server and its config types.
	APIVersion = "v1alpha1"
)

var log = capnslog.NewPackageLogger("github.com/coreos/coreos-baremetal", "api")

// Config configures the api Server.
type Config struct {
	// Store for configs
	Store Store
	// Path to static assets
	AssetsPath string
}

// Server serves matches boot and configuration settings to machines.
type Server struct {
	store      Store
	assetsPath string
}

// NewServer returns a new Server.
func NewServer(config *Config) *Server {
	return &Server{
		store:      config.Store,
		assetsPath: config.AssetsPath,
	}
}

// HTTPHandler returns a HTTP handler for the server.
func (s *Server) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	// API Resources
	newSpecResource(mux, "/spec/", s.store)
	gr := newGroupsResource(s.store)

	// Endpoints
	// Boot via iPXE
	mux.Handle("/boot.ipxe", logRequests(ipxeInspect()))
	mux.Handle("/boot.ipxe.0", logRequests(ipxeInspect()))
	mux.Handle("/ipxe", logRequests(NewHandler(gr.matchSpecHandler(ipxeHandler()))))
	// Boot via Pixiecore
	mux.Handle("/pixiecore/v1/boot/", logRequests(pixiecoreHandler(gr, s.store)))
	// cloud configs
	mux.Handle("/cloud", logRequests(NewHandler(gr.matchSpecHandler(cloudHandler(s.store)))))
	// ignition configs
	mux.Handle("/ignition", logRequests(NewHandler(gr.matchSpecHandler(ignitionHandler(s.store)))))

	// kernel, initrd, and TLS assets
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(s.assetsPath))))
	return mux
}
