/**
 * Analytics Dashboard for Dysg

air ASR Research
 * Visualizes statistical analysis results using Chart.js
 * Provides LaTeX export functionality for academic publications
 */

let analysisData = {};
let charts = {};

const elements = {
    btnLoadAnalytics: $('#btnLoadAnalytics'),
    btnExportLaTeX: $('#btnExportLaTeX'),
    btnExportDissertation: $('#btnExportDissertation'),
    loadingIndicator: $('#loadingIndicator'),
    filterUser: $('#filterUser'),
    tableModelStats: $('#tableModelStats tbody'),
    agreementRate: $('#agreementRate'),
    werAnomaliesSection: $('#werAnomaliesSection'),
    werAnomalyCount: $('#werAnomalyCount'),
    werAnomalyPercentage: $('#werAnomalyPercentage'),
    tableWERanomalies: $('#tableWERanomalies tbody')
};

const COLORS = {
    whisper: '#3498db',      // Blue
    wav2vec2: '#e74c3c',     // Red
    correct: '#2ecc71',      // Green
    asrError: '#e67e22',     // Orange
    userError: '#9b59b6',    // Purple
    ambiguous: '#95a5a6',    // Gray
    gridLines: '#dee2e6'
};

$(() => {
    elements.btnLoadAnalytics.click(() => {
        loadAnalytics();
    });
    elements.btnExportLaTeX.click(() => {
        exportForLaTeX();
    });

    elements.btnExportDissertation.click(() => {
        exportDissertationCharts();
    });

    // Event delegation for export section buttons
    $(document).on('click', '.export-section-btn', function() {
        const sectionName = $(this).data('section');
        exportSection(sectionName);
    });
});

