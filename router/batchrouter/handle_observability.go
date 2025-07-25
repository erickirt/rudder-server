package batchrouter

import (
	"context"
	"fmt"
	"time"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-server/warehouse/router"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/rudderlabs/rudder-go-kit/stats"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
	"github.com/rudderlabs/rudder-server/jobsdb"
	destinationdebugger "github.com/rudderlabs/rudder-server/services/debugger/destination"
	"github.com/rudderlabs/rudder-server/services/diagnostics"
	"github.com/rudderlabs/rudder-server/services/rsources"
	"github.com/rudderlabs/rudder-server/utils/misc"
)

func (brt *Handle) collectMetrics(ctx context.Context) {
	if !diagnostics.EnableBatchRouterMetric {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-brt.diagnosisTicker.C:
			brt.batchRequestsMetricMu.RLock()
			var diagnosisProperties map[string]interface{}
			success := 0
			failed := 0
			for _, batchReqMetric := range brt.batchRequestsMetric {
				success = success + batchReqMetric.batchRequestSuccess
				failed = failed + batchReqMetric.batchRequestFailed
			}
			if len(brt.batchRequestsMetric) > 0 {
				diagnosisProperties = map[string]interface{}{
					brt.destType: map[string]interface{}{
						diagnostics.BatchRouterSuccess: success,
						diagnostics.BatchRouterFailed:  failed,
					},
				}

				brt.Diagnostics.Track(diagnostics.BatchRouterEvents, diagnosisProperties)
			}

			brt.batchRequestsMetric = nil
			brt.batchRequestsMetricMu.RUnlock()
		}
	}
}

func sendDestStatusStats(batchDestination *Connection, jobStateCounts map[string]int, destType string, isWarehouse bool) {
	tags := map[string]string{
		"module":        "batch_router",
		"destType":      destType,
		"isWarehouse":   fmt.Sprintf("%t", isWarehouse),
		"destinationId": misc.GetTagName(batchDestination.Destination.ID, batchDestination.Destination.Name),
		"sourceId":      misc.GetTagName(batchDestination.Source.ID, batchDestination.Source.Name),
	}

	for jobState, count := range jobStateCounts {
		tags["job_state"] = jobState
		stats.Default.NewTaggedStat("event_status", stats.CountType, tags).Count(count)
	}
}

func (brt *Handle) recordAsyncDestinationDeliveryStatus(sourceID, destinationID string, statusList []*jobsdb.JobStatusT) {
	var (
		errorCode    string
		jobState     string
		errorResp    []byte
		successCount int
		failureCount int
		failedReason string
	)

	for _, status := range statusList {
		switch status.JobState {
		case jobsdb.Succeeded.State:
			successCount++
		case jobsdb.Failed.State, jobsdb.Aborted.State:
			failureCount++
			failedReason = string(status.ErrorResponse)
		}
	}
	if failureCount == 0 && successCount == 0 {
		return
	}

	if failureCount > 0 {
		jobState = jobsdb.Failed.State
		errorCode = "500"
		errorResp, _ = jsonrs.Marshal(ErrorResponse{Error: "failed to deliver events. " + failedReason})
	} else {
		jobState = jobsdb.Succeeded.State
		errorCode = "200"
		errorResp = []byte(`{"success":"OK"}`)
	}

	// Emit event_delivery_time metric for successful async destination deliveries
	brt.emitAsyncEventDeliveryTimeMetrics(sourceID, destinationID, statusList)

	// Payload and AttemptNum don't make sense in recording batch router delivery status,
	// So they are set to default values.
	payload, err := sjson.SetBytes([]byte(`{}`), "success", fmt.Sprint(successCount)+" events")
	if err != nil {
		payload = []byte(`{}`)
	}
	payload, err = sjson.SetBytes(payload, "failed", fmt.Sprint(failureCount)+" events")
	if err != nil {
		payload = []byte(`{}`)
	}
	deliveryStatus := destinationdebugger.DeliveryStatusT{
		EventName:     fmt.Sprint(successCount+failureCount) + " events",
		EventType:     "",
		SentAt:        time.Now().Format(misc.RFC3339Milli),
		DestinationID: destinationID,
		SourceID:      sourceID,
		Payload:       payload,
		AttemptNum:    1,
		JobState:      jobState,
		ErrorCode:     errorCode,
		ErrorResponse: errorResp,
	}
	brt.debugger.RecordEventDeliveryStatus(destinationID, &deliveryStatus)
}

