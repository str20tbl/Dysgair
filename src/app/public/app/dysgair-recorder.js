/**
 * Dysgair Audio Recording Module
 * Handles audio recording, playback, and upload functionality
 * ES6+ Compliant
 */

$(() => {
    const elements = {
        startBtn: $('#start-recording'),
        stopBtn: $('#stop-recording'),
        playBtn: $('#playButton'),
        statusText: $('#status-text'),
        statusMessage: $('#status-message'),
        timer: $('#recording-timer'),
        wordDisplay: $('#wordDisplay'),
        feedbackBody: $('#feedback-table-body'),
        compactProgress: $('#compact-progress'),
        compactWordProgress: $('#compact-word-progress'),
        compactOverallBar: $('#compact-overall-bar'),
        overallPctLabel: $('#overall-pct-label'),
        compactWordBar: $('#compact-word-bar'),
        wordPctLabel: $('#word-pct-label'),
        nextBtn: $('#next-recording'),
        prevBtn: $('#previous-recording'),
        rewindBtn: $('#rewind-recording'),
        fastForwardBtn: $('#fast-forward-recording')
    };

    const mediaConstraints = { audio: true };
    let mediaRecorder = null;
    let recordingTimer = null;
    let recordingStartTime = 0;

    const bilingual = (cy, en) => `<span class="cy">${cy}</span><span class="en">${en}</span>`;

    const showError = (errorCy, errorEn) => {
        setRecordingStatus('idle');
        elements.statusMessage.html(`<span class="text-danger">${bilingual(errorCy, errorEn)}</span>`);
        enableRecordingButton();
    };

    function disableRecordingButton() {
        elements.startBtn.prop('disabled', true);
        elements.startBtn.find('i').attr('data-feather', 'loader');

        if (typeof feather !== 'undefined') {
            feather.replace();
        }
    }

    function enableRecordingButton() {
        elements.startBtn.prop('disabled', false);
        elements.startBtn.find('i').attr('data-feather', 'mic');

        if (typeof feather !== 'undefined') {
            feather.replace();
        }
    }

    // Longest common substring algorithm for colored diff display
    function generateColoredDiff(target, attempt) {
        if (!target || !attempt) {
            return '<span class="text-muted">—</span>';
        }

        const str1 = attempt.toLowerCase();
        const str2 = target.toLowerCase();

        let longest = ["", "", ""];
        for (let i = 0; i < str1.length; i++) {
            for (let j = 0; j < str2.length; j++) {
                let k = 0;
                while (i+k < str1.length && j+k < str2.length && str1[i+k] === str2[j+k]) {
                    k++;
                }
                if (k > longest[1].length) {
                    longest = [str1.substring(0, i), str1.substring(i, i+k), str1.substring(i+k)];
                }
            }
        }

        return `<span class="text-danger">${longest[0]}</span>` +
               `<span class="text-success">${longest[1]}</span>` +
               `<span class="text-danger">${longest[2]}</span>`;
    }

    function updateAllFeedback() {
        elements.feedbackBody.find('tr').each((_, row) => {
            const $cell = $(row).find('.coloured-feedback');
            const target = $cell.data('target');
            const attempt = $cell.data('whisper');
            $cell.html(generateColoredDiff(target, attempt));
        });
    }

    // Initialize colored feedback for all rows on page load
    $(document).ready(() => {
        updateAllFeedback();
    });

    function setRecordingStatus(state, message = '') {
        const statusStates = {
            idle: {
                text: bilingual('Barod i recordio', 'Ready to record'),
                message: '',
                timer: 'hide'
            },
            preparing: {
                text: bilingual('Yn paratoi...', 'Preparing...'),
                message: '',
                timer: 'show'
            },
            recording: {
                text: bilingual('Yn recordio...', 'Recording...'),
                message: '',
                timer: 'show'
            },
            uploading: {
                text: bilingual('Yn uwchlwytho...', 'Uploading...'),
                message,
                timer: 'hide',
                stopTimer: true
            },
            complete: {
                text: bilingual('Barod i recordio', 'Ready to record'),
                message,
                timer: 'hide'
            },
            saved: {
                text: bilingual('Barod i recordio', 'Ready to record'),
                message,
                timer: 'hide'
            }
        };

        const config = statusStates[state];
        if (!config) return;

        elements.statusText.html(config.text);
        elements.statusMessage.html(config.message);

        if (config.timer === 'hide') {
            elements.timer.hide();
        } else if (config.timer === 'show') {
            elements.timer.show().text('00:00');
        }

        if (config.stopTimer) {
            stopRecordingTimer();
        }

        // Reapply language toggle
        if (typeof initLanguageToggle === 'function') {
            initLanguageToggle();
        }
    }

    function startRecordingTimer() {
        recordingStartTime = Date.now();
        recordingTimer = setInterval(() => {
            const elapsed = Math.floor((Date.now() - recordingStartTime) / 1000);
            const minutes = Math.floor(elapsed / 60);
            const seconds = elapsed % 60;
            const timeString = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
            elements.timer.text(timeString);
        }, 100);
    }

    function stopRecordingTimer() {
        if (recordingTimer) {
            clearInterval(recordingTimer);
            recordingTimer = null;
        }
        elements.timer.text('00:00');
    }

    elements.startBtn.on('click', () => {
        elements.startBtn.prop('disabled', true);
        elements.stopBtn.prop('disabled', false);
        elements.startBtn.removeClass("btn-danger").addClass("btn-warning");

        setRecordingStatus('preparing');
        captureUserMedia(mediaConstraints, onMediaSuccess, onMediaError);
    });

    elements.stopBtn.on('click', () => {
        elements.stopBtn.prop('disabled', true);
        mediaRecorder.stop();
        mediaRecorder.stream.stop();

        stopRecordingTimer();

        // Reset start button to ready state
        elements.startBtn.removeClass("btn-warning").addClass("btn-danger");
        elements.startBtn.prop('disabled', false);
    });

    elements.playBtn.click(() => {
        const wordId = elements.wordDisplay.data('word-id') || '';
        const audio = new Audio(`/PlayAudio?id=${wordId}`);
        audio.play();
    });

    $(window).on('beforeunload', () => {
        elements.startBtn.prop('disabled', false);
    });

    function captureUserMedia(mediaConstraints, successCallback, errorCallback) {
        navigator.mediaDevices.getUserMedia(mediaConstraints)
            .then(successCallback)
            .catch(errorCallback);
    }

    function onMediaSuccess(stream) {
        mediaRecorder = new MediaStreamRecorder(stream);
        mediaRecorder.stream = stream;

        mediaRecorder.recorderType = StereoAudioRecorder;
        mediaRecorder.mimeType = 'audio/wav';
        mediaRecorder.audioChannels = 1;

        mediaRecorder.ondataavailable = (blob) => {
            uploadRecording(blob);
        };

        const timeInterval = 5000 * 1000;

        mediaRecorder.start(timeInterval);

        // Start timer and update status now that recording has actually begun
        startRecordingTimer();
        setRecordingStatus('recording');

        elements.stopBtn.prop('disabled', false);
    }

    function onMediaError(e) {
        console.error('media error', e);
    }

    const handleUploadSuccess = (data) => {
        console.log('Recording uploaded successfully');

        const statusKey = data.isComplete ? 'complete' : 'saved';
        const message = data.isComplete
            ? bilingual('5/5 recordiad wedi\'u cwblhau', '5/5 recordings completed')
            : bilingual(`Recordiad ${data.recordingCount}/5 wedi'i gadw`, `Recording ${data.recordingCount}/5 saved`);

        setRecordingStatus(statusKey, message);
        updateUI(data);
        enableRecordingButton();

        // Auto-progress to next incomplete word after completing 5 recordings
        if (data.isComplete) {
            setTimeout(() => {
                console.log('Auto-progressing to next incomplete word...');
                navigateWord('jumpNext');
            }, 1500); // 1.5 second delay to let user see completion message
        }
    };

    const handleUploadError = (error) => {
        console.error('Upload error:', error);
        showError('Methwyd uwchlwytho', 'Upload failed');
    };

    async function uploadRecording(blob) {
        disableRecordingButton();
        setRecordingStatus('uploading', bilingual('Prosesu recordiad...', 'Processing recording...'));

        try {
            const fd = new FormData();
            const filename = `${new Date().toISOString()}.wav`;
            fd.append("file", blob, filename);
            fd.append("word", elements.wordDisplay.data('word') || '');
            fd.append("wordID", elements.wordDisplay.data('word-id') || '');

            const response = await fetch("/Upload", {
                method: "POST",
                body: fd
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (!data.success) {
                throw new Error(data.error || 'Upload failed');
            }

            handleUploadSuccess(data);
        } catch (error) {
            handleUploadError(error);
        }
    }

    function updateProgressBars({ completedWordsCount, totalWordsCount, overallPercentage, wordPercentage, recordingCount, isComplete }) {
        // Update text indicators
        elements.compactProgress.text(`${completedWordsCount}/${totalWordsCount}`);
        elements.compactWordProgress.html(
            `<strong>${recordingCount}/5</strong>${isComplete ? ' <i class="bi bi-check-circle-fill text-success ms-1"></i>' : ''}`
        );

        // Update overall dataset progress bar
        elements.compactOverallBar
            .css('width', `${overallPercentage}%`)
            .attr('aria-valuenow', overallPercentage);
        elements.overallPctLabel.text(`${overallPercentage.toFixed(1)}%`);

        // Update current word progress bar
        elements.compactWordBar
            .css('width', `${wordPercentage}%`)
            .attr('aria-valuenow', wordPercentage)
            .attr('class', isComplete ? 'progress-bar bg-success' : 'progress-bar bg-warning');
        elements.wordPctLabel
            .text(`${wordPercentage}%`)
            .attr('class', isComplete ? 'text-success fw-bold' : 'text-warning fw-bold');
    }

    function updateWordDisplay({ word, wordID, english }) {
        elements.wordDisplay
            .data('word', word)
            .data('word-id', wordID)
            .find('.display-6').text(word || '').end()
            .find('.h5').text(english ? `(${english})` : '');
    }

    function updateFeedbackTable({ entries, recordingCount }) {
        elements.feedbackBody.empty();

        if (entries && entries.length > 0) {
            entries.forEach((entry, index) => {
                const attemptNum = recordingCount - index;
                const rowClass = index === 0 ? 'table-success' : '';

                elements.feedbackBody.append(
                    `<tr class="${rowClass}">
                        <td class="fw-bold text-center">${attemptNum}</td>
                        <td class="coloured-feedback"
                            data-target="${entry.Text || ''}"
                            data-whisper="${entry.AttemptWhisper || ''}">
                        </td>
                    </tr>`
                );
            });
            updateAllFeedback();
        } else {
            // Show empty state
            elements.feedbackBody.append(
                `<tr>
                    <td colspan="2" class="text-center text-muted py-4">
                        <i class="bi bi-mic-mute"></i><br>
                        ${bilingual('Dim recordiadau eto', 'No recordings yet')}
                    </td>
                </tr>`
            );
        }
    }

    function updateNextButton(isComplete) {
        if (isComplete) {
            elements.nextBtn.removeClass('disabled opacity-50').attr('title', 'Next word');
        } else {
            elements.nextBtn.addClass('disabled opacity-50').attr('title', 'Complete 5 recordings to unlock');
        }

        // Update fast-forward button tooltip based on language
        const ffTooltipCy = elements.fastForwardBtn.find('.cy').text();
        const ffTooltipEn = elements.fastForwardBtn.find('.en').text();
        const currentLang = getCookie('lang') || 'cy';
        elements.fastForwardBtn.attr('title', currentLang === 'cy' ? ffTooltipCy : ffTooltipEn);
    }

    function getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        return parts.length === 2 ? parts.pop().split(';').shift() : null;
    }

    function handleUIError(data) {
        console.error('Failed to update UI:', data.error);
        if (data.error || data.errorCy) {
            setRecordingStatus('idle');
            elements.statusMessage.html(
                `<span class="text-danger">${bilingual(data.errorCy || data.error, data.error)}</span>`
            );
        }
    }

    function updateUI(data) {
        if (!data.success) {
            return handleUIError(data);
        }

        // Reset recording status to idle for new word
        setRecordingStatus('idle');

        // Destructure data with defaults
        const {
            completedWordsCount = 0,
            totalWordsCount = 979,
            overallPercentage = 0,
            wordPercentage = 0,
            recordingCount = 0,
            isComplete,
            word,
            wordID,
            english,
            entries
        } = data;

        updateProgressBars({ completedWordsCount, totalWordsCount, overallPercentage, wordPercentage, recordingCount, isComplete });
        updateWordDisplay({ word, wordID, english });
        updateFeedbackTable({ entries, recordingCount });
        updateNextButton(isComplete);

        // Apply language to updated content
        if (typeof initLanguageToggle === 'function') {
            initLanguageToggle();
        }
    }

    async function navigateWord(action) {
        const endpoints = {
            reset: '/ResetProgress',
            next: '/IncrementProgress',
            prev: '/DecrementProgress',
            jumpNext: '/JumpToNextIncomplete'
        };

        const endpoint = endpoints[action];
        if (!endpoint) {
            console.error('Invalid navigation action:', action);
            return;
        }

        disableRecordingButton();

        try {
            const response = await fetch(endpoint, { method: 'POST' });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            updateUI(data);
        } catch (error) {
            console.error('Navigation failed:', error);
            showError('Methiant llywio. Ceisiwch eto.', 'Navigation failed. Please try again.');
        } finally {
            enableRecordingButton();
        }
    }

    elements.rewindBtn.click((e) => {
        e.preventDefault();
        navigateWord('reset');
    });

    elements.prevBtn.click((e) => {
        e.preventDefault();
        navigateWord('prev');
    });

    elements.nextBtn.click((e) => {
        e.preventDefault();
        if (elements.nextBtn.hasClass('disabled')) {
            return false;
        }
        navigateWord('next');
    });

    elements.fastForwardBtn.click((e) => {
        e.preventDefault();
        navigateWord('jumpNext');
    });
});