async function loadAnalytics() {
    const userID = elements.filterUser.val();

    elements.loadingIndicator.show();
    elements.btnLoadAnalytics.prop('disabled', true);
    elements.btnExportLaTeX.prop('disabled', true);
    elements.btnExportDissertation.prop('disabled', true);

    try {
        const response = await fetch('/Admin/Analytics/Run', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: new URLSearchParams({ userID })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        if (data.success) {
            analysisData = data.analyses;
            renderAllCharts();
            elements.btnExportLaTeX.prop('disabled', false);
            elements.btnExportDissertation.prop('disabled', false);
        } else {
            showNotification(`Error loading analytics: ${data.error || 'Unknown error'}`, 'error');
        }
    } catch (error) {
        console.error('Load analytics error:', error);
        showNotification('Failed to load analytics data / Methwyd llwytho data dadansoddol', 'error');
    } finally {
        elements.loadingIndicator.hide();
        elements.btnLoadAnalytics.prop('disabled', false);
    }
}

function renderAllCharts() {
    if (!analysisData) return;

    // Render all sections in order
    renderExecutiveSummary();          // Section 0 - NEW
    renderModelComparison();            // Section 1 - Enhanced
    renderPhonologicalAnalysis();       // Section 2 - Comprehensive phonological analysis (merged old Sections 2 & 5)
    renderComprehensiveErrorAnalysis(); // Section 3 - Merged comprehensive error analysis
    renderModelAgreement();             // Section 4 - NEW: Model Agreement Analysis (inter-rater reliability)
    renderTopDifficultWords();          // Section 5 - Most challenging words
    renderWordLength();                 // Section 5B - NEW: Performance by word length
    renderWordDifficulty();             // Section 6 - CER distribution
    renderConsistencyReliability();     // Section 7 - Consistency analysis
    renderQualitativeExamples();        // Section 8 - Qualitative examples
    renderStudyDesign();                // Section 9 - Study design insights
    renderPracticalRecommendations();   // Section 10 - Recommendations
    renderHybridAnalysis();             // Section 11 - MERGED: Comprehensive hybrid system analysis
}

function renderModelComparisonChart(canvasId, metricData, title, chartKey) {
    const data = {
        labels: ['Mean CER'],
        datasets: [
            createBarDataset('Whisper', [metricData.cer_analysis.whisper.mean], COLORS.whisper),
            createBarDataset('Wav2Vec2', [metricData.cer_analysis.wav2vec2.mean], COLORS.wav2vec2)
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            ...getStandardPluginsConfig(title),
            tooltip: {
                callbacks: {
                    label: getCERTooltipFormatter('Mean CER')
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Character Error Rate (%)'),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'ASR Model' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

function renderModelComparisonChartCombined(canvasId, rawData, normalizedData) {
    /**
     * Renders a combined chart showing raw and normalized CER metrics with ghost bar approach
     */
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    const whisperRaw = rawData.cer_analysis.whisper.mean;
    const wav2vec2Raw = rawData.cer_analysis.wav2vec2.mean;
    const whisperNorm = normalizedData.cer_analysis.whisper.mean;
    const wav2vec2Norm = normalizedData.cer_analysis.wav2vec2.mean;

    const chartKey = canvasId;

    const data = {
        labels: ['Whisper', 'Wav2Vec2'],
        datasets: [
            // Raw metrics as ghost bars (lighter/context)
            createBarDataset('Whisper (Raw)', [whisperRaw, null], 'rgba(54, 162, 235, 0.3)', {
                borderColor: 'rgba(54, 162, 235, 0.5)',
                order: 2
            }),
            createBarDataset('Wav2Vec2 (Raw)', [null, wav2vec2Raw], 'rgba(255, 99, 132, 0.3)', {
                borderColor: 'rgba(255, 99, 132, 0.5)',
                order: 2
            }),
            // Normalized metrics as solid bars (primary)
            createBarDataset('Whisper (Normalized)', [whisperNorm, null], COLORS.whisper, {
                borderWidth: 2,
                order: 1
            }),
            createBarDataset('Wav2Vec2 (Normalized)', [null, wav2vec2Norm], COLORS.wav2vec2, {
                borderWidth: 2,
                order: 1
            })
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'Model Performance Comparison (Mean CER)',
                font: { size: 16 }
            },
            subtitle: {
                display: true,
                text: 'Solid bars = normalized (pedagogically relevant), Ghost bars = raw',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 12
                }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const label = context.dataset.label || '';
                        const value = context.parsed.y;
                        return `${label}: ${value.toFixed(2)}%`;
                    }
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Character Error Rate (%)'),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'ASR Model' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

function renderModelComparison() {
    const data = analysisData.model_comparison;
    if (!data || !data.raw || !data.raw.cer_analysis) return;

    // Render combined chart (raw as bars + normalized as line overlay)
    if (data.normalized && data.normalized.cer_analysis) {
        renderModelComparisonChartCombined('chartModelComparison', data.raw, data.normalized);
    } else {
        // Fallback: render raw only if normalized not available
        renderModelComparisonChart('chartModelComparison', data.raw, 'Model Comparison', 'modelComparison');
    }

    // Update statistics table with both RAW and NORMALIZED CER details
    const tbody = elements.tableModelStats;
    tbody.empty();
    const cerAnalysisRaw = data.raw.cer_analysis;
    const cerAnalysisNorm = data.normalized ? data.normalized.cer_analysis : null;

    const rows = [
        ['Sample Size', data.sample_size || data.raw.sample_size],
        ['', ''],
        ['<strong>RAW METRICS (True ASR Quality)</strong>', ''],
        ['', ''],
        ['<em>Whisper</em>', ''],
        ['Mean', cerAnalysisRaw.whisper.mean.toFixed(2) + '%'],
        ['Median', cerAnalysisRaw.whisper.median.toFixed(2) + '%'],
        ['Std Dev', cerAnalysisRaw.whisper.std.toFixed(2) + '%'],
        ['CV (Consistency)', (cerAnalysisRaw.whisper.cv || 0).toFixed(2) + '%'],
        ['', ''],
        ['<em>Wav2Vec2</em>', ''],
        ['Mean', cerAnalysisRaw.wav2vec2.mean.toFixed(2) + '%'],
        ['Median', cerAnalysisRaw.wav2vec2.median.toFixed(2) + '%'],
        ['Std Dev', cerAnalysisRaw.wav2vec2.std.toFixed(2) + '%'],
        ['CV (Consistency)', (cerAnalysisRaw.wav2vec2.cv || 0).toFixed(2) + '%'],
        ['', ''],
        ['<em>Comparison (Enhanced for MSc)</em>', ''],
        ['Mean Difference', cerAnalysisRaw.difference.mean.toFixed(2) + '%'],
        ['Percentage Point Difference', (cerAnalysisRaw.difference.percentage_point_difference || Math.abs(cerAnalysisRaw.difference.mean)).toFixed(2) + '%'],
        ['Relative Improvement', (cerAnalysisRaw.difference.relative_improvement || 0).toFixed(1) + '%'],
        ["Cohen's d", cerAnalysisRaw.effect_size.cohens_d.toFixed(3)],
        ['Interpretation', cerAnalysisRaw.effect_size.interpretation],
        ['Whisper Superiority Rate', (cerAnalysisRaw.whisper_superiority_rate || 0).toFixed(1) + '%']
    ];

    if (cerAnalysisNorm) {
        rows.push(
            ['', ''],
            ['<strong>NORMALIZED METRICS (App Usefulness)</strong>', ''],
            ['', ''],
            ['<em>Whisper</em>', ''],
            ['Mean', cerAnalysisNorm.whisper.mean.toFixed(2) + '%'],
            ['Median', cerAnalysisNorm.whisper.median.toFixed(2) + '%'],
            ['Std Dev', cerAnalysisNorm.whisper.std.toFixed(2) + '%'],
            ['CV (Consistency)', (cerAnalysisNorm.whisper.cv || 0).toFixed(2) + '%'],
            ['', ''],
            ['<em>Wav2Vec2</em>', ''],
            ['Mean', cerAnalysisNorm.wav2vec2.mean.toFixed(2) + '%'],
            ['Median', cerAnalysisNorm.wav2vec2.median.toFixed(2) + '%'],
            ['Std Dev', cerAnalysisNorm.wav2vec2.std.toFixed(2) + '%'],
            ['CV (Consistency)', (cerAnalysisNorm.wav2vec2.cv || 0).toFixed(2) + '%'],
            ['', ''],
            ['<em>Comparison (Enhanced for MSc)</em>', ''],
            ['Mean Difference', cerAnalysisNorm.difference.mean.toFixed(2) + '%'],
            ['Percentage Point Difference', (cerAnalysisNorm.difference.percentage_point_difference || Math.abs(cerAnalysisNorm.difference.mean)).toFixed(2) + '%'],
            ['Relative Improvement', (cerAnalysisNorm.difference.relative_improvement || 0).toFixed(1) + '%'],
            ["Cohen's d", cerAnalysisNorm.effect_size.cohens_d.toFixed(3)],
            ['Interpretation', cerAnalysisNorm.effect_size.interpretation],
            ['Whisper Superiority Rate', (cerAnalysisNorm.whisper_superiority_rate || 0).toFixed(1) + '%']
        );
    }

    rows.forEach((row) => {
        tbody.append(`<tr><td>${row[0]}</td><td>${row[1]}</td></tr>`);
    });

    // Show over-transcription indicator if detected
    const overTranscription = analysisData.over_transcription;
    if (overTranscription && overTranscription.count > 0) {
        tbody.append(`<tr><td colspan="2" class="text-muted"><hr></td></tr>`);
        tbody.append(`<tr><td><strong>WER Anomalies Detected</strong></td><td>${overTranscription.count} cases (${overTranscription.percentage.toFixed(1)}%)</td></tr>`);
        tbody.append(`<tr><td colspan="2" class="small text-muted">See "Over-Transcription Cases" section below</td></tr>`);
    }
}

function renderHybridAnalysis() {
    /**
     * Section 11: Comprehensive Hybrid System Analysis (MERGED)
     *
     * Combines theoretical best-case analysis with practical implications across critical CAPT areas.
     * Shows:
     * 1. Overall performance summary (4 metric cards)
     * 2. Critical areas breakdown (Digraphs, Difficult Words, Error Reduction)
     * 3. Comparison chart (Best Single vs Hybrid)
     * 4. Statistical analysis (Cohen's d effect size)
     * 5. Academic interpretation and deployment recommendation
     *
     * Replaces old renderHybridBestCase() and renderHybridPracticalImplications()
     */
    const bestCaseData = analysisData.hybrid_best_case;
    const implData = analysisData.hybrid_practical_implications;

    if (!bestCaseData || !implData || bestCaseData.error || implData.error) {
        console.warn('[renderHybridAnalysis] Missing hybrid analysis data');
        return;
    }

    console.log('[renderHybridAnalysis] Best-case data:', bestCaseData);
    console.log('[renderHybridAnalysis] Implications data:', implData);

    // =========================================================================
    // 1. OVERALL PERFORMANCE SUMMARY (4 cards)
    // =========================================================================

    // Best Single Model CER
    $('#hybridOverallBestModelCER').text(bestCaseData.best_individual_cer.toFixed(1) + '%');

    // Best model name
    const configDisplayNames = {
        'whisper_raw': 'Whisper (Raw)',
        'whisper_normalized': 'Whisper (Normalized)',
        'wav2vec2_raw': 'Wav2Vec2 (Raw)',
        'wav2vec2_normalized': 'Wav2Vec2 (Normalized)'
    };
    const bestModelName = configDisplayNames[bestCaseData.best_individual_config] || bestCaseData.best_individual_config;
    $('#hybridBestModelName').text(bestModelName);

    // Hybrid Best-Case CER
    $('#hybridOverallHybridCER').text(bestCaseData.hybrid_mean_cer.toFixed(1) + '%');

    // Absolute Improvement (percentage points)
    const absoluteImprovement = bestCaseData.best_individual_cer - bestCaseData.hybrid_mean_cer;
    $('#hybridOverallAbsoluteImprovement').text(absoluteImprovement.toFixed(1) + ' pp');

    // Relative Improvement (% error reduction)
    const relativeImprovement = (absoluteImprovement / bestCaseData.best_individual_cer) * 100;
    $('#hybridOverallRelativeImprovement').text(relativeImprovement.toFixed(1) + '%');

    // =========================================================================
    // 2. CRITICAL AREAS BREAKDOWN (3 cards)
    // =========================================================================

    // Welsh Digraphs
    if (implData.welsh_digraphs) {
        $('#hybridDigraphsBestCER').text(implData.welsh_digraphs.best_single_cer.toFixed(1) + '%');
        $('#hybridDigraphsHybridCER').text(implData.welsh_digraphs.hybrid_cer.toFixed(1) + '%');
        $('#hybridDigraphsImprovement').text(`${implData.welsh_digraphs.improvement_percentage.toFixed(1)}% improvement`);
    }

    // Difficult Words
    if (implData.difficult_words) {
        $('#hybridDifficultBestCER').text(implData.difficult_words.best_single_cer.toFixed(1) + '%');
        $('#hybridDifficultHybridCER').text(implData.difficult_words.hybrid_cer.toFixed(1) + '%');
        $('#hybridDifficultImprovement').text(`${implData.difficult_words.improvement_percentage.toFixed(1)}% improvement`);
    }

    // Error Reduction Potential
    if (implData.error_reduction) {
        $('#hybridErrorsTotal').text(implData.error_reduction.total_asr_errors);
        $('#hybridErrorsPreventable').text(implData.error_reduction.preventable_by_hybrid);
        $('#hybridErrorsReduction').text(`${implData.error_reduction.reduction_rate.toFixed(1)}% preventable`);
    }

    // =========================================================================
    // 3. COMPARISON CHART
    // =========================================================================
    renderHybridAnalysisChart('chartHybridAnalysis', implData);

    // =========================================================================
    // 4. STATISTICAL ANALYSIS
    // =========================================================================

    // Calculate Cohen's d effect size
    // d = (mean1 - mean2) / pooled_std
    // For simplicity, using overall improvement and standard deviations from model comparison
    const cohensD = calculateCohensD(
        bestCaseData.best_individual_cer,
        bestCaseData.hybrid_mean_cer,
        bestCaseData.best_individual_std || 10,  // Fallback if not available
        bestCaseData.hybrid_std || 8
    );

    $('#hybridEffectSize').text(cohensD.toFixed(2));

    // Interpret effect size
    let effectInterpretation = '';
    let effectBadgeClass = '';
    if (Math.abs(cohensD) < 0.2) {
        effectInterpretation = 'Negligible';
        effectBadgeClass = 'bg-secondary';
    } else if (Math.abs(cohensD) < 0.5) {
        effectInterpretation = 'Small';
        effectBadgeClass = 'bg-info';
    } else if (Math.abs(cohensD) < 0.8) {
        effectInterpretation = 'Medium';
        effectBadgeClass = 'bg-warning';
    } else {
        effectInterpretation = 'Large';
        effectBadgeClass = 'bg-success';
    }
    $('#hybridEffectSizeInterpretation').text(effectInterpretation).attr('class', `badge ${effectBadgeClass}`);

    // Sample characteristics
    $('#hybridSampleSize').text(bestCaseData.sample_size || '-');

    // Model preference (when performance differs - excluding ties)
    if (bestCaseData.model_preference) {
        $('#hybridModelPref').text(
            `Whisper ${bestCaseData.model_preference.whisper_pct.toFixed(1)}% ` +
            `(${bestCaseData.model_preference.whisper_count}), ` +
            `Wav2Vec2 ${bestCaseData.model_preference.wav2vec2_pct.toFixed(1)}% ` +
            `(${bestCaseData.model_preference.wav2vec2_count})`
        );

        // Show tie percentage
        $('#hybridTiesPct').text(
            `${bestCaseData.model_preference.tie_pct.toFixed(1)}% ` +
            `(${bestCaseData.model_preference.tie_count} entries with identical CER)`
        );
    } else {
        $('#hybridModelPref').text('-');
        $('#hybridTiesPct').text('-');
    }

    // Normalization preference - REMOVED (hybrid now only uses lenient scores)
    // The hybrid system now only considers lenient (normalized) metrics
    $('#hybridNormPref').text('N/A (Hybrid uses lenient only)');

    // =========================================================================
    // 5. ACADEMIC INTERPRETATION & DEPLOYMENT RECOMMENDATION
    // =========================================================================

    // Calculate average improvement across critical areas
    const avgImprovement = (
        (implData.overall?.improvement_percentage || 0) +
        (implData.welsh_digraphs?.improvement_percentage || 0) +
        (implData.difficult_words?.improvement_percentage || 0) +
        (implData.error_reduction?.reduction_rate || 0)
    ) / 4;

    // Academic interpretation
    let academicInterp = '';
    academicInterp += `This analysis examines the theoretical upper bound of ASR performance achievable through intelligent per-word model selection, `;
    academicInterp += `compared to deploying the best single fixed model (${bestModelName}). `;
    academicInterp += `The hybrid approach achieves a ${absoluteImprovement.toFixed(1)} percentage point absolute improvement `;
    academicInterp += `(${relativeImprovement.toFixed(1)}% relative error reduction), `;
    academicInterp += `with a Cohen's d effect size of ${cohensD.toFixed(2)} (${effectInterpretation.toLowerCase()}). `;

    if (implData.error_reduction && implData.error_reduction.reduction_rate >= 20) {
        academicInterp += `Notably, ${implData.error_reduction.reduction_rate.toFixed(1)}% of all ASR errors are preventable through hybrid selection. `;
    }

    if (implData.welsh_digraphs && implData.welsh_digraphs.improvement_percentage >= 15) {
        academicInterp += `Critical Welsh phonological features (digraphs) show particularly strong gains of ${implData.welsh_digraphs.improvement_percentage.toFixed(1)}%. `;
    }

    $('#hybridAcademicInterpretation').find('span.en').html(academicInterp);
    $('#hybridAcademicInterpretation').find('span.cy').html(academicInterp); // TODO: Welsh translation

    // Deployment recommendation
    let recommendation = '';
    if (avgImprovement >= 20 && cohensD >= 0.5) {
        recommendation = `<strong class="text-success">STRONGLY RECOMMENDED:</strong> Hybrid system demonstrates substantial performance gains (${avgImprovement.toFixed(1)}% avg improvement, d=${cohensD.toFixed(2)}). `;
        recommendation += `Intelligent model selection significantly reduces errors across all critical areas, especially Welsh-specific phonology. `;
        recommendation += `Deploy hybrid architecture for Welsh CAPT systems prioritizing accuracy.`;
    } else if (avgImprovement >= 10 && cohensD >= 0.3) {
        recommendation = `<strong class="text-info">RECOMMENDED:</strong> Hybrid system shows meaningful improvements (${avgImprovement.toFixed(1)}% avg improvement, d=${cohensD.toFixed(2)}). `;
        recommendation += `Notable benefits for difficult words and Welsh digraphs justify implementation complexity. `;
        recommendation += `Hybrid approach beneficial for pedagogically-focused CAPT applications.`;
    } else if (avgImprovement >= 5) {
        recommendation = `<strong class="text-warning">CONSIDER WITH CAUTION:</strong> Hybrid system yields modest gains (${avgImprovement.toFixed(1)}% avg improvement, d=${cohensD.toFixed(2)}). `;
        recommendation += `Evaluate whether improvements justify additional system complexity and computational cost. `;
        recommendation += `May be valuable for research contexts or high-stakes assessment scenarios.`;
    } else {
        recommendation = `<strong class="text-muted">NOT CRITICAL:</strong> Hybrid system shows marginal improvements (${avgImprovement.toFixed(1)}% avg improvement, d=${cohensD.toFixed(2)}). `;
        recommendation += `Best single model (${bestModelName}) appears sufficient for most CAPT applications. `;
        recommendation += `Hybrid complexity not justified by minimal performance gains.`;
    }

    $('#hybridDeploymentRecommendation').find('strong.en').html(recommendation);
    $('#hybridDeploymentRecommendation').find('strong.cy').html(recommendation); // TODO: Welsh translation
}

// Helper function to calculate Cohen's d
function calculateCohensD(mean1, mean2, std1, std2) {
    const pooledStd = Math.sqrt((std1 * std1 + std2 * std2) / 2);
    return (mean1 - mean2) / pooledStd;
}

function renderHybridAnalysisChart(canvasId, data) {
    /**
     * Render grouped bar chart comparing Best Single Model vs Hybrid Best-Case
     * across critical CAPT areas (Overall, Digraphs, Difficult Words)
     */
    const ctx = document.getElementById(canvasId);
    if (!ctx) {
        console.warn('[renderHybridAnalysisChart] Canvas not found:', canvasId);
        return;
    }

    const chartKey = 'hybridAnalysis';

    // Prepare data (CER metrics only - Error Reduction shown in cards above)
    const labels = [
        'Overall CER',
        'Welsh Digraphs',
        'Top 10 Difficult Words'
    ];

    const bestSingleData = [
        data.overall?.best_single_cer || 0,
        data.welsh_digraphs?.best_single_cer || 0,
        data.difficult_words?.best_single_cer || 0
    ];

    const hybridData = [
        data.overall?.hybrid_cer || 0,
        data.welsh_digraphs?.hybrid_cer || 0,
        data.difficult_words?.hybrid_cer || 0
    ];

    const chartData = {
        labels: labels,
        datasets: [
            createBarDataset('Best Single Model', bestSingleData, 'rgba(52, 152, 219, 0.7)', {
                borderColor: 'rgba(52, 152, 219, 1)',
                borderWidth: 2
            }),
            createBarDataset('Hybrid Best-Case', hybridData, 'rgba(46, 204, 113, 0.7)', {
                borderColor: 'rgba(46, 204, 113, 1)',
                borderWidth: 2
            })
        ]
    };

    const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            title: {
                display: true,
                text: 'Hybrid System Impact: CER Comparison Across Critical Areas',
                font: { size: 16, weight: 'bold' }
            },
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 15,
                    font: { size: 12 }
                }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const label = context.dataset.label || '';
                        const value = context.parsed.y;
                        return `${label}: ${value.toFixed(1)}% CER`;
                    }
                }
            }
        },
        scales: {
            x: {
                title: {
                    display: true,
                    text: 'Critical Area',
                    font: { size: 13, weight: 'bold' }
                },
                grid: { display: false }
            },
            y: {
                ...getStandardYAxisConfig('Character Error Rate (%)', 0),
                grid: { color: COLORS.gridLines }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', chartData, options, charts);
}

function renderErrorAttributionChart(canvasId, metricData, title, chartKey) {
    const whisperDist = metricData.whisper_distribution.percentages;
    const wav2vec2Dist = metricData.wav2vec2_distribution.percentages;

    const data = {
        labels: ['Whisper', 'Wav2Vec2'],
        datasets: [
            createBarDataset('Correct', [whisperDist.CORRECT, wav2vec2Dist.CORRECT], COLORS.correct),
            createBarDataset('ASR Error', [whisperDist.ASR_ERROR, wav2vec2Dist.ASR_ERROR], COLORS.asrError),
            createBarDataset('User Error', [whisperDist.USER_ERROR, wav2vec2Dist.USER_ERROR], COLORS.userError),
            createBarDataset('Ambiguous', [whisperDist.AMBIGUOUS, wav2vec2Dist.AMBIGUOUS], COLORS.ambiguous)
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            ...getStandardPluginsConfig(title),
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 10
                }
            },
            tooltip: {
                callbacks: {
                    label: (context) => `${context.dataset.label}: ${context.parsed.y.toFixed(1)}%`
                }
            }
        },
        scales: {
            x: {
                stacked: true,
                title: { display: true, text: 'ASR Model' }
            },
            y: {
                stacked: true,
                ...getStandardYAxisConfig('Percentage (%)', 0, 100)
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

// Comprehensive Error Analysis - Merged from renderErrorAttribution and renderErrorCosts
function renderComprehensiveErrorAnalysis() {
    const attributionData = analysisData.error_attribution;
    const costsData = analysisData.error_costs;

    console.log('[renderComprehensiveErrorAnalysis] Attribution data:', attributionData);
    console.log('[renderComprehensiveErrorAnalysis] Costs data:', costsData);

    if (!attributionData || !attributionData.raw) {
        console.warn('[renderComprehensiveErrorAnalysis] No attribution data available');
        return;
    }

    // === PART A: System Accuracy Overview ===
    const whisperRawCorrect = attributionData.raw.whisper_distribution.percentages.CORRECT;
    const wav2vec2RawCorrect = attributionData.raw.wav2vec2_distribution.percentages.CORRECT;
    const whisperNormCorrect = attributionData.normalized ? attributionData.normalized.whisper_distribution.percentages.CORRECT : 0;
    const wav2vec2NormCorrect = attributionData.normalized ? attributionData.normalized.wav2vec2_distribution.percentages.CORRECT : 0;

    // Populate metric cards
    $('#systemAccuracyWhisperRaw').text(`${whisperRawCorrect.toFixed(1)}%`);
    $('#systemAccuracyWav2Vec2Raw').text(`${wav2vec2RawCorrect.toFixed(1)}%`);
    $('#systemAccuracyWhisperNorm').text(`${whisperNormCorrect.toFixed(1)}%`);
    $('#systemAccuracyWav2Vec2Norm').text(`${wav2vec2NormCorrect.toFixed(1)}%`);

    // Generate CAPT suitability insight
    const avgRawAccuracy = (whisperRawCorrect + wav2vec2RawCorrect) / 2;
    const avgNormAccuracy = (whisperNormCorrect + wav2vec2NormCorrect) / 2;
    const improvement = avgNormAccuracy - avgRawAccuracy;

    let suitabilityHtml = '';
    if (avgNormAccuracy >= 70) {
        suitabilityHtml = `
            <p class="small mb-1 en"><strong>✓ Potentially Suitable for CAPT</strong></p>
            <ul class="small mb-0">
                <li class="en">Whisper achieves ${whisperNormCorrect.toFixed(1)}% correct feedback with normalization</li>
                <li class="en">Wav2Vec2 achieves ${wav2vec2NormCorrect.toFixed(1)}% correct feedback with normalization</li>
                <li class="en">Normalization improves feedback quality by ${improvement.toFixed(1)} percentage points</li>
                ${avgRawAccuracy < 50 ? '<li class="en text-warning">⚠ Raw metrics show only ~' + avgRawAccuracy.toFixed(0) + '% accuracy - lenient matching is ESSENTIAL</li>' : ''}
            </ul>
        `;
    } else if (avgNormAccuracy >= 50) {
        suitabilityHtml = `
            <p class="small mb-1 en"><strong>⚠ Marginal for CAPT - Requires Careful Design</strong></p>
            <ul class="small mb-0">
                <li class="en">Whisper: ${whisperNormCorrect.toFixed(1)}% correct (normalized)</li>
                <li class="en">Wav2Vec2: ${wav2vec2NormCorrect.toFixed(1)}% correct (normalized)</li>
                <li class="en text-warning">⚠ Accuracy of ~${avgNormAccuracy.toFixed(0)}% may frustrate learners</li>
                <li class="en">Normalization improves accuracy by ${improvement.toFixed(1)}pp, but still suboptimal</li>
            </ul>
        `;
    } else {
        suitabilityHtml = `
            <p class="small mb-1 en"><strong>✗ Not Suitable for CAPT</strong></p>
            <ul class="small mb-0">
                <li class="en">Whisper: Only ${whisperNormCorrect.toFixed(1)}% correct (normalized)</li>
                <li class="en">Wav2Vec2: Only ${wav2vec2NormCorrect.toFixed(1)}% correct (normalized)</li>
                <li class="en text-danger">✗ Accuracy of ~${avgNormAccuracy.toFixed(0)}% is too low for effective learning</li>
                <li class="en">Would produce excessive incorrect feedback, demotivating learners</li>
            </ul>
        `;
    }
    $('#captSuitabilityInsight').html(suitabilityHtml);

    // === PART B: Error Attribution Chart ===
    if (attributionData.normalized) {
        renderErrorAttributionStackedChart('chartErrorAttribution', attributionData.raw, attributionData.normalized);
    }

    // === PART C: Error Costs Chart ===
    if (costsData && costsData.raw && costsData.normalized) {
        renderErrorCostsStackedChart('chartErrorCosts', costsData.raw, costsData.normalized);
    }

    // === PART D: Safety Metrics ===
    if (costsData && costsData.raw) {
        $('#errorCostWhisperSafetyRaw').text(costsData.raw.whisper.pedagogical_safety_score.toFixed(1));
        $('#errorCostWav2Vec2SafetyRaw').text(costsData.raw.wav2vec2.pedagogical_safety_score.toFixed(1));

        if (costsData.normalized) {
            $('#errorCostWhisperSafetyNorm').text(costsData.normalized.whisper.pedagogical_safety_score.toFixed(1));
            $('#errorCostWav2Vec2SafetyNorm').text(costsData.normalized.wav2vec2.pedagogical_safety_score.toFixed(1));
        }

        if (costsData.raw.comparison) {
            $('#errorCostSaferModelRaw').text(costsData.raw.comparison.safer_model);
            $('#errorCostRecommendationRaw').text(costsData.raw.comparison.recommendation || '');
        }

        if (costsData.improvement) {
            $('#errorCostWhisperImprovement').html(`
                <strong>Whisper:</strong> ${costsData.improvement.whisper.interpretation || '-'}
            `);
            $('#errorCostWav2Vec2Improvement').html(`
                <strong>Wav2Vec2:</strong> ${costsData.improvement.wav2vec2.interpretation || '-'}
            `);
        }
    }

    // === PART E: Tables ===
    // Error Costs Table
    if (costsData && costsData.raw) {
        populateErrorCostsTable(costsData.raw);
    }

    // Error Attribution Table
    if (attributionData) {
        populateErrorAttributionTable(attributionData);
    }
}

// NEW: Error Attribution Stacked Chart
function renderErrorAttributionStackedChart(canvasId, rawData, normalizedData) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    const chartKey = canvasId;

    // Extract percentages for all categories
    const whisperRaw = rawData.whisper_distribution.percentages;
    const wav2vec2Raw = rawData.wav2vec2_distribution.percentages;
    const whisperNorm = normalizedData.whisper_distribution.percentages;
    const wav2vec2Norm = normalizedData.wav2vec2_distribution.percentages;

    const data = {
        labels: ['Whisper (Raw)', 'Whisper (Norm)', 'Wav2Vec2 (Raw)', 'Wav2Vec2 (Norm)'],
        datasets: [
            createBarDataset('CORRECT', [
                whisperRaw.CORRECT || 0,
                whisperNorm.CORRECT || 0,
                wav2vec2Raw.CORRECT || 0,
                wav2vec2Norm.CORRECT || 0
            ], '#2ecc71', { borderWidth: 0 }),
            createBarDataset('ASR_ERROR', [
                whisperRaw.ASR_ERROR || 0,
                whisperNorm.ASR_ERROR || 0,
                wav2vec2Raw.ASR_ERROR || 0,
                wav2vec2Norm.ASR_ERROR || 0
            ], '#f39c12', { borderWidth: 0 }),
            createBarDataset('USER_ERROR', [
                whisperRaw.USER_ERROR || 0,
                whisperNorm.USER_ERROR || 0,
                wav2vec2Raw.USER_ERROR || 0,
                wav2vec2Norm.USER_ERROR || 0
            ], '#f1c40f', { borderWidth: 0 }),
            createBarDataset('AMBIGUOUS', [
                whisperRaw.AMBIGUOUS || 0,
                whisperNorm.AMBIGUOUS || 0,
                wav2vec2Raw.AMBIGUOUS || 0,
                wav2vec2Norm.AMBIGUOUS || 0
            ], '#95a5a6', { borderWidth: 0 })
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'Error Attribution: Where Do Errors Come From?',
                font: { size: 16 }
            },
            subtitle: {
                display: true,
                text: 'Shows source of each error: ASR fault, user fault, or correct',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => `${context.dataset.label}: ${context.parsed.y.toFixed(1)}%`
                }
            }
        },
        scales: {
            x: {
                stacked: true,
                grid: { display: false }
            },
            y: {
                stacked: true,
                ...getStandardYAxisConfig('Percentage (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

// NEW: Error Costs Stacked Chart
function renderErrorCostsStackedChart(canvasId, rawData, normalizedData) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    const chartKey = canvasId;

    // Extract rates for all error cost categories
    const whisperRaw = rawData.whisper;
    const wav2vec2Raw = rawData.wav2vec2;
    const whisperNorm = normalizedData.whisper;
    const wav2vec2Norm = normalizedData.wav2vec2;

    // Check if False Acceptance has any non-zero values
    const faValues = [
        whisperRaw.false_acceptance?.rate || 0,
        whisperNorm.false_acceptance?.rate || 0,
        wav2vec2Raw.false_acceptance?.rate || 0,
        wav2vec2Norm.false_acceptance?.rate || 0
    ];
    const hasFalseAcceptance = faValues.some(v => v > 0);

    // Build datasets array - always include TP, TN, FR
    const datasets = [
        createBarDataset('True Positive', [
            whisperRaw.true_positive?.rate || 0,
            whisperNorm.true_positive?.rate || 0,
            wav2vec2Raw.true_positive?.rate || 0,
            wav2vec2Norm.true_positive?.rate || 0
        ], '#27ae60', { borderWidth: 0 }),
        createBarDataset('True Negative', [
            whisperRaw.true_negative?.rate || 0,
            whisperNorm.true_negative?.rate || 0,
            wav2vec2Raw.true_negative?.rate || 0,
            wav2vec2Norm.true_negative?.rate || 0
        ], '#95a5a6', { borderWidth: 0 }),
        createBarDataset('False Rejection', [
            whisperRaw.false_rejection?.rate || 0,
            whisperNorm.false_rejection?.rate || 0,
            wav2vec2Raw.false_rejection?.rate || 0,
            wav2vec2Norm.false_rejection?.rate || 0
        ], '#f39c12', { borderWidth: 0 })
    ];

    // Only add False Acceptance if there's data
    if (hasFalseAcceptance) {
        datasets.push(createBarDataset('False Acceptance', faValues, '#e74c3c', { borderWidth: 0 }));
    }

    const data = {
        labels: ['Whisper (Raw)', 'Whisper (Norm)', 'Wav2Vec2 (Raw)', 'Wav2Vec2 (Norm)'],
        datasets: datasets
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'Error Costs: What Is The Pedagogical Impact?',
                font: { size: 16 }
            },
            subtitle: {
                display: true,
                text: hasFalseAcceptance
                    ? 'Red = harmful (false acceptance), Orange = demotivating (false rejection)'
                    : 'Orange = demotivating (false rejection)',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => `${context.dataset.label}: ${context.parsed.y.toFixed(1)}%`
                }
            }
        },
        scales: {
            x: {
                stacked: true,
                grid: { display: false }
            },
            y: {
                stacked: true,
                ...getStandardYAxisConfig('Percentage (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

// NEW: Populate Error Attribution Table
function populateErrorAttributionTable(data) {
    const tbody = $('#tableErrorAttribution');
    if (!tbody.length) return;

    clearTable('tableErrorAttribution');

    if (!data.raw) {
        tbody.append('<tr><td colspan="6" class="text-muted text-center">No data available</td></tr>');
        return;
    }

    const categories = ['CORRECT', 'ASR_ERROR', 'USER_ERROR', 'AMBIGUOUS'];

    categories.forEach(category => {
        const whisperRaw = data.raw.whisper_distribution.percentages[category] || 0;
        const whisperNorm = data.normalized ? (data.normalized.whisper_distribution.percentages[category] || 0) : 0;
        const wav2vec2Raw = data.raw.wav2vec2_distribution.percentages[category] || 0;
        const wav2vec2Norm = data.normalized ? (data.normalized.wav2vec2_distribution.percentages[category] || 0) : 0;

        tbody.append(`
            <tr>
                <td><strong>${category}</strong></td>
                <td class="text-center">${whisperRaw.toFixed(1)}%</td>
                <td class="text-center">${whisperNorm.toFixed(1)}%</td>
                <td class="text-center">${wav2vec2Raw.toFixed(1)}%</td>
                <td class="text-center">${wav2vec2Norm.toFixed(1)}%</td>
            </tr>
        `);
    });
}

function renderSystemAccuracyChart(canvasId, rawData, normalizedData) {
    /**
     * Renders a stacked horizontal bar chart showing system accuracy
     * Binary: CORRECT (green) vs ERROR (red) for easy interpretation
     */
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    // Extract CORRECT percentages (system accuracy)
    const whisperRawCorrect = rawData.whisper_distribution.percentages.CORRECT;
    const whisperRawError = 100 - whisperRawCorrect;
    const wav2vec2RawCorrect = rawData.wav2vec2_distribution.percentages.CORRECT;
    const wav2vec2RawError = 100 - wav2vec2RawCorrect;

    const whisperNormCorrect = normalizedData.whisper_distribution.percentages.CORRECT;
    const whisperNormError = 100 - whisperNormCorrect;
    const wav2vec2NormCorrect = normalizedData.whisper_distribution.percentages.CORRECT;
    const wav2vec2NormError = 100 - wav2vec2NormCorrect;

    const chartKey = canvasId;

    const data = {
        labels: ['Whisper (Raw)', 'Whisper (Norm)', 'Wav2Vec2 (Raw)', 'Wav2Vec2 (Norm)'],
        datasets: [
            createBarDataset('Correct', 
                [whisperRawCorrect, whisperNormCorrect, wav2vec2RawCorrect, wav2vec2NormCorrect],
                '#2ecc71', { borderWidth: 0 }
            ),
            createBarDataset('Error',
                [whisperRawError, whisperNormError, wav2vec2RawError, wav2vec2NormError],
                '#e74c3c', { borderWidth: 0 }
            )
        ]
    };

    const options = {
        indexAxis: 'y', // Horizontal bars
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'System Accuracy: Does ASR Match Target?',
                font: { size: 16 }
            },
            subtitle: {
                display: true,
                text: 'Binary measure: Correct feedback (green) vs Incorrect feedback (red)',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => `${context.dataset.label}: ${context.parsed.x.toFixed(1)}%`
                }
            }
        },
        scales: {
            x: {
                stacked: true,
                ...getStandardYAxisConfig('Percentage (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            y: {
                stacked: true,
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

function renderWordDifficultyCombinedChart(canvasId, rawData, normalizedData, chartKey) {
    /**
     * Renders combined chart showing raw and normalized CER distribution with ghost bar approach
     * Ghost bars (light) = raw metrics (context), Solid bars = normalized (pedagogically relevant)
     */
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    const distRaw = rawData.distribution;
    const distNorm = normalizedData.distribution;

    const chartData = {
        labels: distRaw.buckets, // ["0-10%", "10-20%", ...]
        datasets: [
            // Raw metrics as ghost bars (lighter/context)
            createBarDataset('Whisper (Raw)', distRaw.whisper_counts, 'rgba(52, 162, 235, 0.3)', {
                borderColor: 'rgba(52, 162, 235, 0.5)',
                order: 2 // Render behind normalized bars
            }),
            createBarDataset('Wav2Vec2 (Raw)', distRaw.wav2vec2_counts, 'rgba(231, 76, 60, 0.3)', {
                borderColor: 'rgba(231, 76, 60, 0.5)',
                order: 2
            }),
            // Normalized metrics as solid bars (primary/pedagogically relevant)
            createBarDataset('Whisper (Normalized)', distNorm.whisper_counts, COLORS.whisper, {
                borderWidth: 2,
                order: 1 // Render in front
            }),
            createBarDataset('Wav2Vec2 (Normalized)', distNorm.wav2vec2_counts, COLORS.wav2vec2, {
                borderWidth: 2,
                order: 1
            })
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'CER Distribution Across Vocabulary',
                font: { size: 16 }
            },
            subtitle: {
                display: true,
                text: `Solid bars = normalized (pedagogically relevant), Ghost bars = raw | ${normalizedData.total_unique_words} unique words`,
                font: { size: 11 }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const count = context.parsed.y;
                        const total = normalizedData.total_unique_words;
                        const percentage = ((count / total) * 100).toFixed(1);
                        return `${context.dataset.label}: ${count} words (${percentage}%)`;
                    }
                }
            },
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 12
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Number of Words'),
                ticks: { stepSize: 1 },
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: {
                    display: true,
                    text: 'CER Range (%) - Lower = Better for CAPT'
                },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', chartData, options, charts);
}

function renderWordLength() {
    const data = analysisData.word_length;
    if (!data || !data.normalized) return;

    console.log('[renderWordLength] Data:', data);

    // Convert dict to array for easier iteration
    const categoryKeys = ['one_to_three', 'four_to_six', 'seven_to_nine', 'ten_plus'];
    const categories = categoryKeys.map(key => data.normalized[key]).filter(cat => cat !== null && cat !== undefined);

    if (categories.length === 0) return;

    // Populate summary cards
    const cardKeyMap = {
        'one_to_three': '13',
        'four_to_six': '46',
        'seven_to_nine': '79',
        'ten_plus': '10'
    };
    categoryKeys.forEach(catKey => {
        const cat = data.normalized[catKey];
        if (cat) {
            const cardKey = cardKeyMap[catKey];
            $(`#wordLengthCount${cardKey}`).text(cat.count.toLocaleString());
        }
    });

    // Render chart
    renderWordLengthChart('chartWordLength', data);

    // Populate table
    let tableHTML = '';
    categories.forEach((cat, index) => {
        let trend = '';
        let trendCy = '';

        // Determine trend based on comparison with previous category
        if (index > 0) {
            const prevCat = categories[index - 1];
            const avgCER = (cat.whisper_avg_cer + cat.wav2vec2_avg_cer) / 2;
            const prevAvgCER = (prevCat.whisper_avg_cer + prevCat.wav2vec2_avg_cer) / 2;

            if (avgCER > prevAvgCER + 5) {
                trend = '📈 Harder';
                trendCy = '📈 Anoddach';
            } else if (avgCER < prevAvgCER - 5) {
                trend = '📉 Easier';
                trendCy = '📉 Haws';
            } else {
                trend = '➡️ Similar';
                trendCy = '➡️ Tebyg';
            }
        } else {
            trend = '-';
            trendCy = '-';
        }

        tableHTML += `
            <tr>
                <td class="en">${cat.label} characters</td>
                <td class="cy">${cat.label} nod</td>
                <td class="text-center">${cat.count.toLocaleString()}</td>
                <td class="text-center">${cat.whisper_avg_cer.toFixed(2)}%</td>
                <td class="text-center">${cat.wav2vec2_avg_cer.toFixed(2)}%</td>
                <td class="text-center en">${trend}</td>
                <td class="text-center cy">${trendCy}</td>
            </tr>
        `;
    });

    $('#tableWordLength').html(tableHTML);

    // Generate insight
    let insight = '';
    const shortCER = (categories[0].whisper_avg_cer + categories[0].wav2vec2_avg_cer) / 2;
    const longCER = (categories[3].whisper_avg_cer + categories[3].wav2vec2_avg_cer) / 2;
    const cerDiff = longCER - shortCER;

    if (Math.abs(cerDiff) < 5) {
        insight = `Word length has minimal impact on ASR performance (${Math.abs(cerDiff).toFixed(1)}% difference between shortest and longest). ASR difficulty is driven more by phonological complexity than word length.`;
    } else if (cerDiff > 0) {
        insight = `Longer words show ${cerDiff.toFixed(1)}% higher average CER than short words. This suggests increased phonological complexity or more opportunities for character-level errors in longer Welsh words.`;
    } else {
        insight = `Surprisingly, longer words show ${Math.abs(cerDiff).toFixed(1)}% lower CER than short words. This may indicate that longer Welsh words contain more context cues for ASR models or that the short word category includes particularly challenging phonological patterns.`;
    }

    $('#wordLengthInsight').text(insight);
}

function renderWordLengthChart(canvasId, data) {
    // Convert dict to array
    const categoryKeys = ['one_to_three', 'four_to_six', 'seven_to_nine', 'ten_plus'];
    const categories = categoryKeys.map(key => data.normalized[key]).filter(cat => cat !== null && cat !== undefined);

    const labels = categories.map(c => c.label + ' chars');
    const whisperData = categories.map(c => c.whisper_avg_cer);
    const wav2vec2Data = categories.map(c => c.wav2vec2_avg_cer);

    const chartData = {
        labels: labels,
        datasets: [
            createBarDataset('Whisper CER', whisperData, 'rgba(0, 123, 255, 0.7)', {
                borderColor: 'rgba(0, 123, 255, 1)'
            }),
            createBarDataset('Wav2Vec2 CER', wav2vec2Data, 'rgba(255, 193, 7, 0.7)', {
                borderColor: 'rgba(255, 193, 7, 1)'
            })
        ]
    };

    const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                display: true,
                position: 'top'
            },
            title: {
                display: true,
                text: 'Average CER by Word Length (Normalized Metrics)',
                font: { size: 16 }
            },
            tooltip: {
                callbacks: {
                    label: (context) => `${context.dataset.label}: ${context.parsed.y.toFixed(2)}%`
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Character Error Rate (%)')
            },
            x: {
                title: {
                    display: true,
                    text: 'Word Length Category'
                }
            }
        }
    };

    createOrUpdateChart('wordLength', canvasId, 'bar', chartData, options, charts);
}

function renderWordDifficulty() {
    /**
     * Render combined CER distribution chart with ghost bar approach
     * Shows both raw and normalized metrics in one chart for space efficiency
     */
    const data = analysisData.word_difficulty;
    if (!data || !data.raw || !data.raw.distribution) return;
    if (!data.normalized || !data.normalized.distribution) return;

    // Render combined chart with ghost bars (raw) and solid bars (normalized)
    renderWordDifficultyCombinedChart('chartWordDifficulty', data.raw, data.normalized, 'wordDifficulty');

    // Populate distribution tables
    populateDistributionTable('wordDifficultyRaw', data.raw.distribution);
    populateDistributionTable('wordDifficultyNormalized', data.normalized.distribution);
}

function populateDistributionTable(tableId, distribution) {
    /**
     * Populate CER distribution table with bucket data
     * Shows how many words fall into each CER range
     */
    const tbody = $(`#${tableId} tbody`);
    tbody.empty();

    if (!distribution || !distribution.buckets) {
        tbody.append('<tr><td colspan="3" class="text-muted text-center">No data</td></tr>');
        return;
    }

    distribution.buckets.forEach((bucket, idx) => {
        const whisperCount = distribution.whisper_counts[idx] || 0;
        const wav2vec2Count = distribution.wav2vec2_counts[idx] || 0;

        tbody.append(`
            <tr>
                <td>${bucket}</td>
                <td>${whisperCount}</td>
                <td>${wav2vec2Count}</td>
            </tr>
        `);
    });
}

function renderTopDifficultWordsChart(canvasId, metricData, title, chartKey) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return; // Chart element doesn't exist yet

    const top10 = metricData.most_difficult.slice(0, 10);

    const data = {
        labels: top10.map(w => w.word),
        datasets: [
            createBarDataset('Whisper CER', top10.map(w => w.whisper_avg_cer), COLORS.whisper),
            createBarDataset('Wav2Vec2 CER', top10.map(w => w.wav2vec2_avg_cer), COLORS.wav2vec2)
        ]
    };

    const options = {
        indexAxis: 'y', // Horizontal bar chart
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: title,
                font: { size: 14 }
            },
            subtitle: {
                display: true,
                text: 'Ranked by average CER across all attempts',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const wordData = top10[context.dataIndex];
                        return [
                            `${context.dataset.label}: ${context.parsed.x.toFixed(2)}%`,
                            `Attempts: ${wordData.attempts}`,
                            `Length: ${wordData.word_length} chars`
                        ];
                    }
                }
            }
        },
        scales: {
            x: {
                beginAtZero: true,
                title: {
                    display: true,
                    text: 'Average CER (%)'
                }
            },
            y: {
                title: {
                    display: true,
                    text: 'Word'
                }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

function renderTopDifficultWordsChartCombined(canvasId, rawData, normalizedData) {
    /**
     * Renders chart focusing on TRUE phonological challenges.
     * Ranks by NORMALIZED CER (pedagogically relevant) and shows raw CER as context.
     * Only shows words with normalized CER > 5% (filters out punctuation-only issues).
     */
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    // Create a map for quick lookup between raw and normalized
    const rawMap = {};
    rawData.most_difficult.forEach(w => {
        rawMap[w.word] = w;
    });

    // Filter normalized words to only those with meaningful CER (> 5%)
    // These are TRUE pronunciation challenges, not just punctuation issues
    const MIN_NORMALIZED_CER = 5.0;
    const trulyChallenging = normalizedData.most_difficult.filter(w => {
        const avgNormCER = (w.whisper_avg_cer + w.wav2vec2_avg_cer) / 2;
        return avgNormCER > MIN_NORMALIZED_CER;
    });

    // Take top 10 by normalized CER
    const top10 = trulyChallenging.slice(0, 10);

    if (top10.length === 0) {
        // No truly challenging words - show message instead
        console.log('[renderTopDifficultWordsChartCombined] No words with normalized CER > 5%');
        return;
    }

    const words = top10.map(w => w.word);

    // Extract data for each dataset
    const whisperNormData = top10.map(w => w.whisper_avg_cer);
    const wav2vec2NormData = top10.map(w => w.wav2vec2_avg_cer);
    const whisperRawData = top10.map(w => rawMap[w.word]?.whisper_avg_cer || 0);
    const wav2vec2RawData = top10.map(w => rawMap[w.word]?.wav2vec2_avg_cer || 0);

    const chartKey = canvasId;

    const chartData = {
        labels: words,
        datasets: [
            // Background: Raw CER (ghost/context bars - lighter)
            createBarDataset('Whisper (Raw)', whisperRawData, 'rgba(54, 162, 235, 0.2)', {
                borderColor: 'rgba(54, 162, 235, 0.4)',
                order: 2
            }),
            createBarDataset('Wav2Vec2 (Raw)', wav2vec2RawData, 'rgba(255, 99, 132, 0.2)', {
                borderColor: 'rgba(255, 99, 132, 0.4)',
                order: 2
            }),
            // Foreground: Normalized CER (PRIMARY metric - solid)
            createBarDataset('Whisper (Normalized)', whisperNormData, COLORS.whisper, {
                borderWidth: 2,
                order: 1
            }),
            createBarDataset('Wav2Vec2 (Normalized)', wav2vec2NormData, COLORS.wav2vec2, {
                borderWidth: 2,
                order: 1
            })
        ]
    };

    const options = {
        indexAxis: 'y', // Horizontal bar chart
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            title: {
                display: true,
                text: 'True Phonological Challenges (Normalized CER > 5%)',
                font: { size: 16 }
            },
            subtitle: {
                display: true,
                text: 'Ranked by normalized CER - solid bars show pedagogically relevant errors, ghost bars show raw context',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 12
                }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const wordData = top10[context.dataIndex];
                        return [
                            `${context.dataset.label}: ${context.parsed.x.toFixed(1)}%`,
                            `Attempts: ${wordData.attempts || wordData.count}`,
                            `Length: ${wordData.word_length} chars`
                        ];
                    }
                }
            }
        },
        scales: {
            x: {
                ...getStandardYAxisConfig('Average CER (%)'),
                grid: { color: COLORS.gridLines }
            },
            y: {
                title: { display: true, text: 'Word' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', chartData, options, charts);
}

function renderModelAgreement() {
    const data = analysisData.inter_rater_reliability;
    if (!data || !data.agreement_rate) return;

    console.log('[renderModelAgreement] Data:', data);

    // Populate overall metrics
    const agreementPct = (data.agreement_rate * 100).toFixed(1);
    $('#agreementRate').text(agreementPct + '%');
    $('#agreementSampleSize').text(data.sample_size.toLocaleString());

    // Interpretation based on agreement rate
    let interpretation = '';
    if (data.agreement_rate >= 0.9) {
        interpretation = 'Excellent agreement (>90%). Both models show very similar behavior, suggesting consistent ASR performance.';
    } else if (data.agreement_rate >= 0.75) {
        interpretation = 'Good agreement (75-90%). Models generally agree but some systematic differences exist.';
    } else if (data.agreement_rate >= 0.6) {
        interpretation = 'Moderate agreement (60-75%). Significant disagreements suggest models may respond differently to certain phonological patterns.';
    } else {
        interpretation = 'Low agreement (<60%). Models show substantially different behavior, indicating complementary strengths/weaknesses.';
    }
    $('#agreementInterpretation').text(interpretation);

    // Render agreement pattern chart
    if (data.agreement_counts) {
        renderModelAgreementChart('chartModelAgreement', data);
    }

    // Populate agreement table
    if (data.agreement_counts) {
        const counts = data.agreement_counts;
        const total = data.sample_size;

        const rows = [
            {
                pattern: 'Both Correct',
                patternCy: 'Y Ddau Gywir',
                count: counts.both_correct,
                percentage: ((counts.both_correct / total) * 100).toFixed(1) + '%',
                implication: 'Best case - both models provide correct feedback',
                implicationCy: 'Achos gorau - y ddau fodel yn rhoi adborth cywir'
            },
            {
                pattern: 'Both Incorrect',
                patternCy: 'Y Ddau Anghywir',
                count: counts.both_incorrect,
                percentage: ((counts.both_incorrect / total) * 100).toFixed(1) + '%',
                implication: 'Worst case - systematic ASR failure on certain sounds',
                implicationCy: 'Achos gwaethaf - methiant ASR systematig ar seiniau penodol'
            },
            {
                pattern: 'Whisper Only Correct',
                patternCy: 'Whisper Yn Unig Gywir',
                count: counts.whisper_only_correct,
                percentage: ((counts.whisper_only_correct / total) * 100).toFixed(1) + '%',
                implication: 'Whisper outperforms Wav2Vec2 on these cases',
                implicationCy: 'Whisper yn rhagori ar Wav2Vec2 yn yr achosion hyn'
            },
            {
                pattern: 'Wav2Vec2 Only Correct',
                patternCy: 'Wav2Vec2 Yn Unig Gywir',
                count: counts.wav2vec2_only_correct,
                percentage: ((counts.wav2vec2_only_correct / total) * 100).toFixed(1) + '%',
                implication: 'Wav2Vec2 outperforms Whisper on these cases',
                implicationCy: 'Wav2Vec2 yn rhagori ar Whisper yn yr achosion hyn'
            }
        ];

        let tableHTML = '';
        rows.forEach(row => {
            tableHTML += `
                <tr>
                    <td class="en">${row.pattern}</td>
                    <td class="cy">${row.patternCy}</td>
                    <td class="text-center">${row.count.toLocaleString()}</td>
                    <td class="text-center">${row.percentage}</td>
                    <td class="en">${row.implication}</td>
                    <td class="cy">${row.implicationCy}</td>
                </tr>
            `;
        });

        $('#tableModelAgreement').html(tableHTML);
    }

    // CER Breakdown (if available)
    if (data.cer_breakdown) {
        const cerBreakdown = data.cer_breakdown;
        const cerDiff = cerBreakdown.cer_difference;

        let cerInsight = `When models agree, average CER is ${cerBreakdown.avg_cer_when_agree}%. `;
        cerInsight += `When they disagree, average CER is ${cerBreakdown.avg_cer_when_disagree}%. `;

        if (Math.abs(cerDiff) < 5) {
            cerInsight += 'Agreement has minimal correlation with error rates - suggests models disagree on easy AND difficult words.';
        } else if (cerDiff > 0) {
            cerInsight += `Disagreement correlates with ${Math.abs(cerDiff).toFixed(1)}% higher CER - models disagree more on difficult words.`;
        }

        // Add to interpretation
        interpretation += ' ' + cerInsight;
        $('#agreementInterpretation').text(interpretation);
    }

    // CAPT Suitability Insight
    let captInsight = '';
    const bothCorrectPct = (data.agreement_counts.both_correct / data.sample_size) * 100;
    const bothIncorrectPct = (data.agreement_counts.both_incorrect / data.sample_size) * 100;

    if (bothCorrectPct > 70) {
        captInsight = `High "both correct" rate (${bothCorrectPct.toFixed(1)}%) suggests strong CAPT suitability - most learners will receive correct feedback regardless of model choice.`;
    } else if (bothIncorrectPct > 30) {
        captInsight = `High "both incorrect" rate (${bothIncorrectPct.toFixed(1)}%) indicates systematic ASR challenges with certain Welsh sounds - may require hybrid approach or human verification.`;
    } else {
        captInsight = `Models show complementary strengths. Consider hybrid architecture that selects the better-performing model per word for improved CAPT feedback.`;
    }

    // Add CER context if available
    if (data.cer_breakdown) {
        const bothCorrectCER = data.cer_breakdown.avg_cer_both_correct;
        const bothIncorrectCER = data.cer_breakdown.avg_cer_both_incorrect;
        captInsight += ` Average CER when both correct: ${bothCorrectCER}%, when both incorrect: ${bothIncorrectCER}%.`;
    }

    $('#agreementCAPTInsight').text(captInsight);
}

function renderModelAgreementChart(canvasId, data) {
    const counts = data.agreement_counts;

    const chartData = {
        labels: ['Both Correct', 'Both Incorrect', 'Whisper Only', 'Wav2Vec2 Only'],
        datasets: [{
            label: 'Agreement Pattern',
            data: [
                counts.both_correct,
                counts.both_incorrect,
                counts.whisper_only_correct,
                counts.wav2vec2_only_correct
            ],
            backgroundColor: [
                'rgba(40, 167, 69, 0.7)',   // Green - Both correct
                'rgba(220, 53, 69, 0.7)',   // Red - Both incorrect
                'rgba(0, 123, 255, 0.7)',   // Blue - Whisper only
                'rgba(255, 193, 7, 0.7)'    // Yellow/Orange - Wav2Vec2 only
            ],
            borderColor: [
                'rgba(40, 167, 69, 1)',
                'rgba(220, 53, 69, 1)',
                'rgba(0, 123, 255, 1)',
                'rgba(255, 193, 7, 1)'
            ],
            borderWidth: 1
        }]
    };

    const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 10
                }
            },
            title: {
                display: true,
                text: 'Model Agreement Patterns',
                font: { size: 16 }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const total = data.sample_size;
                        const value = context.parsed.y;
                        const percentage = ((value / total) * 100).toFixed(1);
                        return `${context.label}: ${value.toLocaleString()} (${percentage}%)`;
                    }
                }
            }
        },
        scales: {
            x: {
                title: {
                    display: true,
                    text: 'Agreement Pattern'
                }
            },
            y: {
                ...getStandardYAxisConfig('Number of Cases')
            }
        }
    };

    createOrUpdateChart('modelAgreement', canvasId, 'bar', chartData, options, charts);
}

function renderTopDifficultWords() {
    const data = analysisData.word_difficulty;
    if (!data || !data.raw || !data.raw.most_difficult || data.raw.most_difficult.length === 0) return;

    // Render combined chart (raw + normalized)
    if (data.normalized && data.normalized.most_difficult && data.normalized.most_difficult.length > 0) {
        renderTopDifficultWordsChartCombined('chartTopDifficultWords', data.raw, data.normalized);
    } else {
        // Fallback: render raw only
        renderTopDifficultWordsChart('chartTopDifficultWords', data.raw, 'Most Challenging Words', 'topDifficultWords');
    }

    // Populate drill-down table with top 10 challenging words (pass both raw and normalized)
    if (data.normalized && data.normalized.most_difficult && data.normalized.most_difficult.length > 0) {
        populateChallengingWordsTable(data.raw.most_difficult, data.normalized.most_difficult);
    } else {
        populateChallengingWordsTable(data.raw.most_difficult, null);
    }
}

function populateChallengingWordsTable(rawWords, normalizedWords) {
    const tbody = $('#tableChallengingWords');
    if (!tbody.length) return;

    clearTable('tableChallengingWords');

    // UPDATED: Rank by normalized CER (pedagogically relevant), filter to CER > 5%
    // This matches the chart's filtering logic for consistency
    let top10;
    let rawMap = {};

    if (normalizedWords && normalizedWords.length > 0) {
        // Create map for raw lookup
        if (rawWords) {
            rawWords.forEach(w => {
                rawMap[w.word] = w;
            });
        }

        // Filter normalized words to only those with meaningful CER (> 5%)
        // These are TRUE pronunciation challenges, not just punctuation issues
        const MIN_NORMALIZED_CER = 5.0;
        const trulyChallenging = normalizedWords.filter(w => {
            const avgNormCER = (w.whisper_avg_cer + w.wav2vec2_avg_cer) / 2;
            return avgNormCER > MIN_NORMALIZED_CER;
        });

        // Take top 10 by normalized CER (already sorted by API)
        top10 = trulyChallenging.slice(0, 10);

        if (top10.length === 0) {
            tbody.append('<tr><td colspan="6" class="text-muted text-center">No words with normalized CER > 5% (no true phonological challenges)</td></tr>');
            return;
        }
    } else if (rawWords && rawWords.length > 0) {
        // Fallback: use raw words if normalized not available
        top10 = rawWords.slice(0, 10);
    } else {
        tbody.append('<tr><td colspan="6" class="text-muted text-center">No data available</td></tr>');
        return;
    }

    // Populate table rows
    top10.forEach((word) => {
        const attempts = word.count || word.attempts || 0;

        // Normalized CER values (PRIMARY metric)
        const whisperNorm = word.whisper_avg_cer || 0;
        const wav2vec2Norm = word.wav2vec2_avg_cer || 0;

        // Raw CER values (for context)
        const rawWord = rawMap[word.word];
        const whisperRaw = rawWord ? (rawWord.whisper_avg_cer || 0) : whisperNorm;
        const wav2vec2Raw = rawWord ? (rawWord.wav2vec2_avg_cer || 0) : wav2vec2Norm;

        // Calculate improvement percentage (average across both models)
        const avgRaw = (whisperRaw + wav2vec2Raw) / 2;
        const avgNorm = (whisperNorm + wav2vec2Norm) / 2;
        const improvement = avgRaw - avgNorm;
        const improvementPct = avgRaw > 0 ? ((improvement / avgRaw) * 100) : 0;

        // Categorize word based on normalized CER
        let category, categoryClass;
        if (avgNorm < 5) {
            category = 'Punctuation';
            categoryClass = 'badge bg-info';
        } else if (avgNorm < 15) {
            category = 'Moderate';
            categoryClass = 'badge bg-warning';
        } else {
            category = 'Phonological';
            categoryClass = 'badge bg-danger';
        }

        tbody.append(`
            <tr>
                <td><strong>${word.word}</strong></td>
                <td class="text-center">${attempts}</td>
                <td class="text-center small">${whisperRaw.toFixed(1)}% / <strong>${whisperNorm.toFixed(1)}%</strong></td>
                <td class="text-center small">${wav2vec2Raw.toFixed(1)}% / <strong>${wav2vec2Norm.toFixed(1)}%</strong></td>
                <td class="text-center text-success"><strong>${improvementPct.toFixed(0)}%</strong></td>
                <td class="text-center"><span class="${categoryClass}">${category}</span></td>
            </tr>
        `);
    });

    // Reinitialize language toggle for dynamically added content
    if (typeof initLanguageToggle === 'function') {
        initLanguageToggle();
    }
}

// Character Error Analysis Rendering

function renderCharacterErrors() {
    const data = analysisData.character_errors;
    if (!data || !data.whisper || !data.wav2vec2) return;

    // Render overall character error analysis only
    // Render confusion matrices for both models
    createConfusionMatrixTable(
        data.whisper.confusion_matrix,
        'tableConfusionWhisperOverall',
        'Whisper'
    );
    createConfusionMatrixTable(
        data.wav2vec2.confusion_matrix,
        'tableConfusionWav2Vec2Overall',
        'Wav2Vec2'
    );

    // Render combined error rate chart comparing both models
    createCombinedErrorRateChart(
        data.whisper.per_character,
        data.wav2vec2.per_character,
        'chartCharErrorsCombined',
        'Top 15 Most Problematic Sounds for Welsh CAPT'
    );
}

function createConfusionMatrixTable(confusionMatrix, tableId, modelName) {
    const tbody = $(`#${tableId}`);
    if (!tbody.length) return;

    tbody.empty();

    if (!confusionMatrix || confusionMatrix.length === 0) {
        tbody.append(`
            <tr>
                <td colspan="3" class="empty-state">
                    No errors detected for ${modelName}
                </td>
            </tr>
        `);
        return;
    }

    // Find max count for heatmap scaling
    const maxCount = Math.max(...confusionMatrix.map(item => item.count));

    // Sort by count descending
    const sorted = [...confusionMatrix].sort((a, b) => b.count - a.count);

    // Show all entries (no limit)
    sorted.forEach(item => {
        const heatmapClass = getHeatmapClass(item.count, maxCount);
        tbody.append(`
            <tr>
                <td>${item.expected}</td>
                <td>${item.actual}</td>
                <td class="heatmap-cell ${heatmapClass}">${item.count}</td>
            </tr>
        `);
    });
}

function createErrorRateChart(perCharacterData, chartId, chartTitle, category) {
    const ctx = document.getElementById(chartId);
    if (!ctx) return;

    if (!perCharacterData || Object.keys(perCharacterData).length === 0) {
        // Show empty state
        $(ctx).parent().html(`
            <div class="empty-state">
                No character data available
            </div>
        `);
        return;
    }

    // Convert object to sorted array
    const chartData = Object.entries(perCharacterData)
        .map(([char, stats]) => ({
            char: char,
            errorRate: stats.error_rate,
            totalOccurrences: stats.total_occurrences,
            errorCount: stats.error_count
        }))
        .sort((a, b) => b.errorRate - a.errorRate)
        .slice(0, 15); // Top 15 for readability

    const chartKey = `charErrors_${category}`;

    const data = {
        labels: chartData.map(d => d.char),
        datasets: [
            createBarDataset('Error Rate', chartData.map(d => d.errorRate), COLORS.asrError)
        ]
    };

    const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            title: {
                display: true,
                text: chartTitle,
                font: { size: 14 }
            },
            subtitle: {
                display: true,
                text: 'Top 15 characters by error rate',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const item = chartData[context.dataIndex];
                        return [
                            `Error Rate: ${item.errorRate.toFixed(1)}%`,
                            `Errors: ${item.errorCount}`,
                            `Total: ${item.totalOccurrences}`
                        ];
                    }
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Error Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'Character/Digraph' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, chartId, 'bar', data, options, charts);
}

