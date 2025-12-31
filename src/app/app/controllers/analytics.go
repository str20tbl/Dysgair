package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"app/app/models"
	"app/app/services"

	"github.com/str20tbl/revel"
)

type Analytics struct {
	AuthController
}

// Index displays the analytics dashboard
func (a *Analytics) Index() revel.Result {
	users, err := models.GetNormalUsers(a.Txn)
	if err != nil {
		revel.AppLog.Error(err.Error())
		users = []models.User{} // Fallback to empty slice
	}
	return a.Render(users)
}

// RunAnalysis fetches entries and runs statistical analysis
func (a *Analytics) RunAnalysis(userID string) revel.Result {
	// Delegate to model layer for data fetching
	entries, err := models.GetEntriesForAnalysis(a.Txn, userID)
	if err != nil {
		revel.AppLog.Errorf("RunAnalysis: Failed to fetch entries for user %s: %v", userID, err)
		return a.JSONError("Failed to fetch entries")
	}

	if len(entries) == 0 {
		return a.JSONError("No entries found for analysis")
	}

	// Enrich entries with word-level metrics
	enrichedEntries := services.EnrichWithMetrics(entries)

	// Convert to JSON for Python API
	entriesJSON, err := json.Marshal(enrichedEntries)
	if err != nil {
		revel.AppLog.Errorf("RunAnalysis: Failed to marshal entries: %v", err)
		return a.JSONError("Failed to marshal entries")
	}

	// Call Python API for each analysis type - comprehensive MSc thesis analytics
	analyses := make(map[string]interface{})
	analysisTypes := []string{
		"executive-summary",               // NEW: High-level overview with key findings
		"model-comparison",                // EXISTING: Enhanced with new metrics
		"hybrid-best-case",                // NEW: Best-case performance with intelligent model selection
		"inter-rater-reliability",         // EXISTING: Model agreement metrics
		"linguistic-patterns",             // NEW: Vowels, consonants, digraphs analysis
		"error-costs",                     // NEW: False acceptance/rejection rates
		"error-attribution",               // EXISTING: Enhanced with pedagogical framing
		"word-difficulty",                 // EXISTING: Word-level difficulty rankings
		"word-length",                     // NEW: Performance by word length (1-3, 4-6, 7-9, 10+ chars)
		"character-errors",                // EXISTING: Character-level confusion matrices
		"consistency-reliability",         // NEW: Variance and stability metrics
		"qualitative-examples",            // NEW: Representative examples for thesis
		"study-design",                    // NEW: Methodology and limitations
		"practical-recommendations",       // NEW: Actionable guidance from findings
		"hybrid-practical-implications",   // NEW: Practical implications of hybrid system for critical CAPT areas
	}

	for _, analysisType := range analysisTypes {
		result, err := services.CallPythonAnalysis(analysisType, entriesJSON)
		if err != nil {
			revel.AppLog.Errorf("Analysis %s failed: %v", analysisType, err)
			return a.JSONError(fmt.Sprintf("Analysis %s failed", analysisType))
		}

		// Check if Python API returned an error response
		if success, ok := result["success"].(bool); ok && !success {
			errorMsg := result["error"]
			revel.AppLog.Errorf("Python API error for %s: %v", analysisType, errorMsg)
			return a.JSONError(fmt.Sprintf("Analysis %s failed: %v", analysisType, errorMsg))
		}

		// Extract the actual analysis data (not the wrapper with "success" and "analysis")
		if analysisData, ok := result["analysis"]; ok {
			analyses[strings.Replace(analysisType, "-", "_", -1)] = analysisData
		} else {
			// Fallback to entire result if "analysis" field not found
			analyses[strings.Replace(analysisType, "-", "_", -1)] = result
		}
	}

	return a.JSONSuccess(map[string]interface{}{
		"analyses": analyses,
	})
}

