// Variables globales
let currentSuggestions = [];
let selectedIndex = -1;
let suggestionsContainer;
let searchInput;

// Quand la page est chargée, on initialise tout
document.addEventListener('DOMContentLoaded', function() {
    // Récupérer l'élément avec l'ID
    searchInput = document.getElementById('searchInput');

    if (!searchInput) {
        return;
    }

    // Créer le container des suggestions
    createSuggestionsContainer();

    // Écouter quand l'utilisateur tape quelque chose
    searchInput.addEventListener('input', handleInput);

    // Écouter les touches du clavier
    searchInput.addEventListener('keydown', handleKeydown);

    // Cacher les suggestions si on clique ailleurs
    document.addEventListener('click', function(e) {
        if (!searchInput.contains(e.target) && !suggestionsContainer.contains(e.target)) {
            hideSuggestions();
        }
    });
});

// Créer le container HTML pour les suggestions
function createSuggestionsContainer() {
    suggestionsContainer = document.createElement('div');
    suggestionsContainer.className = 'suggestions-dropdown';
    suggestionsContainer.style.display = 'none';

    // L'insérer juste après le champ de recherche
    const parent = searchInput.parentNode;
    parent.appendChild(suggestionsContainer);
}

// Fonction appelée quand l'utilisateur tape
function handleInput(e) {
    const query = e.target.value.trim();

    // Si la recherche est vide, cacher les suggestions
    if (query.length === 0) {
        hideSuggestions();
        return;
    }

    // Si la recherche fait moins de 2 caractères, ne rien faire
    if (query.length < 2) {
        return;
    }

    // Faire l'appel API pour récupérer les suggestions
    fetchSuggestions(query);
}

// Fonction qui fait l'appel API
function fetchSuggestions(query) {
    // Construire l'URL de l'API
    const url = `/api/suggestions?q=${encodeURIComponent(query)}`;

    // Faire la requête HTTP
    fetch(url)
        .then(response => {
            // Vérifier que la requête a fonctionné
            if (!response.ok) {
                throw new Error('Erreur réseau: ' + response.status);
            }
            return response.json(); // Transformer la réponse en JSON
        })
        .then(data => {
            // Sauvegarder les suggestions
            currentSuggestions = data.suggestions || [];
            selectedIndex = -1; // Remettre la sélection à zéro

            // Afficher les suggestions
            displaySuggestions();
        })
        .catch(error => {
            console.error('Erreur suggestions:', error);
            hideSuggestions();
        });
}

// Afficher les suggestions dans le HTML
function displaySuggestions() {
    // Si pas de suggestions, cacher la dropdown
    if (currentSuggestions.length === 0) {
        hideSuggestions();
        return;
    }

    // Vider le container
    suggestionsContainer.innerHTML = '';

    // Créer chaque suggestion
    currentSuggestions.forEach((suggestion, index) => {
        const item = document.createElement('div');
        item.className = 'suggestion-item';

        // Ajouter le contenu sans émoji
        item.innerHTML = `
            <span class="suggestion-text">${suggestion.text}</span>
            <span class="suggestion-type">${suggestion.type}</span>
        `;

        // Quand on clique sur une suggestion
        item.addEventListener('click', function() {
            selectSuggestion(index);
        });

        suggestionsContainer.appendChild(item);
    });

    // Montrer la dropdown
    suggestionsContainer.style.display = 'block';
}

// Gestion des touches du clavier
function handleKeydown(e) {
    // Si pas de suggestions visibles, ne rien faire
    if (currentSuggestions.length === 0) {
        return;
    }

    switch(e.key) {
        case 'ArrowDown':
            e.preventDefault();
            selectedIndex = Math.min(selectedIndex + 1, currentSuggestions.length - 1);
            updateSelection();
            break;

        case 'ArrowUp':
            e.preventDefault();
            selectedIndex = Math.max(selectedIndex - 1, -1);
            updateSelection();
            break;

        case 'Enter':
            e.preventDefault();
            if (selectedIndex >= 0) {
                selectSuggestion(selectedIndex);
            }
            break;

        case 'Escape':
            hideSuggestions();
            searchInput.blur();
            break;
    }
}

// Mettre à jour la sélection visuelle
function updateSelection() {
    const items = suggestionsContainer.querySelectorAll('.suggestion-item');

    // Enlever la classe "selected" de tous les éléments
    items.forEach(item => item.classList.remove('selected'));

    // Ajouter la classe "selected" à l'élément sélectionné
    if (selectedIndex >= 0 && selectedIndex < items.length) {
        items[selectedIndex].classList.add('selected');
    }
}

// Sélectionner une suggestion (redirection)
function selectSuggestion(index) {
    const suggestion = currentSuggestions[index];

    if (suggestion) {
        // Rediriger vers la page de l'artiste
        window.location.href = `/artist/${suggestion.artist_id}`;
    }
}

// Cacher les suggestions
function hideSuggestions() {
    suggestionsContainer.style.display = 'none';
    currentSuggestions = [];
    selectedIndex = -1;
}

