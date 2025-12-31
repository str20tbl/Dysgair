"""
Statistical Analysis Module for Dysgair ASR Research
Descriptive statistics for single-participant MSc study
Focus: Comparing Whisper vs Wav2Vec2 for Welsh pronunciation feedback

REFACTORED: This module now acts as a facade, delegating to specialized analysis modules.
"""

import numpy as np
from typing import List, Dict, Tuple
import warnings

# Import specialized analysis modules
from analysis import (
    WelshTextProcessor,
    MetricsCalculator,
    ErrorAnalyzer,
    LinguisticAnalyzer,
    ModelComparator,
    ReportGenerator,
    WordAnalyzer
)

warnings.filterwarnings('ignore')

# Welsh character classification constants (maintained for backward compatibility)
WELSH_DIGRAPHS = ['ll', 'ch', 'dd', 'ff', 'ng', 'rh', 'ph', 'th']
WELSH_VOWELS = set('aeiouyẃâêîôûŵŷàèìòùẁỳäëïöüẅÿ')  # includes all diacritics
WELSH_CONSONANTS = set('bcdfghjklmnpqrstvxz')


class StatisticalAnalyzer:
    """
    Descriptive statistical analysis for single-participant ASR model comparison.
    No inferential statistics - focus on practical differences and patterns.

    This class now acts as a facade, delegating to specialized analysis modules:
    - WelshTextProcessor: Text normalization and tokenization
    - MetricsCalculator: Statistical metrics and effect sizes
    - ErrorAnalyzer: Character-level error analysis
    - LinguisticAnalyzer: Phonological pattern analysis
    - ModelComparator: Model comparison and hybrid analysis
    - ReportGenerator: High-level reporting and recommendations
    """

    def __init__(self):
        """Initialize StatisticalAnalyzer with specialized analysis components."""
        # Initialize specialized modules
        self.text_processor = WelshTextProcessor()
        self.metrics_calc = MetricsCalculator()
        self.error_analyzer = ErrorAnalyzer(self.text_processor)
        self.linguistic_analyzer = LinguisticAnalyzer(
            self.text_processor, self.error_analyzer, self.metrics_calc
        )
        self.model_comparator = ModelComparator(self.text_processor, self.metrics_calc)
        self.report_generator = ReportGenerator(self.text_processor, self.metrics_calc)
        self.word_analyzer = WordAnalyzer(self.metrics_calc)

    # ========== Backward Compatibility Static Methods ==========
    # These are maintained for backward compatibility with existing code

    @staticmethod
    def _safe_float(value, default=0.0) -> float:
        """Backward compatibility wrapper for MetricsCalculator.safe_float"""
        return MetricsCalculator.safe_float(value, default)

    @staticmethod
    def _safe_divide(numerator, denominator, default=0.0) -> float:
        """Backward compatibility wrapper for MetricsCalculator.safe_divide"""
        return MetricsCalculator.safe_divide(numerator, denominator, default)

    @staticmethod
    def _normalize_text(text: str) -> str:
        """Backward compatibility wrapper for WelshTextProcessor.normalize_text"""
        return WelshTextProcessor.normalize_text(text)

    @staticmethod
    def _apply_lenient_normalization(target: str, transcription: str) -> Tuple[str, str]:
        """Backward compatibility wrapper for WelshTextProcessor.apply_lenient_normalization"""
        return WelshTextProcessor.apply_lenient_normalization(target, transcription)

    @staticmethod
    def _tokenize_welsh(text: str) -> List[str]:
        """Backward compatibility wrapper for WelshTextProcessor.tokenize_welsh"""
        return WelshTextProcessor.tokenize_welsh(text)

    @staticmethod
    def _classify_token(token: str) -> str:
        """Backward compatibility wrapper for WelshTextProcessor.classify_token"""
        return WelshTextProcessor.classify_token(token)

    @staticmethod
    def _compute_character_alignment(target_tokens: List[str],
                                    asr_tokens: List[str]) -> List[Tuple[str, str, str]]:
        """Backward compatibility wrapper for ErrorAnalyzer.compute_character_alignment"""
        return ErrorAnalyzer.compute_character_alignment(target_tokens, asr_tokens)

    def _build_error_stats(self, alignments: List[Tuple[str, str, str]],
                          token_filter=None) -> Dict:
        """Backward compatibility wrapper for ErrorAnalyzer.build_error_stats"""
        return self.error_analyzer.build_error_stats(alignments, token_filter)

    # ========== Metrics Calculator Methods ==========

    def calculate_cohens_d(self, data1: List[float], data2: List[float]) -> float:
        """Calculate Cohen's d effect size between two groups."""
        return self.metrics_calc.calculate_cohens_d(data1, data2)

    def _compute_metric_statistics(self, whisper_wer: List[float], wav2vec2_wer: List[float],
                                   whisper_cer: List[float], wav2vec2_cer: List[float]) -> Dict:
        """Compute statistics for CER and WER metrics."""
        return self.metrics_calc.compute_metric_statistics(
            whisper_wer, wav2vec2_wer, whisper_cer, wav2vec2_cer
        )

    def _interpret_cohens_d(self, cohens_d: float) -> str:
        """Interpret Cohen's d effect size value."""
        return self.metrics_calc.interpret_cohens_d(cohens_d)

    # ========== Model Comparison Methods ==========

    def model_comparison_analysis(self, entries: List[Dict]) -> Dict:
        """Descriptive comparison between Whisper and Wav2Vec2 models."""
        return self.model_comparator.model_comparison_analysis(entries)

    def hybrid_best_case_analysis(self, entries: List[Dict]) -> Dict:
        """Calculate best-case hybrid performance."""
        return self.model_comparator.hybrid_best_case_analysis(entries)

    def hybrid_practical_implications_analysis(self, entries: List[Dict]) -> Dict:
        """Calculate hybrid system performance for critical CAPT areas."""
        return self.model_comparator.hybrid_practical_implications_analysis(entries)

    def _calculate_hybrid_subset(self, entries: List[Dict]) -> Dict:
        """Helper to calculate hybrid performance for a subset of entries."""
        return self.model_comparator._calculate_hybrid_subset(entries)

    def _calculate_error_reduction(self, entries: List[Dict]) -> Dict:
        """Calculate potential ASR error reduction with hybrid system."""
        return self.model_comparator._calculate_error_reduction(entries)

    def inter_rater_reliability(self, entries: List[Dict]) -> Dict:
        """Calculate simple agreement between Whisper and Wav2Vec2 models."""
        return self.model_comparator.inter_rater_reliability(entries)

    def _compute_error_attribution_stats(self, entries: List[Dict],
                                        whisper_key: str, wav2vec2_key: str) -> Dict:
        """Helper function to compute error attribution statistics."""
        return self.model_comparator._compute_error_attribution_stats(
            entries, whisper_key, wav2vec2_key
        )

    def error_attribution_analysis(self, entries: List[Dict]) -> Dict:
        """Analyze error attribution patterns for both models."""
        return self.model_comparator.error_attribution_analysis(entries)

    # ========== Error Analysis Methods ==========

    def _analyze_character_errors_for_model(self, entries: List[Dict],
                                           model_attempt_key: str) -> Dict:
        """Analyze character errors for a single ASR model."""
        return self.error_analyzer.analyze_character_errors_for_model(
            entries, model_attempt_key
        )

    def character_error_analysis(self, entries: List[Dict]) -> Dict:
        """Character-level error analysis with Welsh digraph support."""
        return self.error_analyzer.character_error_analysis(entries)

    # ========== Linguistic Analysis Methods ==========

    def linguistic_pattern_analysis(self, entries: List[Dict]) -> Dict:
        """Analyze error patterns by linguistic category."""
        return self.linguistic_analyzer.linguistic_pattern_analysis(entries)

    # ========== Reporting Methods ==========

    def executive_summary_analysis(self, entries: List[Dict]) -> Dict:
        """Generate executive summary with key metrics and findings."""
        return self.report_generator.executive_summary_analysis(entries)

    def study_design_metadata(self, entries: List[Dict]) -> Dict:
        """Provide study design and methodological metadata."""
        return self.report_generator.study_design_metadata(entries)

    def practical_recommendations(self, entries: List[Dict]) -> Dict:
        """Generate practical recommendations for Welsh CAPT system design."""
        # Get additional analysis data
        comparison = self.model_comparison_analysis(entries)
        error_costs = None  # Placeholder for error_cost_analysis if implemented

        return self.report_generator.practical_recommendations(
            entries, comparison, error_costs
        )

    # ========== Word Analysis Methods ==========

    def word_difficulty_analysis(self, entries: List[Dict]) -> Dict:
        """Analyze ASR model performance by word with both raw and normalized metrics."""
        return self.word_analyzer.word_difficulty_analysis(entries)

    def word_length_analysis(self, entries: List[Dict]) -> Dict:
        """Analyze ASR performance by word length (character count)."""
        return self.word_analyzer.word_length_analysis(entries)

    def detect_over_transcription(self, entries: List[Dict], threshold: float = 20.0) -> Dict:
        """Detect cases where ASR models predict multiple words for single-word targets."""
        return self.word_analyzer.detect_over_transcription(entries, threshold)

    # ========== Additional Analysis Methods ==========

    def error_cost_analysis(self, entries: List[Dict]) -> Dict:
        """Analyze pedagogical costs of ASR errors."""
        return self.model_comparator.error_cost_analysis(entries)

    def qualitative_examples_selection(self, entries: List[Dict]) -> Dict:
        """Select representative qualitative examples for MSc thesis discussion."""
        return self.report_generator.qualitative_examples_selection(entries)

    def consistency_reliability_analysis(self, entries: List[Dict]) -> Dict:
        """Analyze consistency and reliability of ASR models across attempts."""
        return self.report_generator.consistency_reliability_analysis(entries)
