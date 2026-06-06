package dashboard

type SummaryQuery struct {
	SprintID string
}

type SummaryResponse struct {
	TotalTasks        int64   `json:"total_tasks"`
	TodoTasks         int64   `json:"todo_tasks"`
	InProgressTasks   int64   `json:"in_progress_tasks"`
	ReadyToCheckTasks int64   `json:"ready_to_check_tasks"`
	CheckedByQATasks  int64   `json:"checked_by_qa_tasks"`
	DoneTasks         int64   `json:"done_tasks"`
	BlockedTasks      int64   `json:"blocked_tasks"`
	CompletionRate    float64 `json:"completion_rate"`
	TotalDevelopers   int64   `json:"total_developers"`
	TotalProjects     int64   `json:"total_projects"`
}
