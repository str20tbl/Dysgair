package models

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

// normalizeForMetrics normalizes text for consistent metric calculations by:
//   - Converting to lowercase
//   - Replacing hyphens with spaces (e.g., "well-known" → "well known")
//   - Removing all Unicode punctuation and symbols (e.g., .,!?;:"' … — – '' "")
//   - Normalizing whitespace (trim and collapse multiple spaces)
//
// This ensures fair comparison between ASR models, particularly when one model
// (Whisper) produces punctuation while another (Wav2Vec2) does not.
func normalizeForMetrics(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace hyphens with spaces to split hyphenated words
	text = strings.ReplaceAll(text, "-", " ")

	// Remove all Unicode punctuation and symbols
	text = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return -1 // remove character
		}
		return r
	}, text)

	// Normalize whitespace: trim and collapse multiple spaces
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	return text
}

// applyLenientNormalization applies space-insensitive matching between reference and hypothesis.
//
// This function checks if the reference and hypothesis are exactly equal after removing
// all spaces. For single-word references, it also matches if the hypothesis's first word
// matches the reference (ignoring ASR over-transcription).
//
// Used in "Normalized" metrics to show if ASR output is useful for the app after
// ignoring punctuation (already removed by normalizeForMetrics), spacing differences,
// and over-transcription errors.
//
// Args:
//
//	reference: Normalized reference text (should already be normalized)
//	hypothesis: Normalized hypothesis text (should already be normalized)
//
// Returns:
//
//	Tuple of (reference, hypothesis) where both are set to reference if lenient match found,
//	otherwise returns original values
//
// Examples:
//   - Reference "hello world" matches hypothesis "hel lo wor ld" → ("hello world", "hello world")
//   - Reference "rhyw" matches hypothesis "rhyw reswm wrth..." → ("rhyw", "rhyw") [first word match]
//   - Reference "rhyw" does NOT match "rhywbeth" → ("rhyw", "rhywbeth") [different words]
//   - Reference "test" doesn't match "demo" → ("test", "demo")
func applyLenientNormalization(reference, hypothesis string) (string, string) {
	// Remove all spaces from both strings for comparison
	referenceNoSpaces := strings.ReplaceAll(reference, " ", "")
	hypothesisNoSpaces := strings.ReplaceAll(hypothesis, " ", "")

	// Check for EXACT MATCH after removing spaces (not substring!)
	if referenceNoSpaces != "" && referenceNoSpaces == hypothesisNoSpaces {
		// Exact match found! Return reference for both to get perfect metrics
		return reference, reference
	}

	// For single-word references, check if hypothesis's first word matches
	// This handles ASR over-transcription (e.g., "rhyw" → "rhyw reswm wrth...")
	if reference != "" && !strings.Contains(strings.TrimSpace(reference), " ") {
		// Reference is a single word
		hypothesisWords := strings.Fields(hypothesis)
		if len(hypothesisWords) > 0 && hypothesisWords[0] == reference {
			// First word of hypothesis matches reference exactly
			return reference, reference
		}
	}

	// No match - return original values
	return reference, hypothesis
}

// calculateWER is the internal WER calculation function that operates on pre-normalized strings.
// It assumes the inputs have already been normalized according to the desired strategy.
//
// Formula: WER = (S + D + I) / N × 100
// Where:
//   - S = number of substitutions
//   - D = number of deletions
//   - I = number of insertions
//   - N = number of words in reference
//
// Returns 0.0 if both strings are empty, 100.0 if reference is empty but hypothesis is not.
func calculateWER(reference, hypothesis string) float64 {
	// Both empty is perfect match
	if reference == "" && hypothesis == "" {
		return 0.0
	}

	refWords := strings.Fields(reference)
	if len(refWords) == 0 {
		// Empty reference with non-empty hypothesis = 100% error
		if hypothesis == "" {
			return 0.0
		}
		return 100.0
	}

	// Calculate word-level edit distance
	distance, _ := WordDistance(hypothesis, reference)

	// WER = (edit distance / reference word count) × 100
	wer := float64(distance) / float64(len(refWords)) * 100.0

	return roundToTwoDecimals(wer)
}

