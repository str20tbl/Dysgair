import io
import os
import requests
from fastapi import FastAPI, Response, UploadFile, File, Form
from fastapi.middleware.cors import CORSMiddleware
from faster_whisper import WhisperModel
from starlette.background import BackgroundTask
from transformers import pipeline
from typing import List, Dict
from pydantic import BaseModel

from statistical_analysis import StatisticalAnalyzer

app = FastAPI(title="techiaith/whisper-large-v3-ft-cv-cy",
              summary="Bilingual Welsh-English Speech-to-Text",
              description="Adnabod Leferydd Dwyieithog Cymraeg-Saesneg",
              version="1.0.0",
              debug=True)

app.add_middleware(
    CORSMiddleware,
    allow_origins=['*'],
    allow_credentials=True,
    allow_methods=['*'],
    allow_headers=['*'],
)

whisper_transcriber = pipeline("automatic-speech-recognition", model="techiaith/whisper-large-v3-ft-cv-cy")
wav2vec2_transcriber = pipeline("automatic-speech-recognition", model="techiaith/wav2vec2-btb-cv-ft-cv-cy")

# Initialize statistical analyzer
analyzer = StatisticalAnalyzer()


# Request/Response models for statistical analysis
class AnalysisRequest(BaseModel):
    entries: List[Dict]


@app.get("/")
async def root():
    return {"success": True, "message": "Dysgair techiaith/whisper-large-v3-ft-cv-cy API"}


@app.get("/transcribe")
async def transcribe(filename: str):
    # Transcribe with both models
    whisper_segments = whisper_transcriber(f'{filename}')
    wav2vec2_segments = wav2vec2_transcriber(f'{filename}')

    # Extract text from both results
    whisper_result = whisper_segments["text"].strip()
    wav2vec2_result = wav2vec2_segments["text"].strip()

    return {
        "success": True,
        "results": {
            "whisper": whisper_result,
            "wav2vec2": wav2vec2_result
        }
    }


# Statistical Analysis Endpoints

@app.post("/analysis/model-comparison")
async def model_comparison(request: AnalysisRequest):
    """
    Comprehensive comparison between Whisper and Wav2Vec2 models

    Includes:
    - Paired t-test and Wilcoxon test
    - Effect sizes (Cohen's d)
    - Confidence intervals
    - Descriptive statistics

    For MSc research in Language Technologies
    """
    try:
        results = analyzer.model_comparison_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/hybrid-best-case")
async def hybrid_best_case(request: AnalysisRequest):
    """
    Calculate best-case hybrid performance

    For each word, selects minimum CER across all 4 combinations:
    - Whisper Raw, Whisper Normalized, Wav2Vec2 Raw, Wav2Vec2 Normalized

    Answers: "What if we had a smart hybrid CAPT system that could
    intelligently choose the best ASR model and normalization approach
    for each word?"

    Includes:
    - Hybrid mean CER (theoretical best-case)
    - Improvement vs best individual model
    - Selection distribution (which approach was best most often)
    - Model preference (Whisper vs Wav2Vec2)
    - Normalization preference (Raw vs Normalized)

    Critical for evaluating potential of hybrid CAPT architectures
    """
    try:
        results = analyzer.hybrid_best_case_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/inter-rater-reliability")
async def inter_rater_reliability(request: AnalysisRequest):
    """
    Calculate inter-rater reliability between models

    Includes:
    - Cohen's Kappa
    - Agreement rates
    - Confusion matrix

    Useful for assessing model consistency
    """
    try:
        results = analyzer.inter_rater_reliability(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/error-attribution")
async def error_attribution(request: AnalysisRequest):
    """
    Analyze error attribution patterns

    Includes:
    - Error type distributions for both models
    - Chi-square test for distribution differences
    - Percentage breakdowns

    Helps identify systematic differences in error patterns
    """
    try:
        results = analyzer.error_attribution_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}




@app.post("/analysis/word-difficulty")
async def word_difficulty(request: AnalysisRequest):
    """
    Identify difficult words for ASR models (CER-based)

    Includes:
    - Word-level CER averages (primary metric)
    - Difficulty rankings by CER
    - Most/least difficult words for ASR
    - CER distribution histogram

    Focuses on character-level errors (more meaningful for single-word pronunciation)
    """
    try:
        results = analyzer.word_difficulty_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/word-length")
async def word_length(request: AnalysisRequest):
    """
    Analyze ASR performance by word length (character count)

    Groups words into length categories:
    - 1-3 characters (short words)
    - 4-6 characters (medium words)
    - 7-9 characters (long words)
    - 10+ characters (very long words)

    For each category, calculates:
    - Average CER for both models
    - Average WER for both models
    - Word count in category
    - Both raw (strict) and normalized (lenient) metrics

    Helps identify if word length correlates with ASR difficulty
    """
    try:
        results = analyzer.word_length_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/over-transcription")
