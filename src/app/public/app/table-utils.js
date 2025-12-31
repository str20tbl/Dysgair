/**
 * Table Rendering Utilities for Dysgair Analytics
 * Provides reusable table population and rendering helpers
 * Eliminates repetitive jQuery DOM manipulation across analytics dashboard
 */

/**
 * Clear all rows from a table body
 * @param {string} tableId - Table element ID or tbody selector
 */
function clearTable(tableId) {
    const selector = tableId.startsWith('#') ? tableId : `#${tableId}`;
    $(selector).empty();
}

/**
 * Populate a simple key-value table
 * @param {string} tableId - Table body element ID
 * @param {Array<Array>} rows - Array of [key, value] pairs
 * @param {Object} options - Formatting options
 */
function populateKeyValueTable(tableId, rows, options = {}) {
    const tbody = $(`#${tableId}`);
    clearTable(tableId);

    if (!rows || rows.length === 0) {
        tbody.append('<tr><td colspan="2" class="text-center text-muted">No data available</td></tr>');
        return;
    }

    rows.forEach(([key, value]) => {
        const valueFormatted = options.formatValue ? options.formatValue(value) : value;
        tbody.append(`<tr><td>${key}</td><td>${valueFormatted}</td></tr>`);
    });
}

/**
 * Populate a multi-column data table
 * @param {string} tableId - Table body element ID
 * @param {Array<string>} headers - Column headers (for reference, assumes thead exists)
 * @param {Array<Array>} rows - Array of row data arrays
 * @param {Object} options - Formatting options
 */
function populateDataTable(tableId, headers, rows, options = {}) {
    const tbody = $(`#${tableId}`);
    clearTable(tableId);

    if (!rows || rows.length === 0) {
        const colspan = headers.length;
        tbody.append(`<tr><td colspan="${colspan}" class="text-center text-muted">No data available</td></tr>`);
        return;
    }

    rows.forEach(row => {
        const cells = row.map((cell, idx) => {
            const formatted = options.formatters && options.formatters[idx]
                ? options.formatters[idx](cell)
                : cell;
            return `<td>${formatted}</td>`;
        }).join('');
        tbody.append(`<tr>${cells}</tr>`);
    });
}

/**
 * Populate a distribution table (buckets and counts)
 * Common pattern for CER/WER distribution histograms
 * @param {string} tableId - Table body element ID
 * @param {Object} distribution - Distribution data with buckets and counts
 */
