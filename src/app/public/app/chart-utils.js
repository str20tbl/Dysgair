/**
 * Chart.js Utilities for Dysgair Analytics
 * Provides reusable chart creation, configuration, and export helpers
 * Eliminates repetitive Chart.js boilerplate across analytics dashboard
 *
 * Note: Uses COLORS object from analytics.js (loaded after this file)
 */

/**
 * Creates or updates a Chart.js instance
 * Handles the common pattern: check if exists -> update OR create new
 *
 * @param {string} chartKey - Key in global charts object
 * @param {string} canvasId - Canvas element ID
 * @param {string} type - Chart type (bar, line, doughnut, etc.)
 * @param {Object} data - Chart data object
 * @param {Object} options - Chart options object
 * @param {Object} charts - Global charts object to store chart instances
 */
function createOrUpdateChart(chartKey, canvasId, type, data, options, charts) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) {
        console.warn(`Canvas element not found: ${canvasId}`);
        return null;
    }

    const ctx = canvas.getContext('2d');

    if (charts[chartKey]) {
        // Update existing chart
        charts[chartKey].data = data;
        charts[chartKey].options = options;
        charts[chartKey].update();
    } else {
        // Create new chart
        charts[chartKey] = new Chart(ctx, {
            type: type,
            data: data,
            options: options
        });
    }

    return charts[chartKey];
}

/**
 * Get standard responsive chart options
 */
function getResponsiveOptions() {
    return {
        responsive: true,
        maintainAspectRatio: true
    };
}

/**
 * Get standard tooltip configuration
 * @param {Function} labelCallback - Optional custom label formatter
 */
function getStandardTooltipConfig(labelCallback) {
    const config = {
        enabled: true,
        mode: 'index',
        intersect: false
    };

    if (labelCallback) {
        config.callbacks = { label: labelCallback };
    }

    return config;
}

/**
 * Get standard Y-axis configuration
 * @param {string} label - Y-axis label
 * @param {number} min - Minimum value (default: 0)
 * @param {number} max - Maximum value (optional)
 * @param {Function} tickCallback - Optional tick formatter
 */
function getStandardYAxisConfig(label = '', min = 0, max = null, tickCallback = null) {
    const config = {
        beginAtZero: true,
        min: min,
        title: {
            display: !!label,
            text: label
        }
    };

    if (max !== null) {
        config.max = max;
    }

    if (tickCallback) {
        config.ticks = { callback: tickCallback };
    }

    return config;
}

/**
 * Get standard scales configuration
 * @param {string} yLabel - Y-axis label
 * @param {Object} yConfig - Additional Y-axis config
 */
function getStandardScalesConfig(yLabel = '', yConfig = {}) {
    return {
        y: {
            ...getStandardYAxisConfig(yLabel),
            ...yConfig
        },
        x: {
            grid: {
                display: false
            }
        }
    };
}

/**
 * Get standard plugins configuration
 * @param {string} title - Chart title
 * @param {boolean} showLegend - Whether to show legend (default: true)
 * @param {string} legendPosition - Legend position (default: 'top')
 */
function getStandardPluginsConfig(title = '', showLegend = true, legendPosition = 'top') {
    return {
        title: {
            display: !!title,
            text: title,
            font: { size: 14 }
        },
        legend: {
            display: showLegend,
            position: legendPosition
        }
    };
}

/**
 * Get complete base configuration for bar charts
 * @param {Object} data - Chart data
 * @param {string} title - Chart title
 * @param {string} yLabel - Y-axis label
 * @param {Object} customOptions - Additional custom options
 */
function getBaseBarChartConfig(data, title = '', yLabel = '', customOptions = {}) {
    return {
        type: 'bar',
        data: data,
        options: {
            ...getResponsiveOptions(),
            plugins: getStandardPluginsConfig(title),
            scales: getStandardScalesConfig(yLabel),
            ...customOptions
        }
    };
}

/**
 * Get complete base configuration for line charts
 * @param {Object} data - Chart data
 * @param {string} title - Chart title
 * @param {string} yLabel - Y-axis label
 * @param {Object} customOptions - Additional custom options
 */
function getBaseLineChartConfig(data, title = '', yLabel = '', customOptions = {}) {
    return {
        type: 'line',
        data: data,
        options: {
            ...getResponsiveOptions(),
            plugins: getStandardPluginsConfig(title),
            scales: getStandardScalesConfig(yLabel),
            tension: 0.4, // Smooth lines
            ...customOptions
        }
    };
}

/**
 * Get complete base configuration for doughnut charts
 * @param {Object} data - Chart data
 * @param {string} title - Chart title
 * @param {Object} customOptions - Additional custom options
 */
function getBaseDoughnutChartConfig(data, title = '', customOptions = {}) {
    return {
        type: 'doughnut',
        data: data,
        options: {
            ...getResponsiveOptions(),
            plugins: {
                ...getStandardPluginsConfig(title),
                tooltip: {
                    callbacks: {
                        label: (context) => {
                            const label = context.label || '';
                            const value = context.parsed || 0;
                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                            const percentage = ((value / total) * 100).toFixed(1);
                            return `${label}: ${value} (${percentage}%)`;
                        }
                    }
                }
            },
            ...customOptions
        }
    };
}

/**
 * Export multiple charts to base64 images
 * @param {Array<string>} chartKeys - Array of chart keys to export
 * @param {Object} charts - Global charts object
 * @returns {Object} - Map of chart key to base64 image
 */
function exportChartsToBase64(chartKeys, charts) {
    const images = {};

    chartKeys.forEach(key => {
        if (charts[key]) {
            images[key] = charts[key].toBase64Image('image/png', 1);
        }
    });

    return images;
}

/**
 * Export charts with custom key mapping
 * @param {Object} chartMapping - Map of export key to chart key
 * @param {Object} charts - Global charts object
 * @returns {Object} - Map of export key to base64 image
 */
function collectChartImages(chartMapping, charts) {
    const images = {};

    Object.entries(chartMapping).forEach(([exportKey, chartKey]) => {
        if (charts[chartKey]) {
            images[exportKey] = charts[chartKey].toBase64Image('image/png', 1);
        }
    });

    return images;
}

/**
 * Create dataset for bar chart with standard styling
 * @param {string} label - Dataset label
 * @param {Array} data - Data values
 * @param {string} color - Bar color
 * @param {Object} customConfig - Additional dataset config
 */
function createBarDataset(label, data, color, customConfig = {}) {
    return {
        label: label,
        data: data,
        backgroundColor: color,
        borderColor: color,
        borderWidth: 1,
        ...customConfig
    };
}

/**
 * Create dataset for line chart with standard styling
 * @param {string} label - Dataset label
 * @param {Array} data - Data values
 * @param {string} color - Line color
 * @param {Object} customConfig - Additional dataset config
 */
function createLineDataset(label, data, color, customConfig = {}) {
    return {
        label: label,
        data: data,
        borderColor: color,
        backgroundColor: color + '33', // Add transparency
        fill: false,
        tension: 0.4,
        ...customConfig
    };
}

/**
 * Create CER/WER percentage formatter for tooltips
 * @param {string} metricName - Metric name (CER, WER, etc.)
 */
function getCERTooltipFormatter(metricName = 'CER') {
    return (context) => {
        const label = context.dataset.label || '';
        const value = context.parsed.y;
        return `${label} ${metricName}: ${value.toFixed(2)}%`;
    };
}

/**
 * Create percentage tick formatter for Y-axis
 */
function getPercentageTickFormatter() {
    return (value) => `${value}%`;
}