async def over_transcription(request: AnalysisRequest):
    """
    Detect over-transcription cases (ASR predicts multiple words for single-word target)

    Identifies statistically meaningful WER cases where WER >> CER,
    indicating ASR hallucinated additional words rather than character-level errors.

    Includes:
    - List of anomalous cases with transcriptions
    - WER-CER delta for each model
    - Count and percentage of anomalies

    WER analysis is only meaningful in these specific failure modes
    """
    try:
        results = analyzer.detect_over_transcription(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/character-errors")
async def character_errors(request: AnalysisRequest):
    """
    Analyze character-level errors with Welsh digraph support

    Four separate analyses:
    - Overall: All character/digraph errors
    - Vowels: Vowel substitutions only (a,e,i,o,u,w,y + diacritics)
    - Consonants: Single consonant substitutions
    - Digraphs: Welsh digraph errors (ll,ch,dd,ff,ng,rh,ph,th)

    For each category:
    - Confusion matrix (all non-zero confusions)
    - Per-letter error rates
    - Separate for Whisper and Wav2Vec2

    Diacritics tracked separately (â ≠ a)
    """
    try:
        results = analyzer.character_error_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}



@app.post("/analysis/executive-summary")
async def executive_summary(request: AnalysisRequest):
    """
    Generate executive summary with key metrics and findings for MSc thesis
    
    Provides high-level overview including:
    - Sample size and scope
    - Overall model winner  
    - Key quantitative findings
    - Percentage improvements
    """
    try:
        results = analyzer.executive_summary_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/linguistic-patterns")
async def linguistic_patterns(request: AnalysisRequest):
    """
    Analyze error patterns by linguistic category for MSc thesis
    
    Breaks down character errors into:
    - Vowel errors (a, e, i, o, u, w, y + diacritics)
    - Consonant errors (single consonants)
    - Digraph errors (ll, ch, dd, ff, ng, rh, ph, th)
    
    Critical for Language Technologies/Applied Linguistics focus
    """
    try:
        results = analyzer.linguistic_pattern_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/error-costs")
async def error_costs(request: AnalysisRequest):
    """
    Analyze pedagogical costs of ASR errors
    
    Calculates false acceptance and false rejection rates:
    - False Acceptance (Type I): ASR accepts incorrect pronunciation (harmful)
    - False Rejection (Type II): ASR rejects correct pronunciation (demotivating)
    
    Critical for understanding pedagogical implications
    """
    try:
        results = analyzer.error_cost_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/qualitative-examples")
async def qualitative_examples(request: AnalysisRequest):
    """
    Select representative qualitative examples for MSc thesis presentation
    
    Provides real-world examples of:
    - Best case (both models correct)
    - Whisper success, Wav2Vec2 failure
    - Wav2Vec2 success, Whisper failure
    - Both models failed
    - Pedagogically critical errors
    """
    try:
        results = analyzer.qualitative_examples_selection(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/consistency-reliability")
async def consistency_reliability(request: AnalysisRequest):
    """
    Analyze consistency and reliability of ASR models
    
    Beyond average accuracy, examines:
    - Variance/stability of error rates
    - Interquartile range (IQR)
    - Percentage of reliable predictions
    - Worst-case performance (95th percentile)
    """
    try:
        results = analyzer.consistency_reliability_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/study-design")
async def study_design(request: AnalysisRequest):
    """
    Provide study design and methodological metadata for MSc thesis
    
    Returns structured information about:
    - Study type and design
    - Limitations
    - Statistical approach
    - Data collection context
    """
    try:
        results = analyzer.study_design_metadata(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/practical-recommendations")
async def practical_recommendations(request: AnalysisRequest):
    """
    Generate practical recommendations for Welsh CAPT system design
    
    Synthesizes all findings into actionable guidance for:
    - Model selection decisions
    - System architecture design
    - When to use human verification
    - Confidence thresholds
    - Pedagogical considerations
    """
    try:
        results = analyzer.practical_recommendations(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.post("/analysis/hybrid-practical-implications")
async def hybrid_practical_implications(request: AnalysisRequest):
    """
    Analyze practical implications of hybrid CAPT system for critical areas

    Compares Best Single Model vs Hybrid Best-Case across:
    - Overall CER performance
    - Welsh Digraphs (ll, ch, dd, ff, ng, rh, ph, th)
    - Most Difficult Words (top 10 hardest)
    - ASR Error Reduction Potential

    Answers: "Is it worth building a hybrid system for Welsh CAPT?"
    Provides data for Discussion/Implications section of MSc thesis
    """
    try:
        results = analyzer.hybrid_practical_implications_analysis(request.entries)
        return {"success": True, "analysis": results}
    except Exception as e:
        return {"success": False, "error": str(e)}


@app.get("/analysis/info")
async def analysis_info():
    """
    Get information about available statistical analyses

    Returns descriptions of all available endpoints
    """
    return {
        "success": True,
        "available_analyses": {
            "model_comparison": "Descriptive statistics comparing Whisper vs Wav2Vec2 (CER-focused, WER secondary)",
            "inter_rater_reliability": "Model agreement metrics - how often both models agree on correctness",
            "error_attribution": "Distribution of error types (ASR_ERROR, USER_ERROR, CORRECT, AMBIGUOUS)",
            "word_difficulty": "Word-level difficulty rankings based on average CER (primary) for ASR models",
            "over_transcription": "Detect ASR over-transcription failures (WER >> CER cases)",
            "character_errors": "Character-level error analysis with Welsh digraph support - 4 categories (overall, vowels, consonants, digraphs)"
        },
        "intended_use": "Single-participant MSc study - comparing ASR models for Welsh pronunciation feedback"
    }
