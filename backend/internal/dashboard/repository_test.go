package dashboard

import "testing"

func TestApplyCompletionRate(t *testing.T) {
	summary := &SummaryResponse{
		TotalTasks: 10,
		DoneTasks:  4,
	}

	applyCompletionRate(summary)

	if summary.CompletionRate != 40 {
		t.Fatalf("expected completion rate 40, got %v", summary.CompletionRate)
	}
}

func TestApplyCompletionRateWithNoTasks(t *testing.T) {
	summary := &SummaryResponse{}

	applyCompletionRate(summary)

	if summary.CompletionRate != 0 {
		t.Fatalf("expected completion rate 0, got %v", summary.CompletionRate)
	}
}
