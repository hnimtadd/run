package actrs

import (
	"log/slog"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/hnimtadd/run/internal/message"
	"github.com/hnimtadd/run/internal/store"
)

type (
	MetricAggregator struct {
		store         store.MetricStore
		metadataStore store.Store
		version       string
	}

	MetricAggregatorConfig struct {
		MetricStore   store.MetricStore
		MetadataStore store.Store
		Version       string
	}
)

func (m *MetricAggregator) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *cluster.ClusterInit:
		slog.Info("received cluster init message", "node", "metricAggregator")
	case *message.MetricMessage:
		slog.Info("hello from handle metric message", "msg", msg)
		// if err := m.HandleMetricMessage(msg); err != nil {
		// 	// we could retries
		// 	slog.Info("cannoot handle new metric message")
		// }
	default:
		slog.Info("message type not support", "node", "metricAggregator", "msg", msg)
	}
}

func (m *MetricAggregator) HandleMetricMessage(msg *message.MetricMessage) error {
	deployment, err := m.metadataStore.GetDeploymentByID(msg.DeploymentID)
	if err != nil {
		return err
	}
	if err := m.store.AddEndpointMetric(deployment.EndpointID.String(), msg.Metric); err != nil {
		return err
	}
	return nil
}

func NewMetricAggregator(cfg MetricAggregatorConfig) actor.Producer {
	return func() actor.Actor {
		aggregator := &MetricAggregator{
			version:       cfg.Version,
			store:         cfg.MetricStore,
			metadataStore: cfg.MetadataStore,
		}
		return aggregator
	}
}

var KindMetricAggregator = "kind-metric-aggregator"

func NewMetricAggregatorKind(cfg *MetricAggregatorConfig, opts ...actor.PropsOption) *cluster.Kind {
	return cluster.NewKind(KindMetricAggregator, actor.PropsFromProducer(NewMetricAggregator(*cfg), opts...))
}
