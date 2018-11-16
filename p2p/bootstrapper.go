package p2p

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

// Bootstrapper implements Bootstrapper with p2p details
type Bootstrapper struct{}

// Bootstrap initiates p2p server and client into context
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[bootstrap.BootstrappedConfig]; !ok {
		return fmt.Errorf("config not initialised")
	}

	cfg := ctx[bootstrap.BootstrappedConfig].(*config.Configuration)
	srv := &p2pServer{config: cfg}
	ctx[bootstrap.BootstrappedP2PServer] = srv
	ctx[bootstrap.BootstrappedP2PClient] = srv
	return nil
}