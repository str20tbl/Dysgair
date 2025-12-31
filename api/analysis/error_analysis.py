"""
Error Analysis Module
Handles character-level error analysis, alignment, and confusion matrices
"""

from typing import List, Dict, Tuple
from .text_processing import WelshTextProcessor


class ErrorAnalyzer:
    """
    Analyzes character-level errors in ASR transcriptions.

    Provides Levenshtein alignment, confusion matrix generation,
    and detailed error statistics for Welsh text.
    """

    def __init__(self, text_processor: WelshTextProcessor = None):
        """
        Initialize ErrorAnalyzer.

        Args:
            text_processor: Optional WelshTextProcessor instance (creates new if None)
        """
        self.text_processor = text_processor or WelshTextProcessor()

    @staticmethod
    def compute_character_alignment(target_tokens: List[str],
                                    asr_tokens: List[str]) -> List[Tuple[str, str, str]]:
        """
        Compute character-level alignment using Levenshtein distance.

        Returns list of (operation, expected, actual) tuples:
        - ("match", "a", "a")
        - ("substitution", "a", "e")
        - ("deletion", "a", "")
        - ("insertion", "", "e")

        Args:
            target_tokens: Tokenized target text
            asr_tokens: Tokenized ASR transcription

        Returns:
            List of alignment tuples
        """
        # Compute Levenshtein distance matrix
        m, n = len(target_tokens), len(asr_tokens)
        dp = [[0] * (n + 1) for _ in range(m + 1)]

        # Initialize base cases
        for i in range(m + 1):
            dp[i][0] = i
        for j in range(n + 1):
            dp[0][j] = j

        # Fill matrix
        for i in range(1, m + 1):
            for j in range(1, n + 1):
                if target_tokens[i-1] == asr_tokens[j-1]:
                    dp[i][j] = dp[i-1][j-1]
                else:
                    dp[i][j] = 1 + min(
                        dp[i-1][j],     # deletion
                        dp[i][j-1],     # insertion
                        dp[i-1][j-1]    # substitution
                    )

        # Backtrack to find operations
        alignments = []
        i, j = m, n
        while i > 0 or j > 0:
            if i > 0 and j > 0 and target_tokens[i-1] == asr_tokens[j-1]:
                # Match
                alignments.append(("match", target_tokens[i-1], asr_tokens[j-1]))
                i -= 1
                j -= 1
            elif i > 0 and j > 0 and dp[i][j] == dp[i-1][j-1] + 1:
                # Substitution
                alignments.append(("substitution", target_tokens[i-1], asr_tokens[j-1]))
                i -= 1
                j -= 1
            elif i > 0 and dp[i][j] == dp[i-1][j] + 1:
                # Deletion
                alignments.append(("deletion", target_tokens[i-1], ""))
                i -= 1
            elif j > 0 and dp[i][j] == dp[i][j-1] + 1:
                # Insertion
                alignments.append(("insertion", "", asr_tokens[j-1]))
                j -= 1

        # Reverse to get correct order
        alignments.reverse()
        return alignments

    def build_error_stats(self, alignments: List[Tuple[str, str, str]],
                         token_filter: str = None) -> Dict:
        """
        Build confusion matrix and per-character error statistics.

        Args:
            alignments: List of (operation, expected, actual) tuples
            token_filter: Optional filter: "vowel", "consonant", "digraph", or None for all

        Returns:
            Dictionary with confusion_matrix and per_character stats
        """
        confusion_matrix = {}
        per_character = {}

        for operation, expected, actual in alignments:
            # Skip if filtering by token type
            if token_filter:
                expected_type = self.text_processor.classify_token(expected) if expected else None
                actual_type = self.text_processor.classify_token(actual) if actual else None

                # For substitutions/deletions, check expected token
                if operation in ["substitution", "deletion"]:
                    if expected_type != token_filter:
                        continue
                # For insertions, check actual token
                elif operation == "insertion":
                    if actual_type != token_filter:
                        continue
                # For matches, check both (should be same type)
                elif operation == "match":
                    if expected_type != token_filter:
                        continue

            # Clean up digraph markers for display
            expected_display = expected.strip('⟨⟩') if expected else ""
            actual_display = actual.strip('⟨⟩') if actual else ""

            # Track per-character totals and errors
            if expected_display:
                if expected_display not in per_character:
                    per_character[expected_display] = {"total": 0, "errors": 0}
                per_character[expected_display]["total"] += 1

                if operation != "match":
                    per_character[expected_display]["errors"] += 1

            # Build confusion matrix (only for non-matches)
            if operation != "match":
                key = (expected_display, actual_display)
                confusion_matrix[key] = confusion_matrix.get(key, 0) + 1

        # Calculate error rates
        for char, stats in per_character.items():
            if stats["total"] > 0:
                stats["error_rate"] = round((stats["errors"] / stats["total"]) * 100, 2)
            else:
                stats["error_rate"] = 0.0

        # Convert confusion matrix to JSON-serializable list format
        confusion_matrix_list = [
            {"expected": exp, "actual": act, "count": count}
            for (exp, act), count in confusion_matrix.items()
        ]

        return {
            "confusion_matrix": confusion_matrix_list,
            "per_character": per_character
        }

    def analyze_character_errors_for_model(self, entries: List[Dict],
                                          model_attempt_key: str) -> Dict:
        """
        Analyze character errors for a single ASR model.

        Computes overall character/digraph error analysis combining all error types.

        Args:
            entries: List of entry dictionaries
            model_attempt_key: Key for ASR transcription ("AttemptWhisper" or "AttemptWav2Vec2")

        Returns:
            Dictionary with confusion_matrix and per_character error statistics
        """
        # Collect all alignments across entries
        all_alignments = []

        for entry in entries:
            # For CAPT: Use TARGET word as ground truth
            # Confusion matrix should show: "Which target characters get misrecognized?"
            # NOT: "Did ASR correctly transcribe what was in the audio?"
            reference = self.text_processor.normalize_text(entry.get('Text', ''))
            asr_text = self.text_processor.normalize_text(entry.get(model_attempt_key, ''))

            # Apply lenient space-insensitive matching
            reference, asr_text = self.text_processor.apply_lenient_normalization(
                reference, asr_text
            )

            if not reference or not asr_text:
                continue

            # Tokenize both
            reference_tokens = self.text_processor.tokenize_welsh(reference)
            asr_tokens = self.text_processor.tokenize_welsh(asr_text)

            # Compute alignment
            alignments = self.compute_character_alignment(reference_tokens, asr_tokens)
            all_alignments.extend(alignments)

        # Build overall error statistics (no filtering)
        return self.build_error_stats(all_alignments, token_filter=None)

    def character_error_analysis(self, entries: List[Dict]) -> Dict:
        """
        Character-level error analysis with Welsh digraph support.

        Analyzes all character and digraph errors for both ASR models, providing:
        - Confusion matrix: Which letters/digraphs confused with which (all entries)
        - Per-character stats: Total occurrences, errors, error rate %

        Welsh digraphs (ll, ch, dd, ff, ng, rh, ph, th) are treated as single units.
        Diacritics are tracked separately.

        Args:
            entries: List of entry dictionaries with target text and ASR transcriptions

        Returns:
            Dictionary with error statistics for Whisper and Wav2Vec2:
            {
              "whisper": {confusion_matrix: [...], per_character: {...}},
              "wav2vec2": {confusion_matrix: [...], per_character: {...}}
            }
        """
        return {
            "whisper": self.analyze_character_errors_for_model(entries, "AttemptWhisper"),
            "wav2vec2": self.analyze_character_errors_for_model(entries, "AttemptWav2Vec2")
        }
