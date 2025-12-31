package models

import (
	"encoding/json"
	"strings"
)

// EditOperation represents a single edit operation in the Levenshtein distance calculation
type EditOperation struct {
	Type     string `json:"type"`     // "substitution", "deletion", "insertion"
	Position int    `json:"position"` // Position in the source string
	Expected string `json:"expected"` // Expected character/word
	Actual   string `json:"actual"`   // Actual character/word
}

// EditOperations represents the complete set of edit operations
type EditOperations struct {
	Substitutions []EditOperation `json:"substitutions"`
	Deletions     []EditOperation `json:"deletions"`
	Insertions    []EditOperation `json:"insertions"`
	Total         int             `json:"total"`
}

// levenshteinDistance calculates the edit distance between two sequences using dynamic programming.
// This generic implementation works with any comparable type ([]rune, []string, etc.)
// Returns the minimum edit distance and the distance matrix for backtracking.
func levenshteinDistance[T comparable](source, target []T) (int, [][]int) {
	sourceLen := len(source)
	targetLen := len(target)

	// Edge cases
	if sourceLen == 0 {
		return targetLen, nil
	}
	if targetLen == 0 {
		return sourceLen, nil
	}

	// Create distance matrix
	matrix := make([][]int, sourceLen+1)
	for i := range matrix {
		matrix[i] = make([]int, targetLen+1)
	}

	// Initialize first row and column
	for i := 0; i <= sourceLen; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= targetLen; j++ {
		matrix[0][j] = j
	}

	// Fill the matrix using dynamic programming
	for i := 1; i <= sourceLen; i++ {
		for j := 1; j <= targetLen; j++ {
			if source[i-1] == target[j-1] {
				// Characters match, no operation needed
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				// Take minimum of three operations
				matrix[i][j] = min3(
					matrix[i-1][j]+1,   // deletion
					matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1, // substitution
				)
			}
		}
	}

	return matrix[sourceLen][targetLen], matrix
}

// backtrackOperations traces back through the distance matrix to identify individual edit operations.
// The toString function converts elements of type T to their string representation.
func backtrackOperations[T comparable](source, target []T, matrix [][]int, toString func(T) string) *EditOperations {
	// Pre-allocate slices with estimated capacity to reduce allocations
	estimatedOps := (len(source) + len(target)) / 2
	ops := &EditOperations{
		Substitutions: make([]EditOperation, 0, estimatedOps),
		Deletions:     make([]EditOperation, 0, estimatedOps),
		Insertions:    make([]EditOperation, 0, estimatedOps),
	}

	i := len(source)
	j := len(target)

	// Backtrack through the matrix from bottom-right to top-left
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && source[i-1] == target[j-1] {
			// Characters match, no operation needed
			i--
			j--
			continue
		}

		if i > 0 && j > 0 && matrix[i][j] == matrix[i-1][j-1]+1 {
			// Substitution operation
			ops.Substitutions = append(ops.Substitutions, EditOperation{
				Type:     "substitution",
				Position: i - 1,
				Expected: toString(target[j-1]),
				Actual:   toString(source[i-1]),
			})
			ops.Total++
			i--
			j--
		} else if i > 0 && matrix[i][j] == matrix[i-1][j]+1 {
			// Deletion operation
			ops.Deletions = append(ops.Deletions, EditOperation{
				Type:     "deletion",
				Position: i - 1,
				Expected: "",
				Actual:   toString(source[i-1]),
			})
			ops.Total++
			i--
		} else if j > 0 && matrix[i][j] == matrix[i][j-1]+1 {
			// Insertion operation
			ops.Insertions = append(ops.Insertions, EditOperation{
				Type:     "insertion",
				Position: i,
				Expected: toString(target[j-1]),
				Actual:   "",
			})
			ops.Total++
			j--
		}
	}

	return ops
}

// CharDistance calculates the Levenshtein distance at character level.
// Comparison is case-insensitive.
// Returns the distance and detailed edit operations.
func CharDistance(source, target string) (int, *EditOperations) {
	// Normalize to lowercase for case-insensitive comparison
	sourceRunes := []rune(strings.ToLower(source))
	targetRunes := []rune(strings.ToLower(target))

	// Handle empty strings
	if len(sourceRunes) == 0 && len(targetRunes) == 0 {
		return 0, &EditOperations{
			Substitutions: []EditOperation{},
			Deletions:     []EditOperation{},
			Insertions:    []EditOperation{},
		}
	}

	// Calculate distance using generic algorithm
	distance, matrix := levenshteinDistance(sourceRunes, targetRunes)

	// Backtrack to find operations
	runeToString := func(r rune) string { return string(r) }
	operations := backtrackOperations(sourceRunes, targetRunes, matrix, runeToString)

	return distance, operations
}

// WordDistance calculates the Levenshtein distance at word level.
// Words are split by whitespace and compared case-insensitively.
// Returns the distance and detailed edit operations.
func WordDistance(source, target string) (int, *EditOperations) {
	sourceWords := strings.Fields(strings.ToLower(source))
	targetWords := strings.Fields(strings.ToLower(target))

	// Handle empty strings
	if len(sourceWords) == 0 && len(targetWords) == 0 {
		return 0, &EditOperations{
			Substitutions: []EditOperation{},
			Deletions:     []EditOperation{},
			Insertions:    []EditOperation{},
		}
	}

	// Calculate distance using generic algorithm
	distance, matrix := levenshteinDistance(sourceWords, targetWords)

	// Backtrack to find operations
	wordToString := func(w string) string { return w }
	operations := backtrackOperations(sourceWords, targetWords, matrix, wordToString)

	return distance, operations
}

// GetEditOperationsJSON converts EditOperations to a compact JSON string.
// Returns an empty JSON object if operations is nil.
func GetEditOperationsJSON(ops *EditOperations) string {
	if ops == nil {
		return "{}"
	}

	// Use json.Marshal for safe serialization
	jsonBytes, err := json.Marshal(ops)
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}

// min3 returns the minimum of three integers.
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
