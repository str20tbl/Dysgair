/**
 * Recording Recovery Dashboard
 * Handles orphaned recording recovery with auto-verification and manual matching
 */

// State
let currentRecording = null;
let selectedWordID = null;
let autoVerifiedThisSession = 0;
let manualVerifiedThisSession = 0;
let skippedThisSession = 0;
let allWords = []; // All 880 words for search filtering
let currentMatches = []; // Current ASR-based matches
let progressPollInterval = null; // Interval for polling progress
let manualReviewList = []; // List of recordings needing manual review
let manualReviewIndex = 0; // Current index in manual review list

// Elements
const elements = {
    // Session setup
    sessionSetupPanel: $('#sessionSetupPanel'),
    selectUser: $('#selectUser'),
    btnStartSession: $('#btnStartSession'),

    // Recovery panel
    recoveryPanel: $('#recoveryPanel'),
    progressBar: $('#progressBar'),
    progressPercent: $('#progressPercent'),
    progressText: $('#progressText'),
    autoVerifiedBadge: $('#autoVerifiedBadge'),
    manualVerifiedBadge: $('#manualVerifiedBadge'),
    skippedBadge: $('#skippedBadge'),
    btnResetSession: $('#btnResetSession'),

    // Auto-verify notification
    autoVerifyNotification: $('#autoVerifyNotification'),
    autoVerifyCount: $('#autoVerifyCount'),

    // Completion
    completionMessage: $('#completionMessage'),
    recordingArea: $('#recordingArea'),

    // Recording area
    audioPlayer: $('#audioPlayer'),
    recordingFileName: $('#recordingFileName'),
    whisperText: $('#whisperText'),
    wav2vec2Text: $('#wav2vec2Text'),

    // Matches
    matchesTableBody: $('#matchesTableBody'),
    wordSearch: $('#wordSearch'),
    selectedWordPanel: $('#selectedWordPanel'),
    editWordText: $('#editWordText'),
    editWordEnglish: $('#editWordEnglish'),
    btnUpdateWord: $('#btnUpdateWord'),

    // Actions
    btnVerifyMatch: $('#btnVerifyMatch'),
    btnSkip: $('#btnSkip')
};

/**
 * Initialize the dashboard
 */
function init() {
    // Session setup
    elements.btnStartSession.on('click', startSession);
    elements.btnResetSession.on('click', resetSession);

    // Actions
    elements.btnVerifyMatch.on('click', verifyMatch);
    elements.btnSkip.on('click', skipRecording);
    elements.btnUpdateWord.on('click', updateWord);

    // Keyboard shortcuts
    $(document).on('keydown', handleKeyboard);

    // Table row click
    $(document).on('click', '#matchesTableBody tr', function() {
        const wordID = $(this).data('word-id');
        if (wordID) {
            selectWord(wordID, $(this));
        }
    });

    // Word search input handler
    elements.wordSearch.on('input', function() {
        const query = $(this).val().trim();
        searchWords(query);
    });

    // Load all words for searching
    loadAllWords();

    // Resume session if recovery panel is visible (page refresh during active session)
    if (elements.recoveryPanel.is(':visible')) {
        resumeSession();
    }
}

/**
 * Resume session after page refresh
 */
function resumeSession() {
    // Check progress to see what state we're in
    fetch('/Admin/Recovery/Progress')
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                if (result.status === 'initializing' || (result.status === 'processing' && !result.complete)) {
                    // Background job still running - show processing message and poll
                    showProcessingMessage();
                    pollProgress();
                } else if (result.complete) {
                    // Job complete - load manual review list
                    autoVerifiedThisSession = result.auto_matched || 0;
                    updateBadges();
                    // Remove any processing messages from previous state
                    $('.alert-info').remove();
                    loadManualReviewList();
                } else {
                    // Unknown state - show error
                    showNotification('Unable to resume session. Please reset and start over.', 'error');
                }
            } else {
                // No active progress - session might be expired
                console.log('No active recovery session found');
            }
        })
        .catch(err => {
            console.error('Failed to resume session:', err);
            showNotification('Failed to resume session', 'error');
        });
}