function createCombinedErrorRateChart(whisperData, wav2vec2Data, chartId, chartTitle) {
    const ctx = document.getElementById(chartId);
    if (!ctx) return;

    // Check if both datasets are empty
    const whisperEmpty = !whisperData || Object.keys(whisperData).length === 0;
    const wav2vec2Empty = !wav2vec2Data || Object.keys(wav2vec2Data).length === 0;

    if (whisperEmpty && wav2vec2Empty) {
        $(ctx).parent().html(`
            <div class="empty-state">
                No character data available
            </div>
        `);
        return;
    }

    // Merge all unique characters from both datasets
    const allChars = new Set([
        ...Object.keys(whisperData || {}),
        ...Object.keys(wav2vec2Data || {})
    ]);

    // Create combined data with both models' stats for each character
    const combinedData = Array.from(allChars).map(char => {
        const whisperStats = whisperData?.[char] || { error_rate: 0, error_count: 0, total_occurrences: 0 };
        const wav2vec2Stats = wav2vec2Data?.[char] || { error_rate: 0, error_count: 0, total_occurrences: 0 };

        // Calculate max error rate for sorting
        const maxErrorRate = Math.max(whisperStats.error_rate, wav2vec2Stats.error_rate);

        return {
            char: char,
            whisper: whisperStats,
            wav2vec2: wav2vec2Stats,
            maxErrorRate: maxErrorRate
        };
    });

    // Sort by maximum error rate and take top 15
    const topChars = combinedData
        .sort((a, b) => b.maxErrorRate - a.maxErrorRate)
        .slice(0, 15);

    const chartKey = `charErrors_combined`;

    const data = {
        labels: topChars.map(d => d.char),
        datasets: [
            createBarDataset('Whisper', topChars.map(d => d.whisper.error_rate), COLORS.whisper, {
                charData: topChars.map(d => d.whisper) // Store full data for tooltips
            }),
            createBarDataset('Wav2Vec2', topChars.map(d => d.wav2vec2.error_rate), COLORS.wav2vec2, {
                charData: topChars.map(d => d.wav2vec2) // Store full data for tooltips
            })
        ]
    };

    const options = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            title: {
                display: true,
                text: chartTitle,
                font: { size: 14 }
            },
            subtitle: {
                display: true,
                text: 'Top 15 characters by highest error rate (either model)',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const dataset = context.dataset;
                        const dataIndex = context.dataIndex;
                        const stats = dataset.charData[dataIndex];

                        return [
                            `${dataset.label}:`,
                            `  Error Rate: ${stats.error_rate.toFixed(1)}%`,
                            `  Errors: ${stats.error_count}`,
                            `  Total: ${stats.total_occurrences}`
                        ];
                    }
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Error Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'Character/Digraph' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, chartId, 'bar', data, options, charts);
}