// ExportLaTeX generates a LaTeX export package with charts and tables
func (a *Analytics) ExportLaTeX() revel.Result {
	var requestData struct {
		Images map[string]string      `json:"images"`
		Data   map[string]interface{} `json:"data"`
	}

	body := a.Request.GetBody()

	if err := json.NewDecoder(body).Decode(&requestData); err != nil {
		revel.AppLog.Errorf("ExportLaTeX: Failed to parse request body: %v", err)
		return a.JSONError("Failed to parse request")
	}

	// Create temporary directory for ZIP contents
	timestamp := time.Now().Format("20060102_150405")
	tempDir := filepath.Join("/tmp", fmt.Sprintf("latex_export_%s", timestamp))
	if err := os.MkdirAll(filepath.Join(tempDir, "charts"), 0755); err != nil {
		revel.AppLog.Errorf("ExportLaTeX: Failed to create export directory %s: %v", tempDir, err)
		return a.JSONError("Failed to create export directory")
	}
	defer func() {
		// Best-effort cleanup: log errors but don't fail the request
		// (response has already been sent to user)
		if err := os.RemoveAll(tempDir); err != nil {
			revel.AppLog.Errorf("Failed to clean up temp directory %s: %v", tempDir, err)
		}
	}()

	// Decode and save chart images
	if err := services.DecodeAndSaveImages(requestData.Images, filepath.Join(tempDir, "charts")); err != nil {
		revel.AppLog.Errorf("ExportLaTeX: Failed to save chart images: %v", err)
		return a.JSONError("Failed to save chart images")
	}

	// Convert map data to typed struct
	analysisResponse, err := convertMapToAnalysisResponse(requestData.Data)
	if err != nil {
		return a.JSONErrorf("Failed to parse analysis data: %v", err)
	}

	// Generate LaTeX document
	exporter := services.NewLaTeXExporter()
	latexContent, err := exporter.GenerateDocument(analysisResponse)
	if err != nil {
		return a.JSONErrorf("Failed to generate LaTeX document: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "analysis_summary.tex"), []byte(latexContent), 0644); err != nil {
		revel.AppLog.Errorf("ExportLaTeX: Failed to write LaTeX document: %v", err)
		return a.JSONError("Failed to write LaTeX document")
	}

	// Generate LaTeX tables
	tablesContent, err := exporter.GenerateTables(analysisResponse)
	if err != nil {
		return a.JSONErrorf("Failed to generate LaTeX tables: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "tables.tex"), []byte(tablesContent), 0644); err != nil {
		revel.AppLog.Errorf("ExportLaTeX: Failed to write tables document: %v", err)
		return a.JSONError("Failed to write tables document")
	}

	// Create ZIP archive
	zipPath := filepath.Join("/data", fmt.Sprintf("dysgair_analysis_%s.zip", timestamp))
	if err := services.CreateZIPArchive(tempDir, zipPath); err != nil {
		revel.AppLog.Errorf("ExportLaTeX: Failed to create ZIP archive %s: %v", zipPath, err)
		return a.JSONError("Failed to create ZIP archive")
	}

	// Return download URL
	downloadURL := fmt.Sprintf("/Admin/DownloadExport?file=dysgair_analysis_%s.zip", timestamp)
	return a.JSONSuccess(map[string]interface{}{
		"download_url": downloadURL,
	})
}

// DownloadExport serves the generated export file
func (a *Analytics) DownloadExport(file string) revel.Result {
	file = filepath.Base(file) // Sanitize
	filePath := filepath.Join("/data", file)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return a.RenderText("File not found")
	}

	exportFile, err := os.Open(filePath)
	if err != nil {
		return a.RenderText("Error opening file")
	}

	a.Response.Out.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file))
	a.Response.Out.Header().Set("Content-Type", "application/zip")

	return a.RenderFile(exportFile, revel.Attachment)
}

// RecoverOrphanedRecordings processes audio files in /data/recordings that don't have Entry records
// It transcribes them and matches to closest word using Levenshtein distance
func (a *Analytics) RecoverOrphanedRecordings(userID int64) revel.Result {
	revel.AppLog.Infof("RecoverOrphanedRecordings: Starting recovery for user %d", userID)

	// Validate userID
	if userID == 0 {
		return a.JSONError("Invalid userID")
	}

	// Run recovery process
	result, err := services.RecoverOrphanedRecordings(a.Txn, userID)
	if err != nil {
		revel.AppLog.Errorf("RecoverOrphanedRecordings: Recovery failed: %v", err)
		return a.JSONError(fmt.Sprintf("Recovery failed: %v", err))
	}

	// Return detailed results
	return a.JSONSuccess(map[string]interface{}{
		"orphaned_count":  result.OrphanedCount,
		"processed_count": result.ProcessedCount,
		"failed_count":    result.FailedCount,
		"entries_created": result.CreatedEntries,
		"failures":        result.Failures,
		"message": fmt.Sprintf("Processed %d orphaned recordings: %d succeeded, %d failed",
			result.OrphanedCount, result.ProcessedCount, result.FailedCount),
	})
}

// convertMapToAnalysisResponse converts untyped map data from frontend to typed struct
// This enables type-safe access throughout the LaTeX generation pipeline
func convertMapToAnalysisResponse(data map[string]interface{}) (*services.AnalysisResponse, error) {
	// Marshal map back to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Unmarshal into typed struct
	var analysisResponse services.AnalysisResponse
	if err := json.Unmarshal(jsonBytes, &analysisResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to AnalysisResponse: %w", err)
	}

	return &analysisResponse, nil
}

