/**
 * Transcription Review & Analysis - Streamlined Inline Editing
 * Manages the admin interface for reviewing transcriptions with inline editing
 * Note: Uses showNotification() and showConfirmation() from common.js
 * ES6+ Compliant
 */

let dataTable = null;
let drillDownContext = {}; // Stores additional drill-down parameters from Analytics

const elements = {
    tbody: $('#transcriptionsBody'),
    table: $('#transcriptionsTable'),
    overlay: $('#tableLoadingOverlay'),
    loadingMessage: $('#loadingMessage'),
    filterUser: $('#filterUser'),
    filterWord: $('#filterWord'),
    filterAttribution: $('#filterAttribution'),
    filterReview: $('#filterReview'),
    btnApplyFilter: $('#btnApplyFilter'),
    btnExportLaTeX: $('#btnExportLaTeX'),
    btnRecalculateAll: $('#btnRecalculateAll')
};

const ATTRIBUTION_CONFIG = {
    'ASR_ERROR': { label: 'ASR', class: 'bg-danger', title: 'ASR Error' },
    'USER_ERROR': { label: 'USER', class: 'bg-primary', title: 'User Error' },
    'AMBIGUOUS': { label: 'AMB', class: 'bg-warning text-dark', title: 'Ambiguous' },
    'CORRECT': { label: 'OK', class: 'bg-success', title: 'Correct' }
};

const showTableOverlay = (message) => {
    elements.loadingMessage.html(message);
    elements.overlay.show();
};

const hideTableOverlay = () => {
    elements.overlay.hide();
};

const formatAttributionBadge = (attribution) => {
    if (!attribution) {
        return '<small class="text-muted">-</small>';
    }

    const config = ATTRIBUTION_CONFIG[attribution];
    if (!config) {
        return `<small>${attribution}</small>`;
    }

    return `<span class="badge badge-sm ${config.class}" title="${config.title}">${config.label}</span>`;
};

function getFilters() {
    return {
        userID: elements.filterUser.val(),
        wordText: elements.filterWord.val(),
        errorAttribution: elements.filterAttribution.val(),
        reviewStatus: elements.filterReview.val()
    };
}

function loadData() {
    // If DataTable already exists, reload it with current filters
    if (dataTable) {
        dataTable.ajax.reload();
    } else {
        // Initialize DataTable on first load
        initializeDataTable();
    }
}