function getHeatmapClass(count, maxCount) {
    if (maxCount === 0) return '';

    const percentage = (count / maxCount) * 100;

    if (percentage >= 90) return 'heatmap-10';
    if (percentage >= 80) return 'heatmap-9';
    if (percentage >= 70) return 'heatmap-8';
    if (percentage >= 60) return 'heatmap-7';
    if (percentage >= 50) return 'heatmap-6';
    if (percentage >= 40) return 'heatmap-5';
    if (percentage >= 30) return 'heatmap-4';
    if (percentage >= 20) return 'heatmap-3';
    if (percentage >= 10) return 'heatmap-2';
    return 'heatmap-1';
}


// =============================================================================
// NEW RENDER FUNCTIONS FOR MSC THESIS ANALYTICS
// =============================================================================

function renderExecutiveSummary() {
    const data = analysisData.executive_summary;
    if (!data) return;

    // Populate metric cards
    $('#summaryTotalRecordings').text(data.total_recordings ? data.total_recordings.toLocaleString() : '-');
    $('#summaryTotalWords').text(data.total_words ? data.total_words.toLocaleString() : '-');

    // Recordings per word statistics
    $('#summaryAvgRecordingsPerWord').text(data.avg_recordings_per_word ? data.avg_recordings_per_word.toFixed(2) : '-');
    $('#summaryMinRecordingsPerWord').text(data.min_recordings_per_word !== undefined ? data.min_recordings_per_word : '-');
    $('#summaryMaxRecordingsPerWord').text(data.max_recordings_per_word !== undefined ? data.max_recordings_per_word : '-');

    // Raw CER values
    $('#summaryWhisperCER').text(data.raw_metrics ? `${data.raw_metrics.whisper_cer_mean.toFixed(1)}%` : '-');
    $('#summaryWav2Vec2CER').text(data.raw_metrics ? `${data.raw_metrics.wav2vec2_cer_mean.toFixed(1)}%` : '-');
    
    // Normalized CER values (NEW)
    $('#summaryWhisperCERNorm').text(data.normalized_metrics ? `${data.normalized_metrics.whisper_cer_mean.toFixed(1)}%` : '-');
    $('#summaryWav2Vec2CERNorm').text(data.normalized_metrics ? `${data.normalized_metrics.wav2vec2_cer_mean.toFixed(1)}%` : '-');

    // Winner and key finding
    $('#summaryWinner').text(data.overall_winner || '-');
    $('#summaryWinnerCy').text(data.overall_winner || '-');
    $('#summaryKeyFinding').text(data.key_finding || '-');

    // Raw metrics advantage
    if (data.raw_metrics) {
        $('#summaryRawPctDiff').text(`${data.raw_metrics.percentage_point_difference.toFixed(1)}%`);
        $('#summaryRawRelImprovement').text(`${data.raw_metrics.relative_improvement.toFixed(1)}%`);
    }

    // Normalized metrics advantage
    if (data.normalized_metrics) {
        $('#summaryNormPctDiff').text(`${data.normalized_metrics.percentage_point_difference.toFixed(1)}%`);
        $('#summaryNormRelImprovement').text(`${data.normalized_metrics.relative_improvement.toFixed(1)}%`);
    }
}

