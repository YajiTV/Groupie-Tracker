/**
 * ==============================================
 * GROUPIE TRACKER - FILTER LOGIC WITH INPUTS
 * ==============================================
 */

'use strict';

class FilterManager {
    constructor() {
        this.filterPanel = document.getElementById('filterPanel');
        this.filterToggle = document.getElementById('filterToggle');
        this.filterChevron = document.getElementById('filterChevron');
        this.filterForm = document.getElementById('filterForm');
        this.resetButton = document.getElementById('resetFilters');
        this.activeBadgesContainer = document.getElementById('activeBadges');

        this.isOpen = false;
        this.debounceTimer = null;

        this.init();
    }

    init() {
        this.setupToggleButton();
        this.setupRangeSlidersWithInputs();
        this.setupResetButton();
        this.setupFormListeners();
        this.initializeFromURL();
    }

    setupToggleButton() {
        this.filterToggle.addEventListener('click', () => this.togglePanel());

        this.filterToggle.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                this.togglePanel();
            }
        });
    }

    togglePanel() {
        this.isOpen = !this.isOpen;

        if (this.isOpen) {
            this.filterPanel.style.maxHeight = `${this.filterPanel.scrollHeight}px`;
            this.filterPanel.style.opacity = '1';
            this.filterPanel.setAttribute('aria-hidden', 'false');
            this.filterChevron.style.transform = 'rotate(180deg)';
            this.filterToggle.setAttribute('aria-expanded', 'true');
        } else {
            this.filterPanel.style.maxHeight = '0';
            this.filterPanel.style.opacity = '0';
            this.filterPanel.setAttribute('aria-hidden', 'true');
            this.filterChevron.style.transform = 'rotate(0deg)';
            this.filterToggle.setAttribute('aria-expanded', 'false');
        }
    }

    setupRangeSlidersWithInputs() {
        const sliders = [
            { sliderId: 'creationMin', inputId: 'creationMinInput', showPlus: false },
            { sliderId: 'creationMax', inputId: 'creationMaxInput', showPlus: false },
            { sliderId: 'albumMin', inputId: 'albumMinInput', showPlus: false },
            { sliderId: 'albumMax', inputId: 'albumMaxInput', showPlus: false },
            { sliderId: 'memberMin', inputId: 'memberMinInput', showPlus: false },
            { sliderId: 'memberMax', inputId: 'memberMaxInput', showPlus: true },
        ];

        sliders.forEach(({ sliderId, inputId, showPlus }) => {
            const slider = document.getElementById(sliderId);
            const input = document.getElementById(inputId);

            if (!slider || !input) return;

            // Initialiser
            this.updateSliderProgress(slider);

            // Slider change -> update input
            slider.addEventListener('input', () => {
                const value = slider.value;
                const displayValue = (showPlus && value == slider.max) ? value + '+' : value;
                input.value = value;
                this.updateSliderProgress(slider);
                this.debouncedUpdateBadges();
            });

            // Input change -> update slider
            input.addEventListener('input', () => {
                let value = parseInt(input.value) || parseInt(slider.min);

                // Valider les limites
                if (value < parseInt(slider.min)) value = parseInt(slider.min);
                if (value > parseInt(slider.max)) value = parseInt(slider.max);

                input.value = value;
                slider.value = value;
                this.updateSliderProgress(slider);
                this.debouncedUpdateBadges();
            });

            // Validation au blur (perte de focus)
            input.addEventListener('blur', () => {
                let value = parseInt(input.value);
                if (isNaN(value)) {
                    value = parseInt(slider.value);
                    input.value = value;
                }
            });
        });
    }

    updateSliderProgress(slider) {
        const { value, min, max } = slider;
        const percentage = ((value - min) / (max - min)) * 100;
        slider.style.setProperty('--slider-progress', `${percentage}%`);
    }

    setupResetButton() {
        this.resetButton.addEventListener('click', () => this.resetFilters());
    }

    resetFilters() {
        // Reset sliders et inputs
        const pairs = [
            { sliderId: 'creationMin', inputId: 'creationMinInput', isMin: true },
            { sliderId: 'creationMax', inputId: 'creationMaxInput', isMin: false },
            { sliderId: 'albumMin', inputId: 'albumMinInput', isMin: true },
            { sliderId: 'albumMax', inputId: 'albumMaxInput', isMin: false },
            { sliderId: 'memberMin', inputId: 'memberMinInput', isMin: true },
            { sliderId: 'memberMax', inputId: 'memberMaxInput', isMin: false },
        ];

        pairs.forEach(({ sliderId, inputId, isMin }) => {
            const slider = document.getElementById(sliderId);
            const input = document.getElementById(inputId);

            if (slider && input) {
                const value = isMin ? slider.min : slider.max;
                slider.value = value;
                input.value = value;
                this.updateSliderProgress(slider);
            }
        });

        this.updateActiveBadges();
        window.location.href = window.location.pathname;
    }

    setupFormListeners() {
        this.filterForm.addEventListener('change', () => {
            this.debouncedUpdateBadges();
        });
    }

    debouncedUpdateBadges() {
        clearTimeout(this.debounceTimer);
        this.debounceTimer = setTimeout(() => {
            this.updateActiveBadges();
        }, 300);
    }

    updateActiveBadges() {
        const badges = [];
        const formData = new FormData(this.filterForm);

        // Creation year
        const creationMin = formData.get('creation_year_min');
        const creationMax = formData.get('creation_year_max');
        const creationMinSlider = document.getElementById('creationMin');
        const creationMaxSlider = document.getElementById('creationMax');

        if (creationMin && creationMin !== creationMinSlider.min) {
            badges.push(this.createBadge(`Création ≥ ${creationMin}`));
        }
        if (creationMax && creationMax !== creationMaxSlider.max) {
            badges.push(this.createBadge(`Création ≤ ${creationMax}`));
        }

        // Album year
        const albumMin = formData.get('album_year_min');
        const albumMax = formData.get('album_year_max');
        const albumMinSlider = document.getElementById('albumMin');
        const albumMaxSlider = document.getElementById('albumMax');

        if (albumMin && albumMin !== albumMinSlider.min) {
            badges.push(this.createBadge(`Album ≥ ${albumMin}`));
        }
        if (albumMax && albumMax !== albumMaxSlider.max) {
            badges.push(this.createBadge(`Album ≤ ${albumMax}`));
        }

        // Member count
        const memberMin = formData.get('member_count_min');
        const memberMax = formData.get('member_count_max');
        const memberMinSlider = document.getElementById('memberMin');
        const memberMaxSlider = document.getElementById('memberMax');

        if (memberMin && memberMin !== memberMinSlider.min) {
            badges.push(this.createBadge(`Membres ≥ ${memberMin}`));
        }
        if (memberMax && memberMax !== memberMaxSlider.max) {
            const suffix = memberMax == 8 ? '+' : '';
            badges.push(this.createBadge(`Membres ≤ ${memberMax}${suffix}`));
        }

        if (badges.length > 0) {
            this.activeBadgesContainer.innerHTML = badges.join('');
            this.activeBadgesContainer.classList.remove('hidden');
            this.activeBadgesContainer.classList.add('flex');
        } else {
            this.activeBadgesContainer.classList.add('hidden');
            this.activeBadgesContainer.classList.remove('flex');
        }
    }

    createBadge(text) {
        return `<span class="animate-fade-in-up px-3 py-1.5 bg-neutral-700 border border-neutral-600 text-white rounded-full text-xs font-medium">${text}</span>`;
    }

    initializeFromURL() {
        const params = new URLSearchParams(window.location.search);

        if (params.toString()) {
            setTimeout(() => this.togglePanel(), 100);
            this.updateActiveBadges();
        }
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new FilterManager();
});