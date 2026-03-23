package handler

import (
	"errors"
	"testing"
)

func TestExtractStrategyPublishConflictDetail(t *testing.T) {
	detail, ok := extractStrategyPublishConflictDetail(errors.New(`strategy-engine returned 409 when publishing job job_demo: {"detail":"发布策略拦截: 警告数量 5 超过阈值 3"}`))
	if !ok {
		t.Fatalf("expected conflict to be detected")
	}
	if detail != "发布策略拦截: 警告数量 5 超过阈值 3" {
		t.Fatalf("unexpected detail: %s", detail)
	}
}

func TestExtractStrategyPublishConflictDetailNonConflict(t *testing.T) {
	if _, ok := extractStrategyPublishConflictDetail(errors.New("database unavailable")); ok {
		t.Fatalf("expected non-conflict error to be ignored")
	}
}