function renderPhonologicalAnalysis() {
    /**
     * Master function for comprehensive Phonological Error Analysis (merged Sections 2 & 5)
     * Shows CAPT-relevant phonological patterns across:
     * - Overview cards with key metrics
     * - Category-level error rates (vowels, consonants, digraphs, overall)
     * - Welsh digraph analysis (ll, ch, dd, etc.)
     * - Confusion matrices (4 tabs: overall, vowels, consonants, digraphs)
     * - Top 15 problematic characters
     * - Normalization impact table
     */
    console.log('[renderPhonologicalAnalysis] Full analysisData:', analysisData);
    const data = analysisData.linguistic_patterns;

    console.log('[renderPhonologicalAnalysis] Linguistic patterns data:', data);
    console.log('[renderPhonologicalAnalysis] Data keys:', data ? Object.keys(data) : 'null');

    if (!data) {
        console.warn('[renderPhonologicalAnalysis] No linguistic_patterns data available');
        console.warn('[renderPhonologicalAnalysis] Available analysis keys:', Object.keys(analysisData));
        return;
    }

    if (!data.raw || !data.raw.whisper || !data.raw.wav2vec2) {
        console.warn('[renderPhonologicalAnalysis] Missing raw data structure');
        console.log('[renderPhonologicalAnalysis] Data keys:', Object.keys(data));
        console.log('[renderPhonologicalAnalysis] data.raw:', data.raw);
        return;
    }

    if (!data.normalized || !data.normalized.whisper || !data.normalized.wav2vec2) {
        console.warn('[renderPhonologicalAnalysis] Missing normalized data structure');
        console.log('[renderPhonologicalAnalysis] data.normalized:', data.normalized);
        return;
    }

    console.log('[renderPhonologicalAnalysis] Data validated. Rendering all parts...');
    console.log('[renderPhonologicalAnalysis] Raw keys:', Object.keys(data.raw));
    console.log('[renderPhonologicalAnalysis] Normalized keys:', Object.keys(data.normalized));

    // Part A: Populate overview cards
    try {
        console.log('[renderPhonologicalAnalysis] Calling populatePhonologicalOverview...');
        populatePhonologicalOverview(data);
    } catch (e) {
        console.error('[renderPhonologicalAnalysis] Error in populatePhonologicalOverview:', e);
    }

    // Part B: Category-level comparison charts
    // Combined overview chart
    try {
        console.log('[renderPhonologicalAnalysis] Calling renderCategoryComparisonChart...');
        renderCategoryComparisonChart(data);
    } catch (e) {
        console.error('[renderPhonologicalAnalysis] Error in renderCategoryComparisonChart:', e);
    }

    // Part C: Welsh digraph analysis chart
    try {
        console.log('[renderPhonologicalAnalysis] Calling renderDigraphAnalysisChart...');
        renderDigraphAnalysisChart(data);

        // Generate digraph CAPT insight dynamically
        if (data.normalized && data.normalized.whisper && data.normalized.wav2vec2) {
            const digraphWhisper = data.normalized.whisper.digraphs?.error_rate || 0;
            const digraphWav2vec2 = data.normalized.wav2vec2.digraphs?.error_rate || 0;
            const avgDigraphCER = (digraphWhisper + digraphWav2vec2) / 2;

            let insight = '';
            if (avgDigraphCER < 10) {
                insight = `Average CER: ${avgDigraphCER.toFixed(1)}%. <strong class="text-success">Suitable for CAPT</strong> - ASR can reliably recognize Welsh digraphs.`;
            } else if (avgDigraphCER < 25) {
                insight = `Average CER: ${avgDigraphCER.toFixed(1)}%. <strong class="text-warning">Marginal for CAPT</strong> - ASR shows moderate difficulty with Welsh digraphs. May require human verification.`;
            } else {
                insight = `Average CER: ${avgDigraphCER.toFixed(1)}%. <strong class="text-danger">Unsuitable for CAPT</strong> - High error rates indicate ASR cannot reliably recognize these phonemes.`;
            }

            $('#digraphInsightText').html(insight);
            $('#digraphInsightTextCy').html(insight); // TODO: Welsh translation
        }
    } catch (e) {
        console.error('[renderPhonologicalAnalysis] Error in renderDigraphAnalysisChart:', e);
    }

    // Part D: Populate confusion matrix tabs
    try {
        console.log('[renderPhonologicalAnalysis] Calling populateConfusionMatrices...');
        populateConfusionMatrices(data);
    } catch (e) {
        console.error('[renderPhonologicalAnalysis] Error in populateConfusionMatrices:', e);
    }

    // Part E: Top 15 problematic characters (uses character_errors data)
    const charErrorData = analysisData.character_errors;
    console.log('[renderPhonologicalAnalysis] Character errors data:', charErrorData);
    console.log('[renderPhonologicalAnalysis] Character errors keys:', charErrorData ? Object.keys(charErrorData) : 'null');

    if (charErrorData && charErrorData.whisper && charErrorData.wav2vec2) {
        console.log('[renderPhonologicalAnalysis] Creating combined error rate chart...');
        console.log('[renderPhonologicalAnalysis] Whisper per_character:', charErrorData.whisper.per_character);
        createCombinedErrorRateChart(
            charErrorData.whisper.per_character,
            charErrorData.wav2vec2.per_character,
            'chartCharErrorsCombined',
            'Top 15 Most Problematic Sounds for Welsh CAPT'
        );

        // Generate confusion matrix CAPT insight dynamically
        let confusionInsight = 'Systematic confusions (e.g., \'e\' → \'a\') may be acceptable for beginners, but random errors or digraph confusions (e.g., \'ll\' → \'l\') are pedagogically harmful.';

        // Check if we have digraph confusion data
        if (charErrorData.digraphs && (charErrorData.digraphs.whisper || charErrorData.digraphs.wav2vec2)) {
            const whisperDigraphConfusions = charErrorData.digraphs.whisper?.confusion_matrix ?
                Object.keys(charErrorData.digraphs.whisper.confusion_matrix).length : 0;
            const wav2vec2DigraphConfusions = charErrorData.digraphs.wav2vec2?.confusion_matrix ?
                Object.keys(charErrorData.digraphs.wav2vec2.confusion_matrix).length : 0;
            const totalDigraphConfusions = whisperDigraphConfusions + wav2vec2DigraphConfusions;

            if (totalDigraphConfusions > 0) {
                confusionInsight = `<strong class="text-warning">Warning:</strong> ${totalDigraphConfusions} digraph confusion patterns detected. Digraph confusions (e.g., 'll' → 'l') are pedagogically critical for Welsh, indicating ASR difficulty with unique Welsh phonemes.`;
            }
        }

        $('#confusionInsightText').html(confusionInsight);
        $('#confusionInsightTextCy').html(confusionInsight); // TODO: Welsh translation
    } else {
        console.warn('[renderPhonologicalAnalysis] Character error data incomplete');
    }

    // Part F: Normalization impact table
    populateNormalizationImpactTable(data);
}

