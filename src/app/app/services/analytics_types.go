package services

// analytics_types.go
// Type-safe Go structs for Python Statistical Analysis API responses
// Simplified for single-participant descriptive statistics

// ===== Root Analysis Response =====

// AnalysisResponse contains all analysis results from Python API
type AnalysisResponse struct {
	ExecutiveSummary         *ExecutiveSummaryAnalysis      `json:"executive_summary"`
	ModelComparison          *ModelComparisonAnalysis       `json:"model_comparison"`
	LinguisticPatterns       *LinguisticPatternsAnalysis    `json:"linguistic_patterns"`
	ErrorCosts               *ErrorCostsAnalysis            `json:"error_costs"`
	InterRaterReliability    *InterRaterReliabilityAnalysis `json:"inter_rater_reliability"`
	ErrorAttribution         *ErrorAttributionAnalysis      `json:"error_attribution"`
	WordDifficulty           *WordDifficultyAnalysis        `json:"word_difficulty"`
	CharacterErrors          *CharacterErrorsAnalysis       `json:"character_errors"`
	ConsistencyReliability   *ConsistencyReliabilityAnalysis `json:"consistency_reliability"`
	QualitativeExamples      *QualitativeExamplesAnalysis   `json:"qualitative_examples"`
	StudyDesign              *StudyDesignAnalysis           `json:"study_design"`
	PracticalRecommendations *PracticalRecommendationsAnalysis `json:"practical_recommendations"`
	WordLength               *WordLengthAnalysis            `json:"word_length"`
}

// ===== 1. Model Comparison Analysis =====

// ModelComparisonAnalysis contains descriptive comparison between Whisper and Wav2Vec2
// Focus on CER (more meaningful for single-word prompts)
// Contains both raw (strict) and normalized (pedagogical) metrics
type ModelComparisonAnalysis struct {
	SampleSize int                       `json:"sample_size"`
	Raw        *ModelComparisonMetrics   `json:"raw"`        // Raw (strict) metrics
	Normalized *ModelComparisonMetrics   `json:"normalized"` // Normalized (lenient) metrics
}

// ModelComparisonMetrics contains CER and WER analysis for a specific metric type
type ModelComparisonMetrics struct {
	CERAnalysis *ErrorRateAnalysis `json:"cer_analysis"` // Primary metric
	WERAnalysis *ErrorRateAnalysis `json:"wer_analysis"` // Secondary (binary for single words)
}

// ErrorRateAnalysis contains descriptive statistics for error rates (WER or CER)
type ErrorRateAnalysis struct {
	Whisper    *ModelStats      `json:"whisper"`
	Wav2Vec2   *ModelStats      `json:"wav2vec2"`
	Difference *DifferenceStats `json:"difference"`
	EffectSize *EffectSize      `json:"effect_size"`
}

