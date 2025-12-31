"""
Statistical Analysis Module for Dysgair ASR Research
Modular structure for analyzing Welsh ASR transcription quality
"""

from .text_processing import WelshTextProcessor
from .metrics import MetricsCalculator
from .error_analysis import ErrorAnalyzer
from .linguistic_analysis import LinguisticAnalyzer
from .model_comparison import ModelComparator
from .reporting import ReportGenerator
from .word_analysis import WordAnalyzer

__all__ = [
    'WelshTextProcessor',
    'MetricsCalculator',
    'ErrorAnalyzer',
    'LinguisticAnalyzer',
    'ModelComparator',
    'ReportGenerator',
    'WordAnalyzer',
]