/**
 * Load all words for search filtering
 */
function loadAllWords() {
    fetch('/Admin/Recovery/Words')
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                allWords = result.words || [];
                console.log(`Loaded ${allWords.length} words for searching`);
            } else {
                console.error('Failed to load words:', result.error);
            }
        })
        .catch(err => {
            console.error('Failed to load words:', err);
        });
}

/**
 * Search/filter words based on query
 */
function searchWords(query) {
    if (!query || query.length < 2) {
        // Show ASR-based matches again
        displayMatches(currentMatches);
        return;
    }

    // Filter all words by search query (case-insensitive)
    const lowerQuery = query.toLowerCase();
    const filtered = allWords.filter(word =>
        word.Text.toLowerCase().includes(lowerQuery) ||
        word.English.toLowerCase().includes(lowerQuery)
    );

    // Convert to match format and display
    const searchResults = filtered.map(word => ({
        WordID: word.ID,
        Word: word.Text,
        English: word.English,
        Distance: 0,
        Score: 1.0
    }));

    displayMatches(searchResults);
}

/**
 * Start recovery session
 */
function startSession() {
    const userID = elements.selectUser.val();
    if (!userID) {
        showNotification('Please select a user', 'error');
        return;
    }

    const params = new URLSearchParams({ userID });

    elements.btnStartSession.prop('disabled', true).text('Starting...');

    fetch(`/Admin/Recovery/Start?${params}`, { method: 'POST' })
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                showNotification(`Recovery session started for ${result.userName} - processing recordings in background`, 'success');

                // Switch panels
                elements.sessionSetupPanel.hide();
                elements.recoveryPanel.show();

                // Reset counters
                autoVerifiedThisSession = 0;
                manualVerifiedThisSession = 0;
                skippedThisSession = 0;
                manualReviewList = [];
                manualReviewIndex = 0;
                updateBadges();

                // Show processing message
                elements.recordingArea.hide();
                elements.completionMessage.hide();
                showProcessingMessage();

                // Start polling progress
                pollProgress();
            } else {
                showNotification(result.error || 'Failed to start session', 'error');
            }
        })
        .catch(err => {
            console.error('Failed to start session:', err);
            showNotification('Failed to start session', 'error');
        })
        .finally(() => {
            elements.btnStartSession.prop('disabled', false).html('<span class="en">Start Recovery Session</span><span class="cy">Dechrau Sesiwn Adfer</span>');
        });
}

/**
 * Load next orphaned recording
 */
function loadNextRecording() {
    // Clear current state
    currentRecording = null;
    selectedWordID = null;
    elements.btnVerifyMatch.prop('disabled', true);
    elements.selectedWordPanel.hide();
    elements.autoVerifyNotification.hide();

    fetch('/Admin/Recovery/Next')
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                if (result.completed) {
                    // All done!
                    showCompletion(result.autoVerifiedCount);
                    return;
                }

                if (result.autoVerifyLimitReached) {
                    // Auto-verified batch completed, load next
                    showAutoVerifyNotification(result.autoVerifiedCount);
                    autoVerifiedThisSession += result.autoVerifiedCount;
                    updateBadges();
                    // Automatically load next
                    setTimeout(() => loadNextRecording(), 1000);
                    return;
                }

                // Show auto-verify count if any
                if (result.autoVerifiedCount > 0) {
                    showAutoVerifyNotification(result.autoVerifiedCount);
                    autoVerifiedThisSession += result.autoVerifiedCount;
                    updateBadges();
                }

                // Update progress
                if (result.progress) {
                    updateProgress(result.progress);
                }

                // Load recording
                if (result.recording) {
                    currentRecording = result.recording;
                    displayRecording(result.recording);

                    // Store ASR-based matches and display them
                    currentMatches = result.matches || [];
                    displayMatches(currentMatches);

                    // Clear search input for new recording
                    elements.wordSearch.val('');
                }
            } else {
                showNotification(result.error || 'Failed to load next recording', 'error');
            }
        })
        .catch(err => {
            console.error('Failed to load next recording:', err);
            showNotification('Failed to load next recording', 'error');
        });
}

