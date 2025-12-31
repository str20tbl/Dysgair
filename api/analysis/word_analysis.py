"""
Word Analysis Module
Handles word-based analysis including difficulty, length patterns, and over-transcription detection
"""

import numpy as np
from typing import List, Dict
from .metrics import MetricsCalculator


class WordAnalyzer:
    """
    Analyzes ASR performance patterns at the word level.

    Provides word difficulty rankings, word length correlation analysis,
    and over-transcription detection for Welsh CAPT evaluation.
    """

    def __init__(self, metrics_calc: MetricsCalculator = None):
        """
        Initialize WordAnalyzer.

        Args:
            metrics_calc: Optional MetricsCalculator instance
        """
        self.metrics_calc = metrics_calc or MetricsCalculator()

    def _compute_word_difficulty_stats(self, entries: List[Dict],
                                      whisper_wer_key: str, wav2vec2_wer_key: str,
                                      whisper_cer_key: str, wav2vec2_cer_key: str) -> Dict:
        """
        Helper function to compute word difficulty statistics for a set of metrics.

        Args:
            entries: List of entry dictionaries
            whisper_wer_key: Key for Whisper WER field
            wav2vec2_wer_key: Key for Wav2Vec2 WER field
            whisper_cer_key: Key for Whisper CER field
            wav2vec2_cer_key: Key for Wav2Vec2 CER field

        Returns:
            Dictionary with word rankings and distribution
        """
        word_stats = {}

        for e in entries:
            word = e.get('Text', '').strip()
            if not word:
                continue

            if word not in word_stats:
                word_stats[word] = {
                    'whisper_wers': [],
                    'wav2vec2_wers': [],
                    'whisper_cers': [],
                    'wav2vec2_cers': [],
                    'count': 0
                }

            word_stats[word]['whisper_wers'].append(e.get(whisper_wer_key, 0))
            word_stats[word]['wav2vec2_wers'].append(e.get(wav2vec2_wer_key, 0))
            word_stats[word]['whisper_cers'].append(e.get(whisper_cer_key, 0))
            word_stats[word]['wav2vec2_cers'].append(e.get(wav2vec2_cer_key, 0))
            word_stats[word]['count'] += 1

        # Calculate averages
        word_rankings = []
        for word, stats in word_stats.items():
            word_rankings.append({
                "word": word,
                "attempts": stats['count'],
                "whisper_avg_wer": self.metrics_calc.safe_float(np.mean(stats['whisper_wers'])) if stats['whisper_wers'] else 0.0,
                "wav2vec2_avg_wer": self.metrics_calc.safe_float(np.mean(stats['wav2vec2_wers'])) if stats['wav2vec2_wers'] else 0.0,
                "whisper_avg_cer": self.metrics_calc.safe_float(np.mean(stats['whisper_cers'])) if stats['whisper_cers'] else 0.0,
                "wav2vec2_avg_cer": self.metrics_calc.safe_float(np.mean(stats['wav2vec2_cers'])) if stats['wav2vec2_cers'] else 0.0,
                "word_length": len(word)
            })

        # Sort by average CER (descending = highest ASR error rates first)
        word_rankings.sort(key=lambda x: (x['whisper_avg_cer'] + x['wav2vec2_avg_cer']) / 2, reverse=True)

        # Calculate CER distribution (histogram buckets)
        distribution = self._calculate_error_distribution(word_rankings)

        return {
            "total_unique_words": len(word_rankings),
            "most_difficult": word_rankings[:10],
            "easiest": word_rankings[-10:][::-1],
            "all_words": word_rankings,
            "distribution": distribution
        }

    def word_difficulty_analysis(self, entries: List[Dict]) -> Dict:
        """
        Analyze ASR model performance by word with both raw and normalized metrics.

        Identifies words where ASR models have highest error rates.
        This evaluates ASR limitations, not learner proficiency.
        Focus on CER (more meaningful for single-word Welsh pronunciation).

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with ASR error rate rankings by word for both raw and normalized
        """
        return {
            "raw": self._compute_word_difficulty_stats(
                entries,
                'WERWhisper',
                'WERWav2Vec2',
                'CERWhisper',
                'CERWav2Vec2'
            ),
            "normalized": self._compute_word_difficulty_stats(
                entries,
                'WERWhisperLenient',
                'WERWav2Vec2Lenient',
                'CERWhisperLenient',
                'CERWav2Vec2Lenient'
            )
        }

    def _compute_word_length_stats(self, entries: List[Dict],
                                   whisper_cer_key: str, wav2vec2_cer_key: str,
                                   whisper_wer_key: str, wav2vec2_wer_key: str) -> Dict:
        """
        Helper to compute word length statistics for a specific metric set.

        Args:
            entries: List of entry dictionaries
            whisper_cer_key: Key for Whisper CER metric
            wav2vec2_cer_key: Key for Wav2Vec2 CER metric
            whisper_wer_key: Key for Whisper WER metric
            wav2vec2_wer_key: Key for Wav2Vec2 WER metric

        Returns:
            Dictionary with length categories and their average metrics
        """
        # Define length categories
        length_categories = {
            "one_to_three": {"min": 1, "max": 3, "label": "1-3", "whisper_cers": [], "wav2vec2_cers": [], "whisper_wers": [], "wav2vec2_wers": [], "count": 0},
            "four_to_six": {"min": 4, "max": 6, "label": "4-6", "whisper_cers": [], "wav2vec2_cers": [], "whisper_wers": [], "wav2vec2_wers": [], "count": 0},
            "seven_to_nine": {"min": 7, "max": 9, "label": "7-9", "whisper_cers": [], "wav2vec2_cers": [], "whisper_wers": [], "wav2vec2_wers": [], "count": 0},
            "ten_plus": {"min": 10, "max": 9999, "label": "10+", "whisper_cers": [], "wav2vec2_cers": [], "whisper_wers": [], "wav2vec2_wers": [], "count": 0}
        }

        for e in entries:
            target_word = e.get('Text', '')
            if not target_word:
                continue

            word_length = len(target_word)

            # Find appropriate category
            category = None
            for cat_name, cat_data in length_categories.items():
                if cat_data['min'] <= word_length <= cat_data['max']:
                    category = cat_name
                    break

            if not category:
                continue

            # Get metric values
            whisper_cer = e.get(whisper_cer_key, 0)
            wav2vec2_cer = e.get(wav2vec2_cer_key, 0)
            whisper_wer = e.get(whisper_wer_key, 0)
            wav2vec2_wer = e.get(wav2vec2_wer_key, 0)

            # Add to category
            length_categories[category]['whisper_cers'].append(whisper_cer)
            length_categories[category]['wav2vec2_cers'].append(wav2vec2_cer)
            length_categories[category]['whisper_wers'].append(whisper_wer)
            length_categories[category]['wav2vec2_wers'].append(wav2vec2_wer)
            length_categories[category]['count'] += 1

        # Calculate averages for each category
        results = {}
        for cat_key in ["one_to_three", "four_to_six", "seven_to_nine", "ten_plus"]:
            cat_data = length_categories[cat_key]
            if cat_data['count'] == 0:
                results[cat_key] = {
                    "label": cat_data['label'],
                    "count": 0,
                    "whisper_avg_cer": 0,
                    "wav2vec2_avg_cer": 0,
                    "whisper_avg_wer": 0,
                    "wav2vec2_avg_wer": 0
                }
                continue

            avg_whisper_cer = sum(cat_data['whisper_cers']) / len(cat_data['whisper_cers'])
            avg_wav2vec2_cer = sum(cat_data['wav2vec2_cers']) / len(cat_data['wav2vec2_cers'])
            avg_whisper_wer = sum(cat_data['whisper_wers']) / len(cat_data['whisper_wers'])
            avg_wav2vec2_wer = sum(cat_data['wav2vec2_wers']) / len(cat_data['wav2vec2_wers'])

            results[cat_key] = {
                "label": cat_data['label'],
                "count": cat_data['count'],
                "whisper_avg_cer": round(avg_whisper_cer, 2),
                "wav2vec2_avg_cer": round(avg_wav2vec2_cer, 2),
                "whisper_avg_wer": round(avg_whisper_wer, 2),
                "wav2vec2_avg_wer": round(avg_wav2vec2_wer, 2)
            }

        return results

    def word_length_analysis(self, entries: List[Dict]) -> Dict:
        """
        Analyze ASR performance by word length (character count).
        Groups words into length categories and calculates average CER/WER for each.

        Args:
            entries: List of entry dictionaries

        Returns:
            Dictionary with length-based analysis for both raw and normalized metrics
        """
        return {
            "raw": self._compute_word_length_stats(
                entries,
                'CERWhisper',
                'CERWav2Vec2',
                'WERWhisper',
                'WERWav2Vec2'
            ),
            "normalized": self._compute_word_length_stats(
                entries,
                'CERWhisperLenient',
                'CERWav2Vec2Lenient',
                'WERWhisperLenient',
                'WERWav2Vec2Lenient'
            )
        }

    @staticmethod
    def _calculate_error_distribution(word_rankings: List[Dict]) -> Dict:
        """
        Calculate CER distribution for histogram visualization.

        Buckets words into 10% CER ranges (0-10%, 10-20%, etc.)
        for both Whisper and Wav2Vec2 models.

        Args:
            word_rankings: List of word difficulty items with avg_cer fields

        Returns:
            Dictionary with buckets and counts for each model (CER-based)
        """
        # Define 10% buckets
        buckets = [
            "0-10%", "10-20%", "20-30%", "30-40%", "40-50%",
            "50-60%", "60-70%", "70-80%", "80-90%", "90-100%"
        ]

        whisper_counts = [0] * 10
        wav2vec2_counts = [0] * 10

        for word_data in word_rankings:
            whisper_cer = word_data['whisper_avg_cer']
            wav2vec2_cer = word_data['wav2vec2_avg_cer']

            # Determine bucket index (0-9)
            whisper_bucket = min(int(whisper_cer / 10), 9)
            wav2vec2_bucket = min(int(wav2vec2_cer / 10), 9)

            whisper_counts[whisper_bucket] += 1
            wav2vec2_counts[wav2vec2_bucket] += 1

        return {
            "buckets": buckets,
            "whisper_counts": whisper_counts,
            "wav2vec2_counts": wav2vec2_counts
        }

    def detect_over_transcription(self, entries: List[Dict], threshold: float = 20.0) -> Dict:
        """
        Detect cases where ASR models predict multiple words for single-word targets.
        Identified by WER significantly exceeding CER (WER-CER delta > threshold).

        This indicates a specific ASR failure mode: hallucinating additional words
        instead of character-level errors.

        Args:
            entries: List of entry dictionaries
            threshold: Minimum WER-CER difference to flag (default 20%)

        Returns:
            Dictionary containing anomalies, count, and percentage
        """
        anomalies = []

        for e in entries:
            target = e.get('Text', '').strip()
            if not target:
                continue

            # Get WER and CER for both models
            whisper_wer = e.get('WERWhisper', e.get('WER', 0))
            whisper_cer = e.get('CERWhisper', e.get('CER', 0))
            wav2vec2_wer = e.get('WERWav2Vec2', 0)
            wav2vec2_cer = e.get('CERWav2Vec2', 0)

            # Calculate deltas
            whisper_delta = whisper_wer - whisper_cer
            wav2vec2_delta = wav2vec2_wer - wav2vec2_cer

            # Flag if either model shows significant over-transcription
            if whisper_delta > threshold or wav2vec2_delta > threshold:
                anomalies.append({
                    "target": target,
                    "whisper_transcription": e.get('AttemptWhisper', e.get('Attempt', '')),
                    "wav2vec2_transcription": e.get('AttemptWav2Vec2', ''),
                    "whisper_wer": self.metrics_calc.safe_float(whisper_wer),
                    "whisper_cer": self.metrics_calc.safe_float(whisper_cer),
                    "whisper_delta": self.metrics_calc.safe_float(whisper_delta),
                    "wav2vec2_wer": self.metrics_calc.safe_float(wav2vec2_wer),
                    "wav2vec2_cer": self.metrics_calc.safe_float(wav2vec2_cer),
                    "wav2vec2_delta": self.metrics_calc.safe_float(wav2vec2_delta)
                })

        total_entries = len([e for e in entries if e.get('Text', '').strip()])
        percentage = self.metrics_calc.safe_divide(len(anomalies), total_entries, default=0.0) * 100

        return {
            "anomalies": anomalies,
            "count": len(anomalies),
            "percentage": self.metrics_calc.safe_float(percentage),
            "total_entries": total_entries
        }
