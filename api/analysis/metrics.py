"""
Statistical Metrics Module
Handles calculation of statistical measures, effect sizes, and error rate metrics
"""

import numpy as np
from typing import List, Dict, Tuple


class MetricsCalculator:
    """
    Calculates statistical metrics for ASR evaluation.

    Provides safe numeric operations, effect size calculations (Cohen's d),
    and comprehensive metric statistics for model comparison.
    """

    @staticmethod
    def safe_float(value, default=0.0) -> float:
        """
        Convert to float, replacing NaN/Inf with default value.
        Ensures JSON serialization compatibility.

        Args:
            value: Value to convert
            default: Default value for NaN/Inf cases

        Returns:
            Safe float value
        """
        if value is None or np.isnan(value) or np.isinf(value):
            return default
        return float(value)

    @staticmethod
    def safe_divide(numerator, denominator, default=0.0) -> float:
        """
        Safe division that handles zero denominators and NaN values.

        Args:
            numerator: Numerator value
            denominator: Denominator value
            default: Default value for undefined cases

        Returns:
            Safe division result
        """
        if denominator == 0 or np.isnan(denominator) or np.isnan(numerator):
            return default
        result = numerator / denominator
        if np.isinf(result) or np.isnan(result):
            return default
        return float(result)

    @staticmethod
    def calculate_cohens_d(data1: List[float], data2: List[float]) -> float:
        """
        Calculate Cohen's d effect size between two groups.

        Cohen's d measures the standardized difference between two means,
        useful for quantifying practical significance beyond statistical significance.

        Args:
            data1: First group of values (e.g., Whisper error rates)
            data2: Second group of values (e.g., Wav2Vec2 error rates)

        Returns:
            Cohen's d value (positive if data1 > data2)

        Interpretation:
            |d| < 0.2: Negligible
            0.2 ≤ |d| < 0.5: Small
            0.5 ≤ |d| < 0.8: Medium
            |d| ≥ 0.8: Large
        """
        if len(data1) < 2 or len(data2) < 2:
            return 0.0

        mean1 = np.mean(data1)
        mean2 = np.mean(data2)
        std1 = np.std(data1, ddof=1)
        std2 = np.std(data2, ddof=1)

        # Pooled standard deviation
        n1, n2 = len(data1), len(data2)
        pooled_std = np.sqrt(((n1 - 1) * std1**2 + (n2 - 1) * std2**2) / (n1 + n2 - 2))

        if pooled_std == 0:
            return 0.0

        cohens_d = (mean1 - mean2) / pooled_std
        return MetricsCalculator.safe_float(cohens_d)

    @staticmethod
    def interpret_cohens_d(cohens_d: float) -> str:
        """
        Interpret Cohen's d effect size value.

        Args:
            cohens_d: Cohen's d value

        Returns:
            Interpretation string
        """
        abs_d = abs(cohens_d)
        if abs_d < 0.2:
            return "negligible"
        elif abs_d < 0.5:
            return "small"
        elif abs_d < 0.8:
            return "medium"
        else:
            return "large"

    @staticmethod
    def _compute_single_metric_stats(data1: List[float], data2: List[float],
                                     metric_name: str) -> Dict:
        """
        Compute comprehensive statistics for a single metric (CER or WER).
        Consolidates duplicate code from original _compute_metric_statistics.

        Args:
            data1: Values for first model (e.g., Whisper)
            data2: Values for second model (e.g., Wav2Vec2)
            metric_name: Name of metric for logging ("CER" or "WER")

        Returns:
            Dictionary with model1/model2 stats, differences, and effect size
        """
        if len(data1) == 0 or len(data2) == 0:
            return {}

        cohens_d = MetricsCalculator.calculate_cohens_d(data1, data2)

        mean1 = np.mean(data1)
        mean2 = np.mean(data2)
        std1 = np.std(data1, ddof=1)
        std2 = np.std(data2, ddof=1)

        # Calculate additional metrics for MSc thesis
        mean_diff = mean1 - mean2
        percentage_point_diff = abs(mean_diff)
        baseline_mean = max(mean1, mean2)
        relative_improvement = MetricsCalculator.safe_divide(
            percentage_point_diff, baseline_mean
        ) * 100

        # Coefficient of variation (consistency metric)
        cv1 = MetricsCalculator.safe_divide(std1, mean1) * 100 if mean1 > 0 else 0
        cv2 = MetricsCalculator.safe_divide(std2, mean2) * 100 if mean2 > 0 else 0

        # Superiority rate (how often model1 is better)
        model1_better_count = sum(1 for d1, d2 in zip(data1, data2) if d1 < d2)
        model1_superiority_rate = MetricsCalculator.safe_divide(
            model1_better_count, len(data1)
        ) * 100

        return {
            "model1": {
                "mean": MetricsCalculator.safe_float(mean1),
                "median": MetricsCalculator.safe_float(np.median(data1)),
                "std": MetricsCalculator.safe_float(std1),
                "min": MetricsCalculator.safe_float(np.min(data1)),
                "max": MetricsCalculator.safe_float(np.max(data1)),
                "cv": MetricsCalculator.safe_float(cv1)
            },
            "model2": {
                "mean": MetricsCalculator.safe_float(mean2),
                "median": MetricsCalculator.safe_float(np.median(data2)),
                "std": MetricsCalculator.safe_float(std2),
                "min": MetricsCalculator.safe_float(np.min(data2)),
                "max": MetricsCalculator.safe_float(np.max(data2)),
                "cv": MetricsCalculator.safe_float(cv2)
            },
            "difference": {
                "mean": MetricsCalculator.safe_float(mean_diff),
                "percentage_point_difference": MetricsCalculator.safe_float(percentage_point_diff),
                "relative_improvement": MetricsCalculator.safe_float(relative_improvement)
            },
            "effect_size": {
                "cohens_d": MetricsCalculator.safe_float(cohens_d),
                "interpretation": MetricsCalculator.interpret_cohens_d(cohens_d)
            },
            "model1_superiority_rate": MetricsCalculator.safe_float(model1_superiority_rate)
        }

    @staticmethod
    def compute_metric_statistics(whisper_wer: List[float], wav2vec2_wer: List[float],
                                  whisper_cer: List[float], wav2vec2_cer: List[float]) -> Dict:
        """
        Compute statistics for CER and WER metrics.
        Refactored to eliminate code duplication using _compute_single_metric_stats.

        Args:
            whisper_wer: List of WER values for Whisper
            wav2vec2_wer: List of WER values for Wav2Vec2
            whisper_cer: List of CER values for Whisper
            wav2vec2_cer: List of CER values for Wav2Vec2

        Returns:
            Dictionary with CER and WER analysis
        """
        results = {
            "cer_analysis": {},
            "wer_analysis": {}
        }

        # CER Analysis (primary metric for single-word prompts)
        if len(whisper_cer) > 0 and len(wav2vec2_cer) > 0:
            cer_stats = MetricsCalculator._compute_single_metric_stats(
                whisper_cer, wav2vec2_cer, "CER"
            )
            # Rename generic keys to specific model names
            results["cer_analysis"] = {
                "whisper": cer_stats["model1"],
                "wav2vec2": cer_stats["model2"],
                "difference": cer_stats["difference"],
                "effect_size": cer_stats["effect_size"],
                "whisper_superiority_rate": cer_stats["model1_superiority_rate"]
            }

        # WER Analysis (secondary - less informative for single words)
        if len(whisper_wer) > 0 and len(wav2vec2_wer) > 0:
            wer_stats = MetricsCalculator._compute_single_metric_stats(
                whisper_wer, wav2vec2_wer, "WER"
            )
            # Rename generic keys to specific model names
            results["wer_analysis"] = {
                "whisper": wer_stats["model1"],
                "wav2vec2": wer_stats["model2"],
                "difference": wer_stats["difference"],
                "effect_size": wer_stats["effect_size"],
                "whisper_superiority_rate": wer_stats["model1_superiority_rate"]
            }

        return results