// CalculateWERStrict calculates Word Error Rate (WER) using raw text with minimal normalization.
//
// This is the strict version that uses actual transcription output with only lowercase
// conversion for fair comparison. Shows true transcription quality including punctuation,
// spacing, and other formatting differences.
//
// Returns 0.0 if both strings are empty, 100.0 if reference is empty but hypothesis is not.
func CalculateWERStrict(reference, hypothesis string) float64 {
	// Minimal normalization: lowercase and trim only (preserve punctuation, spacing, hyphens)
	reference = strings.ToLower(strings.TrimSpace(reference))
	hypothesis = strings.ToLower(strings.TrimSpace(hypothesis))

	return calculateWER(reference, hypothesis)
}

// CalculateWER calculates Word Error Rate (WER) between reference and hypothesis transcriptions.
//
// WER is a standard metric for evaluating ASR systems, measuring the percentage of words
// that need to be substituted, deleted, or inserted to transform the hypothesis into the reference.
//
// This version applies full normalization (punctuation removal, hyphen handling) and lenient
// space-insensitive matching for a more forgiving comparison.
//
// Returns 0.0 if both strings are empty, 100.0 if reference is empty but hypothesis is not.
func CalculateWER(reference, hypothesis string) float64 {
	// Normalize text to remove punctuation and handle hyphens
	reference = normalizeForMetrics(reference)
	hypothesis = normalizeForMetrics(hypothesis)

	// Apply lenient space-insensitive matching
	reference, hypothesis = applyLenientNormalization(reference, hypothesis)

	return calculateWER(reference, hypothesis)
}

// calculateCER is the internal CER calculation function that operates on pre-normalized strings.
// It assumes the inputs have already been normalized according to the desired strategy.
//
// Formula: CER = (S + D + I) / N × 100
// Where:
//   - S = number of character substitutions
//   - D = number of character deletions
//   - I = number of character insertions
//   - N = number of characters in reference
//
// Returns 0.0 if both strings are empty, 100.0 if reference is empty but hypothesis is not.
func calculateCER(reference, hypothesis string) float64 {
	// Both empty is perfect match
	if reference == "" && hypothesis == "" {
		return 0.0
	}

	// Count characters (runes, not bytes) for proper Unicode support
	refLen := utf8.RuneCountInString(reference)
	if refLen == 0 {
		// Empty reference with non-empty hypothesis = 100% error
		if hypothesis == "" {
			return 0.0
		}
		return 100.0
	}

	// Calculate character-level edit distance
	distance, _ := CharDistance(hypothesis, reference)

	// CER = (edit distance / reference character count) × 100
	cer := float64(distance) / float64(refLen) * 100.0

	return roundToTwoDecimals(cer)
}

// CalculateCERStrict calculates Character Error Rate (CER) with minimal normalization.
//
// This is the strict version that uses actual transcription output with only lowercase
// conversion for fair comparison. Shows true transcription quality including punctuation,
// spacing, and other formatting differences.
//
// Returns 0.0 if both strings are empty, 100.0 if reference is empty but hypothesis is not.
func CalculateCERStrict(reference, hypothesis string) float64 {
	// Minimal normalization: lowercase and trim only (preserve punctuation, spacing, hyphens)
	reference = strings.ToLower(strings.TrimSpace(reference))
	hypothesis = strings.ToLower(strings.TrimSpace(hypothesis))

	return calculateCER(reference, hypothesis)
}

// CalculateCER calculates Character Error Rate (CER) between reference and hypothesis transcriptions.
//
// CER is similar to WER but operates at the character level, which can be more sensitive to
// small transcription errors and is useful for languages with complex morphology.
//
// This version applies full normalization (punctuation removal, hyphen handling) and lenient
// space-insensitive matching for a more forgiving comparison.
//
// Returns 0.0 if both strings are empty, 100.0 if reference is empty but hypothesis is not.
func CalculateCER(reference, hypothesis string) float64 {
	// Normalize text to remove punctuation and handle hyphens
	reference = normalizeForMetrics(reference)
	hypothesis = normalizeForMetrics(hypothesis)

	// Apply lenient space-insensitive matching
	reference, hypothesis = applyLenientNormalization(reference, hypothesis)

	return calculateCER(reference, hypothesis)
}

