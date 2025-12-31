package models

// EnrichedEntry extends Entry with calculated metrics for analytics
type EnrichedEntry struct {
	Entry // Embed all Entry fields

	// Attempt tracking
	AttemptNumber int `json:"AttemptNumber"`

	// Per-word aggregated metrics - STRICT (average across 5 attempts)
	AvgWERWhisper  float64 `json:"AvgWERWhisper"`
	AvgCERWhisper  float64 `json:"AvgCERWhisper"`
	AvgWERWav2Vec2 float64 `json:"AvgWERWav2Vec2"`
	AvgCERWav2Vec2 float64 `json:"AvgCERWav2Vec2"`

	// Per-word aggregated metrics - LENIENT (normalized, pedagogically relevant)
	AvgWERWhisperLenient  float64 `json:"AvgWERWhisperLenient"`
	AvgCERWhisperLenient  float64 `json:"AvgCERWhisperLenient"`
	AvgWERWav2Vec2Lenient float64 `json:"AvgWERWav2Vec2Lenient"`
	AvgCERWav2Vec2Lenient float64 `json:"AvgCERWav2Vec2Lenient"`

	// Improvement metrics - STRICT (attempt 1 - attempt 5)
	ImprovementWERWhisper  float64 `json:"ImprovementWERWhisper"`
	ImprovementWERWav2Vec2 float64 `json:"ImprovementWERWav2Vec2"`

	// Improvement metrics - LENIENT
	ImprovementWERWhisperLenient  float64 `json:"ImprovementWERWhisperLenient"`
	ImprovementWERWav2Vec2Lenient float64 `json:"ImprovementWERWav2Vec2Lenient"`

	// Best attempt metrics - STRICT
	BestWERWhisper  float64 `json:"BestWERWhisper"`
	BestWERWav2Vec2 float64 `json:"BestWERWav2Vec2"`

	// Best attempt metrics - LENIENT
	BestWERWhisperLenient  float64 `json:"BestWERWhisperLenient"`
	BestWERWav2Vec2Lenient float64 `json:"BestWERWav2Vec2Lenient"`

	// First and last attempt metrics - STRICT
	FirstAttemptWERWhisper  float64 `json:"FirstAttemptWERWhisper"`
	FirstAttemptWERWav2Vec2 float64 `json:"FirstAttemptWERWav2Vec2"`
	LastAttemptWERWhisper   float64 `json:"LastAttemptWERWhisper"`
	LastAttemptWERWav2Vec2  float64 `json:"LastAttemptWERWav2Vec2"`

	// First and last attempt metrics - LENIENT
	FirstAttemptWERWhisperLenient  float64 `json:"FirstAttemptWERWhisperLenient"`
	FirstAttemptWERWav2Vec2Lenient float64 `json:"FirstAttemptWERWav2Vec2Lenient"`
	LastAttemptWERWhisperLenient   float64 `json:"LastAttemptWERWhisperLenient"`
	LastAttemptWERWav2Vec2Lenient  float64 `json:"LastAttemptWERWav2Vec2Lenient"`

	// Completion tracking
	WordCompletionCount int `json:"WordCompletionCount"`
}

// WordMetrics contains aggregated metrics for a specific word
type WordMetrics struct {
	UserID  int64
	WordID  int64
	Word    string
	Entries []Entry

	// Calculated metrics - STRICT (raw)
	AvgWERWhisper          float64
	AvgCERWhisper          float64
	AvgWERWav2Vec2         float64
	AvgCERWav2Vec2         float64
	ImprovementWERWhisper  float64
	ImprovementWERWav2Vec2 float64
	BestWERWhisper         float64
	BestWERWav2Vec2        float64

	// Calculated metrics - LENIENT (normalized, pedagogically relevant)
	AvgWERWhisperLenient          float64
	AvgCERWhisperLenient          float64
	AvgWERWav2Vec2Lenient         float64
	AvgCERWav2Vec2Lenient         float64
	ImprovementWERWhisperLenient  float64
	ImprovementWERWav2Vec2Lenient float64
	BestWERWhisperLenient         float64
	BestWERWav2Vec2Lenient        float64
}
