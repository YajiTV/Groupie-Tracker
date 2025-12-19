/**
 * GROUPIE TRACKER - Filters JS
 * Sync slider <-> number input + toggle panel + reset
 */

'use strict';

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}

function init() {
    setupSlider('creationMin', 'creationMinInput');
    setupSlider('creationMax', 'creationMaxInput');
    setupSlider('albumMin', 'albumMinInput');
    setupSlider('albumMax', 'albumMaxInput');
    setupSlider('memberMin', 'memberMinInput');
    setupSlider('memberMax', 'memberMaxInput');

    setupToggle('filterToggle', 'filterPanel', 'filterChevron');
    setupReset('resetFilters');
}

function setupSlider(sliderId, inputId) {
    const slider = document.getElementById(sliderId);
    const input = document.getElementById(inputId);
    if (!slider || !input) return;

    function updateProgress() {
        const min = Number(slider.min);
        const max = Number(slider.max);
        const val = Number(slider.value);
        const percent = ((val - min) / (max - min)) * 100;
        slider.style.setProperty('--slider-progress', percent + '%');
    }

    // init
    input.value = slider.value;
    updateProgress();

    // slider -> input (temps réel)
    slider.addEventListener('input', () => {
        input.value = slider.value;
        updateProgress();
    });

    // input -> slider
    input.addEventListener('input', () => {
        const v = parseInt(input.value, 10);
        if (Number.isNaN(v)) return;

        const min = parseInt(slider.min, 10);
        const max = parseInt(slider.max, 10);
        const clamped = Math.min(Math.max(v, min), max);

        input.value = clamped;
        slider.value = clamped;
        updateProgress();
    });

    // si vide, on remet la valeur du slider
    input.addEventListener('blur', () => {
        if (input.value === '' || Number.isNaN(parseInt(input.value, 10))) {
            input.value = slider.value;
        }
    });
}

function setupToggle(toggleId, panelId, chevronId) {
    const toggle = document.getElementById(toggleId);
    const panel = document.getElementById(panelId);
    const chevron = document.getElementById(chevronId);
    if (!toggle || !panel) return;

    let open = false;

    function apply() {
        if (open) {
            panel.style.maxHeight = panel.scrollHeight + 'px';
            panel.style.opacity = '1';
            panel.setAttribute('aria-hidden', 'false');
            toggle.setAttribute('aria-expanded', 'true');
            if (chevron) chevron.style.transform = 'rotate(180deg)';
        } else {
            panel.style.maxHeight = '0';
            panel.style.opacity = '0';
            panel.setAttribute('aria-hidden', 'true');
            toggle.setAttribute('aria-expanded', 'false');
            if (chevron) chevron.style.transform = 'rotate(0deg)';
        }
    }

    toggle.addEventListener('click', () => {
        open = !open;
        apply();
    });

    toggle.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            open = !open;
            apply();
        }
    });
}

function setupReset(resetId) {
    const btn = document.getElementById(resetId);
    if (!btn) return;

    btn.addEventListener('click', () => {
        window.location.href = window.location.pathname;
    });
}