/**
 * Display recording details
 */
function displayRecording(recording) {
    elements.recordingFileName.text(recording.fileName);
    elements.whisperText.text(recording.whisperText || '-');
    elements.wav2vec2Text.text(recording.wav2vec2Text || '-');

    // Load audio
    elements.audioPlayer.attr('src', `/PlayRecording?filename=${encodeURIComponent(recording.fileName)}`);
    elements.audioPlayer[0].load();
}

/**
 * Display word matches
 */
function displayMatches(matches) {
    elements.matchesTableBody.empty();

    if (matches.length === 0) {
        elements.matchesTableBody.append(`
            <tr>
                <td colspan="3" class="text-center text-muted">
                    <span class="en">No matches found</span>
                    <span class="cy">Dim parau wedi'u darganfod</span>
                </td>
            </tr>
        `);
        return;
    }

    matches.forEach((match, index) => {
        const scorePercent = (match.Score * 100).toFixed(1);
        const rowClass = match.Distance === 0 ? 'table-success' : '';
        const exactBadge = match.Distance === 0 ? '<span class="badge badge-success badge-sm ml-2">EXACT</span>' : '';

        const row = $(`
            <tr class="${rowClass}" data-word-id="${match.WordID}" style="cursor: pointer;">
                <td><strong>${match.Word}</strong>${exactBadge}</td>
                <td>${match.English}</td>
                <td>
                    <span class="badge badge-${scorePercent >= 90 ? 'success' : scorePercent >= 70 ? 'warning' : 'secondary'}">
                        ${scorePercent}%
                    </span>
                </td>
            </tr>
        `);
        elements.matchesTableBody.append(row);
        // Auto-select first match
        if (index === 0) {
            row.addClass('table-active');
            selectWord(match.WordID, row, false);
        }
    });
}

/**
 * Select a word from matches
 */
function selectWord(wordID, row, scrollToRow = true) {
    selectedWordID = wordID;

    // Highlight row
    $('#matchesTableBody tr').removeClass('table-active');
    row.addClass('table-active');

    // Scroll to row
    if (scrollToRow) {
        row[0].scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }

    // Populate edit fields
    const wordText = row.find('td:eq(0) strong').text();
    const english = row.find('td:eq(1)').text();

    elements.editWordText.val(wordText);
    elements.editWordEnglish.val(english);
    elements.selectedWordPanel.show();

    // Enable verify button and reset text with correct language
    const currentLang = getCookie('lang') || 'cy';
    const verifyText = currentLang === 'en' ? '✓ Verify & Next' : '✓ Dilysu & Nesaf';
    elements.btnVerifyMatch.prop('disabled', false).text(verifyText);
}

/**
 * Verify match and create entry
 */
function verifyMatch() {
    if (!currentRecording || !selectedWordID) {
        showNotification('No word selected', 'error');
        return;
    }

    const params = new URLSearchParams({
        recordingPath: currentRecording.filePath,
        wordID: selectedWordID
    });

    elements.btnVerifyMatch.prop('disabled', true).text('Verifying...');

    fetch(`/Admin/Recovery/Verify?${params}`, { method: 'POST' })
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                manualVerifiedThisSession++;
                updateBadges();

                // Load next manual review item
                loadManualReviewItem(manualReviewIndex + 1);
            } else {
                showNotification(result.error || 'Failed to verify match', 'error');
                const currentLang = getCookie('lang') || 'cy';
                const verifyText = currentLang === 'en' ? '✓ Verify & Next' : '✓ Dilysu & Nesaf';
                elements.btnVerifyMatch.prop('disabled', false).text(verifyText);
            }
        })
        .catch(err => {
            console.error('Failed to verify match:', err);
            showNotification('Failed to verify match', 'error');
            const currentLang = getCookie('lang') || 'cy';
            const verifyText = currentLang === 'en' ? '✓ Verify & Next' : '✓ Dilysu & Nesaf';
            elements.btnVerifyMatch.prop('disabled', false).text(verifyText);
        });
}

