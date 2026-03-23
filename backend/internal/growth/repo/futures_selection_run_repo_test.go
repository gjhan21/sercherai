package repo

import "testing"

func TestBuildPreviousFuturesPublishDiff(t *testing.T) {
	diff := buildPreviousFuturesPublishDiff("IF2606", []map[string]any{
		{
			"contract":       "IF2606",
			"rank":           2,
			"reason_summary": "上一版仍在组合中",
		},
	})
	if got := asString(diff["status"]); got != "UNCHANGED" {
		t.Fatalf("expected unchanged futures diff, got %#v", diff)
	}
	if got := asInt(diff["previous_rank"]); got != 2 {
		t.Fatalf("expected previous rank 2, got %#v", diff)
	}
	if got := asString(diff["previous_reason"]); got != "上一版仍在组合中" {
		t.Fatalf("expected previous reason, got %#v", diff)
	}

	added := buildPreviousFuturesPublishDiff("CU2606", []map[string]any{
		{
			"contract":       "IF2606",
			"rank":           2,
			"reason_summary": "上一版仍在组合中",
		},
	})
	if got := asString(added["status"]); got != "ADDED" {
		t.Fatalf("expected added futures diff, got %#v", added)
	}
}
