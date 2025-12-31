package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// ErrorClassification represents the source of transcription errors in the ASR evaluation system.
// This helps distinguish between ASR model errors and user pronunciation errors.
type ErrorClassification string

const (
	// ErrorClassCorrect indicates both human and ASR transcribed the target correctly
	ErrorClassCorrect ErrorClassification = "CORRECT"

	// ErrorClassASR indicates the ASR made an error while the human pronounced correctly
	ErrorClassASR ErrorClassification = "ASR_ERROR"

	// ErrorClassUser indicates the user mispronounced and the ASR correctly transcribed the mispronunciation
	ErrorClassUser ErrorClassification = "USER_ERROR"

	// ErrorClassAmbiguous indicates both are wrong but differ, requiring manual review
	ErrorClassAmbiguous ErrorClassification = "AMBIGUOUS"
)

// ClassifyError determines the source of transcription errors using decision tree logic.
//
// Decision tree (evaluated in order):
//   - Human = Target AND ASR = Target → CORRECT (both correct)
//   - Human = Target (AND ASR ≠ Target) → ASR_ERROR (ASR failed to transcribe correct pronunciation)
//   - ASR = Human (AND Human ≠ Target) → USER_ERROR (ASR correctly transcribed mispronunciation)
//   - Otherwise → AMBIGUOUS (human wrong AND ASR doesn't match human - unclear attribution)
//
// The AMBIGUOUS case covers:
//   - Human ≠ Target AND ASR ≠ Target AND ASR ≠ Human (both wrong, different errors)
//   - Human ≠ Target AND ASR = Target AND ASR ≠ Human (ASR guessed right despite wrong input)
//
// Parameters:
//   - target: The intended word/phrase to be pronounced
//   - asrTranscription: What the ASR system transcribed
//   - humanTranscription: What the human reviewer heard
//
// Returns empty string if humanTranscription is not yet available.
// ClassifyErrorStrict classifies whether an ASR error was caused by user pronunciation or ASR failure
// using strict normalization (NO lenient space-insensitive matching).
//
// This uses strict metric-level normalization (same as used in WER/CER strict calculations)
// instead of the simple lowercase/trim normalization.
//
// Args:
//   - target: The intended word/phrase to be pronounced
//   - asrTranscription: What the ASR system transcribed
//   - humanTranscription: What the human reviewer heard
//
// Returns empty string if humanTranscription is not yet available.
func ClassifyErrorStrict(target, asrTranscription, humanTranscription string) ErrorClassification {
	// If no human transcription yet, return empty (not yet reviewed)
	if humanTranscription == "" {
		return ""
	}

	// Minimal normalization: lowercase and trim only (preserve punctuation, spacing, hyphens)
	targetNorm := strings.ToLower(strings.TrimSpace(target))
	asrNorm := strings.ToLower(strings.TrimSpace(asrTranscription))
	humanNorm := strings.ToLower(strings.TrimSpace(humanTranscription))

	// Apply decision tree logic with strict (raw text) comparison
	humanCorrect := humanNorm == targetNorm
	asrCorrect := asrNorm == targetNorm
	asrMatchesHuman := asrNorm == humanNorm

	switch {
	case humanCorrect && asrCorrect:
		// Both transcribed correctly
		return ErrorClassCorrect

	case humanCorrect:
		// Human pronounced correctly, but ASR failed (!asrCorrect is implied)
		return ErrorClassASR

	case asrMatchesHuman:
		// Human mispronounced, ASR correctly transcribed the mispronunciation (!humanCorrect is implied)
		return ErrorClassUser

	default:
		// Either both wrong but different, or ASR guessed right despite mispronunciation
		// (!humanCorrect && !asrMatchesHuman - ambiguous either way)
		return ErrorClassAmbiguous
	}
}

func ClassifyError(target, asrTranscription, humanTranscription string) ErrorClassification {
	// If no human transcription yet, return empty (not yet reviewed)
	if humanTranscription == "" {
		return ""
	}

	// Normalize all inputs with metrics normalization (includes punctuation removal, etc.)
	targetNorm := normalizeForMetrics(target)
	asrNorm := normalizeForMetrics(asrTranscription)
	humanNorm := normalizeForMetrics(humanTranscription)

	// Apply lenient space-insensitive matching for each comparison
	targetNormVsASR, asrNormVsTarget := applyLenientNormalization(targetNorm, asrNorm)
	targetNormVsHuman, humanNormVsTarget := applyLenientNormalization(targetNorm, humanNorm)
	asrNormVsHuman, humanNormVsASR := applyLenientNormalization(asrNorm, humanNorm)

	// Apply decision tree logic with lenient comparison
	humanCorrect := humanNormVsTarget == targetNormVsHuman
	asrCorrect := asrNormVsTarget == targetNormVsASR
	asrMatchesHuman := asrNormVsHuman == humanNormVsASR

	switch {
	case humanCorrect && asrCorrect:
		// Both transcribed correctly
		return ErrorClassCorrect

	case humanCorrect:
		// Human pronounced correctly, but ASR failed (!asrCorrect is implied)
		return ErrorClassASR

	case asrMatchesHuman:
		// Human mispronounced, ASR correctly transcribed the mispronunciation (!humanCorrect is implied)
		return ErrorClassUser

	default:
		// Either both wrong but different, or ASR guessed right despite mispronunciation
		// (!humanCorrect && !asrMatchesHuman - ambiguous either way)
		return ErrorClassAmbiguous
	}
}

// normalizeForComparison prepares a string for comparison by converting to lowercase and trimming whitespace.
// This ensures consistent comparison regardless of case or surrounding whitespace.
func normalizeForComparison(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Scan implements the sql.Scanner interface for ErrorClassification.
// This allows the database driver to scan string values from the database into ErrorClassification types.
func (e *ErrorClassification) Scan(value interface{}) error {
	if value == nil {
		*e = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*e = ErrorClassification(v)
		return nil
	case []byte:
		*e = ErrorClassification(string(v))
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into ErrorClassification", value)
	}
}

// Value implements the driver.Valuer interface for ErrorClassification.
// This allows the database driver to convert ErrorClassification values to strings when saving to the database.
func (e *ErrorClassification) Value() (driver.Value, error) {
	if e == nil {
		return "", nil
	}
	return string(*e), nil
}