// calculateAccuracy is the internal accuracy calculation function that operates on pre-normalized strings.
// It assumes the inputs have already been normalized according to the desired strategy.
//
// Formula: Accuracy = (1 - distance/maxLength) × 100
// Where maxLength is the length of the longer string.
//
// Returns:
//   - 100.0 for perfect match
//   - 0.0 if one or both strings are empty
//   - Value between 0-100 indicating percentage similarity
func calculateAccuracy(humanTranscription, asrTranscription string) float64 {
	// Empty strings have no accuracy
	if humanTranscription == "" || asrTranscription == "" {
		return 0.0
	}

	// Calculate character-level edit distance
	distance, _ := CharDistance(asrTranscription, humanTranscription)

	// Use the longer string as the maximum possible distance
	humanLen := utf8.RuneCountInString(humanTranscription)
	asrLen := utf8.RuneCountInString(asrTranscription)
	maxLen := humanLen
	if asrLen > maxLen {
		maxLen = asrLen
	}

	// Avoid division by zero (shouldn't happen due to empty check above)
	if maxLen == 0 {
		return 100.0
	}

	// Accuracy = (1 - normalized_distance) × 100
	accuracy := (1.0 - float64(distance)/float64(maxLen)) * 100.0

	// Clamp to [0, 100] range
	if accuracy < 0 {
		accuracy = 0
	}
	if accuracy > 100 {
		accuracy = 100
	}

	return roundToTwoDecimals(accuracy)
}

// CalculateAccuracyStrict calculates accuracy with minimal normalization.
//
// This is the strict version that uses actual transcription output with only lowercase
// conversion for fair comparison. Shows true transcription quality including punctuation,
// spacing, and other formatting differences.
//
// Returns:
//   - 100.0 for perfect match
//   - 0.0 if one or both strings are empty
//   - Value between 0-100 indicating percentage similarity
func CalculateAccuracyStrict(humanTranscription, asrTranscription string) float64 {
	// Minimal normalization: lowercase and trim only (preserve punctuation, spacing, hyphens)
	humanTranscription = strings.ToLower(strings.TrimSpace(humanTranscription))
	asrTranscription = strings.ToLower(strings.TrimSpace(asrTranscription))

	return calculateAccuracy(humanTranscription, asrTranscription)
}

// CalculateAccuracy calculates the similarity between human and ASR transcriptions.
//
// Unlike WER/CER which compare against target pronunciation, accuracy measures how closely
// the ASR output matches what the human reviewer heard. This is useful for validating ASR
// performance on actual user recordings (which may contain mispronunciations).
//
// This version applies full normalization (punctuation removal, hyphen handling) and lenient
// space-insensitive matching for a more forgiving comparison.
//
// Returns:
//   - 100.0 for perfect match
//   - 0.0 if one or both strings are empty
//   - Value between 0-100 indicating percentage similarity
func CalculateAccuracy(humanTranscription, asrTranscription string) float64 {
	// Normalize text to remove punctuation and handle hyphens
	humanTranscription = normalizeForMetrics(humanTranscription)
	asrTranscription = normalizeForMetrics(asrTranscription)

	// Apply lenient space-insensitive matching
	humanTranscription, asrTranscription = applyLenientNormalization(humanTranscription, asrTranscription)

	return calculateAccuracy(humanTranscription, asrTranscription)
}