/**
 * Skip current recording
 */
function skipRecording() {
    if (!currentRecording) {
        return;
    }

    const params = new URLSearchParams({
        recordingPath: currentRecording.filePath
    });

    fetch(`/Admin/Recovery/Skip?${params}`, { method: 'POST' })
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                skippedThisSession++;
                updateBadges();

                // Load next manual review item
                loadManualReviewItem(manualReviewIndex + 1);
            } else {
                showNotification(result.error || 'Failed to skip recording', 'error');
            }
        })
        .catch(err => {
            console.error('Failed to skip recording:', err);
            showNotification('Failed to skip recording', 'error');
        });
}

/**
 * Update word Text and English fields
 */
function updateWord() {
    if (!selectedWordID) {
        return;
    }

    const text = elements.editWordText.val().trim();
    const english = elements.editWordEnglish.val().trim();

    if (!text || !english) {
        showNotification('Both Welsh text and English translation are required', 'error');
        return;
    }

    const params = new URLSearchParams({
        wordID: selectedWordID,
        text: text,
        english: english
    });

    elements.btnUpdateWord.prop('disabled', true).text('Updating...');

    fetch(`/Admin/Recovery/UpdateWord?${params}`, { method: 'POST' })
        .then(response => response.json())
        .then(result => {
            if (result.success) {
                showNotification('Word updated successfully', 'success');

                // Update the displayed match
                const selectedRow = $('#matchesTableBody tr.table-active');
                selectedRow.find('td:eq(0) strong').text(result.word.text);
                selectedRow.find('td:eq(1)').text(result.word.english);
            } else {
                showNotification(result.error || 'Failed to update word', 'error');
            }
        })
        .catch(err => {
            console.error('Failed to update word:', err);
            showNotification('Failed to update word', 'error');
        })
        .finally(() => {
            elements.btnUpdateWord.prop('disabled', false).html('<span class="en">Update Word</span><span class="cy">Diweddaru Gair</span>');
        });
}

/**
 * Reset recovery session
 */
function resetSession() {
    showConfirmation('Reset recovery session? This will clear all progress and you will start over.', () => {
        // Stop progress polling
        if (progressPollInterval) {
            clearInterval(progressPollInterval);
            progressPollInterval = null;
        }

        fetch('/Admin/Recovery/Reset', { method: 'POST' })
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    showNotification('Session reset successfully', 'success');

                    // Reset counters
                    autoVerifiedThisSession = 0;
                    manualVerifiedThisSession = 0;
                    skippedThisSession = 0;
                    manualReviewList = [];
                    manualReviewIndex = 0;
                    updateBadges();

                    // Switch panels
                    elements.recoveryPanel.hide();
                    elements.sessionSetupPanel.show();

                    // Clear recording area
                    currentRecording = null;
                    selectedWordID = null;
                    elements.audioPlayer.attr('src', '');
                    elements.matchesTableBody.empty();

                    // Remove any processing messages
                    $('.alert-info').remove();
                } else {
                    showNotification(result.error || 'Failed to reset session', 'error');
                }
            })
            .catch(err => {
                console.error('Failed to reset session:', err);
                showNotification('Failed to reset session', 'error');
            });
    });
}

/**
 * Update progress bar
 */
function updateProgress(progress) {
    const percent = progress.total > 0 ? (progress.processed / progress.total * 100).toFixed(1) : 0;

    elements.progressBar.css('width', percent + '%');
    elements.progressPercent.text(percent + '%');
    elements.progressText.text(`${progress.processed} / ${progress.total}`);
}

/**
 * Update badge counts
 */