// emitAsyncEventDeliveryTimeMetrics emits event_delivery_time metrics for successful async destination deliveries
func (brt *Handle) emitAsyncEventDeliveryTimeMetrics(sourceID, destinationID string, statusList []*jobsdb.JobStatusT) {
	// Get the async destination struct to access original job parameters
	asyncDestStruct := brt.asyncDestinationStruct[destinationID]
	if asyncDestStruct == nil {
		brt.logger.Errorn("Async destination struct not found for destinationID: %s", obskit.DestinationID(destinationID))
		return
	}

	// Process each successful job status to emit event_delivery_time metric
	for _, status := range statusList {
		if status.JobState != jobsdb.Succeeded.State {
			continue // Only emit metrics for successful deliveries
		}

		// Get original job parameters for this job
		originalParams, exists := asyncDestStruct.OriginalJobParameters[status.JobID]
		if !exists {
			brt.logger.Debugn("Original job parameters not found for jobID: %d", logger.NewIntField("jobID", status.JobID))
			continue
		}

		// Extract receivedAt from original job parameters
		receivedAtStr := gjson.GetBytes(originalParams, "received_at").String()
		if receivedAtStr == "" {
			brt.logger.Debugn("ReceivedAt not found in job parameters for jobID: %d", logger.NewIntField("jobID", status.JobID))
			continue
		}

		// Parse receivedAt time
		receivedTime, err := time.Parse(misc.RFC3339Milli, receivedAtStr)
		if err != nil {
			brt.logger.Debugn("Failed to parse receivedAt time for jobID: %d, receivedAt: %s, error: %v", logger.NewIntField("jobID", status.JobID), logger.NewStringField("receivedAt", receivedAtStr), obskit.Error(err))
			continue
		}

		// Extract source category from original job parameters
		sourceCategory := gjson.GetBytes(originalParams, "source_category").String()

		// Create and emit the event_delivery_time metric
		eventDeliveryTimeStat := stats.Default.NewTaggedStat("event_delivery_time", stats.TimerType, map[string]string{
			"module":         "batch_router",
			"destType":       brt.destType,
			"workspaceId":    status.WorkspaceId,
			"sourceId":       sourceID,
			"destID":         destinationID,
			"sourceCategory": sourceCategory,
		})

		// Send the timing metric (time from received to delivered)
		eventDeliveryTimeStat.SendTiming(time.Since(receivedTime))
	}
}

func (brt *Handle) recordDeliveryStatus(batchDestination Connection, output UploadResult, isWarehouse bool) {
	var (
		errorCode string
		jobState  string
		errorResp []byte
	)

	err := output.Error
	if err != nil {
		jobState = jobsdb.Failed.State
		errorCode = "500"
		if isWarehouse {
			jobState = router.GeneratingStagingFileFailedState
		}
		errorResp, _ = jsonrs.Marshal(ErrorResponse{Error: err.Error()})
	} else {
		jobState = jobsdb.Succeeded.State
		errorCode = "200"
		if isWarehouse {
			jobState = router.GeneratedStagingFileState
		}
		errorResp = []byte(`{"success":"OK"}`)
	}

	// Payload and AttemptNum don't make sense in recording batch router delivery status,
	// So they are set to default values.
	payload, err := sjson.SetBytes([]byte(`{}`), "location", output.FileLocation)
	if err != nil {
		payload = []byte(`{}`)
	}
	deliveryStatus := destinationdebugger.DeliveryStatusT{
		EventName:     fmt.Sprint(output.TotalEvents) + " events",
		EventType:     "",
		SentAt:        time.Now().Format(misc.RFC3339Milli),
		DestinationID: batchDestination.Destination.ID,
		SourceID:      batchDestination.Source.ID,
		Payload:       payload,
		AttemptNum:    1,
		JobState:      jobState,
		ErrorCode:     errorCode,
		ErrorResponse: errorResp,
	}
	brt.debugger.RecordEventDeliveryStatus(batchDestination.Destination.ID, &deliveryStatus)
}

