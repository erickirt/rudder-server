package enricher

import (
	backendconfig "github.com/rudderlabs/rudder-server/backend-config"
	"github.com/rudderlabs/rudder-server/processor/types"
)

// PipelineEnricher is a new paradigm under which the gateway events in
// processing pipeline are enriched with new information based on the handler passed.
type PipelineEnricher interface {
	Enrich(source *backendconfig.SourceT, request *types.GatewayBatchRequest, eventParams *types.EventParams) error
	Close() error
}
