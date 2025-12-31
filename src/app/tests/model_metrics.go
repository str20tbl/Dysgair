package tests

import (
	"app/app/models"

	"github.com/str20tbl/revel"
	"github.com/str20tbl/revel/testing"
)

type MetricsTest struct {
	testing.TestSuite
}

func (t *MetricsTest) Before() {
	revel.AppLog.Info("MetricsTest: Set up")
}

// TestCharDistance_EmptyStrings tests empty string edge cases
func (t *MetricsTest) TestCharDistance_EmptyStrings() {
	distance, ops := models.CharDistance("", "")
	t.Assert(distance == 0)
	t.Assert(ops.Total == 0)

	distance, ops = models.CharDistance("hello", "")
	t.Assert(distance == 5)
	t.Assert(ops.Total == 5)

	distance, ops = models.CharDistance("", "world")
	t.Assert(distance == 5)
	t.Assert(ops.Total == 5)
}

// TestCharDistance_IdenticalStrings tests matching strings
func (t *MetricsTest) TestCharDistance_IdenticalStrings() {
	distance, ops := models.CharDistance("hello", "hello")
	t.Assert(distance == 0)
	t.Assert(ops.Total == 0)

	// Case insensitive
	distance, ops = models.CharDistance("Hello", "hello")
	t.Assert(distance == 0)
	t.Assert(ops.Total == 0)

	distance, ops = models.CharDistance("CYMRAEG", "cymraeg")
	t.Assert(distance == 0)
	t.Assert(ops.Total == 0)
}

// TestCharDistance_CompletelyDifferent tests completely different strings
func (t *MetricsTest) TestCharDistance_CompletelyDifferent() {
	distance, ops := models.CharDistance("abc", "xyz")
	t.Assert(distance == 3)
	t.Assert(len(ops.Substitutions) == 3)
}

// TestCharDistance_SingleOperations tests individual operation types
func (t *MetricsTest) TestCharDistance_SingleOperations() {
	// Single substitution
	distance, ops := models.CharDistance("cat", "bat")
	t.Assert(distance == 1)
	t.Assert(len(ops.Substitutions) == 1)

	// Single insertion
	distance, ops = models.CharDistance("cat", "cats")
	t.Assert(distance == 1)
	t.Assert(len(ops.Insertions) == 1)

	// Single deletion
	distance, ops = models.CharDistance("cats", "cat")
	t.Assert(distance == 1)
	t.Assert(len(ops.Deletions) == 1)
}

// TestCharDistance_WelshWords tests with actual Welsh words
func (t *MetricsTest) TestCharDistance_WelshWords() {
	// Common Welsh pronunciation error
	distance, _ := models.CharDistance("cymraeg", "cymrag")
	t.Assert(distance == 1) // One deletion ('e')

	// Completely wrong
	distance, _ = models.CharDistance("bore", "hwyl")
	t.Assert(distance == 4)
}

// TestCharDistance_UnicodeSupport tests Unicode/diacritics
func (t *MetricsTest) TestCharDistance_UnicodeSupport() {
	distance, _ := models.CharDistance("café", "cafe")
	t.Assert(distance == 1) // é vs e

	distance, _ = models.CharDistance("Cymraeg", "Cymraeg")
	t.Assert(distance == 0)
}

// TestWordDistance_EmptyStrings tests word-level empty cases
func (t *MetricsTest) TestWordDistance_EmptyStrings() {
	distance, ops := models.WordDistance("", "")
	t.Assert(distance == 0)
	t.Assert(ops.Total == 0)

	distance, ops = models.WordDistance("hello world", "")
	t.Assert(distance == 2)

	distance, ops = models.WordDistance("", "hello world")
	t.Assert(distance == 2)
}

// TestWordDistance_IdenticalPhrases tests matching phrases
func (t *MetricsTest) TestWordDistance_IdenticalPhrases() {
	distance, ops := models.WordDistance("hello world", "hello world")
	t.Assert(distance == 0)
	t.Assert(ops.Total == 0)

	// Case insensitive
	distance, ops = models.WordDistance("Hello World", "hello world")
	t.Assert(distance == 0)
}

