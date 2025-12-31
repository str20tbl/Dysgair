/**
 * Common JavaScript functionality across all pages
 * Handles language toggle, copyright year, and shared utilities
 * ES6+ Compliant
 */

let cachedNotificationModal = null;
let cachedConfirmationModal = null;

function getCookie(name) {
    try {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        return parts.length === 2 ? parts.pop().split(';').shift() : null;
    } catch (e) {
        console.error('Error reading cookie:', name, e);
        return null;
    }
}

function setCookie(name, value, days = 365, path = '/') {
    let expires = '';
    if (days) {
        const date = new Date();
        date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
        expires = `; expires=${date.toUTCString()}`;
    }
    document.cookie = `${name}=${value}${expires}; path=${path}; SameSite=Lax`;
}

function applyLanguage(lang) {
    const isEnglish = (lang === 'en');
    $('.cy').toggle(!isEnglish);
    $('.en').toggle(isEnglish);
    $('#lang-sel').text(isEnglish ? 'Cymraeg' : 'English');
}

function setLanguage(lang) {
    applyLanguage(lang);
    setCookie('lang', lang);
}

function initLanguageToggle() {
    const lang = getCookie('lang') || 'cy';
    applyLanguage(lang);
}

function renderIconsAndLanguage() {
    if (typeof feather !== 'undefined') {
        feather.replace();
    }
    const lang = getCookie('lang') || 'cy';
    applyLanguage(lang);
}

function showNotification(message, type = 'info') {
    const modalElement = document.getElementById('notificationModal');
    if (!modalElement) {
        console.warn('Notification modal not found on page');
        return;
    }
    const types = {
        success: {
            icon: 'check-circle',
            title: 'Success / Llwyddiant',
            headerClass: 'bg-success text-white'
        },
        error: {
            icon: 'x-circle',
            title: 'Error / Gwall',
            headerClass: 'bg-danger text-white'
        },
        warning: {
            icon: 'alert-triangle',
            title: 'Warning / Rhybudd',
            headerClass: 'bg-warning'
        },
        info: {
            icon: 'info',
            title: 'Information / Gwybodaeth',
            headerClass: 'bg-info text-white'
        }
    };
    const config = types[type] || types.info;
    $('#notificationIcon').attr('data-feather', config.icon).attr('class', 'me-2');
    $('#notificationTitle').text(config.title);
    $('#notificationHeader').attr('class', `modal-header ${config.headerClass}`);
    $('#notificationMessage').text(message);
    if (!cachedNotificationModal) {
        cachedNotificationModal = new bootstrap.Modal(modalElement);
    }
    cachedNotificationModal.show();
    if (typeof feather !== 'undefined') {
        feather.replace();
    }
    if (type === 'success') {
        setTimeout(() => {
            cachedNotificationModal.hide();
        }, 2000);
    }
}

function showConfirmation(message, onConfirm) {
    const modalElement = document.getElementById('confirmationModal');
    if (!modalElement) {
        console.warn('Confirmation modal not found on page');
        return;
    }
    $('#confirmationMessage').text(message);
    if (!cachedConfirmationModal) {
        cachedConfirmationModal = new bootstrap.Modal(modalElement);
    }
    $('#btnConfirmAction').off('click').one('click', () => {
        cachedConfirmationModal.hide();
        if (typeof onConfirm === 'function') {
            onConfirm();
        }
    });
    cachedConfirmationModal.show();
}

$(() => {
    $('#cc-year').text(new Date().getFullYear());
    if (typeof feather !== 'undefined') {
        feather.replace();
    }
    $('#lang-sel').click(() => {
        const currentLang = getCookie('lang') || 'cy';
        setLanguage(currentLang === 'en' ? 'cy' : 'en');
    });
    initLanguageToggle();
});
