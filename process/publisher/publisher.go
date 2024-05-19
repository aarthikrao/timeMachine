// Publihser is responsible for publishing jobs to appropriate routes.
package publisher

import (
	"errors"
	"net/http"
	"sync"

	"github.com/aarthikrao/timeMachine/components/routestore"
	"github.com/aarthikrao/timeMachine/models/jobmodels"
	"github.com/aarthikrao/timeMachine/models/routemodels"
	"github.com/aarthikrao/timeMachine/utils/httpclient"
	"github.com/aarthikrao/timeMachine/utils/kafkaclient"
	"go.uber.org/zap"
)

var (
	// ErrReturnedNon200 is returned when the HTTP response code is not 200
	ErrReturnedNon200 = errors.New("HTTP response code is not 200")
)

// Publisher is responsible for publishing jobs to appropriate routes.
type Publihser struct {
	httpClient  *httpclient.HTTPClient
	kafkaClient *kafkaclient.KafkaClient
	routeStore  *routestore.RouteStore
	wg          sync.WaitGroup

	log *zap.Logger
}

func NewPublisher(
	httpClient *httpclient.HTTPClient,
	kafkaClient *kafkaclient.KafkaClient,
	routeStore *routestore.RouteStore,
	jobch chan *jobmodels.Job,
	publisherCount int,
	log *zap.Logger,
) *Publihser {
	pub := &Publihser{
		httpClient:  httpClient,
		kafkaClient: kafkaClient,
		routeStore:  routeStore,
		log:         log,
	}

	for i := 0; i < publisherCount; i++ {
		pub.wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for job := range jobch {
				if err := pub.Publish(job); err != nil {
					log.Error("failed to publish job",
						zap.String("job_id", job.ID),
						zap.String("route", job.Route),
						zap.Error(err))
				}
			}
		}(&pub.wg)
	}

	return pub
}

// Publish publishes the given job to the appropriate route.
// It retrieves the routing information based on the job's route ID,
// and then publishes the job to either an HTTP endpoint or a Kafka topic,
// depending on the route type.
// For HTTP routes, it sends a POST request to the webhook URL with the job metadata.
// For Kafka routes, it publishes the job metadata and ID to the specified Kafka topic on the given host.
// Returns an error if the publishing fails.
func (p *Publihser) Publish(j *jobmodels.Job) error {
	// Get the routing information
	route := p.routeStore.GetRoute(j.Route)
	if route == nil {
		return routemodels.ErrInvalidRouteID
	}

	switch route.Type {
	case routemodels.Http:
		// Publish the job to the HTTP endpoint
		by, code, err := p.httpClient.Post(route.WebhookURL, j.Meta)
		if err != nil {
			return err
		}
		if code != http.StatusOK {
			p.log.Error("HTTP response code is not 200",
				zap.Int("code", code),
				zap.String("job_id", j.ID),
				zap.String("msg", string(by)),
				zap.String("route", route.ID))
			return ErrReturnedNon200
		}

	case routemodels.Kafka:
		// Publish the job to the Kafka topic
		return p.kafkaClient.Publish(route.Host, route.Topic, []byte(j.ID), j.Meta)
	}

	return nil
}

// Wait waits for all the publishers to finish.
func (p *Publihser) Wait() {
	p.wg.Wait()
}