// TestWordDistance_WordOperations tests word-level operations
func (t *MetricsTest) TestWordDistance_WordOperations() {
	// One word substitution
	distance, ops := models.WordDistance("hello world", "hello there")
	t.Assert(distance == 1)
	t.Assert(len(ops.Substitutions) == 1)

	// One word insertion
	distance, ops = models.WordDistance("hello world", "hello beautiful world")
	t.Assert(distance == 1)
	t.Assert(len(ops.Insertions) == 1)

	// One word deletion
	distance, ops = models.WordDistance("hello beautiful world", "hello world")
	t.Assert(distance == 1)
	t.Assert(len(ops.Deletions) == 1)
}

// TestCalculateWER_EdgeCases tests Word Error Rate edge cases
func (t *MetricsTest) TestCalculateWER_EdgeCases() {
	// Both empty - 0% error
	wer := models.CalculateWER("", "")
	t.Assert(wer == 0.0)

	// Empty reference, non-empty hypothesis - 100% error
	wer = models.CalculateWER("", "hello")
	t.Assert(wer == 100.0)

	// Perfect match - 0% error
	wer = models.CalculateWER("hello world", "hello world")
	t.Assert(wer == 0.0)

	// Complete mismatch
	wer = models.CalculateWER("hello world", "foo bar")
	t.Assert(wer == 100.0)
}

// TestCalculateWER_PartialErrors tests partial WER
func (t *MetricsTest) TestCalculateWER_PartialErrors() {
	// 1 error out of 2 words = 50%
	wer := models.CalculateWER("hello world", "hello there")
	t.Assert(wer == 50.0)

	// 1 error out of 3 words = 33.33%
	wer = models.CalculateWER("the quick brown", "the slow brown")
	t.Assert(wer == 33.33)

	// Extra word (insertion)
	wer = models.CalculateWER("hello", "hello world")
	t.Assert(wer == 100.0) // 1 edit / 1 reference word
}

// TestCalculateWER_WelshPhrases tests WER with Welsh language
func (t *MetricsTest) TestCalculateWER_WelshPhrases() {
	// Perfect Welsh transcription
	wer := models.CalculateWER("bore da", "bore da")
	t.Assert(wer == 0.0)

	// One word wrong
	wer = models.CalculateWER("bore da", "bore dda")
	t.Assert(wer == 50.0)
}

// TestCalculateCER_EdgeCases tests Character Error Rate edge cases
func (t *MetricsTest) TestCalculateCER_EdgeCases() {
	// Both empty - 0% error
	cer := models.CalculateCER("", "")
	t.Assert(cer == 0.0)

	// Empty reference, non-empty hypothesis - 100% error
	cer = models.CalculateCER("", "hello")
	t.Assert(cer == 100.0)

	// Perfect match - 0% error
	cer = models.CalculateCER("hello", "hello")
	t.Assert(cer == 0.0)

	// Mostly different (4 out of 5 chars different, one 'l' matches)
	cer = models.CalculateCER("hello", "world")
	t.Assert(cer == 80.0)
}

// TestCalculateCER_PartialErrors tests partial CER
func (t *MetricsTest) TestCalculateCER_PartialErrors() {
	// 1 error out of 3 chars = 33.33%
	cer := models.CalculateCER("cat", "bat")
	t.Assert(cer == 33.33)

	// 1 error out of 5 chars = 20% (e→a substitution)
	cer = models.CalculateCER("hello", "hallo")
	t.Assert(cer == 20.0)
}

// TestClassifyError_Correct tests CORRECT classification
func (t *MetricsTest) TestClassifyError_Correct() {
	// Both human and ASR correct
	attr := models.ClassifyError("bore da", "bore da", "bore da")
	t.Assert(attr == models.ErrorClassCorrect)

	// Case insensitive
	attr = models.ClassifyError("bore da", "Bore Da", "BORE DA")
	t.Assert(attr == models.ErrorClassCorrect)
}

// TestClassifyError_ASRError tests ASR_ERROR classification
func (t *MetricsTest) TestClassifyError_ASRError() {
	// Human correct, ASR wrong
	attr := models.ClassifyError("cymraeg", "cymrag", "cymraeg")
	t.Assert(attr == models.ErrorClassASR)

	attr = models.ClassifyError("hello", "helo", "hello")
	t.Assert(attr == models.ErrorClassASR)
}

// TestClassifyError_UserError tests USER_ERROR classification
func (t *MetricsTest) TestClassifyError_UserError() {
	// Human wrong, ASR matches human (correct transcription of mispronunciation)
	attr := models.ClassifyError("cymraeg", "cymrag", "cymrag")
	t.Assert(attr == models.ErrorClassUser)

	attr = models.ClassifyError("hello", "helo", "helo")
	t.Assert(attr == models.ErrorClassUser)
}