func (brt *Handle) trackRequestMetrics(batchReqDiagnostics batchRequestMetric) {
	if diagnostics.EnableBatchRouterMetric {
		brt.batchRequestsMetricMu.Lock()
		brt.batchRequestsMetric = append(brt.batchRequestsMetric, batchReqDiagnostics)
		brt.batchRequestsMetricMu.Unlock()
	}
}

func (brt *Handle) recordUploadStats(destination Connection, output UploadResult) {
	destinationTag := misc.GetTagName(destination.Destination.ID, destination.Destination.Name)
	eventDeliveryStat := stats.Default.NewTaggedStat("event_delivery", stats.CountType, map[string]string{
		"module":      "batch_router",
		"destType":    brt.destType,
		"destination": destinationTag,
		"workspaceId": destination.Source.WorkspaceID,
		"source":      destination.Source.ID,
	})
	eventDeliveryStat.Count(output.TotalEvents)

	if receivedTime, err := time.Parse(misc.RFC3339Milli, output.FirstEventAt); err == nil {
		eventDeliveryTimeStat := stats.Default.NewTaggedStat("event_delivery_time", stats.TimerType, map[string]string{
			"module":         "batch_router",
			"destType":       brt.destType,
			"destination":    destinationTag,
			"workspaceId":    destination.Source.WorkspaceID,
			"sourceId":       destination.Source.ID,
			"destID":         destination.Destination.ID,
			"sourceCategory": destination.Source.SourceDefinition.Category,
		})
		eventDeliveryTimeStat.SendTiming(time.Since(receivedTime))
	}
}

func (brt *Handle) sendRetryStoreStats(attempt int) {
	brt.logger.Warnf("Timeout during store jobs in batch router module, attempt %d", attempt)
	stats.Default.NewTaggedStat("jobsdb_store_timeout", stats.CountType, stats.Tags{"attempt": fmt.Sprint(attempt), "module": "batch_router"}).Count(1)
}

func (brt *Handle) sendRetryUpdateStats(attempt int) {
	brt.logger.Warnf("Timeout during update job status in batch router module, attempt %d", attempt)
	stats.Default.NewTaggedStat("jobsdb_update_timeout", stats.CountType, stats.Tags{"attempt": fmt.Sprint(attempt), "module": "batch_router"}).Count(1)
}

func (brt *Handle) sendQueryRetryStats(attempt int) {
	brt.logger.Warnf("Timeout during query jobs in batch router module, attempt %d", attempt)
	stats.Default.NewTaggedStat("jobsdb_query_timeout", stats.CountType, stats.Tags{"attempt": fmt.Sprint(attempt), "module": "batch_router"}).Count(1)
}

func (brt *Handle) updateRudderSourcesStats(
	ctx context.Context,
	tx jobsdb.UpdateSafeTx,
	jobs []*jobsdb.JobT,
	jobStatuses []*jobsdb.JobStatusT,
) error {
	rsourcesStats := rsources.NewStatsCollector(brt.rsourcesService, "brt", stats.Default)
	rsourcesStats.BeginProcessing(jobs)
	rsourcesStats.CollectStats(jobStatuses)
	rsourcesStats.CollectFailedRecords(jobStatuses)
	err := rsourcesStats.Publish(ctx, tx.SqlTx())
	if err != nil {
		return fmt.Errorf("publishing rsources stats: %w", err)
	}
	return nil
}

// pipelineDelayStats reports the delay of the pipeline as a range:
//
// - max - time elapsed since the first job was created
//
// - min - time elapsed since the last job was created
func (brt *Handle) pipelineDelayStats(partition string, first, last *jobsdb.JobT) {
	var firstJobDelay float64
	var lastJobDelay float64
	if first != nil {
		firstJobDelay = time.Since(first.CreatedAt).Seconds()
	}
	if last != nil {
		lastJobDelay = time.Since(last.CreatedAt).Seconds()
	}
	stats.Default.NewTaggedStat("pipeline_delay_min_seconds", stats.GaugeType, stats.Tags{"destType": brt.destType, "partition": partition, "module": "batch_router"}).Gauge(lastJobDelay)
	stats.Default.NewTaggedStat("pipeline_delay_max_seconds", stats.GaugeType, stats.Tags{"destType": brt.destType, "partition": partition, "module": "batch_router"}).Gauge(firstJobDelay)
}