function initializeDataTable() {
    // Destroy existing DataTable if any
    if (dataTable) {
        dataTable.destroy();
        dataTable = null;
    }
    elements.tbody.empty();

    // Initialize DataTable with server-side processing
    dataTable = elements.table.DataTable({
        serverSide: true,
        processing: true,
        ajax: function(data, callback, settings) {
            const filters = getFilters();
            const params = new URLSearchParams({
                ...filters,
                ...drillDownContext, // Include drill-down parameters (cerRange, modelWinner)
                start: data.start,
                length: data.length
            });

            showTableOverlay('<span class="en">Loading transcriptions...</span><span class="cy">Llwytho trawsgrifiadau...</span>');

            fetch(`/Admin/Transcriptions/Get?${params}`)
                .then(response => response.json())
                .then(result => {
                    hideTableOverlay();
                    if (result.success) {
                        callback({
                            draw: data.draw,
                            recordsTotal: result.recordsTotal,
                            recordsFiltered: result.recordsFiltered,
                            data: result.data
                        });
                    } else {
                        showNotification(`Error loading data: ${result.error}`, 'error');
                        callback({ draw: data.draw, recordsTotal: 0, recordsFiltered: 0, data: [] });
                    }
                })
                .catch(error => {
                    hideTableOverlay();
                    console.error('Load data error:', error);
                    showNotification('Failed to load transcriptions', 'error');
                    callback({ draw: data.draw, recordsTotal: 0, recordsFiltered: 0, data: [] });
                });
        },
        columns: [
            { // Audio player
                data: null,
                orderable: false,
                render: function(data, type, row) {
                    const filename = row.Recording ? row.Recording.split('/').pop() : '';
                    const audioSrc = filename ? `/PlayRecording?filename=${filename}` : '';
                    return audioSrc
                        ? `<audio controls preload="none" src="${audioSrc}" style="width: 40px; height: 30px;"></audio>`
                        : '<span class="text-muted">-</span>';
                }
            },
            { // Target word - simplified (no reassignment)
                data: 'Text',
                render: function(data, type, row) {
                    return `<span class="badge bg-light text-dark">${data}</span>`;
                }
            },
            { // Human transcription (editable)
                data: null,
                orderable: false,
                render: function(data, type, row) {
                    const humanText = row.HumanTranscription || '';
                    const isComplete = humanText.trim() !== '';
                    const cellClass = isComplete ? 'table-success' : 'table-warning';
                    const currentLang = getCookie('lang') || 'cy';
                    const placeholderText = currentLang === 'en'
                        ? '<em class="text-muted">Click to add</em>'
                        : '<em class="text-muted">Cliciwch i ychwanegu</em>';
                    const checkmarkBtn = !isComplete
                        ? `<button class="btn btn-sm btn-success btn-mark-correct me-2" data-entry-id="${row.ID}" data-target-word="${row.Text}" title="Mark as correct">✓</button>`
                        : '';
                    return checkmarkBtn + (humanText || placeholderText);
                },
                createdCell: function(td, cellData, rowData) {
                    const humanText = rowData.HumanTranscription || '';
                    const isComplete = humanText.trim() !== '';
                    $(td).addClass('editable-transcription').addClass(isComplete ? 'table-success' : 'table-warning');
                    $(td).attr('data-entry-id', rowData.ID).css('cursor', 'pointer').attr('title', 'Click to edit');
                }
            },
            // Whisper STRICT columns
            {
                data: 'AttemptWhisper',
                render: (data) => `<small>${data || '-'}</small>`,
                createdCell: (td) => $(td).addClass('asr-col')
            },
            { data: 'WERWhisper', render: (data) => `<small>${data ? data.toFixed(1) : '-'}</small>` },
            { data: 'CERWhisper', render: (data) => `<small>${data ? data.toFixed(1) : '-'}</small>` },
            { data: 'ErrorAttributionWhisper', render: (data) => formatAttributionBadge(data) },
            // Whisper LENIENT columns
            {
                data: 'NormalizedWhisper',
                render: (data) => `<small class="text-success">${data || '-'}</small>`,
                createdCell: (td) => $(td).addClass('lenient-metric-col asr-col')
            },
            {
                data: 'WERWhisperLenient',
                render: (data) => `<small class="text-success">${data ? data.toFixed(1) : '-'}</small>`,
                createdCell: (td) => $(td).addClass('lenient-metric-col')
            },
            {
                data: 'CERWhisperLenient',
                render: (data) => `<small class="text-success">${data ? data.toFixed(1) : '-'}</small>`,
                createdCell: (td) => $(td).addClass('lenient-metric-col')
            },
            {
                data: 'ErrorAttributionWhisperLenient',
                render: (data) => formatAttributionBadge(data),
                createdCell: (td) => $(td).addClass('lenient-metric-col')
            },
            // Wav2Vec2 STRICT columns
            {
                data: 'AttemptWav2Vec2',
                render: (data) => `<small>${data || '-'}</small>`,
                createdCell: (td) => $(td).addClass('asr-col')
            },
            { data: 'WERWav2Vec2', render: (data) => `<small>${data ? data.toFixed(1) : '-'}</small>` },
            { data: 'CERWav2Vec2', render: (data) => `<small>${data ? data.toFixed(1) : '-'}</small>` },
            { data: 'ErrorAttributionWav2Vec2', render: (data) => formatAttributionBadge(data) },
            // Wav2Vec2 LENIENT columns
            {
                data: 'NormalizedWav2Vec2',
                render: (data) => `<small class="text-success">${data || '-'}</small>`,
                createdCell: (td) => $(td).addClass('lenient-metric-col asr-col')
            },
            {
                data: 'WERWav2Vec2Lenient',
                render: (data) => `<small class="text-success">${data ? data.toFixed(1) : '-'}</small>`,
                createdCell: (td) => $(td).addClass('lenient-metric-col')
            },
            {
                data: 'CERWav2Vec2Lenient',
                render: (data) => `<small class="text-success">${data ? data.toFixed(1) : '-'}</small>`,
                createdCell: (td) => $(td).addClass('lenient-metric-col')
            },
            {
                data: 'ErrorAttributionWav2Vec2Lenient',
                render: (data) => formatAttributionBadge(data),
                createdCell: (td) => $(td).addClass('lenient-metric-col')
            },
            { // Actions column
                data: null,
                orderable: false,
                render: function(data, type, row) {
                    return `<button class="btn btn-sm btn-danger btn-delete-entry"
                                    data-entry-id="${row.ID}"
                                    title="Delete entry">
                                <i class="bi bi-trash"></i>
                            </button>`;
                }
            }
        ],
        paging: true,
        pageLength: 25,
        lengthMenu: [[10, 25, 50, 100], [10, 25, 50, 100]],
        searching: false,
        order: [], // Server-side sorting: entries needing review (empty HumanTranscription) first, then by ID desc
        deferRender: true,
        language: {
            emptyTable: "No entries found / Dim cofnodion",
            info: "Showing _START_ to _END_ of _TOTAL_ entries",
            infoEmpty: "Showing 0 to 0 of 0 entries",
            lengthMenu: "Show _MENU_ entries",
            processing: '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div>',
            paginate: {
                first: "First",
                last: "Last",
                next: "Next",
                previous: "Previous"
            }
        }
    });

    // Render icons and reinitialize language toggle after adding dynamic content
    renderIconsAndLanguage();
}