// TestClassifyError_Ambiguous tests AMBIGUOUS classification
func (t *MetricsTest) TestClassifyError_Ambiguous() {
	// Both wrong but different
	attr := models.ClassifyError("cymraeg", "kumrag", "simrag")
	t.Assert(attr == models.ErrorClassAmbiguous)

	attr = models.ClassifyError("hello", "helo", "hallo")
	t.Assert(attr == models.ErrorClassAmbiguous)
}

// TestClassifyError_EmptyHumanTranscription tests empty human transcription
func (t *MetricsTest) TestClassifyError_EmptyHumanTranscription() {
	// No human transcription yet
	attr := models.ClassifyError("cymraeg", "cymrag", "")
	t.Assert(attr == "")
}

// TestCalculateAccuracy_EdgeCases tests accuracy edge cases
func (t *MetricsTest) TestCalculateAccuracy_EdgeCases() {
	// Empty strings
	acc := models.CalculateAccuracy("", "")
	t.Assert(acc == 0.0)

	acc = models.CalculateAccuracy("hello", "")
	t.Assert(acc == 0.0)

	acc = models.CalculateAccuracy("", "world")
	t.Assert(acc == 0.0)

	// Perfect match
	acc = models.CalculateAccuracy("hello", "hello")
	t.Assert(acc == 100.0)
}

// TestCalculateAccuracy_PartialMatch tests partial accuracy
func (t *MetricsTest) TestCalculateAccuracy_PartialMatch() {
	// 4/5 characters match = 80%
	acc := models.CalculateAccuracy("hello", "hallo")
	t.Assert(acc == 80.0)

	// Completely different = 0%
	acc = models.CalculateAccuracy("cat", "dog")
	t.Assert(acc == 0.0)
}

// TestRecalculateWhisperMetrics tests Entry.RecalculateWhisperMetrics integration
func (t *MetricsTest) TestRecalculateWhisperMetrics() {
	entry := models.Entry{
		Text:           "cymraeg",
		AttemptWhisper: "cymrag",
	}

	entry.RecalculateWhisperMetrics()

	// Should have calculated WER and CER
	t.Assert(entry.WERWhisper > 0)
	t.Assert(entry.CERWhisper > 0)

	// Without human transcription, accuracy should be 0
	t.Assert(entry.TranscriptionAccuracyWhisper == 0)
	t.Assert(entry.ErrorAttributionWhisper == "")
}

// TestRecalculateWhisperMetrics_WithHumanTranscription tests with human input
func (t *MetricsTest) TestRecalculateWhisperMetrics_WithHumanTranscription() {
	entry := models.Entry{
		Text:               "cymraeg",
		AttemptWhisper:     "cymrag",
		HumanTranscription: "cymraeg",
	}

	entry.RecalculateWhisperMetrics()

	// Should have calculated all metrics
	t.Assert(entry.WERWhisper > 0)
	t.Assert(entry.CERWhisper > 0)
	t.Assert(entry.TranscriptionAccuracyWhisper > 0)
	t.Assert(entry.ErrorAttributionWhisper == models.ErrorClassASR)
	t.Assert(entry.EditOperations != "")
}

// TestRecalculateAllMetrics tests dual model metrics
func (t *MetricsTest) TestRecalculateAllMetrics() {
	entry := models.Entry{
		Text:               "cymraeg",
		AttemptWhisper:     "cymrag",
		AttemptWav2Vec2:    "cymraeg",
		HumanTranscription: "cymraeg",
	}

	entry.RecalculateAllMetrics()

	// Whisper metrics (ASR error)
	t.Assert(entry.WERWhisper > 0)
	t.Assert(entry.ErrorAttributionWhisper == models.ErrorClassASR)

	// Wav2Vec2 metrics (correct)
	t.Assert(entry.WERWav2Vec2 == 0.0)
	t.Assert(entry.ErrorAttributionWav2Vec2 == models.ErrorClassCorrect)
}

// TestGetEditOperationsJSON tests JSON conversion
func (t *MetricsTest) TestGetEditOperationsJSON() {
	_, ops := models.CharDistance("cat", "bat")
	json := models.GetEditOperationsJSON(ops)

	t.Assert(json != "")
	t.Assert(json != "{}")

	// Nil operations
	json = models.GetEditOperationsJSON(nil)
	t.Assert(json == "{}")
}

func (t *MetricsTest) After() {
	revel.AppLog.Info("MetricsTest: Tear down")
}