// RecalculateWhisperMetrics updates all Whisper ASR metrics for this entry.
//
// This method calculates:
//   - WER and CER (comparing target vs Whisper ASR output)
//   - Transcription accuracy (comparing human vs Whisper ASR output)
//   - Error classification (ASR error vs user error)
//   - Edit operations (for detailed analysis)
//
// Call this after updating HumanTranscription to recalculate human-dependent metrics.
func (e *Entry) RecalculateWhisperMetrics() {
	// Calculate ASR performance metrics (Target vs Whisper) - both strict and lenient
	e.WERWhisper = CalculateWERStrict(e.Text, e.AttemptWhisper)
	e.CERWhisper = CalculateCERStrict(e.Text, e.AttemptWhisper)

	// For lenient metrics, cap Whisper hypothesis to prevent hallucination inflation
	// Cap at max(target_length, Wav2Vec2_length) if Wav2Vec2 is available
	whisperHypothesis := e.AttemptWhisper
	if e.AttemptWav2Vec2 != "" {
		capLength := len(e.Text) // Start with target length
		if len(e.AttemptWav2Vec2) > capLength {
			capLength = len(e.AttemptWav2Vec2)
		}
		if len(whisperHypothesis) > capLength {
			whisperHypothesis = truncateToByteLength(whisperHypothesis, capLength)
		}
	}

	e.WERWhisperLenient = CalculateWER(e.Text, whisperHypothesis)
	e.CERWhisperLenient = CalculateCER(e.Text, whisperHypothesis)

	// If human transcription is available, calculate human-ASR comparison metrics
	if e.HumanTranscription != "" {
		// Strict versions
		e.TranscriptionAccuracyWhisper = CalculateAccuracyStrict(e.HumanTranscription, e.AttemptWhisper)
		e.ErrorAttributionWhisper = ClassifyErrorStrict(e.Text, e.AttemptWhisper, e.HumanTranscription)

		// Lenient versions
		e.TranscriptionAccuracyWhisperLenient = CalculateAccuracy(e.HumanTranscription, e.AttemptWhisper)
		e.ErrorAttributionWhisperLenient = ClassifyError(e.Text, e.AttemptWhisper, e.HumanTranscription)

		// Detailed edit operations between human and Whisper transcriptions
		_, ops := CharDistance(e.AttemptWhisper, e.HumanTranscription)
		e.EditOperations = GetEditOperationsJSON(ops)
	}

	// Also update Wav2Vec2 metrics if available
	if e.AttemptWav2Vec2 != "" {
		e.WERWav2Vec2 = CalculateWERStrict(e.Text, e.AttemptWav2Vec2)
		e.CERWav2Vec2 = CalculateCERStrict(e.Text, e.AttemptWav2Vec2)
		e.WERWav2Vec2Lenient = CalculateWER(e.Text, e.AttemptWav2Vec2)
		e.CERWav2Vec2Lenient = CalculateCER(e.Text, e.AttemptWav2Vec2)

		if e.HumanTranscription != "" {
			e.TranscriptionAccuracyWav2Vec2 = CalculateAccuracyStrict(e.HumanTranscription, e.AttemptWav2Vec2)
			e.ErrorAttributionWav2Vec2 = ClassifyErrorStrict(e.Text, e.AttemptWav2Vec2, e.HumanTranscription)
			e.TranscriptionAccuracyWav2Vec2Lenient = CalculateAccuracy(e.HumanTranscription, e.AttemptWav2Vec2)
			e.ErrorAttributionWav2Vec2Lenient = ClassifyError(e.Text, e.AttemptWav2Vec2, e.HumanTranscription)
		}
	}
}