// ModelStats contains descriptive statistics for a single model
type ModelStats struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	Std    float64 `json:"std"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// DifferenceStats contains mean difference between models
type DifferenceStats struct {
	Mean float64 `json:"mean"`
}

// EffectSize contains Cohen's d effect size (practical significance)
type EffectSize struct {
	CohensD        float64 `json:"cohens_d"`
	Interpretation string  `json:"interpretation"` // "negligible", "small", "medium", "large"
}

// ===== 2. Inter-Rater Reliability Analysis =====

// InterRaterReliabilityAnalysis measures simple agreement between models
type InterRaterReliabilityAnalysis struct {
	SampleSize      int             `json:"sample_size"`
	AgreementRate   float64         `json:"agreement_rate"`
	AgreementCounts *AgreementCounts `json:"agreement_counts"`
}

// AgreementCounts shows model agreement patterns
type AgreementCounts struct {
	BothCorrect         int `json:"both_correct"`
	BothIncorrect       int `json:"both_incorrect"`
	WhisperOnlyCorrect  int `json:"whisper_only_correct"`
	Wav2Vec2OnlyCorrect int `json:"wav2vec2_only_correct"`
}

// ===== 3. Error Attribution Analysis =====

// ErrorAttributionAnalysis shows error type distributions
type ErrorAttributionAnalysis struct {
	Raw        *ErrorAttributionMetrics `json:"raw"`
	Normalized *ErrorAttributionMetrics `json:"normalized"`
}

// ErrorAttributionMetrics contains error distributions for both models
type ErrorAttributionMetrics struct {
	WhisperDistribution  *ErrorDistribution `json:"whisper_distribution"`
	Wav2Vec2Distribution *ErrorDistribution `json:"wav2vec2_distribution"`
}

// ErrorDistribution contains counts and percentages for error types
type ErrorDistribution struct {
	Counts      map[string]int     `json:"counts"`      // "ASR_ERROR", "USER_ERROR", "AMBIGUOUS", "CORRECT", "NONE"
	Percentages map[string]float64 `json:"percentages"` // Same keys as Counts
}

// ===== 4. Word Difficulty Analysis =====

// WordDifficultyAnalysis ranks words by difficulty
type WordDifficultyAnalysis struct {
	Raw        *WordDifficultyMetrics `json:"raw"`
	Normalized *WordDifficultyMetrics `json:"normalized"`
}

type WordDifficultyMetrics struct {
	TotalUniqueWords int                    `json:"total_unique_words"`
	MostDifficult    []WordDifficultyItem   `json:"most_difficult"` // Top 10
	Easiest          []WordDifficultyItem   `json:"easiest"`        // Top 10
	AllWords         []WordDifficultyItem   `json:"all_words"`      // Full list
	Distribution     *ErrorDistributionData `json:"distribution"`   // Histogram data
}

// WordDifficultyItem contains metrics for a single word
type WordDifficultyItem struct {
	Word           string  `json:"word"`
	Count          int     `json:"count"`          // Number of recordings
	Attempts       int     `json:"attempts"`       // Alias for count
	WhisperAvgWER  float64 `json:"whisper_avg_wer"`
	Wav2Vec2AvgWER float64 `json:"wav2vec2_avg_wer"`
	WhisperAvgCER  float64 `json:"whisper_avg_cer"`
	Wav2Vec2AvgCER float64 `json:"wav2vec2_avg_cer"`
	AverageCER     float64 `json:"average_cer"`     // Average across both models
	WordLength     int     `json:"word_length"`
}

// ErrorDistributionData contains histogram data for error rate distribution
type ErrorDistributionData struct {
	Buckets        []string `json:"buckets"`         // ["0-10%", "10-20%", ...]
	WhisperCounts  []int    `json:"whisper_counts"`  // Word counts per bucket
	Wav2Vec2Counts []int    `json:"wav2vec2_counts"` // Word counts per bucket
}

// ===== NEW ANALYSIS TYPES =====

// ExecutiveSummaryAnalysis contains high-level summary
type ExecutiveSummaryAnalysis struct {
	TotalRecordings   int                       `json:"total_recordings"`
	TotalWords        int                       `json:"total_words"`
	OverallWinner     string                    `json:"overall_winner"`
	KeyFinding        string                    `json:"key_finding"`
	RawMetrics        *SummaryMetricsComparison `json:"raw_metrics"`
	NormalizedMetrics *SummaryMetricsComparison `json:"normalized_metrics"`
}

// SummaryMetricsComparison contains CER comparison metrics
type SummaryMetricsComparison struct {
	WhisperCERMean            float64 `json:"whisper_cer_mean"`
	Wav2Vec2CERMean           float64 `json:"wav2vec2_cer_mean"`
	PercentagePointDifference float64 `json:"percentage_point_difference"`
	RelativeImprovement       float64 `json:"relative_improvement"`
	Winner                    string  `json:"winner"`
}

// LinguisticPatternsAnalysis contains phonetic error analysis
type LinguisticPatternsAnalysis struct {
	Raw        *LinguisticPatternMetrics `json:"raw"`
	Normalized *LinguisticPatternMetrics `json:"normalized"`
	Improvement map[string]interface{} `json:"improvement"`
}

// LinguisticPatternMetrics contains pattern stats for both models
type LinguisticPatternMetrics struct {
	Whisper    *LinguisticPatternStats `json:"whisper"`
	Wav2Vec2   *LinguisticPatternStats `json:"wav2vec2"`
	Comparison map[string]interface{} `json:"comparison"`
}

type LinguisticPatternStats struct {
	Vowels     *PatternErrorStats `json:"vowels"`
	Consonants *PatternErrorStats `json:"consonants"`
	Digraphs   *PatternErrorStats `json:"digraphs"`
	Overall    *PatternErrorStats `json:"overall"`
}

type PatternErrorStats struct {
	Total           int                      `json:"total_characters"`
	Errors          int                      `json:"error_count"`
	ErrorRate       float64                  `json:"error_rate"`
	ConfusionMatrix []ConfusionMatrixEntry `json:"confusion_matrix"`
}

// ErrorCostsAnalysis contains pedagogical error cost metrics
type ErrorCostsAnalysis struct {
	Raw        *ErrorCostsMetrics `json:"raw"`
	Normalized *ErrorCostsMetrics `json:"normalized"`
	Improvement map[string]interface{} `json:"improvement"`
}

// ErrorCostsMetrics contains costs for both models
type ErrorCostsMetrics struct {
	Whisper    *PedagogicalCosts `json:"whisper"`
	Wav2Vec2   *PedagogicalCosts `json:"wav2vec2"`
	Comparison map[string]interface{} `json:"comparison"`
}

type PedagogicalCosts struct {
	FalseAcceptance *ErrorCostDetail `json:"false_acceptance"`
	FalseRejection  *ErrorCostDetail `json:"false_rejection"`
	TruePositive    *ErrorCostDetail `json:"true_positive"`
	TrueNegative    *ErrorCostDetail `json:"true_negative"`
	TotalCases      int              `json:"total_cases"`
	PedagogicalRiskScore   float64 `json:"pedagogical_risk_score"`
	PedagogicalSafetyScore float64 `json:"pedagogical_safety_score"`
}

// ErrorCostDetail contains count and rate for an error type
type ErrorCostDetail struct {
	Count int      `json:"count"`
	Rate  float64  `json:"rate"`
	Examples []map[string]interface{} `json:"examples,omitempty"`
}

// CharacterErrorsAnalysis contains character-level error analysis
type CharacterErrorsAnalysis struct {
	Whisper  *ModelCharacterErrors `json:"whisper"`
	Wav2Vec2 *ModelCharacterErrors `json:"wav2vec2"`
}

type ModelCharacterErrors struct {
	ConfusionMatrix []ConfusionMatrixEntry      `json:"confusion_matrix"`
	PerCharacter    map[string]CharacterStats `json:"per_character"`
}

type ConfusionMatrixEntry struct {
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Count    int    `json:"count"`
}

type CharacterStats struct {
	Total     int     `json:"total"`
	Errors    int     `json:"errors"`
	ErrorRate float64 `json:"error_rate"`
}

// ConsistencyReliabilityAnalysis contains variance and stability metrics
type ConsistencyReliabilityAnalysis struct {
	Raw        *ConsistencyMetricsSet `json:"raw"`
	Normalized *ConsistencyMetricsSet `json:"normalized"`
}

// ConsistencyMetricsSet contains consistency metrics for both models
type ConsistencyMetricsSet struct {
	Whisper  *ConsistencyMetrics `json:"whisper"`
	Wav2Vec2 *ConsistencyMetrics `json:"wav2vec2"`
}

type ConsistencyMetrics struct {
	CV          float64 `json:"cv"`           // Coefficient of Variation
	IQR         float64 `json:"iqr"`          // Interquartile Range
	Percentile95 float64 `json:"percentile_95"` // 95th percentile
}

// QualitativeExamplesAnalysis contains representative examples
type QualitativeExamplesAnalysis struct {
	BothCorrect            []TranscriptionExample `json:"both_correct"`
	WhisperOnly            []TranscriptionExample `json:"whisper_only"`
	Wav2Vec2Only           []TranscriptionExample `json:"wav2vec2_only"`
	PedagogicallyCritical  []TranscriptionExample `json:"pedagogically_critical"`
}

type TranscriptionExample struct {
	Target         string `json:"target"`
	Human          string `json:"human"`
	Whisper        string `json:"whisper"`
	Wav2Vec2       string `json:"wav2vec2"`
	CriticalModels string `json:"critical_models,omitempty"`
}

// StudyDesignAnalysis contains methodology description
type StudyDesignAnalysis struct {
	StudyType              string                  `json:"study_type"`
	ResearchDesign         *ResearchDesign         `json:"research_design"`
	DataCollection         *DataCollection         `json:"data_collection"`
	StatisticalApproach    *StatisticalApproach    `json:"statistical_approach"`
	Limitations            []string                `json:"limitations"`
	Strengths              []string                `json:"strengths"`
	EthicalConsiderations  *EthicalConsiderations  `json:"ethical_considerations"`
}

type ResearchDesign struct {
	Type             string `json:"type"`
	Participants     int    `json:"participants"`
	AsrModels        int    `json:"asr_models"`
	WordsPerModel    int    `json:"words_per_model"`
	UniqueWords      int    `json:"unique_words"`
	TotalRecordings  int    `json:"total_recordings"`
}

type DataCollection struct {
	Task                string   `json:"task"`
	RecordingsPerWord   string   `json:"recordings_per_word"`
	AsrModelsCompared   []string `json:"asr_models_compared"`
	DateRange           string   `json:"date_range"`
}

type StatisticalApproach struct {
	Paradigm          string   `json:"paradigm"`
	Rationale         string   `json:"rationale"`
	PrimaryMetrics    []string `json:"primary_metrics"`
	SecondaryMetrics  []string `json:"secondary_metrics"`
	EffectSizeMeasure string   `json:"effect_size_measure"`
}

type EthicalConsiderations struct {
	DataPrivacy     string `json:"data_privacy"`
	InformedConsent string `json:"informed_consent"`
	DataUsage       string `json:"data_usage"`
}

// PracticalRecommendationsAnalysis contains actionable guidance
type PracticalRecommendationsAnalysis struct {
	PrimaryRecommendation                *PrimaryRecommendation `json:"primary_recommendation"`
	UseCases                             *UseCases              `json:"use_cases"`
	FeaturesRequiringHumanVerification   []string               `json:"features_requiring_human_verification"`
	ConfidenceThresholds                 *ConfidenceThresholds  `json:"confidence_thresholds"`
	SystemDesignImplications             []string               `json:"system_design_implications"`
	PedagogicalConsiderations            []string               `json:"pedagogical_considerations"`
}

type PrimaryRecommendation struct {
	RecommendedModel string `json:"recommended_model"`
	Rationale        string `json:"rationale"`
	Confidence       string `json:"confidence"`
}

type UseCases struct {
	WhisperPreferred  []string `json:"whisper_preferred"`
	Wav2Vec2Preferred []string `json:"wav2vec2_preferred"`
	BothAcceptable    []string `json:"both_acceptable"`
}

type ConfidenceThresholds struct {
	HighConfidenceThreshold   string `json:"high_confidence_threshold"`
	MediumConfidenceThreshold string `json:"medium_confidence_threshold"`
	LowConfidenceThreshold    string `json:"low_confidence_threshold"`
	Rationale                 string `json:"rationale"`
}

// ===== Word Length Analysis =====

// WordLengthAnalysis contains performance metrics grouped by word length categories
type WordLengthAnalysis struct {
	Raw        *WordLengthMetrics `json:"raw"`
	Normalized *WordLengthMetrics `json:"normalized"`
}

// WordLengthMetrics contains analysis for each length category
type WordLengthMetrics struct {
	OneToThree  *LengthCategoryStats `json:"one_to_three"`
	FourToSix   *LengthCategoryStats `json:"four_to_six"`
	SevenToNine *LengthCategoryStats `json:"seven_to_nine"`
	TenPlus     *LengthCategoryStats `json:"ten_plus"`
}

// LengthCategoryStats contains metrics for a specific length category
type LengthCategoryStats struct {
	Label       string  `json:"label"`
	Count       int     `json:"count"`
	WhisperCER  float64 `json:"whisper_avg_cer"`
	Wav2Vec2CER float64 `json:"wav2vec2_avg_cer"`
	WhisperWER  float64 `json:"whisper_avg_wer"`
	Wav2Vec2WER float64 `json:"wav2vec2_avg_wer"`
}