function populatePhonologicalOverview(data) {
    /**
     * Part A: Populate overview metric cards
     * - Hardest category (vowels/consonants/digraphs)
     * - Hardest Welsh digraph
     * - CAPT suitability verdict
     */
    console.log('[populatePhonologicalOverview] Starting...');
    console.log('[populatePhonologicalOverview] data.normalized:', data.normalized);

    const categories = ['vowels', 'consonants', 'digraphs'];
    const categoryLabelsEn = {
        vowels: 'Vowels',
        consonants: 'Consonants',
        digraphs: 'Digraphs'
    };
    const categoryLabelsCy = {
        vowels: 'Llafariaid',
        consonants: 'Cytseiniaid',
        digraphs: 'Deuseiniau'
    };

    // Find hardest category (average of both models, normalized)
    let hardestCategory = null;
    let highestCategoryRate = 0;

    categories.forEach(cat => {
        const whisperRate = data.normalized.whisper[cat]?.error_rate || 0;
        const wav2vec2Rate = data.normalized.wav2vec2[cat]?.error_rate || 0;
        const avgRate = (whisperRate + wav2vec2Rate) / 2;

        console.log(`[populatePhonologicalOverview] ${cat}: Whisper ${whisperRate}%, Wav2Vec2 ${wav2vec2Rate}%, Avg ${avgRate}%`);

        if (avgRate > highestCategoryRate) {
            highestCategoryRate = avgRate;
            hardestCategory = cat;
        }
    });

    console.log(`[populatePhonologicalOverview] Hardest category: ${hardestCategory} (${highestCategoryRate}%)`);


    // Find hardest Welsh digraph
    const welshDigraphs = ['ll', 'ch', 'dd', 'ff', 'ng', 'rh', 'ph', 'th'];
    let hardestDigraph = null;
    let highestDigraphRate = 0;

    // Use character_errors data which has pre-calculated per-character error rates
    const charErrorData = analysisData.character_errors;
    console.log('[populatePhonologicalOverview] Character errors available:', !!charErrorData);

    if (charErrorData && charErrorData.whisper && charErrorData.wav2vec2) {
        welshDigraphs.forEach(digraph => {
            const whisperStats = charErrorData.whisper.per_character[digraph];
            const wav2vec2Stats = charErrorData.wav2vec2.per_character[digraph];

            const whisperRate = whisperStats?.error_rate || 0;
            const wav2vec2Rate = wav2vec2Stats?.error_rate || 0;
            const avgRate = (whisperRate + wav2vec2Rate) / 2;

            console.log(`[populatePhonologicalOverview] ${digraph}: Whisper ${whisperRate}%, Wav2Vec2 ${wav2vec2Rate}%, Avg ${avgRate}%`);

            if (avgRate > highestDigraphRate) {
                highestDigraphRate = avgRate;
                hardestDigraph = digraph;
            }
        });

        console.log(`[populatePhonologicalOverview] Hardest digraph: ${hardestDigraph} (${highestDigraphRate}%)`);
    } else {
        console.warn('[populatePhonologicalOverview] Character errors data not available for digraph analysis');
    }


    // CAPT suitability verdict based on overall normalized error rate
    const overallWhisper = data.normalized.whisper.overall?.error_rate || 0;
    const overallWav2vec2 = data.normalized.wav2vec2.overall?.error_rate || 0;
    const avgOverall = (overallWhisper + overallWav2vec2) / 2;

    let captVerdictEn = '';
    let captVerdictCy = '';
    let captVerdictClass = '';
    if (avgOverall < 10) {
        captVerdictEn = 'Excellent CAPT suitability';
        captVerdictCy = 'Addasrwydd CAPT rhagorol';
        captVerdictClass = 'bg-success text-white';
    } else if (avgOverall < 20) {
        captVerdictEn = 'Good CAPT suitability';
        captVerdictCy = 'Addasrwydd CAPT da';
        captVerdictClass = 'bg-info text-white';
    } else if (avgOverall < 30) {
        captVerdictEn = 'Moderate CAPT suitability';
        captVerdictCy = 'Addasrwydd CAPT cymedrol';
        captVerdictClass = 'bg-warning';
    } else {
        captVerdictEn = 'Poor CAPT suitability';
        captVerdictCy = 'Addasrwydd CAPT gwael';
        captVerdictClass = 'bg-danger text-white';
    }

    // Populate cards
    $('#phonoHardestCategory').html(`
        <div class="card h-100 border-danger">
            <div class="card-body">
                <h6 class="card-subtitle mb-2 text-muted">
                    <span class="en">Hardest Category</span>
                    <span class="cy">Categori Anoddaf</span>
                </h6>
                <div class="h4 mb-0">
                    <span class="en">${categoryLabelsEn[hardestCategory]}</span>
                    <span class="cy">${categoryLabelsCy[hardestCategory]}</span>
                </div>
                <small class="text-muted">${highestCategoryRate.toFixed(1)}% avg error rate</small>
            </div>
        </div>
    `);

    $('#phonoHardestDigraph').html(`
        <div class="card h-100 border-warning">
            <div class="card-body">
                <h6 class="card-subtitle mb-2 text-muted">
                    <span class="en">Hardest Welsh Digraph</span>
                    <span class="cy">Deuseiniog Cymraeg Anoddaf</span>
                </h6>
                <div class="h4 mb-0 font-monospace">${hardestDigraph || 'N/A'}</div>
                <small class="text-muted">${highestDigraphRate.toFixed(1)}% avg error rate</small>
            </div>
        </div>
    `);

    $('#phonoCaptVerdict').html(`
        <div class="card h-100 ${captVerdictClass}">
            <div class="card-body">
                <h6 class="card-subtitle mb-2">
                    <span class="en">CAPT Suitability Verdict</span>
                    <span class="cy">Dyfarniad Addasrwydd CAPT</span>
                </h6>
                <div class="h4 mb-0">
                    <span class="en">${captVerdictEn}</span>
                    <span class="cy">${captVerdictCy}</span>
                </div>
                <small>
                    <span class="en">Based on ${avgOverall.toFixed(1)}% avg phonological error rate</span>
                    <span class="cy">Yn seiliedig ar gyfradd gwall ffonolegol gyfartalog o ${avgOverall.toFixed(1)}%</span>
                </small>
            </div>
        </div>
    `);

    console.log('[populatePhonologicalOverview] Overview cards populated successfully');
}

function renderCategoryComparisonChart(data) {
    /**
     * Part B: Horizontal bar chart showing error rates by category
     * Categories: Vowels, Consonants, Digraphs, Overall
     * Uses ghost bar approach (light = raw context, solid = normalized)
     */
    console.log('[renderCategoryComparisonChart] Starting...');
    const ctx = document.getElementById('chartPhonologicalCategories');
    if (!ctx) {
        console.warn('[renderCategoryComparisonChart] Canvas not found: chartPhonologicalCategories');
        return;
    }

    const categories = ['vowels', 'consonants', 'digraphs', 'overall'];
    const categoryLabels = {
        vowels: 'Vowels',
        consonants: 'Consonants',
        digraphs: 'Digraphs',
        overall: 'Overall'
    };

    // Extract data for each category
    const labels = categories.map(cat => categoryLabels[cat]);
    const whisperRaw = categories.map(cat => data.raw.whisper[cat]?.error_rate || 0);
    const wav2vec2Raw = categories.map(cat => data.raw.wav2vec2[cat]?.error_rate || 0);
    const whisperNorm = categories.map(cat => data.normalized.whisper[cat]?.error_rate || 0);
    const wav2vec2Norm = categories.map(cat => data.normalized.wav2vec2[cat]?.error_rate || 0);

    console.log('[renderCategoryComparisonChart] Whisper raw:', whisperRaw);
    console.log('[renderCategoryComparisonChart] Whisper norm:', whisperNorm);
    console.log('[renderCategoryComparisonChart] Wav2Vec2 raw:', wav2vec2Raw);
    console.log('[renderCategoryComparisonChart] Wav2Vec2 norm:', wav2vec2Norm);

    const chartKey = 'phonologicalCategories';

    const chartData = {
        labels: labels,
        datasets: [
            // Raw metrics as ghost bars (context)
            createBarDataset('Whisper (Raw)', whisperRaw, 'rgba(54, 162, 235, 0.3)', {
                borderColor: 'rgba(54, 162, 235, 0.5)',
                order: 2
            }),
            createBarDataset('Wav2Vec2 (Raw)', wav2vec2Raw, 'rgba(255, 99, 132, 0.3)', {
                borderColor: 'rgba(255, 99, 132, 0.5)',
                order: 2
            }),
            // Normalized metrics as solid bars (primary)
            createBarDataset('Whisper (Normalized)', whisperNorm, COLORS.whisper, {
                borderWidth: 2,
                order: 1
            }),
            createBarDataset('Wav2Vec2 (Normalized)', wav2vec2Norm, COLORS.wav2vec2, {
                borderWidth: 2,
                order: 1
            })
        ]
    };

    const options = {
        indexAxis: 'y', // Horizontal bars
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'Error Rates by Phonological Category',
                font: { size: 14 }
            },
            subtitle: {
                display: true,
                text: 'Solid bars = normalized (pedagogically relevant), Ghost bars = raw',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 12
                }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const label = context.dataset.label || '';
                        const value = context.parsed.x;
                        return `${label}: ${value.toFixed(1)}%`;
                    }
                }
            }
        },
        scales: {
            x: {
                ...getStandardYAxisConfig('Error Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            y: {
                title: { display: true, text: 'Phonological Category' },
                grid: { display: false }
            }
        }
    };

    console.log('[renderCategoryComparisonChart] Creating/updating chart');
    createOrUpdateChart(chartKey, 'chartPhonologicalCategories', 'bar', chartData, options, charts);
    console.log('[renderCategoryComparisonChart] Chart rendered successfully');
}

function renderDigraphAnalysisChart(data) {
    /**
     * Part C: Welsh digraph-specific analysis
     * Shows error rates for ll, ch, dd, ff, ng, rh, ph, th
     * Calculates error rate per digraph from confusion matrix
     */
    const ctx = document.getElementById('chartDigraphAnalysis');
    if (!ctx) {
        console.warn('[renderDigraphAnalysisChart] Canvas not found');
        return;
    }

    const welshDigraphs = ['ll', 'ch', 'dd', 'ff', 'ng', 'rh', 'ph', 'th'];

    // Use character_errors data which has pre-calculated per-character error rates with actual totals
    // (Confusion matrix only contains top 15 error pairs, not correct recognitions)
    const charErrorData = analysisData.character_errors;
    console.log('[renderDigraphAnalysisChart] Character errors available:', !!charErrorData);

    if (!charErrorData || !charErrorData.whisper || !charErrorData.wav2vec2) {
        console.warn('[renderDigraphAnalysisChart] Character errors data not available');
        return;
    }

    // Extract error rates for each Welsh digraph
    const whisperRates = welshDigraphs.map(digraph => {
        const stats = charErrorData.whisper.per_character[digraph];
        return stats?.error_rate || 0;
    });

    const wav2vec2Rates = welshDigraphs.map(digraph => {
        const stats = charErrorData.wav2vec2.per_character[digraph];
        return stats?.error_rate || 0;
    });

    console.log('[renderDigraphAnalysisChart] Whisper rates:', whisperRates);
    console.log('[renderDigraphAnalysisChart] Wav2Vec2 rates:', wav2vec2Rates);

    const chartKey = 'digraphAnalysis';

    const chartData = {
        labels: welshDigraphs,
        datasets: [
            createBarDataset('Whisper', whisperRates, COLORS.whisper),
            createBarDataset('Wav2Vec2', wav2vec2Rates, COLORS.wav2vec2)
        ]
    };

    const options = {
        indexAxis: 'y', // Horizontal bars
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: 'Welsh Digraph Error Rates (Normalized)',
                font: { size: 14 }
            },
            subtitle: {
                display: true,
                text: 'Critical for L2 Welsh learners - these sounds don\'t exist in English',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top'
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const label = context.dataset.label || '';
                        const value = context.parsed.x;
                        return `${label}: ${value.toFixed(1)}%`;
                    }
                }
            }
        },
        scales: {
            x: {
                ...getStandardYAxisConfig('Error Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            y: {
                title: { display: true, text: 'Welsh Digraph' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, 'chartDigraphAnalysis', 'bar', chartData, options, charts);
}

function populateConfusionMatrices(data) {
    /**
     * Part D: Populate 4 confusion matrix tabs (overall, vowels, consonants, digraphs)
     * Each tab shows 2 tables: Whisper and Wav2Vec2
     * Top 15 character pairs showing which sounds get confused
     */
    console.log('[populateConfusionMatrices] Starting...');
    const categories = [
        { key: 'overall', label: 'Overall' },
        { key: 'vowels', label: 'Vowels' },
        { key: 'consonants', label: 'Consonants' },
        { key: 'digraphs', label: 'Digraphs' }
    ];

    categories.forEach(({ key, label }) => {
        // Whisper confusion matrix
        const whisperMatrix = data.normalized.whisper[key]?.confusion_matrix || [];
        const whisperTableId = `tableConfusionWhisper${key.charAt(0).toUpperCase() + key.slice(1)}`;
        console.log(`[populateConfusionMatrices] Populating ${whisperTableId}, matrix length: ${whisperMatrix.length}`);
        populateConfusionTable(whisperTableId, whisperMatrix);

        // Wav2Vec2 confusion matrix
        const wav2vec2Matrix = data.normalized.wav2vec2[key]?.confusion_matrix || [];
        const wav2vec2TableId = `tableConfusionWav2Vec2${key.charAt(0).toUpperCase() + key.slice(1)}`;
        console.log(`[populateConfusionMatrices] Populating ${wav2vec2TableId}, matrix length: ${wav2vec2Matrix.length}`);
        populateConfusionTable(wav2vec2TableId, wav2vec2Matrix);
    });

    console.log('[populateConfusionMatrices] All matrices populated');
}

function populateConfusionTable(tableId, matrixData) {
    /**
     * Helper: Populate a single confusion matrix table
     * Shows top 15 character pairs (expected → actual) with counts
     */
    const tbody = $(`#${tableId}`);
    if (!tbody.length) {
        console.warn(`[populateConfusionTable] Table not found: ${tableId}`);
        return;
    }

    tbody.empty();

    if (!matrixData || matrixData.length === 0) {
        tbody.append('<tr><td colspan="4" class="text-center text-muted">No confusion data available</td></tr>');
        return;
    }

    // Take top 15
    const top15 = matrixData.slice(0, 15);

    top15.forEach((entry, idx) => {
        const rowClass = idx < 5 ? 'table-danger' : (idx < 10 ? 'table-warning' : '');
        const expected = entry.expected || '-';
        const actual = entry.actual || '-';
        const count = entry.count || 0;

        tbody.append(`
            <tr class="${rowClass}">
                <td>${idx + 1}</td>
                <td class="font-monospace fw-bold">${expected}</td>
                <td class="font-monospace">${actual}</td>
                <td class="text-end">${count}</td>
            </tr>
        `);
    });
}

function populateNormalizationImpactTable(data) {
    /**
     * Part F: Normalization impact table
     * Shows raw vs normalized error rates and improvement for each category
     */
    const tbody = $('#tableNormalizationImpact');
    if (!tbody.length) {
        console.warn('[populateNormalizationImpactTable] Table not found');
        return;
    }

    tbody.empty();

    const categories = ['vowels', 'consonants', 'digraphs', 'overall'];
    const categoryLabels = {
        vowels: 'Vowels / Llafariaid',
        consonants: 'Consonants / Cytseiniaid',
        digraphs: 'Digraphs / Deuseiniau',
        overall: 'Overall / Cyffredinol'
    };

    categories.forEach(cat => {
        // Average across both models for cleaner presentation
        const rawWhisper = data.raw.whisper[cat]?.error_rate || 0;
        const rawWav2vec2 = data.raw.wav2vec2[cat]?.error_rate || 0;
        const normWhisper = data.normalized.whisper[cat]?.error_rate || 0;
        const normWav2vec2 = data.normalized.wav2vec2[cat]?.error_rate || 0;

        const avgRaw = (rawWhisper + rawWav2vec2) / 2;
        const avgNorm = (normWhisper + normWav2vec2) / 2;
        const absImprovement = avgRaw - avgNorm;
        const relImprovement = avgRaw > 0 ? (absImprovement / avgRaw) * 100 : 0;

        const impactClass = relImprovement > 30 ? 'text-success fw-bold' :
                           relImprovement > 15 ? 'text-info' : '';

        tbody.append(`
            <tr>
                <td>${categoryLabels[cat]}</td>
                <td>${avgRaw.toFixed(1)}%</td>
                <td>${avgNorm.toFixed(1)}%</td>
                <td class="${impactClass}">${absImprovement.toFixed(1)}% (${relImprovement.toFixed(0)}% relative)</td>
            </tr>
        `);
    });
}

function renderLinguisticCategoryChart(canvasId, category, data) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) {
        console.warn(`Canvas not found: ${canvasId}`);
        return;
    }

    console.log(`[renderLinguisticCategoryChart] canvasId: ${canvasId}, category: ${category}`);
    console.log(`[renderLinguisticCategoryChart] data.whisper:`, data.whisper);
    console.log(`[renderLinguisticCategoryChart] data.wav2vec2:`, data.wav2vec2);

    // Validate data structure
    if (!data.whisper || !data.wav2vec2) {
        console.warn(`Missing whisper or wav2vec2 data for ${canvasId}`);
        return;
    }
    
    if (!data.whisper[category]) {
        console.warn(`Missing data.whisper[${category}] for ${canvasId}. Available keys:`, Object.keys(data.whisper));
        return;
    }
    
    if (!data.wav2vec2[category]) {
        console.warn(`Missing data.wav2vec2[${category}] for ${canvasId}. Available keys:`, Object.keys(data.wav2vec2));
        return;
    }

    const whisperRate = data.whisper[category].error_rate || 0;
    const wav2vec2Rate = data.wav2vec2[category].error_rate || 0;

    console.log(`[renderLinguisticCategoryChart] ${canvasId} - Whisper: ${whisperRate}%, Wav2Vec2: ${wav2vec2Rate}%`);

    // Use canvasId as the unique chart key (not category, since we now have multiple charts per category)
    const chartKey = canvasId;

    const chartData = {
        labels: [category.charAt(0).toUpperCase() + category.slice(1) + ' Errors'],
        datasets: [
            createBarDataset('Whisper', [whisperRate], COLORS.whisper),
            createBarDataset('Wav2Vec2', [wav2vec2Rate], COLORS.wav2vec2)
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            ...getStandardPluginsConfig('', true, 'top'),
            tooltip: {
                callbacks: {
                    label: (context) => `Error Rate: ${context.parsed.y.toFixed(1)}%`
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Error Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'ASR Model' },
                grid: { display: false }
            }
        }
    };

    console.log(`[renderLinguisticCategoryChart] Creating/updating chart: ${chartKey}`);
    createOrUpdateChart(chartKey, canvasId, 'bar', chartData, options, charts);
    console.log(`[renderLinguisticCategoryChart] Chart rendered successfully: ${chartKey}`);
}