// RecalculateAllMetrics updates metrics for both Whisper and Wav2Vec2 ASR models.
//
// This method calculates all metrics for both models:
//   - WER and CER for both models
//   - Transcription accuracy (if human transcription exists)
//   - Error classification (if human transcription exists)
//
// Use this when you need to ensure both models' metrics are up to date,
// such as after uploading a new recording or updating the human transcription.
func (e *Entry) RecalculateAllMetrics() {
	// Calculate BOTH strict and lenient versions of all metrics
	// Strict metrics provide conservative error estimates
	// Lenient metrics give benefit of the doubt for spacing variations

	// For CAPT evaluation: Always use TARGET word as reference
	// This measures "Would the system give correct feedback to learners?"
	// NOT "Did ASR correctly transcribe the audio?" (which would use HumanTranscription)
	// CAPT systems compare ASR output to target, not to actual pronunciation
	reference := e.Text

	// Calculate Whisper metrics (both strict and lenient)
	if e.AttemptWhisper != "" {
		// Strict versions (without lenient space-insensitive matching)
		e.WERWhisper = CalculateWERStrict(reference, e.AttemptWhisper)
		e.CERWhisper = CalculateCERStrict(reference, e.AttemptWhisper)

		// For lenient metrics, cap Whisper hypothesis to prevent hallucination inflation
		// Cap at max(target_length, Wav2Vec2_length) if Wav2Vec2 is available
		whisperHypothesis := e.AttemptWhisper
		if e.AttemptWav2Vec2 != "" {
			capLength := len(reference) // Start with target length
			if len(e.AttemptWav2Vec2) > capLength {
				capLength = len(e.AttemptWav2Vec2)
			}
			if len(whisperHypothesis) > capLength {
				whisperHypothesis = truncateToByteLength(whisperHypothesis, capLength)
			}
		}

		// Lenient versions (with lenient space-insensitive matching and capping)
		e.WERWhisperLenient = CalculateWER(reference, whisperHypothesis)
		e.CERWhisperLenient = CalculateCER(reference, whisperHypothesis)

		// Compute and store normalized text for analytics (use capped version)
		normalizedRef := normalizeForMetrics(reference)
		normalizedHyp := normalizeForMetrics(whisperHypothesis)
		_, e.AttemptWhisperLenient = applyLenientNormalization(normalizedRef, normalizedHyp)
	}

	// Calculate Wav2Vec2 metrics (both strict and lenient)
	if e.AttemptWav2Vec2 != "" {
		// Strict versions
		e.WERWav2Vec2 = CalculateWERStrict(reference, e.AttemptWav2Vec2)
		e.CERWav2Vec2 = CalculateCERStrict(reference, e.AttemptWav2Vec2)

		// Lenient versions
		e.WERWav2Vec2Lenient = CalculateWER(reference, e.AttemptWav2Vec2)
		e.CERWav2Vec2Lenient = CalculateCER(reference, e.AttemptWav2Vec2)

		// Compute and store normalized text for analytics (eliminates duplicate normalization in Python)
		normalizedRef := normalizeForMetrics(reference)
		normalizedHyp := normalizeForMetrics(e.AttemptWav2Vec2)
		_, e.AttemptWav2Vec2Lenient = applyLenientNormalization(normalizedRef, normalizedHyp)
	}

	// Auto-populate HumanTranscription if either ASR model has perfect lenient CER
	// Using lenient CER for auto-population as it's more forgiving of spacing issues
	if e.HumanTranscription == "" && e.Text != "" {
		if e.CERWhisperLenient == 0 || e.CERWav2Vec2Lenient == 0 {
			e.HumanTranscription = e.Text
		}
	}

	// Calculate human-dependent metrics if HumanTranscription is available
	if e.HumanTranscription != "" {
		if e.AttemptWhisper != "" {
			// Strict versions
			e.TranscriptionAccuracyWhisper = CalculateAccuracyStrict(e.HumanTranscription, e.AttemptWhisper)
			e.ErrorAttributionWhisper = ClassifyErrorStrict(e.Text, e.AttemptWhisper, e.HumanTranscription)

			// Lenient versions
			e.TranscriptionAccuracyWhisperLenient = CalculateAccuracy(e.HumanTranscription, e.AttemptWhisper)
			e.ErrorAttributionWhisperLenient = ClassifyError(e.Text, e.AttemptWhisper, e.HumanTranscription)

			// Note: EditOperations not generated here to avoid database size issues during bulk operations
			// Use RecalculateWhisperMetrics() for detailed edit operations when needed
		}
		if e.AttemptWav2Vec2 != "" {
			// Strict versions
			e.TranscriptionAccuracyWav2Vec2 = CalculateAccuracyStrict(e.HumanTranscription, e.AttemptWav2Vec2)
			e.ErrorAttributionWav2Vec2 = ClassifyErrorStrict(e.Text, e.AttemptWav2Vec2, e.HumanTranscription)

			// Lenient versions
			e.TranscriptionAccuracyWav2Vec2Lenient = CalculateAccuracy(e.HumanTranscription, e.AttemptWav2Vec2)
			e.ErrorAttributionWav2Vec2Lenient = ClassifyError(e.Text, e.AttemptWav2Vec2, e.HumanTranscription)
		}
	}
}

// PopulateNormalizedText populates the display fields for normalized text.
//
// Since normalized text is now stored in the database (AttemptWhisperLenient/AttemptWav2Vec2Lenient),
// this function simply copies those values to the backwards-compatible NormalizedWhisper/NormalizedWav2Vec2 fields.
//
// This function is kept for backwards compatibility with code that expects NormalizedWhisper/NormalizedWav2Vec2.
// Call this after fetching entries from the database when you need to display normalized text
// (e.g., in the transcription review table).
func (e *Entry) PopulateNormalizedText() {
	// Copy from database fields to display fields (backwards compatibility)
	e.NormalizedWhisper = e.AttemptWhisperLenient
	e.NormalizedWav2Vec2 = e.AttemptWav2Vec2Lenient
}

// roundToTwoDecimals rounds a float64 to 2 decimal places.
func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// truncateToByteLength safely truncates a UTF-8 string to a maximum byte length.
// This ensures we don't break multi-byte UTF-8 characters (important for Welsh diacritics).
//
// If the string is already shorter than maxBytes, it is returned unchanged.
// Otherwise, it truncates at maxBytes and walks back to the last valid UTF-8 boundary.
func truncateToByteLength(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Truncate to maxBytes, then walk back to last valid UTF-8 boundary
	result := s[:maxBytes]
	for len(result) > 0 && !utf8.ValidString(result) {
		result = result[:len(result)-1]
	}
	return result
}
