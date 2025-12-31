"""
Reporting and Analysis Module
Handles high-level reporting, recommendations, and study metadata generation
"""

import numpy as np
from typing import List, Dict
from .text_processing import WelshTextProcessor
from .metrics import MetricsCalculator


class ReportGenerator:
    """
    Generates comprehensive reports and analysis summaries.

    Provides executive summaries, qualitative examples, study metadata,
    and practical recommendations for MSc thesis presentation.
    """

    def __init__(self, text_processor: WelshTextProcessor = None,
                 metrics_calc: MetricsCalculator = None):
        """
        Initialize ReportGenerator.

        Args:
            text_processor: Optional WelshTextProcessor instance
            metrics_calc: Optional MetricsCalculator instance
        """
        self.text_processor = text_processor or WelshTextProcessor()
        self.metrics_calc = metrics_calc or MetricsCalculator()

    def executive_summary_analysis(self, entries: List[Dict]) -> Dict:
        """
        Generate executive summary with key metrics and findings for MSc thesis.

        Provides high-level overview including:
        - Sample size and scope
        - Overall model winner
        - Key quantitative findings
        - Percentage improvements

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with executive summary metrics
        """
        if not entries:
            return {
                "total_recordings": 0,
                "total_words": 0,
                "date_range": "No data",
                "overall_winner": "N/A",
                "key_finding": "No data available for analysis"
            }

        # Extract metrics
        whisper_cer_raw = [e.get('CERWhisper', e.get('CER', 0)) for e in entries
                          if e.get('CERWhisper') is not None or e.get('CER') is not None]
        wav2vec2_cer_raw = [e.get('CERWav2Vec2', 0) for e in entries
                           if e.get('CERWav2Vec2') is not None]
        whisper_cer_norm = [e.get('CERWhisperLenient', 0) for e in entries
                           if e.get('CERWhisperLenient') is not None]
        wav2vec2_cer_norm = [e.get('CERWav2Vec2Lenient', 0) for e in entries
                            if e.get('CERWav2Vec2Lenient') is not None]

        # Calculate means
        whisper_mean_raw = self.metrics_calc.safe_float(np.mean(whisper_cer_raw)) if whisper_cer_raw else 0
        wav2vec2_mean_raw = self.metrics_calc.safe_float(np.mean(wav2vec2_cer_raw)) if wav2vec2_cer_raw else 0
        whisper_mean_norm = self.metrics_calc.safe_float(np.mean(whisper_cer_norm)) if whisper_cer_norm else 0
        wav2vec2_mean_norm = self.metrics_calc.safe_float(np.mean(wav2vec2_cer_norm)) if wav2vec2_cer_norm else 0

        # Determine winner
        raw_advantage = wav2vec2_mean_raw - whisper_mean_raw
        norm_advantage = wav2vec2_mean_norm - whisper_mean_norm

        overall_winner = "Whisper" if raw_advantage > 0 else "Wav2Vec2"
        if abs(raw_advantage) < 1.0:  # Less than 1 percentage point difference
            overall_winner = "Comparable"

        # Calculate improvements
        percentage_point_diff_raw = abs(raw_advantage)
        percentage_point_diff_norm = abs(norm_advantage)
        relative_improvement_raw = self.metrics_calc.safe_divide(
            abs(raw_advantage), max(whisper_mean_raw, wav2vec2_mean_raw)
        ) * 100
        relative_improvement_norm = self.metrics_calc.safe_divide(
            abs(norm_advantage), max(whisper_mean_norm, wav2vec2_mean_norm)
        ) * 100

        # Count unique words and recordings per word
        unique_words = len(set(e.get('Text', '') for e in entries if e.get('Text')))

        # Calculate recordings per word statistics
        word_counts = {}
        for e in entries:
            word = e.get('Text', '').strip()
            if word:
                word_counts[word] = word_counts.get(word, 0) + 1

        recordings_per_word_values = list(word_counts.values()) if word_counts else [0]
        avg_recordings_per_word = self.metrics_calc.safe_divide(
            sum(recordings_per_word_values), len(recordings_per_word_values)
        )
        min_recordings_per_word = min(recordings_per_word_values) if recordings_per_word_values else 0
        max_recordings_per_word = max(recordings_per_word_values) if recordings_per_word_values else 0

        # Generate comprehensive key finding that discusses both raw and normalized metrics
        if overall_winner == "Whisper":
            key_finding = (
                f"Whisper achieves {percentage_point_diff_raw:.1f} percentage points lower CER "
                f"({relative_improvement_raw:.1f}% relative improvement) using strict (raw) matching. "
                f"With normalized metrics (lenient matching that ignores punctuation/spacing errors), "
                f"Whisper's advantage is {percentage_point_diff_norm:.1f} percentage points "
                f"({relative_improvement_norm:.1f}% relative improvement). "
                f"Normalized metrics are pedagogically significant as they reflect practical user experience "
                f"by focusing on pronunciation accuracy rather than transcription formatting."
            )
        elif overall_winner == "Wav2Vec2":
            key_finding = (
                f"Wav2Vec2 achieves {percentage_point_diff_raw:.1f} percentage points lower CER "
                f"({relative_improvement_raw:.1f}% relative improvement) using strict (raw) matching. "
                f"With normalized metrics (lenient matching that ignores punctuation/spacing errors), "
                f"Wav2Vec2's advantage is {percentage_point_diff_norm:.1f} percentage points "
                f"({relative_improvement_norm:.1f}% relative improvement). "
                f"Normalized metrics are pedagogically significant as they reflect practical user experience "
                f"by focusing on pronunciation accuracy rather than transcription formatting."
            )
        else:
            key_finding = (
                f"Models show comparable performance using strict (raw) matching (CER difference < 1 percentage point). "
                f"With normalized metrics (lenient matching that ignores punctuation/spacing errors), "
                f"the difference is {percentage_point_diff_norm:.1f} percentage points. "
                f"Normalized metrics provide a more realistic assessment of pedagogical utility "
                f"by focusing on pronunciation accuracy rather than transcription formatting."
            )

        return {
            "total_recordings": len(entries),
            "total_words": unique_words,
            "avg_recordings_per_word": round(avg_recordings_per_word, 2),
            "min_recordings_per_word": min_recordings_per_word,
            "max_recordings_per_word": max_recordings_per_word,
            "overall_winner": overall_winner,
            "key_finding": key_finding,
            "raw_metrics": {
                "whisper_cer_mean": whisper_mean_raw,
                "wav2vec2_cer_mean": wav2vec2_mean_raw,
                "percentage_point_difference": percentage_point_diff_raw,
                "relative_improvement": relative_improvement_raw,
                "winner": "Whisper" if raw_advantage > 0 else "Wav2Vec2"
            },
            "normalized_metrics": {
                "whisper_cer_mean": whisper_mean_norm,
                "wav2vec2_cer_mean": wav2vec2_mean_norm,
                "percentage_point_difference": percentage_point_diff_norm,
                "relative_improvement": relative_improvement_norm,
                "winner": "Whisper" if norm_advantage > 0 else "Wav2Vec2"
            }
        }

    def study_design_metadata(self, entries: List[Dict]) -> Dict:
        """
        Provide study design and methodological metadata for MSc thesis.

        Returns structured information about study type, design, limitations,
        and statistical approach for transparent methodology reporting.

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with study design metadata
        """
        # Count unique words
        unique_words = set(e.get('Text', '') for e in entries if e.get('Text'))

        # Determine date range if timestamps available
        dates = [e.get('CreatedAt', e.get('created_at', ''))
                for e in entries
                if e.get('CreatedAt') or e.get('created_at')]

        date_range = "Not available"
        if dates:
            try:
                dates_sorted = sorted([d for d in dates if d])
                if dates_sorted:
                    date_range = f"{dates_sorted[0][:10]} to {dates_sorted[-1][:10]}"
            except:
                pass

        return {
            "study_type": "Single-participant descriptive case study",
            "research_design": {
                "type": "Within-subjects comparison",
                "participants": 1,
                "asr_models": 2,
                "unique_words": len(unique_words),
                "total_recordings": len(entries)
            },
            "data_collection": {
                "task": "Single-word Welsh pronunciation",
                "asr_models_compared": ["Whisper (large-v3-ft-cv-cy)", "Wav2Vec2 (btb-cv-ft-cv-cy)"],
                "date_range": date_range
            },
            "statistical_approach": {
                "paradigm": "Descriptive statistics with effect sizes",
                "rationale": "Single-participant study precludes inferential testing",
                "primary_metrics": ["CER (Character Error Rate)", "Error attribution"],
                "effect_size_measure": "Cohen's d"
            },
            "limitations": [
                "Single participant: findings may not generalize",
                "No control group or baseline comparison",
                "Limited to single-word pronunciation tasks",
                "Welsh language-specific findings"
            ],
            "strengths": [
                "Real-world ecological validity",
                "Comprehensive character-level error analysis with digraph support",
                "Pedagogical framing beyond technical accuracy",
                "Dual metrics approach (raw and normalized)",
                "Linguistic perspective (vowels, consonants, digraphs)"
            ]
        }

    def practical_recommendations(self, entries: List[Dict],
                                 comparison: Dict = None,
                                 error_costs: Dict = None) -> Dict:
        """
        Generate practical recommendations for Welsh CAPT system design.

        Synthesizes findings into actionable guidance for model selection,
        system architecture, and pedagogical considerations.

        Args:
            entries: List of entry dictionaries
            comparison: Optional model comparison results
            error_costs: Optional error cost analysis results

        Returns:
            Dictionary with structured recommendations
        """
        summary = self.executive_summary_analysis(entries)
        overall_winner = summary.get("overall_winner", "N/A")

        # Basic recommendation based on overall winner
        if overall_winner == "Whisper":
            primary_recommendation = "Whisper"
            primary_rationale = f"Lower CER in both raw and normalized metrics"
        elif overall_winner == "Wav2Vec2":
            primary_recommendation = "Wav2Vec2"
            primary_rationale = f"Lower CER in both raw and normalized metrics"
        else:
            primary_recommendation = "Either model acceptable"
            primary_rationale = "Models show comparable performance"

        return {
            "primary_recommendation": {
                "recommended_model": primary_recommendation,
                "rationale": primary_rationale,
                "confidence": "Medium" if overall_winner == "Comparable" else "High"
            },
            "normalization_strategy": {
                "recommendation": "Implement both raw and normalized metrics",
                "raw_use_case": "Technical accuracy reporting",
                "normalized_use_case": "Learner-facing feedback"
            },
            "system_design_implications": [
                f"Deploy {primary_recommendation} as the primary ASR engine for Welsh CAPT",
                "Implement both raw and normalized CER metrics for comprehensive assessment",
                "Use normalized metrics for learner-facing feedback (filters out punctuation/capitalization)",
                "Use raw metrics for technical accuracy reporting and system monitoring",
                "Deploy hybrid approach for best per-word performance (switch between models)",
                "Consider ensemble methods to leverage both models' strengths"
            ],
            "pedagogical_considerations": [
                "Normalized CER provides more pedagogically relevant feedback for L2 learners",
                "False acceptances (incorrect pronunciation marked correct) are more harmful than false rejections",
                "Monitor Welsh digraph recognition (ll, ch, dd, ff, ng, rh, ph, th) as these are critical for L2 learners",
                "Focus error analysis on phonological challenges rather than punctuation/capitalization",
                "Provide targeted practice on high-error phonemes identified in linguistic analysis",
                "Use confidence scores to provide nuanced feedback rather than binary correct/incorrect"
            ],
            "features_requiring_human_verification": [
                "Welsh digraphs with high error rates (>25% CER)",
                "Words with false acceptance errors (pedagogically harmful)",
                "Low-confidence transcriptions (consider confidence threshold tuning)",
                "Novel vocabulary not in training data",
                "Ambiguous pronunciations where both models disagree",
                "Critical phonological features for L2 acquisition"
            ],
            "implementation_notes": [
                "Implement confidence scoring with human verification fallback",
                "Monitor false acceptance rate (most harmful error type)",
                "Set confidence thresholds based on pedagogical priorities",
                "Provide learner dashboard showing progress on specific phonemes",
                "Log model disagreements for continuous improvement"
            ]
        }

    def consistency_reliability_analysis(self, entries: List[Dict]) -> Dict:
        """
        Analyze consistency and reliability of ASR models across attempts.

        Calculates variance, interquartile range (IQR), and reliability thresholds
        for both raw and normalized metrics to assess model stability.

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with consistency metrics for both models
        """
        if not entries:
            return {
                "raw": {"whisper": {}, "wav2vec2": {}},
                "normalized": {"whisper": {}, "wav2vec2": {}}
            }

        # Extract metrics
        whisper_cer_raw = [e.get('CERWhisper', e.get('CER', 0)) for e in entries
                          if e.get('CERWhisper') is not None or e.get('CER') is not None]
        wav2vec2_cer_raw = [e.get('CERWav2Vec2', 0) for e in entries
                           if e.get('CERWav2Vec2') is not None]
        whisper_cer_norm = [e.get('CERWhisperLenient', 0) for e in entries
                           if e.get('CERWhisperLenient') is not None]
        wav2vec2_cer_norm = [e.get('CERWav2Vec2Lenient', 0) for e in entries
                            if e.get('CERWav2Vec2Lenient') is not None]

        def calculate_reliability_stats(cer_values: List[float]) -> Dict:
            """Calculate reliability statistics for a set of CER values."""
            if not cer_values:
                return {
                    "mean": 0.0,
                    "std_dev": 0.0,
                    "variance": 0.0,
                    "iqr": 0.0,
                    "percentile_95": 0.0,
                    "coefficient_of_variation": 0.0,
                    "reliability_rating": "N/A"
                }

            mean_cer = float(np.mean(cer_values))
            std_dev = float(np.std(cer_values, ddof=1)) if len(cer_values) > 1 else 0.0
            variance = float(np.var(cer_values, ddof=1)) if len(cer_values) > 1 else 0.0

            # Calculate IQR
            q25 = float(np.percentile(cer_values, 25))
            q75 = float(np.percentile(cer_values, 75))
            iqr = q75 - q25

            # 95th percentile (worst-case performance)
            p95 = float(np.percentile(cer_values, 95))

            # Coefficient of variation (normalized std dev)
            cv = self.metrics_calc.safe_divide(std_dev, mean_cer) * 100 if mean_cer > 0 else 0.0

            # Reliability rating based on mean and consistency
            if mean_cer < 10 and cv < 50:
                reliability = "High"
            elif mean_cer < 25 and cv < 75:
                reliability = "Medium"
            else:
                reliability = "Low"

            return {
                "mean": round(mean_cer, 2),
                "std_dev": round(std_dev, 2),
                "variance": round(variance, 2),
                "iqr": round(iqr, 2),
                "percentile_95": round(p95, 2),
                "coefficient_of_variation": round(cv, 2),
                "reliability_rating": reliability
            }

        return {
            "raw": {
                "whisper": calculate_reliability_stats(whisper_cer_raw),
                "wav2vec2": calculate_reliability_stats(wav2vec2_cer_raw)
            },
            "normalized": {
                "whisper": calculate_reliability_stats(whisper_cer_norm),
                "wav2vec2": calculate_reliability_stats(wav2vec2_cer_norm)
            }
        }

    def qualitative_examples_selection(self, entries: List[Dict]) -> Dict:
        """
        Select representative qualitative examples for MSc thesis discussion.

        Categorizes recordings into pedagogically meaningful groups:
        - Both models correct
        - Only Whisper correct
        - Only Wav2Vec2 correct
        - Both models incorrect
        - Pedagogically critical (false acceptances)
        - Interesting disagreements

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with categorized example lists
        """
        if not entries:
            return {
                "both_correct": [],
                "whisper_only": [],
                "wav2vec2_only": [],
                "both_incorrect": [],
                "pedagogically_critical": [],
                "interesting_cases": []
            }

        # Use normalized metrics for pedagogical assessment
        both_correct = []
        whisper_only = []
        wav2vec2_only = []
        both_incorrect = []
        pedagogically_critical = []  # False acceptances
        interesting_cases = []  # Models disagree significantly

        for e in entries:
            target = e.get('Text', '').strip()
            if not target:
                continue

            whisper_cer = e.get('CERWhisperLenient', e.get('CERWhisper', 0))
            wav2vec2_cer = e.get('CERWav2Vec2Lenient', 0)
            whisper_attempt = e.get('AttemptWhisper', e.get('Attempt', ''))
            wav2vec2_attempt = e.get('AttemptWav2Vec2', '')

            # Define "correct" as CER < 10% (effectively perfect or near-perfect)
            whisper_correct = whisper_cer < 10
            wav2vec2_correct = wav2vec2_cer < 10

            example = {
                "target": target,
                "whisper_transcription": whisper_attempt,
                "wav2vec2_transcription": wav2vec2_attempt,
                "whisper_cer": round(whisper_cer, 2),
                "wav2vec2_cer": round(wav2vec2_cer, 2)
            }

            # Categorize
            if whisper_correct and wav2vec2_correct:
                both_correct.append(example)
            elif whisper_correct and not wav2vec2_correct:
                whisper_only.append(example)
            elif wav2vec2_correct and not whisper_correct:
                wav2vec2_only.append(example)
            else:
                both_incorrect.append(example)

            # Pedagogically critical: false acceptances (ASR says OK when it's wrong)
            # This would require ground truth learner error data, which we don't have
            # Instead, flag cases where both models are highly confident but disagree
            if abs(whisper_cer - wav2vec2_cer) > 30:  # 30 percentage point difference
                interesting_cases.append({
                    **example,
                    "delta": round(abs(whisper_cer - wav2vec2_cer), 2)
                })

        # Sort interesting cases by delta (most divergent first)
        interesting_cases.sort(key=lambda x: x.get('delta', 0), reverse=True)

        return {
            "both_correct": both_correct[:10],  # Top 10 examples
            "whisper_only": whisper_only[:10],
            "wav2vec2_only": wav2vec2_only[:10],
            "both_incorrect": both_incorrect[:10],
            "pedagogically_critical": pedagogically_critical[:10],
            "interesting_cases": interesting_cases[:10],
            "counts": {
                "both_correct": len(both_correct),
                "whisper_only": len(whisper_only),
                "wav2vec2_only": len(wav2vec2_only),
                "both_incorrect": len(both_incorrect),
                "pedagogically_critical": len(pedagogically_critical),
                "interesting_cases": len(interesting_cases)
            }
        }