function updateBadges() {
    elements.autoVerifiedBadge.text(`Auto: ${autoVerifiedThisSession}`);
    elements.manualVerifiedBadge.text(`Manual: ${manualVerifiedThisSession}`);
    elements.skippedBadge.text(`Skipped: ${skippedThisSession}`);
}

/**
 * Show auto-verify notification
 */
function showAutoVerifyNotification(count) {
    elements.autoVerifyCount.text(count);
    elements.autoVerifyNotification.fadeIn();

    setTimeout(() => {
        elements.autoVerifyNotification.fadeOut();
    }, 3000);
}

/**
 * Show completion message
 */
function showCompletion(autoVerifiedCount) {
    if (autoVerifiedCount > 0) {
        showAutoVerifyNotification(autoVerifiedCount);
        autoVerifiedThisSession += autoVerifiedCount;
        updateBadges();
    }

    elements.recordingArea.hide();
    elements.completionMessage.show();
}

/**
 * Show processing message while background job runs
 */
function showProcessingMessage() {
    elements.autoVerifyNotification.hide();
    const currentLang = getCookie('lang') || 'cy';
    const message = currentLang === 'en'
        ? '<div class="alert alert-info"><h5>Processing Recordings...</h5><p>Auto-matching recordings in the background. This may take a few minutes.</p></div>'
        : '<div class="alert alert-info"><h5>Prosesu Recordiadau...</h5><p>Paru recordiadau yn awtomatig yn y cefndir. Gall hyn gymryd ychydig funudau.</p></div>';

    elements.recordingArea.before(message);
}

/**
 * Poll progress of background recovery job
 */
function pollProgress() {
    // Clear any existing interval
    if (progressPollInterval) {
        clearInterval(progressPollInterval);
    }

    // Poll every second
    progressPollInterval = setInterval(() => {
        fetch('/Admin/Recovery/Progress')
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    // Update progress bar
                    updateProgress({
                        total: result.total,
                        processed: result.processed
                    });

                    // Update badges
                    autoVerifiedThisSession = result.auto_matched;
                    updateBadges();

                    // Check if complete
                    if (result.complete) {
                        clearInterval(progressPollInterval);
                        progressPollInterval = null;

                        // Remove processing message
                        $('.alert-info').remove();

                        // Load manual review list
                        loadManualReviewList();
                    }
                }
            })
            .catch(err => {
                console.error('Failed to poll progress:', err);
            });
    }, 1000);
}

/**
 * Load list of recordings needing manual review
 */
function loadManualReviewList() {
    // Fetch both the manual review list and processed recordings
    Promise.all([
        fetch('/Admin/Recovery/ManualReviewList').then(r => r.json()),
        fetch('/Admin/Recovery/ProcessedRecordings').then(r => r.json())
    ])
        .then(([reviewResult, processedResult]) => {
            if (reviewResult.success) {
                const allItems = reviewResult.items || [];
                const processedPaths = processedResult.success ? (processedResult.recordings || []) : [];

                // Filter out already processed items
                manualReviewList = allItems.filter(item => !processedPaths.includes(item.file_path));
                manualReviewIndex = 0;

                const skippedCount = allItems.length - manualReviewList.length;
                if (skippedCount > 0) {
                    console.log(`Filtered out ${skippedCount} already-processed items. Remaining: ${manualReviewList.length}`);
                }

                if (manualReviewList.length === 0) {
                    // No manual review needed - all done!
                    showCompletion(0);
                } else {
                    // Show first unprocessed manual review item
                    const message = skippedCount > 0
                        ? `Resuming manual review! ${manualReviewList.length} recordings remaining (skipped ${skippedCount} already processed).`
                        : `Auto-matching complete! ${manualReviewList.length} recordings need manual review.`;
                    showNotification(message, 'info');
                    loadManualReviewItem(0);
                }
            } else {
                showNotification(reviewResult.error || 'Failed to load manual review list', 'error');
            }
        })
        .catch(err => {
            console.error('Failed to load manual review list:', err);
            showNotification('Failed to load manual review list', 'error');
        });
}

