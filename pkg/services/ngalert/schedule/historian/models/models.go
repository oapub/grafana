package models

import (
	"time"

	"github.com/grafana/grafana/pkg/services/ngalert/models"
)

type EvalStatus string

const (
	EvalStatusSuccess      EvalStatus = "success"
	EvalStatusNoData       EvalStatus = "no-data"
	EvalStatusEvalError    EvalStatus = "eval-error"
	EvalStatusProcessError EvalStatus = "process-error"
)

type Record struct {
	Attempt         int
	RuleKey         models.AlertRuleKey
	GroupKey        models.AlertRuleGroupKey
	RuleFingerprint string
	Status          EvalStatus
	Error           string
	Duration        time.Duration
	Tick            time.Time
	EvaluationTime  time.Time
	Version         int64
}
