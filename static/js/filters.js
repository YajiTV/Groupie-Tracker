/**
 * GROUPIE TRACKER - Gestion des filtres
 * Ce fichier gère la synchronisation entre les curseurs et les champs de texte
 */

'use strict';

// Attendre que la page soit complètement chargée
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}

/**
 * Fonction principale d'initialisation
 */
function init() {
    // Configurer tous les curseurs
    setupSlider('creationMin', 'creationMinInput');
    setupSlider('creationMax', 'creationMaxInput');
    setupSlider('albumMin', 'albumMinInput');
    setupSlider('albumMax', 'albumMaxInput');
    setupSlider('memberMin', 'memberMinInput');
    setupSlider('memberMax', 'memberMaxInput');

    // Configurer le bouton d'ouverture/fermeture des filtres
    setupToggle();

    // Configurer le bouton de réinitialisation
    setupReset();
}

/**
 * Configure la synchronisation entre un curseur et son input
 * @param {string} sliderId - ID du curseur (range)
 * @param {string} inputId - ID du champ texte (input number)
 */
function setupSlider(sliderId, inputId) {
    const slider = document.getElementById(sliderId);
    const input = document.getElementById(inputId);

    // Vérifier que les éléments existent
    if (!slider || !input) return;

    // Fonction pour mettre à jour la barre de progression du curseur
    function updateProgress() {
        const percent = ((slider.value - slider.min) / (slider.max - slider.min)) * 100;
        slider.style.setProperty('--slider-progress', percent + '%');
    }

    // Initialiser les valeurs au chargement
    input.value = slider.value;
    updateProgress();

    // Quand on bouge le curseur -> mettre à jour l'input
    slider.addEventListener('input', function() {
        input.value = this.value;
        updateProgress();
    });

    // Quand on tape dans l'input -> mettre à jour le curseur
    input.addEventListener('input', function() {
        let value = parseInt(this.value);

        // Vérifier que la valeur est un nombre valide
        if (isNaN(value)) return;

        // Limiter la valeur entre min et max
        if (value < parseInt(slider.min)) value = parseInt(slider.min);
        if (value > parseInt(slider.max)) value = parseInt(slider.max);

        // Appliquer la nouvelle valeur
        this.value = value;
        slider.value = value;
        updateProgress();
    });

    // Quand on quitte l'input -> vérifier la valeur
    input.addEventListener('blur', function() {
        if (!this.value || isNaN(this.value)) {
            this.value = slider.value;
        }
    });
}

/**
 * Configure le bouton d'ouverture/fermeture du panneau de filtres
 */
function setupToggle() {
    const toggle = document.getElementById('filterToggle');
    const panel = document.getElementById('filterPanel');
    const chevron = document.getElementById('filterChevron');

    // Vérifier que les éléments existent
    if (!toggle || !panel) return;

    let isOpen = false;

    // Au clic sur le bouton
    toggle.addEventListener('click', function() {
        isOpen = !isOpen;

        if (isOpen) {
            // Ouvrir le panneau
            panel.style.maxHeight = panel.scrollHeight + 'px';
            panel.style.opacity = '1';
            panel.setAttribute('aria-hidden', 'false');
            if (chevron) chevron.style.transform = 'rotate(180deg)';
            toggle.setAttribute('aria-expanded', 'true');
        } else {
            // Fermer le panneau
            panel.style.maxHeight = '0';
            panel.style.opacity = '0';
            panel.setAttribute('aria-hidden', 'true');
            if (chevron) chevron.style.transform = 'rotate(0deg)';
            toggle.setAttribute('aria-expanded', 'false');
        }
    });
}

/**
 * Configure le bouton de réinitialisation des filtres
 */
function setupReset() {
    const resetBtn = document.getElementById('resetFilters');

    // Vérifier que l'élément existe
    if (!resetBtn) return;

    // Au clic -> retourner à la page sans paramètres
    resetBtn.addEventListener('click', function() {
        window.location.href = window.location.pathname;
    });
}