/**
 * Load a specific manual review item
 */
function loadManualReviewItem(index) {
    if (index >= manualReviewList.length) {
        // All manual reviews complete!
        showCompletion(0);
        return;
    }

    const item = manualReviewList[index];
    manualReviewIndex = index;

    // Update progress to show manual review progress
    updateProgress({
        processed: index,
        total: manualReviewList.length
    });

    // Set current recording
    currentRecording = {
        filePath: item.file_path,
        fileName: item.file_name,
        whisperText: item.whisper_text,
        wav2vec2Text: item.wav2vec2_text
    };

    // Display recording
    elements.recordingArea.show();
    elements.completionMessage.hide();
    displayRecording(currentRecording);

    // Find top word matches based on transcriptions
    const transcription = item.whisper_text || item.wav2vec2_text;
    findAndDisplayMatches(transcription);
}

/**
 * Find and display word matches for a transcription
 */
function findAndDisplayMatches(transcription) {
    if (!transcription) {
        currentMatches = [];
        displayMatches([]);
        return;
    }

    // Use existing word matching logic (simple substring matching for now)
    const lowerTranscription = transcription.toLowerCase().trim();

    // Calculate match scores for all words
    const matches = allWords.map(word => {
        const lowerWord = word.Text.toLowerCase().trim();
        const lowerEnglish = word.English.toLowerCase().trim();

        // Simple matching: exact match = distance 0, contains = low distance, else high distance
        let distance = 100;
        if (lowerWord === lowerTranscription || lowerEnglish === lowerTranscription) {
            distance = 0;
        } else if (lowerWord.includes(lowerTranscription) || lowerTranscription.includes(lowerWord)) {
            distance = Math.abs(lowerWord.length - lowerTranscription.length);
        } else if (lowerEnglish.includes(lowerTranscription) || lowerTranscription.includes(lowerEnglish)) {
            distance = Math.abs(lowerEnglish.length - lowerTranscription.length);
        }

        const maxLen = Math.max(lowerTranscription.length, lowerWord.length);
        const score = maxLen > 0 ? 1.0 - (distance / maxLen) : 0;

        return {
            WordID: word.ID,
            Word: word.Text,
            English: word.English,
            Distance: distance,
            Score: score
        };
    });

    // Sort by distance (lowest first) and take top 10
    matches.sort((a, b) => a.Distance - b.Distance);
    currentMatches = matches.slice(0, 10);

    displayMatches(currentMatches);

    // Clear search input
    elements.wordSearch.val('');
}

/**
 * Handle keyboard shortcuts
 */
function handleKeyboard(e) {
    // Only handle when recovery panel is visible and not typing in input
    if (!elements.recoveryPanel.is(':visible') || $(e.target).is('input, textarea')) {
        return;
    }

    switch(e.key) {
        case 'Enter':
            e.preventDefault();
            if (!elements.btnVerifyMatch.prop('disabled')) {
                verifyMatch();
            }
            break;

        case 's':
        case 'S':
            e.preventDefault();
            skipRecording();
            break;

        case 'ArrowUp':
            e.preventDefault();
            navigateMatches(-1);
            break;

        case 'ArrowDown':
            e.preventDefault();
            navigateMatches(1);
            break;
    }
}

/**
 * Navigate through matches with arrow keys
 */
function navigateMatches(direction) {
    const rows = $('#matchesTableBody tr[data-word-id]');
    if (rows.length === 0) return;

    const currentIndex = rows.index($('#matchesTableBody tr.table-active'));
    let newIndex = currentIndex + direction;

    // Wrap around
    if (newIndex < 0) newIndex = rows.length - 1;
    if (newIndex >= rows.length) newIndex = 0;

    const newRow = rows.eq(newIndex);
    const wordID = newRow.data('word-id');
    selectWord(wordID, newRow, true);
}

// Initialize on page load
$(document).ready(init);
