"""
Model Comparison Module
Handles comparative analysis between ASR models, hybrid systems, and reliability metrics
"""

import numpy as np
from typing import List, Dict
from .text_processing import WelshTextProcessor
from .metrics import MetricsCalculator


class ModelComparator:
    """
    Compares ASR model performance and analyzes hybrid system potential.

    Provides model comparison metrics, hybrid analysis, inter-rater reliability,
    and error attribution analysis for CAPT system evaluation.
    """

    def __init__(self, text_processor: WelshTextProcessor = None,
                 metrics_calc: MetricsCalculator = None):
        """
        Initialize ModelComparator.

        Args:
            text_processor: Optional WelshTextProcessor instance
            metrics_calc: Optional MetricsCalculator instance
        """
        self.text_processor = text_processor or WelshTextProcessor()
        self.metrics_calc = metrics_calc or MetricsCalculator()

    def model_comparison_analysis(self, entries: List[Dict]) -> Dict:
        """
        Descriptive comparison between Whisper and Wav2Vec2 models with both raw and normalized metrics.
        Focus on CER (more meaningful for single-word prompts)

        Args:
            entries: List of entry dictionaries with WER/CER for both models

        Returns:
            Dictionary containing descriptive statistics and effect sizes for both raw and normalized metrics
        """
        # Extract RAW (strict) metrics for both models
        whisper_wer_raw = [e.get('WERWhisper', e.get('WER', 0)) for e in entries
                          if e.get('WERWhisper') is not None or e.get('WER') is not None]
        wav2vec2_wer_raw = [e.get('WERWav2Vec2', 0) for e in entries
                           if e.get('WERWav2Vec2') is not None]
        whisper_cer_raw = [e.get('CERWhisper', e.get('CER', 0)) for e in entries
                          if e.get('CERWhisper') is not None or e.get('CER') is not None]
        wav2vec2_cer_raw = [e.get('CERWav2Vec2', 0) for e in entries
                           if e.get('CERWav2Vec2') is not None]

        # Extract NORMALIZED (lenient) metrics for both models
        whisper_wer_norm = [e.get('WERWhisperLenient', 0) for e in entries
                           if e.get('WERWhisperLenient') is not None]
        wav2vec2_wer_norm = [e.get('WERWav2Vec2Lenient', 0) for e in entries
                            if e.get('WERWav2Vec2Lenient') is not None]
        whisper_cer_norm = [e.get('CERWhisperLenient', 0) for e in entries
                           if e.get('CERWhisperLenient') is not None]
        wav2vec2_cer_norm = [e.get('CERWav2Vec2Lenient', 0) for e in entries
                            if e.get('CERWav2Vec2Lenient') is not None]

        # Ensure equal lengths for paired comparison (raw metrics)
        min_wer_len_raw = min(len(whisper_wer_raw), len(wav2vec2_wer_raw))
        min_cer_len_raw = min(len(whisper_cer_raw), len(wav2vec2_cer_raw))
        whisper_wer_raw = whisper_wer_raw[:min_wer_len_raw]
        wav2vec2_wer_raw = wav2vec2_wer_raw[:min_wer_len_raw]
        whisper_cer_raw = whisper_cer_raw[:min_cer_len_raw]
        wav2vec2_cer_raw = wav2vec2_cer_raw[:min_cer_len_raw]

        # Ensure equal lengths for paired comparison (normalized metrics)
        min_wer_len_norm = min(len(whisper_wer_norm), len(wav2vec2_wer_norm))
        min_cer_len_norm = min(len(whisper_cer_norm), len(wav2vec2_cer_norm))
        whisper_wer_norm = whisper_wer_norm[:min_wer_len_norm]
        wav2vec2_wer_norm = wav2vec2_wer_norm[:min_wer_len_norm]
        whisper_cer_norm = whisper_cer_norm[:min_cer_len_norm]
        wav2vec2_cer_norm = wav2vec2_cer_norm[:min_cer_len_norm]

        return {
            "sample_size": len(entries),
            "raw": self.metrics_calc.compute_metric_statistics(
                whisper_wer_raw, wav2vec2_wer_raw,
                whisper_cer_raw, wav2vec2_cer_raw
            ),
            "normalized": self.metrics_calc.compute_metric_statistics(
                whisper_wer_norm, wav2vec2_wer_norm,
                whisper_cer_norm, wav2vec2_cer_norm
            )
        }

    def hybrid_best_case_analysis(self, entries: List[Dict]) -> Dict:
        """
        Calculate best-case hybrid performance by selecting minimum CER between lenient models.

        For each word, finds the best result from:
        - CERWhisperLenient (normalized)
        - CERWav2Vec2Lenient (normalized)

        This answers the research question: "What if we had a smart hybrid CAPT system
        that could intelligently choose the best ASR model for each word using lenient metrics?"

        IMPORTANT: Only includes entries with BOTH lenient CER values to ensure fair comparison.

        Args:
            entries: List of entry dictionaries with CER metrics for both models

        Returns:
            Dictionary with hybrid best-case metrics and selection distribution
        """
        hybrid_cers = []
        best_individual_cers = []  # Track best single approach for SAME sample
        selection_counts = {
            "whisper_lenient": 0,
            "wav2vec2_lenient": 0,
            "tie": 0
        }

        for e in entries:
            # Get lenient CER values for this entry
            cer_whisper_norm = e.get('CERWhisperLenient')
            cer_wav2vec2_norm = e.get('CERWav2Vec2Lenient')

            # CRITICAL: Only include entries with BOTH lenient CER values for fair comparison
            if None in [cer_whisper_norm, cer_wav2vec2_norm]:
                continue

            # Find minimum CER (best result) and track which model was selected
            cers_map = {
                "whisper_lenient": float(cer_whisper_norm),
                "wav2vec2_lenient": float(cer_wav2vec2_norm)
            }

            min_cer = min(cers_map.values())
            winners = [k for k, v in cers_map.items() if v == min_cer]

            hybrid_cers.append(min_cer)

            # Only count as a selection if there's a clear winner (not a tie)
            if len(winners) == 1:
                selection_counts[winners[0]] += 1
            else:
                selection_counts["tie"] += 1

        if len(hybrid_cers) == 0:
            return {"error": "No valid entries for hybrid analysis - need entries with both lenient CER values"}

        # Calculate hybrid statistics
        hybrid_mean_cer = self.metrics_calc.safe_float(np.mean(hybrid_cers))

        # For fair comparison: determine which SINGLE FIXED configuration is best
        valid_entries_indices = []
        for idx, e in enumerate(entries):
            cer_whisper_norm = e.get('CERWhisperLenient')
            cer_wav2vec2_norm = e.get('CERWav2Vec2Lenient')

            if None not in [cer_whisper_norm, cer_wav2vec2_norm]:
                valid_entries_indices.append(idx)

        # Calculate mean for each fixed configuration using same sample
        whisper_raw_mean = np.mean([float(entries[i].get('CERWhisper', float('inf')))
                                    for i in valid_entries_indices])
        whisper_norm_mean = np.mean([float(entries[i]['CERWhisperLenient'])
                                     for i in valid_entries_indices])
        wav2vec2_raw_mean = np.mean([float(entries[i].get('CERWav2Vec2', float('inf')))
                                     for i in valid_entries_indices])
        wav2vec2_norm_mean = np.mean([float(entries[i]['CERWav2Vec2Lenient'])
                                      for i in valid_entries_indices])

        # Best individual configuration is the one with lowest mean CER
        config_means = {
            "whisper_raw": whisper_raw_mean,
            "whisper_normalized": whisper_norm_mean,
            "wav2vec2_raw": wav2vec2_raw_mean,
            "wav2vec2_normalized": wav2vec2_norm_mean
        }

        best_config_name = min(config_means.keys(), key=lambda k: config_means[k])
        best_individual_cer = self.metrics_calc.safe_float(config_means[best_config_name])

        # Calculate improvement metrics
        improvement_absolute = best_individual_cer - hybrid_mean_cer
        improvement_percentage = self.metrics_calc.safe_divide(
            improvement_absolute, best_individual_cer
        ) * 100

        # Calculate selection percentages
        total_selections = sum(selection_counts.values())
        selection_percentages = {
            k: self.metrics_calc.safe_float((v / total_selections) * 100) if total_selections > 0 else 0
            for k, v in selection_counts.items()
        }

        # Calculate model preference (Whisper vs Wav2Vec2) - excluding ties
        whisper_total = selection_counts["whisper_lenient"]
        wav2vec2_total = selection_counts["wav2vec2_lenient"]
        tie_count = selection_counts["tie"]
        total_non_tie = whisper_total + wav2vec2_total

        model_preference = {
            "whisper_pct": self.metrics_calc.safe_float(
                (whisper_total / total_non_tie) * 100
            ) if total_non_tie > 0 else 0,
            "wav2vec2_pct": self.metrics_calc.safe_float(
                (wav2vec2_total / total_non_tie) * 100
            ) if total_non_tie > 0 else 0,
            "tie_pct": self.metrics_calc.safe_float(
                (tie_count / total_selections) * 100
            ) if total_selections > 0 else 0,
            "whisper_count": whisper_total,
            "wav2vec2_count": wav2vec2_total,
            "tie_count": tie_count
        }

        return {
            "hybrid_mean_cer": self.metrics_calc.safe_float(hybrid_mean_cer),
            "best_individual_cer": self.metrics_calc.safe_float(best_individual_cer),
            "best_individual_config": best_config_name,
            "improvement_absolute": self.metrics_calc.safe_float(improvement_absolute),
            "improvement_percentage": self.metrics_calc.safe_float(improvement_percentage),
            "selection_counts": selection_counts,
            "selection_percentages": selection_percentages,
            "model_preference": model_preference,
            "sample_size": len(hybrid_cers)
        }

    def _calculate_hybrid_subset(self, entries: List[Dict]) -> Dict:
        """
        Helper to calculate hybrid performance for a subset of entries.
        Hybrid uses only lenient scores (min of lenient models).

        Args:
            entries: Subset of entries to analyze

        Returns:
            Dict with hybrid_cer, best_single_cer, improvement metrics
        """
        if not entries:
            return {
                "hybrid_cer": 0,
                "best_single_cer": 0,
                "improvement_absolute": 0,
                "improvement_percentage": 0,
                "sample_size": 0
            }

        # Calculate hybrid (min lenient CER per entry)
        hybrid_cers = []
        whisper_raw_cers = []
        whisper_norm_cers = []
        wav2vec2_raw_cers = []
        wav2vec2_norm_cers = []

        for e in entries:
            cer_wn = e.get('CERWhisperLenient')
            cer_vn = e.get('CERWav2Vec2Lenient')

            if None in [cer_wn, cer_vn]:
                continue

            # Hybrid: minimum lenient CER
            hybrid_cers.append(min(float(cer_wn), float(cer_vn)))

            # Track each config for best single comparison
            cer_wr = e.get('CERWhisper')
            cer_vr = e.get('CERWav2Vec2')

            whisper_raw_cers.append(float(cer_wr) if cer_wr is not None else float('inf'))
            whisper_norm_cers.append(float(cer_wn))
            wav2vec2_raw_cers.append(float(cer_vr) if cer_vr is not None else float('inf'))
            wav2vec2_norm_cers.append(float(cer_vn))

        if not hybrid_cers:
            return {
                "hybrid_cer": 0,
                "best_single_cer": 0,
                "improvement_absolute": 0,
                "improvement_percentage": 0,
                "sample_size": 0
            }

        hybrid_mean = self.metrics_calc.safe_float(np.mean(hybrid_cers))

        # Find best single configuration
        config_means = {
            "whisper_raw": np.mean(whisper_raw_cers),
            "whisper_normalized": np.mean(whisper_norm_cers),
            "wav2vec2_raw": np.mean(wav2vec2_raw_cers),
            "wav2vec2_normalized": np.mean(wav2vec2_norm_cers)
        }

        best_config = min(config_means.keys(), key=lambda k: config_means[k])
        best_single_mean = self.metrics_calc.safe_float(config_means[best_config])

        improvement = best_single_mean - hybrid_mean
        improvement_pct = self.metrics_calc.safe_divide(improvement, best_single_mean) * 100

        return {
            "hybrid_cer": hybrid_mean,
            "best_single_cer": best_single_mean,
            "best_single_config": best_config,
            "improvement_absolute": self.metrics_calc.safe_float(improvement),
            "improvement_percentage": self.metrics_calc.safe_float(improvement_pct),
            "sample_size": len(hybrid_cers)
        }

    def _calculate_error_reduction(self, entries: List[Dict]) -> Dict:
        """
        Calculate potential ASR error reduction with hybrid system.

        Counts cases where:
        - Best Single Model had ASR_ERROR
        - But the other model had CORRECT
        (These errors could be prevented by hybrid)

        Args:
            entries: All entries

        Returns:
            Dict with error reduction metrics
        """
        # Determine best single model from overall hybrid analysis
        overall = self.hybrid_best_case_analysis(entries)
        best_config = overall.get("best_individual_config", "wav2vec2_normalized")

        preventable_errors = 0
        total_errors = 0

        for e in entries:
            # Get error attributions (use lenient by default)
            whisper_attr = e.get('ErrorAttributionWhisperLenient') or e.get('ErrorAttributionWhisper')
            wav2vec2_attr = e.get('ErrorAttributionWav2Vec2Lenient') or e.get('ErrorAttributionWav2Vec2')

            if not whisper_attr or not wav2vec2_attr:
                continue

            # Determine which is "best single" and which is "alternative"
            if 'whisper' in best_config.lower():
                best_attr = whisper_attr
                alt_attr = wav2vec2_attr
            else:
                best_attr = wav2vec2_attr
                alt_attr = whisper_attr

            # Count ASR_ERROR cases
            if best_attr == 'ASR_ERROR':
                total_errors += 1

                # Could hybrid prevent this?
                if alt_attr == 'CORRECT':
                    preventable_errors += 1

        reduction_rate = self.metrics_calc.safe_divide(preventable_errors, total_errors) * 100 if total_errors > 0 else 0

        return {
            "total_asr_errors": total_errors,
            "preventable_by_hybrid": preventable_errors,
            "still_errors": total_errors - preventable_errors,
            "reduction_rate": round(reduction_rate, 1),
            "best_config_used": best_config
        }

    def hybrid_practical_implications_analysis(self, entries: List[Dict]) -> Dict:
        """
        Calculate hybrid system performance for critical CAPT areas:
        1. Overall CER
        2. Welsh Digraphs (ll, ch, dd, ff, ng, rh, ph, th)
        3. Most Difficult Words
        4. Error Reduction Potential

        Compares Best Single Model vs Hybrid Best-Case to show practical value
        of intelligent model selection for Welsh CAPT deployment.

        Args:
            entries: List of entry dictionaries with all CER metrics

        Returns:
            Dictionary with comparison metrics for each critical area
        """
        # 1. OVERALL CER (reuse hybrid_best_case_analysis)
        overall_hybrid = self.hybrid_best_case_analysis(entries)

        if "error" in overall_hybrid:
            return overall_hybrid

        # 2. WELSH DIGRAPHS PERFORMANCE
        welsh_digraphs = ['ll', 'ch', 'dd', 'ff', 'ng', 'rh', 'ph', 'th']
        digraph_entries = [e for e in entries
                          if any(dg in e.get('Text', '').lower() for dg in welsh_digraphs)]

        digraph_hybrid = self._calculate_hybrid_subset(digraph_entries)

        # 3. MOST DIFFICULT WORDS (Top 10 hardest by normalized CER)
        word_cers = {}
        for e in entries:
            word = e.get('Text', '').strip()
            if not word:
                continue
            cer = e.get('CERWhisperLenient') or e.get('CERWav2Vec2Lenient') or 0
            if word not in word_cers:
                word_cers[word] = []
            word_cers[word].append(cer)

        # Get top 10 hardest words
        word_avg_cers = {w: np.mean(cers) for w, cers in word_cers.items() if cers}
        top_difficult_words = sorted(word_avg_cers.keys(),
                                    key=lambda w: word_avg_cers[w],
                                    reverse=True)[:10]

        difficult_entries = [e for e in entries
                            if e.get('Text', '').strip() in top_difficult_words]
        difficult_hybrid = self._calculate_hybrid_subset(difficult_entries)

        # 4. ERROR REDUCTION POTENTIAL
        error_reduction = self._calculate_error_reduction(entries)

        return {
            "overall": {
                "hybrid_cer": overall_hybrid["hybrid_mean_cer"],
                "best_single_cer": overall_hybrid["best_individual_cer"],
                "best_single_config": overall_hybrid["best_individual_config"],
                "improvement_absolute": overall_hybrid["improvement_absolute"],
                "improvement_percentage": overall_hybrid["improvement_percentage"],
                "sample_size": overall_hybrid["sample_size"]
            },
            "welsh_digraphs": digraph_hybrid,
            "difficult_words": difficult_hybrid,
            "error_reduction": error_reduction
        }

    def inter_rater_reliability(self, entries: List[Dict]) -> Dict:
        """
        Calculate simple agreement between Whisper and Wav2Vec2 models.
        Shows how often both models agree on correctness of pronunciations.
        Also computes average CER for agreement vs disagreement patterns.

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with agreement metrics and CER breakdown
        """
        # Check if we have text to compare against
        if not any('HumanTranscription' in e or 'Text' in e for e in entries):
            return {"error": "No reference text available for agreement analysis"}

        whisper_correct = []
        wav2vec2_correct = []
        both_correct = 0
        both_incorrect = 0
        whisper_only_correct = 0
        wav2vec2_only_correct = 0

        # Track CER values for each agreement pattern (using lenient metrics)
        cer_agree = []
        cer_disagree = []
        cer_both_correct = []
        cer_both_incorrect = []

        for e in entries:
            # Use HumanTranscription as ground truth, fall back to Text if not available
            reference = self.text_processor.normalize_text(
                e.get('HumanTranscription', e.get('Text', ''))
            )
            whisper = self.text_processor.normalize_text(
                e.get('AttemptWhisper', e.get('Attempt', ''))
            )
            wav2vec2 = self.text_processor.normalize_text(
                e.get('AttemptWav2Vec2', '')
            )

            # Apply lenient space-insensitive matching
            reference, whisper = self.text_processor.apply_lenient_normalization(
                reference, whisper
            )
            reference_for_wav2vec2, wav2vec2 = self.text_processor.apply_lenient_normalization(
                reference, wav2vec2
            )

            if not reference:
                continue

            w_correct = (whisper == reference)
            v_correct = (wav2vec2 == reference)

            whisper_correct.append(w_correct)
            wav2vec2_correct.append(v_correct)

            # Get lenient CER values
            whisper_cer = e.get('CERWhisperLenient', e.get('CERWhisper', 0))
            wav2vec2_cer = e.get('CERWav2Vec2Lenient', e.get('CERWav2Vec2', 0))
            avg_cer = (whisper_cer + wav2vec2_cer) / 2

            # Count agreement patterns and collect CER values
            if w_correct and v_correct:
                both_correct += 1
                cer_agree.append(avg_cer)
                cer_both_correct.append(avg_cer)
            elif not w_correct and not v_correct:
                both_incorrect += 1
                cer_agree.append(avg_cer)
                cer_both_incorrect.append(avg_cer)
            elif w_correct and not v_correct:
                whisper_only_correct += 1
                cer_disagree.append(avg_cer)
            else:
                wav2vec2_only_correct += 1
                cer_disagree.append(avg_cer)

        if len(whisper_correct) == 0:
            return {"error": "No valid entries for agreement analysis"}

        # Calculate overall agreement rate
        total = len(whisper_correct)
        agreements = sum(1 for w, v in zip(whisper_correct, wav2vec2_correct) if w == v)
        agreement_rate = agreements / total

        # Calculate average CER for each pattern
        avg_cer_agree = sum(cer_agree) / len(cer_agree) if cer_agree else 0
        avg_cer_disagree = sum(cer_disagree) / len(cer_disagree) if cer_disagree else 0
        avg_cer_both_correct = sum(cer_both_correct) / len(cer_both_correct) if cer_both_correct else 0
        avg_cer_both_incorrect = sum(cer_both_incorrect) / len(cer_both_incorrect) if cer_both_incorrect else 0

        return {
            "sample_size": total,
            "agreement_rate": float(agreement_rate),
            "agreement_counts": {
                "both_correct": both_correct,
                "both_incorrect": both_incorrect,
                "whisper_only_correct": whisper_only_correct,
                "wav2vec2_only_correct": wav2vec2_only_correct
            },
            "cer_breakdown": {
                "avg_cer_when_agree": round(avg_cer_agree, 2),
                "avg_cer_when_disagree": round(avg_cer_disagree, 2),
                "avg_cer_both_correct": round(avg_cer_both_correct, 2),
                "avg_cer_both_incorrect": round(avg_cer_both_incorrect, 2),
                "cer_difference": round(avg_cer_disagree - avg_cer_agree, 2)
            }
        }

    def _compute_error_attribution_stats(self, entries: List[Dict],
                                         whisper_key: str, wav2vec2_key: str) -> Dict:
        """
        Helper function to compute error attribution statistics for a set of metrics.

        Args:
            entries: List of entry dictionaries
            whisper_key: Key for Whisper error attribution field
            wav2vec2_key: Key for Wav2Vec2 error attribution field

        Returns:
            Dictionary with error distribution counts and percentages
        """
        # Count error types for both models
        whisper_errors = {"ASR_ERROR": 0, "USER_ERROR": 0, "AMBIGUOUS": 0, "CORRECT": 0, "NONE": 0}
        wav2vec2_errors = {"ASR_ERROR": 0, "USER_ERROR": 0, "AMBIGUOUS": 0, "CORRECT": 0, "NONE": 0}

        for e in entries:
            whisper_attr = e.get(whisper_key, '')
            wav2vec2_attr = e.get(wav2vec2_key, '')

            if whisper_attr in whisper_errors:
                whisper_errors[whisper_attr] += 1
            else:
                whisper_errors["NONE"] += 1

            if wav2vec2_attr in wav2vec2_errors:
                wav2vec2_errors[wav2vec2_attr] += 1
            else:
                wav2vec2_errors["NONE"] += 1

        # Calculate percentages
        total_whisper = sum(whisper_errors.values())
        total_wav2vec2 = sum(wav2vec2_errors.values())

        whisper_percentages = {
            k: float((v / total_whisper * 100) if total_whisper > 0 else 0)
            for k, v in whisper_errors.items()
        }
        wav2vec2_percentages = {
            k: float((v / total_wav2vec2 * 100) if total_wav2vec2 > 0 else 0)
            for k, v in wav2vec2_errors.items()
        }

        return {
            "whisper_distribution": {
                "counts": whisper_errors,
                "percentages": whisper_percentages
            },
            "wav2vec2_distribution": {
                "counts": wav2vec2_errors,
                "percentages": wav2vec2_percentages
            }
        }

    def error_attribution_analysis(self, entries: List[Dict]) -> Dict:
        """
        Analyze error attribution patterns for both models with pedagogical framing.
        Shows distribution of ASR_ERROR vs USER_ERROR vs CORRECT vs AMBIGUOUS

        Enhanced for MSc thesis with pedagogical utility metrics:
        - Usable feedback percentage (CORRECT + pedagogically safe ASR_ERROR)
        - Harmful error rate (false acceptances)
        - Model trustworthiness scores

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with error distribution, pedagogical metrics for both raw and normalized
        """
        raw_stats = self._compute_error_attribution_stats(
            entries,
            'ErrorAttributionWhisper',
            'ErrorAttributionWav2Vec2'
        )

        normalized_stats = self._compute_error_attribution_stats(
            entries,
            'ErrorAttributionWhisperLenient',
            'ErrorAttributionWav2Vec2Lenient'
        )

        def add_pedagogical_metrics(stats: Dict) -> Dict:
            """Add pedagogical interpretation to attribution stats"""
            w_dist = stats["whisper_distribution"]["percentages"]
            v_dist = stats["wav2vec2_distribution"]["percentages"]

            # Usable feedback: CORRECT + ASR_ERROR (learner gets useful signal)
            whisper_usable = w_dist.get("CORRECT", 0) + w_dist.get("ASR_ERROR", 0)
            wav2vec2_usable = v_dist.get("CORRECT", 0) + v_dist.get("ASR_ERROR", 0)

            # Trust score: How often can learner trust the ASR feedback?
            whisper_trust = w_dist.get("CORRECT", 0) - (w_dist.get("AMBIGUOUS", 0) * 0.5)
            wav2vec2_trust = v_dist.get("CORRECT", 0) - (v_dist.get("AMBIGUOUS", 0) * 0.5)

            stats["pedagogical_metrics"] = {
                "whisper": {
                    "usable_feedback_pct": self.metrics_calc.safe_float(whisper_usable),
                    "trust_score": self.metrics_calc.safe_float(max(0, whisper_trust))
                },
                "wav2vec2": {
                    "usable_feedback_pct": self.metrics_calc.safe_float(wav2vec2_usable),
                    "trust_score": self.metrics_calc.safe_float(max(0, wav2vec2_trust))
                },
                "interpretation": {
                    "more_usable": "Whisper" if whisper_usable > wav2vec2_usable else "Wav2Vec2",
                    "more_trustworthy": "Whisper" if whisper_trust > wav2vec2_trust else "Wav2Vec2",
                    "usable_advantage": self.metrics_calc.safe_float(abs(whisper_usable - wav2vec2_usable)),
                    "trust_advantage": self.metrics_calc.safe_float(abs(whisper_trust - wav2vec2_trust))
                }
            }

            return stats

        return {
            "raw": add_pedagogical_metrics(raw_stats),
            "normalized": add_pedagogical_metrics(normalized_stats)
        }

    def error_cost_analysis(self, entries: List[Dict]) -> Dict:
        """
        Analyze pedagogical costs of ASR errors for MSc thesis.

        Calculates false acceptance and false rejection rates:
        - False Acceptance (Type I): ASR accepts incorrect pronunciation (harmful)
        - False Rejection (Type II): ASR rejects correct pronunciation (demotivating)

        Returns both RAW and NORMALIZED metrics to show how lenient matching affects
        pedagogical safety.

        Args:
            entries: List of entry dictionaries with attribution labels

        Returns:
            Dictionary with error cost metrics for both models and both metrics
        """
        def calculate_error_costs(entries_list: List[Dict], model_prefix: str,
                                 use_lenient: bool = False) -> Dict:
            """Helper to calculate error costs for one model"""
            false_acceptance_count = 0
            false_rejection_count = 0
            true_positive_count = 0
            true_negative_count = 0
            total_count = 0

            # Examples for qualitative analysis
            false_acceptance_examples = []
            false_rejection_examples = []

            # Choose field names based on whether we're using lenient or strict
            if use_lenient:
                attribution_field = f'ErrorAttribution{model_prefix}Lenient'
                asr_field = f'Attempt{model_prefix}Lenient'
            else:
                attribution_field = f'ErrorAttribution{model_prefix}'
                asr_field = f'Attempt{model_prefix}'

            for entry in entries_list:
                attribution = entry.get(attribution_field, '')
                target = entry.get('Text', '')
                human_transcription = entry.get('HumanTranscription', '')
                asr_transcription = entry.get(asr_field, '')

                if not attribution:
                    continue

                total_count += 1

                # For CAPT: Determine if ASR output matches TARGET word
                if use_lenient:
                    asr_matches_target = (asr_transcription == self.text_processor.normalize_text(target))
                else:
                    asr_matches_target = (
                        self.text_processor.normalize_text(asr_transcription) ==
                        self.text_processor.normalize_text(target)
                    )

                # Classify pedagogical error cost based on Attribution field
                if attribution == "USER_ERROR":
                    # User mispronounced
                    if asr_matches_target:
                        # False Acceptance: HARMFUL
                        false_acceptance_count += 1
                        if len(false_acceptance_examples) < 5:
                            false_acceptance_examples.append({
                                "target": target,
                                "human": human_transcription,
                                "asr": asr_transcription
                            })
                    else:
                        # True Negative: CORRECT
                        true_negative_count += 1

                elif attribution == "CORRECT":
                    # User pronounced correctly
                    if asr_matches_target:
                        # True Positive: GOOD
                        true_positive_count += 1
                    else:
                        # False Rejection: DEMOTIVATING
                        false_rejection_count += 1
                        if len(false_rejection_examples) < 5:
                            false_rejection_examples.append({
                                "target": target,
                                "human": human_transcription,
                                "asr": asr_transcription
                            })

                elif attribution == "ASR_ERROR":
                    # User correct, but ASR didn't output target
                    # False Rejection: DEMOTIVATING
                    false_rejection_count += 1
                    if len(false_rejection_examples) < 5:
                        false_rejection_examples.append({
                            "target": target,
                            "human": human_transcription,
                            "asr": asr_transcription
                        })

            # Calculate rates
            false_acceptance_rate = self.metrics_calc.safe_divide(
                false_acceptance_count, total_count
            ) * 100
            false_rejection_rate = self.metrics_calc.safe_divide(
                false_rejection_count, total_count
            ) * 100
            true_positive_rate = self.metrics_calc.safe_divide(
                true_positive_count, total_count
            ) * 100
            true_negative_rate = self.metrics_calc.safe_divide(
                true_negative_count, total_count
            ) * 100

            # Pedagogical risk score (weight false acceptance heavily)
            # False acceptance is 3x worse than false rejection for learning
            risk_score = (
                (false_acceptance_count * 3 + false_rejection_count) / total_count * 100
                if total_count > 0 else 0
            )

            # Safety score (inverse of risk)
            safety_score = 100 - risk_score

            # Log analysis results for debugging
            metric_type = "Lenient" if use_lenient else "Raw"
            print(f"[error_costs] {model_prefix} ({metric_type}): total={total_count}, "
                  f"false_accept={false_acceptance_count}, false_reject={false_rejection_count}, "
                  f"true_pos={true_positive_count}, true_neg={true_negative_count}")

            return {
                "false_acceptance": {
                    "count": false_acceptance_count,
                    "rate": self.metrics_calc.safe_float(false_acceptance_rate),
                    "examples": false_acceptance_examples
                },
                "false_rejection": {
                    "count": false_rejection_count,
                    "rate": self.metrics_calc.safe_float(false_rejection_rate),
                    "examples": false_rejection_examples
                },
                "true_positive": {
                    "count": true_positive_count,
                    "rate": self.metrics_calc.safe_float(true_positive_rate)
                },
                "true_negative": {
                    "count": true_negative_count,
                    "rate": self.metrics_calc.safe_float(true_negative_rate)
                },
                "total_cases": total_count,
                "pedagogical_risk_score": self.metrics_calc.safe_float(risk_score),
                "pedagogical_safety_score": self.metrics_calc.safe_float(safety_score)
            }

        # Calculate for both models - RAW metrics
        raw_results = {
            "whisper": calculate_error_costs(entries, "Whisper", use_lenient=False),
            "wav2vec2": calculate_error_costs(entries, "Wav2Vec2", use_lenient=False)
        }

        # Calculate for both models - NORMALIZED metrics
        normalized_results = {
            "whisper": calculate_error_costs(entries, "Whisper", use_lenient=True),
            "wav2vec2": calculate_error_costs(entries, "Wav2Vec2", use_lenient=True)
        }

        # Determine safer model for raw
        raw_safer_model = (
            "Whisper"
            if raw_results["whisper"]["pedagogical_safety_score"] >
               raw_results["wav2vec2"]["pedagogical_safety_score"]
            else "Wav2Vec2"
        )
        raw_safety_advantage = abs(
            raw_results["whisper"]["pedagogical_safety_score"] -
            raw_results["wav2vec2"]["pedagogical_safety_score"]
        )

        raw_results["comparison"] = {
            "safer_model": raw_safer_model,
            "safety_advantage": self.metrics_calc.safe_float(raw_safety_advantage),
            "recommendation": (
                f"{raw_safer_model} is {raw_safety_advantage:.1f} points safer "
                f"for learners (lower risk of harmful errors)"
            )
        }

        # Determine safer model for normalized
        norm_safer_model = (
            "Whisper"
            if normalized_results["whisper"]["pedagogical_safety_score"] >
               normalized_results["wav2vec2"]["pedagogical_safety_score"]
            else "Wav2Vec2"
        )
        norm_safety_advantage = abs(
            normalized_results["whisper"]["pedagogical_safety_score"] -
            normalized_results["wav2vec2"]["pedagogical_safety_score"]
        )

        normalized_results["comparison"] = {
            "safer_model": norm_safer_model,
            "safety_advantage": self.metrics_calc.safe_float(norm_safety_advantage),
            "recommendation": (
                f"{norm_safer_model} is {norm_safety_advantage:.1f} points safer "
                f"for learners with normalized matching"
            )
        }

        # Calculate improvement from normalization
        improvement = {
            "whisper": {},
            "wav2vec2": {}
        }

        for model in ["whisper", "wav2vec2"]:
            # False acceptance improvement (lower is better)
            raw_fa_rate = raw_results[model]["false_acceptance"]["rate"]
            norm_fa_rate = normalized_results[model]["false_acceptance"]["rate"]
            fa_improvement = raw_fa_rate - norm_fa_rate

            # False rejection improvement (lower is better)
            raw_fr_rate = raw_results[model]["false_rejection"]["rate"]
            norm_fr_rate = normalized_results[model]["false_rejection"]["rate"]
            fr_improvement = raw_fr_rate - norm_fr_rate

            # Safety score improvement (higher is better)
            raw_safety = raw_results[model]["pedagogical_safety_score"]
            norm_safety = normalized_results[model]["pedagogical_safety_score"]
            safety_improvement = norm_safety - raw_safety

            improvement[model] = {
                "false_acceptance_change": self.metrics_calc.safe_float(fa_improvement),
                "false_rejection_change": self.metrics_calc.safe_float(fr_improvement),
                "safety_score_improvement": self.metrics_calc.safe_float(safety_improvement),
                "interpretation": (
                    f"Normalization {'improves' if safety_improvement > 0 else 'reduces'} "
                    f"safety by {abs(safety_improvement):.1f} points. "
                    f"False acceptance {'reduced' if fa_improvement > 0 else 'increased'} "
                    f"by {abs(fa_improvement):.1f}pp, "
                    f"false rejection {'reduced' if fr_improvement > 0 else 'increased'} "
                    f"by {abs(fr_improvement):.1f}pp."
                )
            }

        return {
            "raw": raw_results,
            "normalized": normalized_results,
            "improvement": improvement
        }
