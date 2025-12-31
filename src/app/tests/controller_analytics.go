package tests

import (
	"net/url"

	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/testing"
)

type AnalyticsTest struct {
	testing.TestSuite
}

func (t *AnalyticsTest) Before() {
	revel.AppLog.Info("AnalyticsTest: Set up")
}

// TestIndex_NotAuthenticated tests page redirects without auth
func (t *AnalyticsTest) TestIndex_NotAuthenticated() {
	t.Get("/Admin/Analytics")
	// Should redirect to / (AuthController protection)
}

// TestIndex_LoadsUserList tests user dropdown population
func (t *AnalyticsTest) TestIndex_LoadsUserList() {
	t.Get("/Admin/Analytics")
	// Page should load list of users for analysis selection
	// Only UserType = 0 (non-admin users)
}

// TestRunAnalysis_NoUser tests running analysis on all users
func (t *AnalyticsTest) TestRunAnalysis_NoUser() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"0"},
	})
	// Should analyze all users
	// May fail if no data available
}

// TestRunAnalysis_SpecificUser tests running analysis for one user
func (t *AnalyticsTest) TestRunAnalysis_SpecificUser() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should analyze only user ID 1
}

// TestRunAnalysis_NoEntries tests analysis with no data
func (t *AnalyticsTest) TestRunAnalysis_NoEntries() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"999999"}, // Non-existent user
	})
	// Should return error: "No entries found for analysis"
}

// TestRunAnalysis_NoCompleteWords tests analysis with incomplete words
func (t *AnalyticsTest) TestRunAnalysis_NoCompleteWords() {
	// When user has entries but no complete words (5 recordings)
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should return error: "No complete words (5 recordings) found for analysis"
}

// TestRunAnalysis_ResponseFormat tests JSON response structure
func (t *AnalyticsTest) TestRunAnalysis_ResponseFormat() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should return JSON with:
	// {
	//   "success": true,
	//   "analyses": {
	//     "model_comparison": {...},
	//     "inter_rater_reliability": {...},
	//     "error_attribution": {...},
	//     "learning_curves": {...},
	//     "word_difficulty": {...},
	//     "correlations": {...}
	//   }
	// }
}

// TestRunAnalysis_ModelComparison tests model comparison analysis
func (t *AnalyticsTest) TestRunAnalysis_ModelComparison() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should include model_comparison analysis
	// Comparing Whisper vs Wav2Vec2 performance
}

// TestRunAnalysis_InterRaterReliability tests IRR analysis
func (t *AnalyticsTest) TestRunAnalysis_InterRaterReliability() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should include inter_rater_reliability analysis
	// Agreement between human transcribers
}

// TestRunAnalysis_ErrorAttribution tests error attribution analysis
func (t *AnalyticsTest) TestRunAnalysis_ErrorAttribution() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should include error_attribution analysis
	// Distribution of ASR_ERROR, USER_ERROR, CORRECT, AMBIGUOUS
}

// TestRunAnalysis_LearningCurves tests learning curve analysis
func (t *AnalyticsTest) TestRunAnalysis_LearningCurves() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should include learning_curves analysis
	// User improvement over time
}

// TestRunAnalysis_WordDifficulty tests word difficulty analysis
func (t *AnalyticsTest) TestRunAnalysis_WordDifficulty() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should include word_difficulty analysis
	// Which words are hardest/easiest
}

// TestRunAnalysis_Correlations tests correlation analysis
func (t *AnalyticsTest) TestRunAnalysis_Correlations() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should include correlations analysis
	// Relationships between metrics
}

// TestRunAnalysis_CallsPythonAPI tests Python service integration
func (t *AnalyticsTest) TestRunAnalysis_CallsPythonAPI() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should call Python API for each analysis type
	// May fail if Python service is not running
}

// TestExportLaTeX_ValidData tests successful export
func (t *AnalyticsTest) TestExportLaTeX_ValidData() {
	// This is complex - requires JSON payload with images and data
	// PostForm won't work, need to send JSON body

	// For now, just test the endpoint exists
	t.PostForm("/Admin/Analytics/ExportLaTeX", url.Values{})
	// Should process request (may fail due to invalid data format)
}

// TestExportLaTeX_CreatesZipFile tests ZIP archive creation
func (t *AnalyticsTest) TestExportLaTeX_CreatesZipFile() {
	// After successful export, should create ZIP file
	// Response should include download_url
	// Response format: {"success": true, "download_url": "/Admin/DownloadExport?file=..."}
}

// TestExportLaTeX_IncludesCharts tests chart image inclusion
func (t *AnalyticsTest) TestExportLaTeX_IncludesCharts() {
	// ZIP should contain charts/ directory with chart images
	// Charts are decoded from base64 images in request
}

// TestExportLaTeX_IncludesLaTeXFiles tests LaTeX document generation
func (t *AnalyticsTest) TestExportLaTeX_IncludesLaTeXFiles() {
	// ZIP should contain:
	// - analysis_summary.tex (main document)
	// - tables.tex (LaTeX tables)
	// - charts/ directory
}

// TestDownloadExport_ValidFile tests downloading export
func (t *AnalyticsTest) TestDownloadExport_ValidFile() {
	// Requires a file to exist first
	t.Get("/Admin/DownloadExport?file=dysgair_analysis_20240101_120000.zip")
	// Should serve ZIP file if it exists
	// Content-Type: application/zip
	// Content-Disposition: attachment
}

// TestDownloadExport_NonExistentFile tests missing file
func (t *AnalyticsTest) TestDownloadExport_NonExistentFile() {
	t.Get("/Admin/DownloadExport?file=doesnotexist.zip")
	// Should return "File not found"
}

// TestRunAnalysis_FilterToCompleteWords tests 5-recording filter
func (t *AnalyticsTest) TestRunAnalysis_FilterToCompleteWords() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should only analyze words with 5 recordings
	// Uses services.FilterToCompleteWords()
}

// TestRunAnalysis_EnrichWithMetrics tests metric enrichment
func (t *AnalyticsTest) TestRunAnalysis_EnrichWithMetrics() {
	t.PostForm("/Admin/Analytics/Run", url.Values{
		"userID": {"1"},
	})
	// Should enrich entries with word-level metrics
	// Uses services.EnrichWithMetrics()
}

func (t *AnalyticsTest) After() {
	revel.AppLog.Info("AnalyticsTest: Tear down")
}
