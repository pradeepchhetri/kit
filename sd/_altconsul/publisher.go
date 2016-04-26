package consul

import (
	"fmt"

	stdconsul "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/log"
)

// Publisher publishes service instance liveness information to Consul.
type Publisher struct {
	client       Client
	registration *stdconsul.AgentServiceRegistration
	logger       log.Logger
}

// NewPublisher returns a Consul publisher acting on the provided catalog
// registration.
func NewPublisher(client Client, r *stdconsul.AgentServiceRegistration, logger log.Logger) *Publisher {
	return &Publisher{
		client:       client,
		registration: r,
		logger:       log.NewContext(logger).With("service", r.Name, "tags", fmt.Sprint(r.Tags), "address", r.Address),
	}
}

// Publish implements sd.Publisher interface.
func (p *Publisher) Publish() {
	if err := p.client.Register(p.registration); err != nil {
		p.logger.Log("err", err)
	} else {
		p.logger.Log("action", "publish")
	}
}

// Unpublish implements sd.Publisher interface.
func (p *Publisher) Unpublish() {
	if err := p.client.Deregister(p.registration); err != nil {
		p.logger.Log("err", err)
	} else {
		p.logger.Log("action", "unpublish")
	}
}