function renderLinguisticCategoryChartCombined(canvasId, category, rawData, normalizedData) {
    /**
     * Renders a combined chart showing raw and normalized metrics with ghost bar approach.
     * This makes it easy to compare:
     * 1. Model performance (Whisper vs Wav2Vec2)
     * 2. Impact of normalization (raw vs normalized for each model)
     * Ghost bars (light) = raw context, Solid bars = normalized (pedagogically relevant)
     */
    const ctx = document.getElementById(canvasId);
    if (!ctx) {
        console.warn(`[renderLinguisticCategoryChartCombined] Canvas ${canvasId} not found`);
        return;
    }

    // Extract data
    const whisperRaw = rawData.whisper[category]?.error_rate || 0;
    const wav2vec2Raw = rawData.wav2vec2[category]?.error_rate || 0;
    const whisperNorm = normalizedData.whisper[category]?.error_rate || 0;
    const wav2vec2Norm = normalizedData.wav2vec2[category]?.error_rate || 0;

    console.log(`[renderLinguisticCategoryChartCombined] ${canvasId} - Whisper: ${whisperRaw}% raw, ${whisperNorm}% norm | Wav2Vec2: ${wav2vec2Raw}% raw, ${wav2vec2Norm}% norm`);

    const chartKey = canvasId;
    const categoryLabel = category.charAt(0).toUpperCase() + category.slice(1);

    const chartData = {
        labels: ['Whisper', 'Wav2Vec2'],
        datasets: [
            // Raw metrics as ghost bars (lighter/context)
            createBarDataset('Whisper (Raw)', [whisperRaw, null], 'rgba(54, 162, 235, 0.3)', {
                borderColor: 'rgba(54, 162, 235, 0.5)',
                order: 2
            }),
            createBarDataset('Wav2Vec2 (Raw)', [null, wav2vec2Raw], 'rgba(255, 99, 132, 0.3)', {
                borderColor: 'rgba(255, 99, 132, 0.5)',
                order: 2
            }),
            // Normalized metrics as solid bars (primary)
            createBarDataset('Whisper (Normalized)', [whisperNorm, null], COLORS.whisper, {
                borderWidth: 2,
                order: 1
            }),
            createBarDataset('Wav2Vec2 (Normalized)', [null, wav2vec2Norm], COLORS.wav2vec2, {
                borderWidth: 2,
                order: 1
            })
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            title: {
                display: true,
                text: `ASR Performance on Welsh ${categoryLabel}`,
                font: { size: 14 }
            },
            subtitle: {
                display: true,
                text: 'Solid bars = normalized (pedagogically relevant), Ghost bars = raw',
                font: { size: 11 }
            },
            legend: {
                display: true,
                position: 'top',
                labels: {
                    usePointStyle: true,
                    padding: 12
                }
            },
            tooltip: {
                callbacks: {
                    label: (context) => {
                        const label = context.dataset.label || '';
                        const value = context.parsed.y;
                        return `${label}: ${value.toFixed(1)}%`;
                    }
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Error Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'ASR Model' },
                grid: { display: false }
            }
        }
    };

    console.log(`[renderLinguisticCategoryChartCombined] Creating/updating chart: ${chartKey}`);
    createOrUpdateChart(chartKey, canvasId, 'bar', chartData, options, charts);
    console.log(`[renderLinguisticCategoryChartCombined] Chart created successfully: ${chartKey}`);
}

function renderErrorCosts() {
    const data = analysisData.error_costs;
    
    console.log('[renderErrorCosts] Data:', data);
    
    if (!data) {
        console.warn('[renderErrorCosts] No data available');
        return;
    }
    
    // Check for new structure (raw/normalized split)
    if (!data.raw) {
        console.warn('[renderErrorCosts] No raw data. Data structure:', data);
        // Fallback: if old structure (without raw/normalized split), use it directly
        if (data.whisper && data.wav2vec2) {
            console.log('[renderErrorCosts] Using legacy data structure');
            renderErrorCostChart('chartErrorCostsWhisperRaw', data.whisper, 'Whisper');
            renderErrorCostChart('chartErrorCostsWav2Vec2Raw', data.wav2vec2, 'Wav2Vec2');
            
            $('#errorCostWhisperSafetyRaw').text(data.whisper.pedagogical_safety_score.toFixed(1));
            $('#errorCostWav2Vec2SafetyRaw').text(data.wav2vec2.pedagogical_safety_score.toFixed(1));
            
            if (data.comparison) {
                $('#errorCostSaferModelRaw').text(data.comparison.safer_model);
                $('#errorCostRecommendationRaw').text(data.comparison.recommendation || '');
            }
            
            populateErrorCostsTable(data);
            return;
        }
        return;
    }
    
    if (!data.raw.whisper || !data.raw.wav2vec2) {
        console.warn('[renderErrorCosts] Raw data missing whisper or wav2vec2:', data.raw);
        return;
    }

    console.log('[renderErrorCosts] Rendering RAW charts');
    // Render RAW charts for both models
    renderErrorCostChart('chartErrorCostsWhisperRaw', data.raw.whisper, 'Whisper (Raw)');
    renderErrorCostChart('chartErrorCostsWav2Vec2Raw', data.raw.wav2vec2, 'Wav2Vec2 (Raw)');

    // Render NORMALIZED charts for both models
    if (data.normalized && data.normalized.whisper && data.normalized.wav2vec2) {
        console.log('[renderErrorCosts] Rendering NORMALIZED charts');
        renderErrorCostChart('chartErrorCostsWhisperNorm', data.normalized.whisper, 'Whisper (Normalized)');
        renderErrorCostChart('chartErrorCostsWav2Vec2Norm', data.normalized.wav2vec2, 'Wav2Vec2 (Normalized)');
    } else {
        console.warn('[renderErrorCosts] No normalized data available');
    }

    // Populate RAW safety scores
    $('#errorCostWhisperSafetyRaw').text(data.raw.whisper.pedagogical_safety_score.toFixed(1));
    $('#errorCostWav2Vec2SafetyRaw').text(data.raw.wav2vec2.pedagogical_safety_score.toFixed(1));

    // Populate NORMALIZED safety scores
    if (data.normalized && data.normalized.whisper && data.normalized.wav2vec2) {
        $('#errorCostWhisperSafetyNorm').text(data.normalized.whisper.pedagogical_safety_score.toFixed(1));
        $('#errorCostWav2Vec2SafetyNorm').text(data.normalized.wav2vec2.pedagogical_safety_score.toFixed(1));
    }

    // Show safer model (raw)
    if (data.raw.comparison) {
        $('#errorCostSaferModelRaw').text(data.raw.comparison.safer_model);
        $('#errorCostRecommendationRaw').text(data.raw.comparison.recommendation || '');
    }

    // Show improvement interpretation
    if (data.improvement) {
        $('#errorCostWhisperImprovement').html(`
            <strong>Whisper:</strong> ${data.improvement.whisper.interpretation || '-'}
        `);
        $('#errorCostWav2Vec2Improvement').html(`
            <strong>Wav2Vec2:</strong> ${data.improvement.wav2vec2.interpretation || '-'}
        `);
    } else {
        console.warn('[renderErrorCosts] No improvement data available');
    }

    // Populate error cost drill-down table (using raw data)
    populateErrorCostsTable(data.raw);
}

function populateErrorCostsTable(data) {
    const tbody = $('#tableErrorCosts');
    if (!tbody.length) return;

    clearTable('tableErrorCosts');

    const errorTypes = [
        {
            key: 'false_acceptance',
            labelEn: 'False Acceptance',
            labelCy: 'Derbyniad Ffug',
            harmEn: '⚠️ High (Harmful)',
            harmCy: '⚠️ Uchel (Niweidiol)',
            harmClass: 'table-danger'
        },
        {
            key: 'false_rejection',
            labelEn: 'False Rejection',
            labelCy: 'Gwrthodiad Ffug',
            harmEn: '⚡ Medium (Demotivating)',
            harmCy: '⚡ Canolig (Digalonni)',
            harmClass: 'table-warning'
        }
    ];

    // Add rows for Whisper (only if count > 0)
    errorTypes.forEach(errorType => {
        const whisperData = data.whisper[errorType.key];
        if (whisperData) {
            const count = whisperData.count || 0;
            const rate = whisperData.rate !== undefined ? whisperData.rate.toFixed(1) : '0.0';

            // Only show row if there's actual data (count > 0)
            if (count > 0) {
                tbody.append(`
                    <tr class="${errorType.harmClass}">
                        <td><strong>Whisper</strong></td>
                        <td><span class="en">${errorType.labelEn}</span><span class="cy">${errorType.labelCy}</span></td>
                        <td class="text-center">${count}</td>
                        <td class="text-center">${rate}%</td>
                        <td><span class="en">${errorType.harmEn}</span><span class="cy">${errorType.harmCy}</span></td>
                    </tr>
                `);
            }
        }
    });

    // Add rows for Wav2Vec2 (only if count > 0)
    errorTypes.forEach(errorType => {
        const wav2vec2Data = data.wav2vec2[errorType.key];
        if (wav2vec2Data) {
            const count = wav2vec2Data.count || 0;
            const rate = wav2vec2Data.rate !== undefined ? wav2vec2Data.rate.toFixed(1) : '0.0';

            // Only show row if there's actual data (count > 0)
            if (count > 0) {
                tbody.append(`
                    <tr class="${errorType.harmClass}">
                        <td><strong>Wav2Vec2</strong></td>
                        <td><span class="en">${errorType.labelEn}</span><span class="cy">${errorType.labelCy}</span></td>
                        <td class="text-center">${count}</td>
                        <td class="text-center">${rate}%</td>
                        <td><span class="en">${errorType.harmEn}</span><span class="cy">${errorType.harmCy}</span></td>
                    </tr>
                `);
            }
        }
    });

    // Reinitialize language toggle for dynamically added content
    if (typeof initLanguageToggle === 'function') {
        initLanguageToggle();
    }
}

function renderErrorCostChart(canvasId, modelData, modelName) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    // Validate model data structure
    if (!modelData) {
        console.warn(`Missing error cost data for model: ${modelName}`);
        return;
    }

    const chartKey = `errorCost_${modelName}`;

    const data = {
        labels: ['Error Costs'],
        datasets: [
            createBarDataset('False Acceptance', [modelData.false_acceptance?.rate || 0], '#e74c3c'),
            createBarDataset('False Rejection', [modelData.false_rejection?.rate || 0], '#f39c12'),
            createBarDataset('True Positive', [modelData.true_positive?.rate || 0], '#2ecc71'),
            createBarDataset('True Negative', [modelData.true_negative?.rate || 0], '#95a5a6')
        ]
    };

    const options = {
        ...getResponsiveOptions(),
        plugins: {
            ...getStandardPluginsConfig('', true, 'top'),
            tooltip: {
                callbacks: {
                    label: (context) => `${context.label}: ${context.parsed.y.toFixed(1)}%`
                }
            }
        },
        scales: {
            y: {
                ...getStandardYAxisConfig('Rate (%)', 0, 100),
                grid: { color: COLORS.gridLines }
            },
            x: {
                title: { display: true, text: 'Error Cost Category' },
                grid: { display: false }
            }
        }
    };

    createOrUpdateChart(chartKey, canvasId, 'bar', data, options, charts);
}

function renderConsistencyReliability() {
    const data = analysisData.consistency_reliability;
    
    console.log('[renderConsistencyReliability] Data:', data);
    
    if (!data || !data.raw) {
        console.warn('[renderConsistencyReliability] No data or raw data available');
        return;
    }

    const rawData = data.raw;
    const normData = data.normalized;

    // Show summary interpretation (if available from comparison)
    if (data.comparison) {
        $('#consistencyWinner').text(data.comparison.more_consistent_model || '-');
        $('#reliabilityWinner').text(data.comparison.more_reliable_model || '-');
        $('#consistencyInterpretation').text(data.comparison.interpretation || '-');
    } else {
        // Generate interpretation from the data
        const whisperCV = rawData.whisper?.coefficient_of_variation || 0;
        const wav2vec2CV = rawData.wav2vec2?.coefficient_of_variation || 0;
        const moreConsistent = whisperCV < wav2vec2CV ? 'Whisper' : 'Wav2Vec2';
        
        $('#consistencyWinner').text(moreConsistent);
        $('#reliabilityWinner').text(moreConsistent);
        $('#consistencyInterpretation').text(`${moreConsistent} shows lower variation (CV: ${Math.min(whisperCV, wav2vec2CV).toFixed(1)}% vs ${Math.max(whisperCV, wav2vec2CV).toFixed(1)}%)`);
    }

    // Populate RAW detailed metrics table
    const tbodyRaw = $('#tableConsistencyMetricsRaw');
    clearTable('tableConsistencyMetricsRaw');

    const metricsRaw = [
        { label: 'Mean CER', whisper: rawData.whisper?.mean, wav2vec2: rawData.wav2vec2?.mean, format: v => `${v.toFixed(1)}%` },
        { label: 'Std Dev', whisper: rawData.whisper?.std_dev, wav2vec2: rawData.wav2vec2?.std_dev, format: v => `${v.toFixed(1)}%` },
        { label: 'CV (Coefficient of Variation)', whisper: rawData.whisper?.coefficient_of_variation, wav2vec2: rawData.wav2vec2?.coefficient_of_variation, format: v => `${v.toFixed(1)}%` },
        { label: 'IQR (Interquartile Range)', whisper: rawData.whisper?.iqr, wav2vec2: rawData.wav2vec2?.iqr, format: v => `${v.toFixed(1)}%` },
        { label: 'Worst Case (95th percentile)', whisper: rawData.whisper?.percentile_95, wav2vec2: rawData.wav2vec2?.percentile_95, format: v => `${v.toFixed(1)}%` },
        { label: 'Reliability Rating', whisper: rawData.whisper?.reliability_rating, wav2vec2: rawData.wav2vec2?.reliability_rating, format: v => v }
    ];

    metricsRaw.forEach(metric => {
        if (metric.whisper !== undefined && metric.wav2vec2 !== undefined) {
            tbodyRaw.append(`
                <tr>
                    <td>${metric.label}</td>
                    <td class="text-end">${metric.format(metric.whisper)}</td>
                    <td class="text-end">${metric.format(metric.wav2vec2)}</td>
                </tr>
            `);
        }
    });

    // Populate NORMALIZED detailed metrics table
    if (normData && normData.whisper && normData.wav2vec2) {
        console.log('[renderConsistencyReliability] Rendering normalized metrics table');
        const tbodyNorm = $('#tableConsistencyMetricsNorm');
        clearTable('tableConsistencyMetricsNorm');

        const metricsNorm = [
            { label: 'Mean CER', whisper: normData.whisper?.mean, wav2vec2: normData.wav2vec2?.mean, format: v => `${v.toFixed(1)}%` },
            { label: 'Std Dev', whisper: normData.whisper?.std_dev, wav2vec2: normData.wav2vec2?.std_dev, format: v => `${v.toFixed(1)}%` },
            { label: 'CV (Coefficient of Variation)', whisper: normData.whisper?.coefficient_of_variation, wav2vec2: normData.wav2vec2?.coefficient_of_variation, format: v => `${v.toFixed(1)}%` },
            { label: 'IQR (Interquartile Range)', whisper: normData.whisper?.iqr, wav2vec2: normData.wav2vec2?.iqr, format: v => `${v.toFixed(1)}%` },
            { label: 'Worst Case (95th percentile)', whisper: normData.whisper?.percentile_95, wav2vec2: normData.wav2vec2?.percentile_95, format: v => `${v.toFixed(1)}%` },
            { label: 'Reliability Rating', whisper: normData.whisper?.reliability_rating, wav2vec2: normData.wav2vec2?.reliability_rating, format: v => v }
        ];

        metricsNorm.forEach(metric => {
            if (metric.whisper !== undefined && metric.wav2vec2 !== undefined) {
                tbodyNorm.append(`
                    <tr>
                        <td>${metric.label}</td>
                        <td class="text-end">${metric.format(metric.whisper)}</td>
                        <td class="text-end">${metric.format(metric.wav2vec2)}</td>
                    </tr>
                `);
            }
        });
    } else {
        console.warn('[renderConsistencyReliability] No normalized data available');
    }
}

