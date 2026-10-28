package usecase

import (
	"time"

	"github.com/maximka-44js/lc/internal/projects/entity"
)

type CreateProjectInput struct {
	OwnerID     string
	Name        string
	Description string
	DeadlineAt  *time.Time
	Priority    entity.ProjectPriority
	Metadata    map[string]string
}

type UpdateProjectInput struct {
	OwnerID     string
	ProjectID   string
	Name        string
	Description string
	DeadlineAt  *time.Time
	Priority    entity.ProjectPriority
	Metadata    map[string]string
}

type AddStageInput struct {
	OwnerID    string
	ProjectID  string
	Name       string
	Goal       string
	Order      int
	DeadlineAt *time.Time
}

type UpdateStageInput struct {
	OwnerID    string
	ProjectID  string
	StageID    string
	Name       string
	Goal       string
	Order      int
	DeadlineAt *time.Time
}

type AddTaskInput struct {
	OwnerID    string
	ProjectID  string
	StageID    string
	Outcome    string
	Difficulty int
	Priority   entity.ProjectPriority
	Policy     entity.TaskPolicy
	DeadlineAt *time.Time
	Blocks     []entity.CompetencyBlock
}

type UpdateTaskInput struct {
	OwnerID    string
	ProjectID  string
	StageID    string
	TaskID     string
	Outcome    string
	Difficulty int
	Priority   entity.ProjectPriority
	DeadlineAt *time.Time
}

type ConfigureTaskWeightsInput struct {
	OwnerID   string
	ProjectID string
	StageID   string
	TaskID    string
	Weights   entity.WeightDistribution
}

type StartTaskSessionInput struct {
	OwnerID   string
	ProjectID string
	StageID   string
	TaskID    string
}

type RecordAssessmentSnapshotInput struct {
	OwnerID             string
	ProjectID           string
	StageID             string
	TaskID              string
	BlockType           entity.CompetencyBlockType
	Source              entity.AssessmentSource
	RawScore            float64
	NormalizedScore     float64
	ConfidenceBefore    float64
	ConfidenceAfter     float64
	RecordedAt          time.Time
	TraceID             string
	SourceReference     string
	SourceReferenceType string
}

type RecordAssessmentSnapshotOutput struct {
	Task     entity.Task
	Snapshot entity.AssessmentSnapshot
}

type RecalculateTaskReadinessInput struct {
	OwnerID    string
	ProjectID  string
	StageID    string
	TaskID     string
	At         time.Time
	Profiles   map[entity.CompetencyBlockType]entity.DecayProfile
	RiskPolicy entity.RiskPolicy
}

type RecalculateStageReadinessInput struct {
	OwnerID           string
	ProjectID         string
	StageID           string
	At                time.Time
	AggregationPolicy entity.AggregationPolicy
}

type RecalculateProjectReadinessInput struct {
	OwnerID           string
	ProjectID         string
	At                time.Time
	AggregationPolicy entity.AggregationPolicy
}

type GenerateRecommendationsInput struct {
	OwnerID              string
	ProjectID            string
	StageID              string
	TaskID               string
	RecommendationPolicy entity.RecommendationPolicy
	CalibrationPolicy    entity.CalibrationPolicy
}

type MarkTaskExamReadyInput struct {
	OwnerID   string
	ProjectID string
	StageID   string
	TaskID    string
}

type TransitionTaskStatusInput struct {
	OwnerID    string
	ProjectID  string
	StageID    string
	TaskID     string
	NextStatus entity.TaskStatus
}

type UserPolicies struct {
	Aggregation    entity.AggregationPolicy
	Risk           entity.RiskPolicy
	Recommendation entity.RecommendationPolicy
	Calibration    entity.CalibrationPolicy
}

type ProjectDashboardView struct {
	Project            entity.Project              `json:"project"`
	Coverage           float64                     `json:"coverage"`
	ReadinessNow       float64                     `json:"readinessNow"`
	RiskByStage        map[string]entity.RiskLevel `json:"riskByStage"`
	CriticalToDeadline []CriticalItem              `json:"criticalToDeadline"`
}

type StageBoardView struct {
	Stage              entity.Stage                            `json:"stage"`
	Tasks              []entity.Task                           `json:"tasks"`
	BlockDeficits      map[string][]entity.CompetencyBlockType `json:"blockDeficits"`
	NextActionPriority map[string]entity.ProjectPriority       `json:"nextActionPriority"`
}

type TaskReadinessView struct {
	Task            entity.Task                               `json:"task"`
	BlockScoresNow  []BlockScoreNow                           `json:"blockScoresNow"`
	LastAssessments map[entity.CompetencyBlockType]*time.Time `json:"lastAssessments"`
	DropForecast    []ForecastPoint                           `json:"dropForecast"`
	Recommendations []entity.Recommendation                   `json:"recommendations"`
}

type BlockScoreNow struct {
	Type      entity.CompetencyBlockType `json:"type"`
	Weight    int                        `json:"weight"`
	Mastery   float64                    `json:"mastery"`
	Freshness float64                    `json:"freshness"`
	ScoreNow  float64                    `json:"scoreNow"`
	Risk      entity.RiskLevel           `json:"risk"`
}

type ForecastPoint struct {
	At        time.Time        `json:"at"`
	Readiness float64          `json:"readiness"`
	Risk      entity.RiskLevel `json:"risk"`
}

type CriticalItem struct {
	StageID    string     `json:"stageID"`
	TaskID     string     `json:"taskID"`
	Reason     string     `json:"reason"`
	DeadlineAt *time.Time `json:"deadlineAt"`
}