// ExportSection generates a LaTeX export for a single analytics section
func (a *Analytics) ExportSection() revel.Result {
	// Debug logging
	revel.AppLog.Infof("ExportSection: Request received - Method: %s, Content-Type: %s",
		a.Request.Method, a.Request.Header.Get("Content-Type"))

	var requestData struct {
		Section string                 `json:"section"`
		Images  map[string]string      `json:"images"`
		Data    map[string]interface{} `json:"data"`
	}

	// Try using Params.BindJSON instead of GetBody()
	if err := a.Params.BindJSON(&requestData); err != nil {
		revel.AppLog.Errorf("ExportSection: BindJSON failed: %v", err)
		// Fallback to GetBody() method
		body := a.Request.GetBody()
		if err := json.NewDecoder(body).Decode(&requestData); err != nil {
			revel.AppLog.Errorf("ExportSection: Failed to parse request body: %v", err)
			revel.AppLog.Errorf("ExportSection: Request URL: %s, Method: %s", a.Request.URL, a.Request.Method)
			return a.JSONError(fmt.Sprintf("Failed to parse request: %v", err))
		}
	}

	revel.AppLog.Infof("ExportSection: Successfully parsed request for section '%s' with %d images", requestData.Section, len(requestData.Images))

	// Validate section name
	validSections := map[string]bool{
		"executive-summary": true, "model-comparison": true, "linguistic-patterns": true,
		"inter-rater-reliability": true, "error-costs": true, "error-attribution": true,
		"comprehensive-errors": true, "challenging-words": true,
		"character-errors": true, "cer-distribution": true, "consistency": true,
		"examples": true, "study-design": true, "recommendations": true,
		"word-length": true, "hybrid-analysis": true,
	}
	if !validSections[requestData.Section] {
		return a.JSONError(fmt.Sprintf("Invalid section name: %s", requestData.Section))
	}

	// Create temporary directory for ZIP contents
	timestamp := time.Now().Format("20060102_150405")
	tempDir := filepath.Join("/tmp", fmt.Sprintf("latex_section_%s_%s", requestData.Section, timestamp))
	if err := os.MkdirAll(filepath.Join(tempDir, "charts"), 0755); err != nil {
		revel.AppLog.Errorf("ExportSection: Failed to create export directory %s: %v", tempDir, err)
		return a.JSONError("Failed to create export directory")
	}
	defer func() {
		// Best-effort cleanup
		if err := os.RemoveAll(tempDir); err != nil {
			revel.AppLog.Errorf("Failed to clean up temp directory %s: %v", tempDir, err)
		}
	}()

	// Decode and save chart images
	if err := services.DecodeAndSaveImages(requestData.Images, filepath.Join(tempDir, "charts")); err != nil {
		revel.AppLog.Errorf("ExportSection: Failed to save chart images: %v", err)
		return a.JSONError("Failed to save chart images")
	}

	// Convert map data to typed struct
	analysisResponse, err := convertMapToAnalysisResponse(requestData.Data)
	if err != nil {
		return a.JSONErrorf("Failed to parse analysis data: %v", err)
	}

	// Generate LaTeX document for this section
	exporter := services.NewLaTeXExporter()
	latexContent, err := exporter.GenerateSectionDocument(requestData.Section, analysisResponse)
	if err != nil {
		return a.JSONErrorf("Failed to generate LaTeX document for section %s: %v", requestData.Section, err)
	}

	sectionFileName := fmt.Sprintf("%s.tex", strings.Replace(requestData.Section, "-", "_", -1))
	if err := os.WriteFile(filepath.Join(tempDir, sectionFileName), []byte(latexContent), 0644); err != nil {
		revel.AppLog.Errorf("ExportSection: Failed to write LaTeX document: %v", err)
		return a.JSONError("Failed to write LaTeX document")
	}

	// Create README
	readmeContent := fmt.Sprintf(`# %s Export

Generated: %s

## Contents
- %s - LaTeX source for this section
- charts/ - Chart images referenced in the document

## Compilation
pdflatex %s

Note: You may need to run pdflatex twice to resolve references.
`, strings.Title(strings.Replace(requestData.Section, "-", " ", -1)),
		time.Now().Format("2006-01-02 15:04:05"),
		sectionFileName, sectionFileName)

	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		revel.AppLog.Errorf("ExportSection: Failed to write README: %v", err)
		return a.JSONError("Failed to write README")
	}

	// Create ZIP archive
	zipPath := filepath.Join("/data", fmt.Sprintf("dysgair_%s_%s.zip", requestData.Section, timestamp))
	if err := services.CreateZIPArchive(tempDir, zipPath); err != nil {
		revel.AppLog.Errorf("ExportSection: Failed to create ZIP archive %s: %v", zipPath, err)
		return a.JSONError("Failed to create ZIP archive")
	}

	// Return download URL
	downloadFileName := fmt.Sprintf("dysgair_%s_%s.zip", requestData.Section, timestamp)
	downloadURL := fmt.Sprintf("/Admin/DownloadExport?file=%s", downloadFileName)
	return a.JSONSuccess(map[string]interface{}{
		"download_url": downloadURL,
	})
}