function enableInlineEdit(cell) {
    const entryId = cell.data('entry-id');
    const currentValue = cell.find('em').length ? '' : cell.text().trim();

    // Create input element
    const input = $('<input>')
        .addClass('form-control form-control-sm')
        .val(currentValue)
        .attr('placeholder', 'Type transcription...');

    // Replace cell content with input
    cell.html(input);
    input.focus();

    // Save on blur or Enter
    input.on('blur', () => {
        saveInlineTranscription(entryId, input.val().trim(), cell);
    });

    input.on('keypress', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            saveInlineTranscription(entryId, input.val().trim(), cell);
        }
    });

    // Cancel on Escape
    input.on('keydown', (e) => {
        if (e.key === 'Escape') {
            e.preventDefault();
            cell.html(currentValue || '<em class="text-muted"><span class="en">Click to add</span><span class="cy">Cliciwch i ychwanegu</span></em>');
            cell.removeClass('table-success').addClass(currentValue ? 'table-success' : 'table-warning');
        }
    });
}

async function saveInlineTranscription(entryId, humanText, cell) {
    if (!humanText) {
        // Visual feedback only - cell stays yellow/warning
        cell.html('<em class="text-muted"><span class="en">Click to add</span><span class="cy">Cliciwch i ychwanegu</span></em>');
        cell.removeClass('table-success').addClass('table-warning');
        return;
    }

    // Show spinner in transcription cell
    cell.html('<span class="spinner-border spinner-border-sm" role="status"></span>');
    cell.removeClass('table-warning table-success').addClass('table-light');

    // Show spinners in both attribution columns
    const row = cell.closest('tr');
    const whisperAttrCell = row.find('td').eq(6); // Whisper Attr column
    const wav2vec2AttrCell = row.find('td').eq(10); // Wav2Vec2 Attr column
    whisperAttrCell.html('<span class="spinner-border spinner-border-sm" role="status"></span>');
    wav2vec2AttrCell.html('<span class="spinner-border spinner-border-sm" role="status"></span>');

    try {
        const formData = new URLSearchParams({
            entryID: entryId,
            humanTranscription: humanText
        });

        const response = await fetch('/Admin/Transcriptions/Update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: formData
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        if (data.success) {
            // Display saved value with success styling
            cell.html(humanText);
            cell.removeClass('table-light table-warning').addClass('table-success');

            // Update metric cells in the current row (no need to reload entire table)
            row.find('td').eq(4).html(`<small>${data.werWhisper ? data.werWhisper.toFixed(1) : '-'}</small>`);
            row.find('td').eq(5).html(`<small>${data.cerWhisper ? data.cerWhisper.toFixed(1) : '-'}</small>`);
            row.find('td').eq(6).html(formatAttributionBadge(data.errorAttributionWhisper));
            row.find('td').eq(8).html(`<small>${data.werWav2Vec2 ? data.werWav2Vec2.toFixed(1) : '-'}</small>`);
            row.find('td').eq(9).html(`<small>${data.cerWav2Vec2 ? data.cerWav2Vec2.toFixed(1) : '-'}</small>`);
            row.find('td').eq(10).html(formatAttributionBadge(data.errorAttributionWav2Vec2));
        } else {
            throw new Error(data.error || 'Save failed');
        }
    } catch (error) {
        console.error('Save transcription error:', error);
        cell.html(humanText);
        cell.removeClass('table-light table-success').addClass('table-danger');
        showNotification('Failed to save / Methwyd cadw', 'error');
    }
}

function exportData() {
    const filters = {
        userID: elements.filterUser.val(),
        wordText: elements.filterWord.val(),
        errorAttribution: elements.filterAttribution.val(),
        reviewStatus: elements.filterReview.val()
    };

    const url = `/Admin/Transcriptions/Export?${$.param(filters)}`;
    window.location.href = url;
}

async function recalculateAllMetrics() {
    const confirmMessage = 'This will recalculate WER/CER metrics for all existing entries. This may take a while. Continue?\n\nBydd hyn yn ailgyfrifo metrigau WER/CER ar gyfer pob cofnod presennol. Gall hyn gymryd amser. Parhau?';

    showConfirmation(confirmMessage, async () => {
        showTableOverlay('<span class="en">Recalculating metrics...</span><span class="cy">Ailgyfrifo metrigau...</span>');

        try {
            const response = await fetch('/Admin/Transcriptions/RecalculateAll', {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (data.success) {
                // Success - reload table to show updated entries
                loadData();
            } else {
                throw new Error(data.error || 'Recalculation failed');
            }
        } catch (error) {
            console.error('Recalculate metrics error:', error);
            hideTableOverlay();
            showNotification('Failed to recalculate metrics / Methwyd ailgyfrifo metrigau', 'error');
        }
    });
}

async function deleteEntry(entryId, rowElement) {
    const confirmMessage = 'Are you sure you want to delete this entry? This cannot be undone.\n\nYdych chi\'n siŵr eich bod am ddileu\'r cofnod hwn? Ni ellir dadwneud hyn.';

    showConfirmation(confirmMessage, async () => {
        try {
            const response = await fetch('/Admin/Transcriptions/Delete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                },
                body: new URLSearchParams({ entryID: entryId })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (data.success) {
                // Remove row from DataTable - visual feedback is the row disappearing
                dataTable.row(rowElement).remove().draw();
            } else {
                throw new Error(data.error || 'Delete failed');
            }
        } catch (error) {
            console.error('Delete entry error:', error);
            showNotification('Failed to delete entry / Methwyd dileu cofnod', 'error');
        }
    });
}

function applyURLFilters() {
    // Read URL parameters from hidden inputs (set by Analytics drill-down)
    const urlParamWord = $('#urlParamWord').val();
    const urlParamCerRange = $('#urlParamCerRange').val();
    const urlParamErrorAttribution = $('#urlParamErrorAttribution').val();
    const urlParamModelWinner = $('#urlParamModelWinner').val();

    let filtersApplied = false;

    // Apply word filter if present
    if (urlParamWord && urlParamWord.trim() !== '') {
        elements.filterWord.val(urlParamWord);
        filtersApplied = true;
    }

    // Apply error attribution filter if present
    if (urlParamErrorAttribution && urlParamErrorAttribution !== 'ALL' && urlParamErrorAttribution.trim() !== '') {
        elements.filterAttribution.val(urlParamErrorAttribution);
        filtersApplied = true;
    }

    // Store drill-down context (parameters without UI filter controls)
    drillDownContext = {};
    if (urlParamCerRange && urlParamCerRange.trim() !== '') {
        drillDownContext.cerRange = urlParamCerRange;
        filtersApplied = true;
    }
    if (urlParamModelWinner && urlParamModelWinner.trim() !== '') {
        drillDownContext.modelWinner = urlParamModelWinner;
        filtersApplied = true;
    }

    // If filters were applied, show notification with context
    if (filtersApplied) {
        let messageParts = ['Filters applied from Analytics drill-down / Hidlyddion wedi\'u cymhwyso o Analytics'];

        // Add context about CER range if present
        if (drillDownContext.cerRange) {
            messageParts.push(`CER Range: ${drillDownContext.cerRange} / Ystod CER: ${drillDownContext.cerRange}`);
        }

        // Add context about model winner if present
        if (drillDownContext.modelWinner) {
            const modelName = drillDownContext.modelWinner === 'whisper' ? 'Whisper' : 'Wav2Vec2';
            messageParts.push(`Context: ${modelName} had lower error / Cyd-destun: ${modelName} gyda gwall is`);
        }

        showNotification(messageParts.join('. '), 'info');
    }

    return filtersApplied;
}

$(() => {
    // Apply URL filters from Analytics drill-down (if any)
    applyURLFilters();

    // Load data (with or without URL filters)
    loadData();

    elements.btnApplyFilter.click(() => {
        loadData();
    });

    elements.btnExportLaTeX.click(() => {
        exportData();
    });

    elements.btnRecalculateAll.click(() => {
        recalculateAllMetrics();
    });

    elements.tbody.on('click', '.editable-transcription', (e) => {
        // Don't trigger edit if checkmark button was clicked
        if ($(e.target).hasClass('btn-mark-correct')) {
            return;
        }
        enableInlineEdit($(e.currentTarget));
    });

    elements.tbody.on('click', '.btn-mark-correct', (e) => {
        e.stopPropagation();
        const targetWord = $(e.currentTarget).data('target-word');
        const entryId = $(e.currentTarget).data('entry-id');
        const cell = $(e.currentTarget).closest('td');
        // Replace button with target word and trigger save
        $(e.currentTarget).remove();
        cell.html(targetWord);
        cell.removeClass('table-warning').addClass('table-light');
        saveInlineTranscription(entryId, targetWord, cell);
    });

    elements.tbody.on('click', '.btn-delete-entry', function(e) {
        e.stopPropagation();
        const entryId = $(this).data('entry-id');
        const row = $(this).closest('tr');
        deleteEntry(entryId, row);
    });

});
