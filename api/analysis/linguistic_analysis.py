"""
Linguistic Analysis Module
Analyzes error patterns by linguistic categories (vowels, consonants, digraphs)
"""

from typing import List, Dict
from .text_processing import WelshTextProcessor
from .error_analysis import ErrorAnalyzer
from .metrics import MetricsCalculator


class LinguisticAnalyzer:
    """
    Analyzes linguistic patterns in ASR errors.

    Breaks down errors by phonological categories (vowels, consonants, digraphs)
    to identify which linguistic features are most challenging for ASR models.
    """

    def __init__(self, text_processor: WelshTextProcessor = None,
                 error_analyzer: ErrorAnalyzer = None,
                 metrics_calc: MetricsCalculator = None):
        """
        Initialize LinguisticAnalyzer.

        Args:
            text_processor: Optional WelshTextProcessor instance
            error_analyzer: Optional ErrorAnalyzer instance
            metrics_calc: Optional MetricsCalculator instance
        """
        self.text_processor = text_processor or WelshTextProcessor()
        self.error_analyzer = error_analyzer or ErrorAnalyzer(self.text_processor)
        self.metrics_calc = metrics_calc or MetricsCalculator()

    def _analyze_model_errors_by_category(self, entries: List[Dict], model_field: str,
                                         apply_lenient: bool = False,
                                         use_strict: bool = False) -> Dict:
        """
        Analyze errors for a single model, categorized by linguistic type.
        Refactored from nested function in linguistic_pattern_analysis.

        Args:
            entries: List of entry dictionaries
            model_field: Field name for ASR output (e.g., 'AttemptWhisper')
            apply_lenient: If True, use lenient normalized fields from DB
            use_strict: If True, use raw text without any normalization

        Returns:
            Dictionary with error stats for vowels, consonants, digraphs, and overall
        """
        # Initialize stats for all categories
        category_stats = {
            "vowel": {"total": 0, "errors": 0, "confusion": {}},
            "consonant": {"total": 0, "errors": 0, "confusion": {}},
            "digraph": {"total": 0, "errors": 0, "confusion": {}},
            "all": {"total": 0, "errors": 0, "confusion": {}}
        }
        skipped_count = 0

        # Process each entry ONCE
        for entry in entries:
            # For CAPT: Always compare ASR to TARGET word (what learner should say)
            # NOT to HumanTranscription (what was actually in the audio)
            reference = entry.get('Text', '')

            # Use pre-computed normalized text from database for lenient metrics
            # This eliminates duplicate normalization work (computed once in Go, stored in DB)
            if apply_lenient:
                # Use normalized fields from database (computed by Go's applyLenientNormalization)
                hypothesis = entry.get(f'{model_field}Lenient', '')
            else:
                # Use raw transcription for strict metrics
                hypothesis = entry.get(model_field, '')

            if not reference or not hypothesis:
                skipped_count += 1
                continue

            # Apply normalization based on mode
            if use_strict:
                # STRICT mode: Use raw text without ANY normalization
                # This shows true technical accuracy including punctuation, capitalization
                ref_normalized = reference
                hyp_normalized = hypothesis
            elif apply_lenient:
                # LENIENT mode: Hypothesis already normalized from DB
                # Normalize reference to match
                # This shows pedagogically-relevant errors (ignores formatting)
                ref_normalized = self.text_processor.normalize_text(reference)
                hyp_normalized = hypothesis  # Already normalized in DB
            else:
                # STANDARD mode (shouldn't be used now that we have strict/lenient)
                # Kept for backwards compatibility
                ref_normalized = self.text_processor.normalize_text(reference)
                hyp_normalized = self.text_processor.normalize_text(hypothesis)

            ref_tokens = self.text_processor.tokenize_welsh(ref_normalized)
            hyp_tokens = self.text_processor.tokenize_welsh(hyp_normalized)

            # Compute alignment ONCE per entry
            alignment = self.error_analyzer.compute_character_alignment(
                ref_tokens, hyp_tokens
            )

            # Classify and accumulate stats for ALL categories
            for operation, ref_char, hyp_char in alignment:
                ref_category = self.text_processor.classify_token(ref_char)
                is_error = (ref_char != hyp_char)

                # Update specific category
                if ref_category in category_stats:
                    category_stats[ref_category]["total"] += 1
                    if is_error:
                        category_stats[ref_category]["errors"] += 1
                        pair = (ref_char, hyp_char)
                        category_stats[ref_category]["confusion"][pair] = \
                            category_stats[ref_category]["confusion"].get(pair, 0) + 1

                # Update overall category
                category_stats["all"]["total"] += 1
                if is_error:
                    category_stats["all"]["errors"] += 1
                    pair = (ref_char, hyp_char)
                    category_stats["all"]["confusion"][pair] = \
                        category_stats["all"]["confusion"].get(pair, 0) + 1

        # Build results for each category
        results = {}
        category_names = {
            "vowel": "vowels",
            "consonant": "consonants",
            "digraph": "digraphs",
            "all": "overall"
        }

        for cat_key, cat_name in category_names.items():
            stats = category_stats[cat_key]
            error_rate = self.metrics_calc.safe_divide(
                stats["errors"], stats["total"]
            ) * 100

            # Build confusion matrix (top 15)
            confusion_matrix = []
            if stats["confusion"]:
                sorted_pairs = sorted(
                    stats["confusion"].items(),
                    key=lambda x: x[1],
                    reverse=True
                )
                for (expected, actual), count in sorted_pairs[:15]:
                    confusion_matrix.append({
                        "expected": expected,
                        "actual": actual,
                        "count": count
                    })

            results[cat_name] = {
                "error_rate": self.metrics_calc.safe_float(error_rate),
                "total_characters": stats["total"],
                "error_count": stats["errors"],
                "confusion_matrix": confusion_matrix
            }

            # Log results
            if use_strict:
                metric_type = "Strict (no normalization)"
            elif apply_lenient:
                metric_type = "Lenient (normalized)"
            else:
                metric_type = "Standard (normalized)"
            print(f"[linguistic_patterns] {model_field} ({metric_type}) {cat_name}: "
                  f"processed {len(entries)} entries, skipped {skipped_count}, "
                  f"total_chars={stats['total']}, error_chars={stats['errors']}")

        return results

    def linguistic_pattern_analysis(self, entries: List[Dict]) -> Dict:
        """
        Analyze error patterns by linguistic category for MSc thesis.

        Breaks down character errors into:
        - Vowel errors (a, e, i, o, u, w, y + diacritics)
        - Consonant errors (single consonants)
        - Digraph errors (ll, ch, dd, ff, ng, rh, ph, th)

        This linguistic perspective is critical for Language Technologies/Applied Linguistics
        focus, showing which phonological features are challenging for ASR.

        Returns both RAW and NORMALIZED metrics to show how lenient matching affects
        different phonological categories.

        Args:
            entries: List of entry dictionaries with transcriptions

        Returns:
            Dictionary with error rates by linguistic category for both models and both metrics:
            {
                "raw": {"whisper": {...}, "wav2vec2": {...}, "comparison": {...}},
                "normalized": {"whisper": {...}, "wav2vec2": {...}, "comparison": {...}},
                "improvement": {"whisper": {...}, "wav2vec2": {...}}
            }
        """
        # Analyze both models with RAW transcriptions (STRICT - no normalization at all)
        # This shows true technical accuracy including punctuation, capitalization
        raw_results = {
            "whisper": self._analyze_model_errors_by_category(
                entries, "AttemptWhisper", apply_lenient=False, use_strict=True
            ),
            "wav2vec2": self._analyze_model_errors_by_category(
                entries, "AttemptWav2Vec2", apply_lenient=False, use_strict=True
            )
        }

        # Analyze both models with NORMALIZED transcriptions (LENIENT - pedagogically relevant)
        # This shows practical utility for learners (ignores formatting differences)
        normalized_results = {
            "whisper": self._analyze_model_errors_by_category(
                entries, "AttemptWhisper", apply_lenient=True, use_strict=False
            ),
            "wav2vec2": self._analyze_model_errors_by_category(
                entries, "AttemptWav2Vec2", apply_lenient=True, use_strict=False
            )
        }

        # Add comparative metrics for raw
        raw_results["comparison"] = {}
        for cat_name in ["vowels", "consonants", "digraphs", "overall"]:
            w_rate = raw_results["whisper"][cat_name]["error_rate"]
            v_rate = raw_results["wav2vec2"][cat_name]["error_rate"]
            diff = abs(w_rate - v_rate)
            winner = "Whisper" if w_rate < v_rate else "Wav2Vec2"
            if diff < 1.0:
                winner = "Comparable"

            raw_results["comparison"][cat_name] = {
                "difference": self.metrics_calc.safe_float(diff),
                "winner": winner
            }

        # Add comparative metrics for normalized
        normalized_results["comparison"] = {}
        for cat_name in ["vowels", "consonants", "digraphs", "overall"]:
            w_rate = normalized_results["whisper"][cat_name]["error_rate"]
            v_rate = normalized_results["wav2vec2"][cat_name]["error_rate"]
            diff = abs(w_rate - v_rate)
            winner = "Whisper" if w_rate < v_rate else "Wav2Vec2"
            if diff < 1.0:
                winner = "Comparable"

            normalized_results["comparison"][cat_name] = {
                "difference": self.metrics_calc.safe_float(diff),
                "winner": winner
            }

        # Calculate improvement from normalization
        improvement = {
            "whisper": {},
            "wav2vec2": {}
        }

        for model in ["whisper", "wav2vec2"]:
            for cat_name in ["vowels", "consonants", "digraphs", "overall"]:
                raw_rate = raw_results[model][cat_name]["error_rate"]
                norm_rate = normalized_results[model][cat_name]["error_rate"]
                improvement_value = raw_rate - norm_rate
                improvement_pct = self.metrics_calc.safe_divide(
                    improvement_value, raw_rate
                ) * 100 if raw_rate > 0 else 0

                improvement[model][cat_name] = {
                    "absolute_improvement": self.metrics_calc.safe_float(improvement_value),
                    "relative_improvement": self.metrics_calc.safe_float(improvement_pct)
                }

        return {
            "raw": raw_results,
            "normalized": normalized_results,
            "improvement": improvement
        }