function renderQualitativeExamples() {
    const data = analysisData.qualitative_examples;
    if (!data) return;

    const summary = data.summary;

    // Update counts
    $('#exampleBothCorrectCount, #exampleBothCorrectCountCy').text(summary?.both_correct_count || 0);
    $('#exampleWhisperOnlyCount, #exampleWhisperOnlyCountCy').text(summary?.whisper_only_count || 0);
    $('#exampleWav2Vec2OnlyCount, #exampleWav2Vec2OnlyCountCy').text(summary?.wav2vec2_only_count || 0);
    $('#exampleCriticalCount, #exampleCriticalCountCy').text(summary?.pedagogically_critical_count || 0);

    // Populate tables (data is now flat, not nested under "examples")
    populateExampleTable('tableExamplesBothCorrect', data.both_correct);
    populateExampleTable('tableExamplesWhisperOnly', data.whisper_only);
    populateExampleTable('tableExamplesWav2Vec2Only', data.wav2vec2_only);
    populateCriticalExampleTable('tableExamplesCritical', data.pedagogically_critical);
}

function populateExampleTable(tableId, examples) {
    const tbody = $(`#${tableId} tbody`);
    tbody.empty();

    if (!examples || examples.length === 0) {
        tbody.append('<tr><td colspan="5" class="text-muted text-center">No examples</td></tr>');
        return;
    }

    examples.forEach(ex => {
        tbody.append(`
            <tr>
                <td>${ex.target || '-'}</td>
                <td>${ex.human || '-'}</td>
                <td>${ex.whisper || '-'}</td>
                <td>${ex.wav2vec2 || '-'}</td>
                <td class="text-center">
                    <a href="/Admin/Transcriptions?word=${encodeURIComponent(ex.target || '')}"
                       class="btn btn-sm btn-outline-primary">
                        <span class="en">📋 View</span>
                        <span class="cy">📋 Gweld</span>
                    </a>
                </td>
            </tr>
        `);
    });

    // Reinitialize language toggle for dynamically added content
    if (typeof initLanguageToggle === 'function') {
        initLanguageToggle();
    }
}

function populateCriticalExampleTable(tableId, examples) {
    const tbody = $(`#${tableId} tbody`);
    tbody.empty();

    if (!examples || examples.length === 0) {
        tbody.append('<tr><td colspan="6" class="text-muted text-center">No critical errors</td></tr>');
        return;
    }

    examples.forEach(ex => {
        const criticalModels = ex.critical_models || '-';
        tbody.append(`
            <tr class="table-warning">
                <td>${ex.target || '-'}</td>
                <td>${ex.human || '-'}</td>
                <td>${ex.whisper || '-'}</td>
                <td>${ex.wav2vec2 || '-'}</td>
                <td class="fw-bold">${criticalModels}</td>
                <td class="text-center">
                    <a href="/Admin/Transcriptions?word=${encodeURIComponent(ex.target || '')}"
                       class="btn btn-sm btn-outline-primary">
                        <span class="en">📋 View</span>
                        <span class="cy">📋 Gweld</span>
                    </a>
                </td>
            </tr>
        `);
    });

    // Reinitialize language toggle for dynamically added content
    if (typeof initLanguageToggle === 'function') {
        initLanguageToggle();
    }
}

function renderStudyDesign() {
    const data = analysisData.study_design;
    if (!data) return;

    // Study type
    $('#studyType').text(data.study_type || '-');

    // Statistical approach
    if (data.statistical_approach) {
        $('#studyStatisticalApproach').text(
            `${data.statistical_approach.paradigm}: ${data.statistical_approach.rationale}`
        );
    }

    // Data collection
    if (data.data_collection) {
        const dc = data.data_collection;
        const html = `
            <li><strong>Task:</strong> ${dc.task || '-'}</li>
            <li><strong>Models:</strong> ${dc.asr_models_compared ? dc.asr_models_compared.join(', ') : '-'}</li>
            <li><strong>Date Range:</strong> ${dc.date_range || '-'}</li>
        `;
        $('#studyDataCollection').html(html);
    }

    // Limitations
    if (data.limitations && data.limitations.length > 0) {
        const html = data.limitations.map(l => `<li>${l}</li>`).join('');
        $('#studyLimitations').html(html);
    }

    // Strengths
    if (data.strengths && data.strengths.length > 0) {
        const html = data.strengths.map(s => `<li>${s}</li>`).join('');
        $('#studyStrengths').html(html);
    }
}

function renderPracticalRecommendations() {
    const data = analysisData.practical_recommendations;
    if (!data) return;

    // Primary recommendation
    if (data.primary_recommendation) {
        $('#recommendedModel').text(data.primary_recommendation.recommended_model || '-');
        $('#recommendationRationale').text(data.primary_recommendation.rationale || '-');
    }

    // System design implications
    if (data.system_design_implications && data.system_design_implications.length > 0) {
        const html = data.system_design_implications.map(item => `<li>${item}</li>`).join('');
        $('#recommendationSystemDesign').html(html);
    }

    // Pedagogical considerations
    if (data.pedagogical_considerations && data.pedagogical_considerations.length > 0) {
        const html = data.pedagogical_considerations.map(item => `<li>${item}</li>`).join('');
        $('#recommendationPedagogical').html(html);
    }

    // Features requiring human verification
    if (data.features_requiring_human_verification && data.features_requiring_human_verification.length > 0) {
        const html = data.features_requiring_human_verification.map(item => `<li>${item}</li>`).join('');
        $('#recommendationHumanVerification').html(html);
    }
}

// renderHybridPracticalImplications() and renderHybridImplicationsChart() have been merged
// into renderHybridAnalysis() and renderHybridAnalysisChart() above

// =============================================================================
// ENHANCED EXISTING RENDER FUNCTIONS
// =============================================================================

async function exportForLaTeX() {
    if (!analysisData || Object.keys(analysisData).length === 0) {
        showNotification('No data to export. Please load analytics first. / Dim data i allforio. Llwythwch ddadansoddiad yn gyntaf.', 'warning');
        return;
    }

    // Collect chart images
    const chartImages = {};

    if (charts.modelComparison) {
        chartImages['model_comparison'] = charts.modelComparison.toBase64Image('image/png', 1);
    }
    if (charts.confusionMatrix) {
        chartImages['confusion_matrix'] = charts.confusionMatrix.toBase64Image('image/png', 1);
    }
    if (charts.errorAttribution) {
        chartImages['error_attribution'] = charts.errorAttribution.toBase64Image('image/png', 1);
    }
    if (charts.wordDifficulty) {
        chartImages['word_difficulty'] = charts.wordDifficulty.toBase64Image('image/png', 1);
    }

    // Character error charts
    if (charts.whisper_overall) {
        chartImages['char_errors_whisper_overall'] = charts.whisper_overall.toBase64Image('image/png', 1);
    }
    if (charts.wav2vec2_overall) {
        chartImages['char_errors_wav2vec2_overall'] = charts.wav2vec2_overall.toBase64Image('image/png', 1);
    }
    if (charts.whisper_vowels) {
        chartImages['char_errors_whisper_vowels'] = charts.whisper_vowels.toBase64Image('image/png', 1);
    }
    if (charts.wav2vec2_vowels) {
        chartImages['char_errors_wav2vec2_vowels'] = charts.wav2vec2_vowels.toBase64Image('image/png', 1);
    }
    if (charts.whisper_consonants) {
        chartImages['char_errors_whisper_consonants'] = charts.whisper_consonants.toBase64Image('image/png', 1);
    }
    if (charts.wav2vec2_consonants) {
        chartImages['char_errors_wav2vec2_consonants'] = charts.wav2vec2_consonants.toBase64Image('image/png', 1);
    }
    if (charts.whisper_digraphs) {
        chartImages['char_errors_whisper_digraphs'] = charts.whisper_digraphs.toBase64Image('image/png', 1);
    }
    if (charts.wav2vec2_digraphs) {
        chartImages['char_errors_wav2vec2_digraphs'] = charts.wav2vec2_digraphs.toBase64Image('image/png', 1);
    }

    elements.btnExportLaTeX.prop('disabled', true).html('<i data-feather="loader"></i> Generating...');

    if (typeof feather !== 'undefined') {
        feather.replace();
    }

    try {
        const response = await fetch('/Admin/Analytics/ExportLaTeX', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                images: chartImages,
                data: analysisData
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        if (data.success && data.download_url) {
            window.location.href = data.download_url;
            showNotification('LaTeX export downloaded successfully! / Allforio LaTeX wedi\'i lwytho\'n llwyddiannus!', 'success');
        } else {
            throw new Error(data.error || 'Unknown error');
        }
    } catch (error) {
        console.error('Export LaTeX error:', error);
        showNotification('Failed to generate LaTeX export / Methwyd cynhyrchu allforio LaTeX', 'error');
    } finally {
        elements.btnExportLaTeX.prop('disabled', false).html('<i data-feather="download"></i> <span class="en">Export for LaTeX</span><span class="cy">Allforio ar gyfer LaTeX</span>');

        if (typeof feather !== 'undefined') {
            feather.replace();
        }
    }
}


// Export individual section for LaTeX
async function exportSection(sectionName) {
    if (!analysisData || Object.keys(analysisData).length === 0) {
        showNotification('No data to export. Please load analytics first. / Dim data i allforio. Llwythwch ddadansoddiad yn gyntaf.', 'warning');
        return;
    }

    // Map section names to chart keys
    const sectionCharts = {
        'executive-summary': [],
        'model-comparison': ['modelComparisonRaw', 'modelComparisonNormalized'],
        'linguistic-patterns': ['linguistic_vowels', 'linguistic_consonants', 'linguistic_digraphs'],
        'error-costs': ['errorCost_Whisper', 'errorCost_Wav2Vec2'],
        'error-attribution': ['errorAttributionRaw', 'errorAttributionNormalized'],
        'challenging-words': ['topDifficultWordsRaw', 'topDifficultWordsNormalized'],
        'character-errors': ['charErrors_combined'],
        'cer-distribution': ['wordDifficulty'],  // Chart key for CER distribution
        'consistency': [],
        'examples': [],
        'study-design': [],
        'recommendations': []
    };

    // Collect chart images for this section
    const chartImages = {};
    const chartsToCapture = sectionCharts[sectionName] || [];
    
    chartsToCapture.forEach(chartKey => {
        if (charts[chartKey]) {
            chartImages[chartKey] = charts[chartKey].toBase64Image('image/png', 1);
        }
    });

    // Show loading state on the button
    const $button = $(`.export-section-btn[data-section="${sectionName}"]`);
    const originalHtml = $button.html();
    $button.prop('disabled', true).html('<i data-feather="loader"></i> Exporting...');
    
    if (typeof feather !== 'undefined') {
        feather.replace();
    }

    try {
        const response = await fetch('/Admin/Analytics/ExportSection', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                section: sectionName,
                images: chartImages,
                data: analysisData  // Send all analysis data
            })
        });

        console.log('[exportSection] Response Status:', response.status, 'Content-Type:', response.headers.get('Content-Type'));

        if (!response.ok) {
            const errorText = await response.text();
            console.error('[exportSection] Server error:', errorText);
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        // Parse response
        const responseText = await response.text();
        let data;
        try {
            data = JSON.parse(responseText);
        } catch (parseError) {
            console.error('[exportSection] JSON parse error:', parseError);
            console.error('[exportSection] Response preview:', responseText.substring(0, 200));
            throw new Error(`Invalid JSON response: ${parseError.message}`);
        }

        if (data.success && data.download_url) {
            window.location.href = data.download_url;
            showNotification(`Section "${sectionName}" exported successfully! / Adran "${sectionName}" wedi'i hallforio'n llwyddiannus!`, 'success');
        } else {
            throw new Error(data.error || 'Unknown error');
        }
    } catch (error) {
        console.error('Export section error:', error);
        showNotification(`Failed to export section "${sectionName}" / Methwyd allforio adran "${sectionName}"`, 'error');
    } finally {
        // Restore button state
        $button.prop('disabled', false).html(originalHtml);
        if (typeof feather !== 'undefined') {
            feather.replace();
        }
    }
}

/**
 * Export charts for dissertation Results chapter
 * Downloads 5 specific charts as individual PNG files with dissertation-ready naming
 * Uses sequential downloads with delays to avoid browser blocking
 */
async function exportDissertationCharts() {
    if (!analysisData || Object.keys(analysisData).length === 0) {
        showNotification('No data to export. Please load analytics first. / Dim data i allforio. Llwythwch ddadansoddiad yn gyntaf.', 'warning');
        return;
    }

    // Map dissertation chart names to chart keys (as stored in charts object)
    const dissertationCharts = {
        'modelComparison.png': 'chartModelComparison',
        'phonologicalCategories.png': 'phonologicalCategories',  // No "chart" prefix in storage
        'errorAttribution.png': 'chartErrorAttribution',
        'digraphAnalysis.png': 'digraphAnalysis',  // No "chart" prefix in storage
        'difficultWords.png': 'chartTopDifficultWords'
    };

    let successCount = 0;
    let failCount = 0;
    const missingCharts = [];

    // Export each chart sequentially with delay (browsers block multiple simultaneous downloads)
    for (const [filename, canvasId] of Object.entries(dissertationCharts)) {
        // Get Chart.js instance from charts object
        if (!charts[canvasId]) {
            console.warn(`[exportDissertationCharts] Chart not found: ${canvasId}`);
            missingCharts.push(filename);
            failCount++;
            continue;
        }

        try {
            // Get base64 PNG data from Chart.js instance
            const base64Data = charts[canvasId].toBase64Image('image/png', 1);

            // Convert base64 to blob
            const byteString = atob(base64Data.split(',')[1]);
            const mimeString = base64Data.split(',')[0].split(':')[1].split(';')[0];
            const ab = new ArrayBuffer(byteString.length);
            const ia = new Uint8Array(ab);
            for (let i = 0; i < byteString.length; i++) {
                ia[i] = byteString.charCodeAt(i);
            }
            const blob = new Blob([ab], { type: mimeString });

            // Trigger download
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(link.href);

            successCount++;
            console.log(`[exportDissertationCharts] Exported ${filename} (${successCount}/${Object.keys(dissertationCharts).length})`);

            // Wait 300ms before next download to avoid browser blocking
            if (successCount < Object.keys(dissertationCharts).length) {
                await new Promise(resolve => setTimeout(resolve, 300));
            }
        } catch (error) {
            console.error(`[exportDissertationCharts] Failed to export ${filename}:`, error);
            failCount++;
            missingCharts.push(filename);
        }
    }

    // Show result notification
    if (successCount === Object.keys(dissertationCharts).length) {
        showNotification(`Successfully exported ${successCount} dissertation charts! / Wedi allforio ${successCount} siart traethawd hir yn llwyddiannus!`, 'success');
    } else if (successCount > 0) {
        showNotification(`Exported ${successCount} charts, ${failCount} failed. Missing: ${missingCharts.join(', ')} / Wedi allforio ${successCount} siart, ${failCount} wedi methu.`, 'warning');
    } else {
        showNotification(`Failed to export charts. Please ensure analytics data is loaded. / Methwyd allforio siartiau.`, 'error');
    }
}
