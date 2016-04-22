package etcd

import (
	etcd "github.com/coreos/etcd/client"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/internal/cache"
	"github.com/go-kit/kit/service"
)

// Subscriber yield endpoints stored in a certain etcd keyspace. Any kind of
// change in that keyspace is watched and will update the Subscriber endpoints.
type Subscriber struct {
	client Client
	prefix string
	cache  *cache.Cache
	logger log.Logger
	quit   chan struct{}
}

var _ sd.Subscriber = &Subscriber{}

// NewSubscriber returns an etcd subscriber. It will start watching the given
// prefix for changes, and update the Subscriber endpoints.
func NewSubscriber(c Client, prefix string, factory sd.Factory, logger log.Logger) (*Subscriber, error) {
	s := &Subscriber{
		client: c,
		prefix: prefix,
		cache:  cache.New(factory, logger),
		logger: logger,
		quit:   make(chan struct{}),
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err == nil {
		logger.Log("prefix", s.prefix, "instances", len(instances))
	} else {
		logger.Log("prefix", s.prefix, "err", err)
	}
	s.cache.Update(instances)

	go s.loop()
	return s, nil
}

func (s *Subscriber) loop() {
	responseChan := make(chan *etcd.Response)
	go s.client.WatchPrefix(s.prefix, responseChan)
	for {
		select {
		case <-responseChan:
			instances, err := s.client.GetEntries(s.prefix)
			if err != nil {
				s.logger.Log("msg", "failed to retrieve entries", "err", err)
				continue
			}
			s.cache.Update(instances)

		case <-s.quit:
			return
		}
	}
}

// Services implements the Subscriber interface.
func (s *Subscriber) Services() ([]service.Service, error) {
	return s.cache.Services()
}

// Stop terminates the Subscriber.
func (s *Subscriber) Stop() {
	close(s.quit)
}