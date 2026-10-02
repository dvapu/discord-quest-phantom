package api

// UserMe represents the current authenticated Discord user (/users/@me).
type UserMe struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	GlobalName *string `json:"global_name"`
	MFAEnabled bool    `json:"mfa_enabled"`
	Email      *string `json:"email"`
}

// QuestResponse represents the top-level payload from GET /api/v9/quests/@me.
type QuestResponse struct {
	Quests                      []Quest `json:"quests"`
	ExcludedQuests              []Quest `json:"excluded_quests"`
	QuestEnrollmentBlockedUntil *string `json:"quest_enrollment_blocked_until"`
}

// Quest represents a single Discord quest object.
type Quest struct {
	ID                    string       `json:"id"`
	Config                QuestConfig  `json:"config"`
	UserStatus            *UserStatus  `json:"user_status"`
	TrafficMetadataRaw    *string      `json:"traffic_metadata_raw"`
	TrafficMetadataSealed *string      `json:"traffic_metadata_sealed"`
}

// QuestConfig contains campaign metadata, timeframes, and task specifications.
type QuestConfig struct {
	ID           string           `json:"id"`
	ConfigVersion int             `json:"config_version"`
	StartsAt     string           `json:"starts_at"`
	ExpiresAt    string           `json:"expires_at"`
	Features     []int            `json:"features"`
	Application  *ApplicationMeta `json:"application"`
	Messages     QuestMessages    `json:"messages"`
	TaskConfig   *TaskConfig      `json:"task_config"`
	TaskConfigV2 *TaskConfig      `json:"task_config_v2"`
}

// ApplicationMeta describes the primary game or Discord application associated with the quest.
type ApplicationMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}

// QuestMessages contains user-visible campaign strings.
type QuestMessages struct {
	QuestName     *string `json:"quest_name"`
	GameTitle     *string `json:"game_title"`
	GamePublisher *string `json:"game_publisher"`
}

// TaskConfig maps task types to requirement specifications.
type TaskConfig struct {
	Tasks map[string]TaskRequirement `json:"tasks"`
}

// TaskRequirement defines criteria (target seconds, target apps) for a task type.
type TaskRequirement struct {
	Type         string           `json:"type"`
	Target       int              `json:"target"` // Total target seconds (e.g. 900)
	Applications []ApplicationRef `json:"applications"`
}

// ApplicationRef references a detectable application ID.
type ApplicationRef struct {
	ID string `json:"id"`
}

// UserStatus stores the authenticated user's progress and enrollment state.
type UserStatus struct {
	UserID             string                  `json:"user_id"`
	QuestID            string                  `json:"quest_id"`
	EnrolledAt         *string                 `json:"enrolled_at"`
	CompletedAt        *string                 `json:"completed_at"`
	ClaimedAt          *string                 `json:"claimed_at"`
	Progress           map[string]TaskProgress `json:"progress"`
	OrbQuantityClaimed int                     `json:"orb_quantity_claimed"`
}

// TaskProgress records elapsed time or progress units for a specific task type.
type TaskProgress struct {
	Value       float64 `json:"value"` // Elapsed seconds
	EventName   string  `json:"event_name"`
	UpdatedAt   *string `json:"updated_at"`
	CompletedAt *string `json:"completed_at"`
}

// EnrollPayload is sent to POST /api/v9/quests/{qid}/enroll.
type EnrollPayload struct {
	Location              int     `json:"location"`
	IsTargeted            bool    `json:"is_targeted"`
	MetadataRaw           *string `json:"metadata_raw"`
	MetadataSealed        *string `json:"metadata_sealed"`
	TrafficMetadataRaw    *string `json:"traffic_metadata_raw"`
	TrafficMetadataSealed *string `json:"traffic_metadata_sealed"`
}

// RateLimitResponse parses HTTP 429 JSON payload.
type RateLimitResponse struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

// VideoProgressPayload is sent to POST /api/v9/quests/{qid}/video-progress.
type VideoProgressPayload struct {
	Timestamp float64 `json:"timestamp"`
}

// HeartbeatPayload is sent to POST /api/v9/quests/{qid}/heartbeat.
type HeartbeatPayload struct {
	StreamKey string `json:"stream_key"`
	Terminal  bool   `json:"terminal"`
}

// ProgressUpdateResponse is returned by heartbeat and video-progress endpoints.
type ProgressUpdateResponse struct {
	UserID      string                  `json:"user_id"`
	QuestID     string                  `json:"quest_id"`
	CompletedAt *string                 `json:"completed_at"`
	Progress    map[string]TaskProgress `json:"progress"`
}
