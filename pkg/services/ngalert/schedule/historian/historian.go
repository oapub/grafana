package historian

import (
	"context"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/lokiclient"
	"github.com/grafana/grafana/pkg/services/ngalert/metrics"
	historianModels "github.com/grafana/grafana/pkg/services/ngalert/schedule/historian/models"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	LokiClientSpanName  = "ngalert.evaluation-historian.client"
	historyWriteTimeout = time.Minute
	HistoryKey          = "from"
	HistoryLabelValue   = "evaluation-history"
	OrgIDLabel          = "orgID"
	GroupLabel          = "group"
	FolderUIDLabel      = "folderUID"
)

type EvaluationHistoryLokiEntry struct {
	SchemaVersion int    `json:"schemaVersion"`
	RuleUID       string `json:"ruleUID"`
	Version       string `json:"version"`
	Tick          int64  `json:"tick"`
	FingerPrint   string `json:"fingerprint"`
	Attempt       int64  `json:"attempt"`
	Error         string `json:"error,omitempty"`
	Duration      int64  `json:"duration"`
	Status        string `json:"status"`
}

type LokiClient interface {
	Push(context.Context, []lokiclient.Stream) error
}

type Historian struct {
	client         LokiClient
	externalLabels map[string]string
	metrics        *metrics.EvalHistorian
	log            log.Logger
	timeout        time.Duration
}

func NewHistorian(
	logger log.Logger,
	lokiClient LokiClient,
	metrics *metrics.EvalHistorian,
	externalLabels map[string]string,
	timeout time.Duration,
) *Historian {
	if timeout <= 0 {
		timeout = historyWriteTimeout
	}
	return &Historian{
		client:         lokiClient,
		externalLabels: externalLabels,
		metrics:        metrics,
		log:            logger,
		timeout:        timeout,
	}
}

func (h *Historian) Record(ctx context.Context, opts historianModels.RecordOpts) {
	logger := h.log.FromContext(ctx)
	stream, err := h.prepareStream(opts)
	if err != nil {
		logger.Error("Failed to convert evaluation history to stream", "error", err)
	}
	org := fmt.Sprint(opts.GroupKey.OrgID)

	writeCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	defer h.metrics.WritesTotal.WithLabelValues(org).Inc()

	if err := h.recordStream(writeCtx, stream); err != nil {
		logger.Error("Failed to save evaluation history", "error", err)
		h.metrics.WritesFailed.WithLabelValues(org).Inc()
	}
}

func (h *Historian) prepareStream(opts historianModels.RecordOpts) (lokiclient.Stream, error) {
	entry := EvaluationHistoryLokiEntry{
		SchemaVersion: 1,
		RuleUID:       opts.RuleKey.UID,
		Version:       strconv.FormatInt(opts.Version, 10),
		FingerPrint:   opts.RuleFingerprint,
		Attempt:       int64(opts.Attempt),
		Duration:      opts.Duration.Milliseconds(),
		Status:        string(opts.Status),
	}
	if opts.Error != nil {
		entry.Error = opts.Error.Error()
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return lokiclient.Stream{}, err
	}

	streamLabels := make(map[string]string)
	for k, v := range h.externalLabels {
		streamLabels[k] = v
	}
	streamLabels[HistoryKey] = HistoryLabelValue
	streamLabels[OrgIDLabel] = fmt.Sprint(opts.GroupKey.OrgID)
	streamLabels[GroupLabel] = fmt.Sprint(opts.GroupKey.RuleGroup)
	streamLabels[FolderUIDLabel] = fmt.Sprint(opts.GroupKey.NamespaceUID)

	return lokiclient.Stream{
		Stream: streamLabels,
		Values: []lokiclient.Sample{
			{
				T: opts.Tick,
				V: string(entryJSON),
			}},
	}, nil
}

func (h *Historian) recordStream(ctx context.Context, stream lokiclient.Stream) error {
	if err := h.client.Push(ctx, []lokiclient.Stream{stream}); err != nil {
		return err
	}
	return nil
}