function populateDistributionTable(tableId, distribution) {
    const tbody = $(`#${tableId}`);
    clearTable(tableId);

    if (!distribution || !distribution.buckets || !distribution.whisper_counts) {
        tbody.append('<tr><td colspan="3" class="text-center text-muted">No distribution data</td></tr>');
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

/**
 * Populate challenging words table with both raw and normalized metrics
 * @param {string} tableId - Table body element ID
 * @param {Array} rawWords - Raw metric word data
 * @param {Array} normalizedWords - Normalized metric word data
 */
function populateChallengingWordsTable(tableId, rawWords, normalizedWords) {
    const tbody = $(`#${tableId}`);
    clearTable(tableId);

    if (!rawWords || rawWords.length === 0) {
        tbody.append('<tr><td colspan="7" class="text-center text-muted">No difficult words found</td></tr>');
        return;
    }

    // Take top 10
    const top10Raw = rawWords.slice(0, 10);
    const top10Norm = normalizedWords ? normalizedWords.slice(0, 10) : null;

    top10Raw.forEach((wordData, idx) => {
        const normData = top10Norm ? top10Norm[idx] : null;

        tbody.append(`
            <tr>
                <td><strong>${wordData.word}</strong></td>
                <td>${wordData.attempts}</td>
                <td>${wordData.whisper_avg_cer.toFixed(1)}%</td>
                <td>${wordData.wav2vec2_avg_cer.toFixed(1)}%</td>
                <td>${normData ? normData.whisper_avg_cer.toFixed(1) : '-'}%</td>
                <td>${normData ? normData.wav2vec2_avg_cer.toFixed(1) : '-'}%</td>
                <td>${wordData.word_length}</td>
            </tr>
        `);
    });
}

/**
 * Populate qualitative examples table
 * @param {string} tableId - Table body element ID
 * @param {Array} examples - Example data objects
 * @param {number} limit - Maximum examples to show (default: 10)
 */
function populateQualitativeExamplesTable(tableId, examples, limit = 10) {
    const tbody = $(`#${tableId}`);
    clearTable(tableId);

    if (!examples || examples.length === 0) {
        tbody.append('<tr><td colspan="5" class="text-center text-muted">No examples available</td></tr>');
        return;
    }

    examples.slice(0, limit).forEach(ex => {
        tbody.append(`
            <tr>
                <td><strong>${ex.target}</strong></td>
                <td>${ex.whisper_transcription || '-'}</td>
                <td>${ex.wav2vec2_transcription || '-'}</td>
                <td>${ex.whisper_cer.toFixed(1)}%</td>
                <td>${ex.wav2vec2_cer.toFixed(1)}%</td>
            </tr>
        `);
    });
}

/**
 * Populate a metrics comparison table
 * Common pattern for showing Whisper vs Wav2Vec2 metrics
 * @param {string} tableId - Table body element ID
 * @param {Object} whisperMetrics - Whisper metrics object
 * @param {Object} wav2vec2Metrics - Wav2Vec2 metrics object
 * @param {Array<string>} metricKeys - Keys to display
 * @param {Object} labels - Display labels for metrics
 */
function populateMetricsComparisonTable(tableId, whisperMetrics, wav2vec2Metrics, metricKeys, labels) {
    const tbody = $(`#${tableId}`);
    clearTable(tableId);

    metricKeys.forEach(key => {
        const label = labels[key] || key;
        const whisperValue = whisperMetrics[key];
        const wav2vec2Value = wav2vec2Metrics[key];

        const whisperFormatted = typeof whisperValue === 'number' ? whisperValue.toFixed(2) : whisperValue;
        const wav2vec2Formatted = typeof wav2vec2Value === 'number' ? wav2vec2Value.toFixed(2) : wav2vec2Value;

        tbody.append(`
            <tr>
                <td>${label}</td>
                <td>${whisperFormatted}</td>
                <td>${wav2vec2Formatted}</td>
            </tr>
        `);
    });
}

/**
 * Populate a confusion matrix table with heatmap coloring
 * @param {string} tableId - Table element ID (includes thead and tbody)
 * @param {Object} confusionData - Confusion matrix data
 */
function populateConfusionMatrixTable(tableId, confusionData) {
    const table = $(`#${tableId}`);
    table.empty();

    if (!confusionData || !confusionData.matrix) {
        table.append('<tr><td class="text-center text-muted">No confusion matrix data</td></tr>');
        return;
    }

    const matrix = confusionData.matrix;
    const characters = confusionData.characters || [];

    // Create header row
    let headerRow = '<tr><th>Target→<br>ASR↓</th>';
    characters.forEach(char => {
        headerRow += `<th>${char}</th>`;
    });
    headerRow += '</tr>';

    const thead = $('<thead></thead>').append(headerRow);
    const tbody = $('<tbody></tbody>');

    // Find max value for heatmap scaling
    const maxVal = Math.max(...matrix.flat());

    // Create data rows
    matrix.forEach((row, rowIdx) => {
        let dataRow = `<td><strong>${characters[rowIdx]}</strong></td>`;
        row.forEach(val => {
            const heatmapClass = val > 0 ? `heatmap-${Math.min(10, Math.ceil((val / maxVal) * 10))}` : '';
            dataRow += `<td class="${heatmapClass}">${val}</td>`;
        });
        tbody.append(`<tr>${dataRow}</tr>`);
    });

    table.append(thead).append(tbody);
}

/**
 * Update element text with formatted value
 * @param {string} selector - jQuery selector
 * @param {any} value - Value to display
 * @param {Function} formatter - Optional formatter function
 */
function updateElementText(selector, value, formatter = null) {
    const formatted = formatter ? formatter(value) : value;
    $(selector).text(formatted);
}

/**
 * Update multiple elements with data mapping
 * @param {Object} elementMap - Map of selector to value
 * @param {Function} defaultFormatter - Default formatter for all values
 */
function updateMultipleElements(elementMap, defaultFormatter = null) {
    Object.entries(elementMap).forEach(([selector, value]) => {
        updateElementText(selector, value, defaultFormatter);
    });
}

/**
 * Format number as percentage with decimal places
 * @param {number} value - Numeric value
 * @param {number} decimals - Decimal places (default: 1)
 */
function formatPercentage(value, decimals = 1) {
    if (value === null || value === undefined || isNaN(value)) return '-';
    return `${value.toFixed(decimals)}%`;
}

/**
 * Format number with locale-specific thousands separators
 * @param {number} value - Numeric value
 */
function formatNumber(value) {
    if (value === null || value === undefined || isNaN(value)) return '-';
    return value.toLocaleString();
}

/**
 * Format decimal number with fixed places
 * @param {number} value - Numeric value
 * @param {number} decimals - Decimal places (default: 2)
 */
function formatDecimal(value, decimals = 2) {
    if (value === null || value === undefined || isNaN(value)) return '-';
    return value.toFixed(decimals);
}
