package event

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

type PanaceaSubscriber struct {
	client *rpchttp.HTTP
}

// NewSubscriber generates a rpc http client with websocket address.
func NewSubscriber(wsAddr string) (*PanaceaSubscriber, error) {
	client, err := rpchttp.New(wsAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	err = client.Start()
	if err != nil {
		return nil, err
	}

	return &PanaceaSubscriber{
		client: client,
	}, nil
}

func (s *PanaceaSubscriber) Run(events ...Event) error {
	log.Infof("Panacea event subscriber is started")

	for _, e := range events {
		if err := s.subscribe(e); err != nil {
			return err
		}
	}

	return nil
}

func (s *PanaceaSubscriber) subscribe(event Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	query := event.GetEventQuery()

	log.Infof("subscribe %s. query: %s", event.Name(), query)
	txs, err := s.client.Subscribe(ctx, "", query)
	if err != nil {
		return err
	}

	go func(e Event) {
		for tx := range txs {
			log.Infof("received event a %s", e.Name())
			if err := e.EventHandler(tx); err != nil {
				log.Errorf("failed to handle event '%s': %v", query, err)
			}
		}
	}(event)

	return nil
}

func (s *PanaceaSubscriber) Close() error {
	log.Infof("closing Panacea event subscriber")
	return s.client.Stop()
}
