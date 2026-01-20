const el = document.getElementById("map");
const artistId = Number(el.dataset.artistId);

const map = L.map("map").setView([20, 0], 2);
L.tileLayer("https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png", {
    attribution: "© OpenStreetMap contributors · © CARTO",
    maxZoom: 19
}).addTo(map);

// Transforme "north_carolina-usa" -> "north carolina, united states"
function normalizeLocationKey(key) {
    const parts = String(key).split("-");
    const cityOrRegion = (parts[0] || "").replace(/_/g, " ");
    let country = (parts[1] || "").replace(/_/g, " ");

    // Petites corrections utiles pour Nominatim
    if (country === "usa") country = "united states";
    if (country === "uk") country = "united kingdom";

    return (cityOrRegion + ", " + country).trim();
}

async function fetchJSON(url) {
    const r = await fetch(url);
    if (!r.ok) throw new Error(url + " => HTTP " + r.status);
    return r.json();
}

// Pause simple pour éviter de spam Nominatim (sinon 429 possible)
function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

(async() => {
    // 1) Locations
    const locPayload = await fetchJSON("https://groupietrackers.herokuapp.com/api/locations"); // {index:[...]} [attached_file:3]
    const locEntry = locPayload.index.find(x => Number(x.id) === artistId);

    // 2) Dates par lieu (relation)
    const relPayload = await fetchJSON("https://groupietrackers.herokuapp.com/api/relation"); // {index:[...]} [attached_file:2]
    const relEntry = relPayload.index.find(x => Number(x.id) === artistId);
    const datesLocations = (relEntry && relEntry.datesLocations) ? relEntry.datesLocations : {};

    if (!locEntry || !locEntry.locations || !locEntry.locations.length) return;

    const layers = [];

    for (const locKey of locEntry.locations) {
        const query = normalizeLocationKey(locKey);

        // cache navigateur
        const cacheKey = "geo:" + query.toLowerCase();
        let cached = localStorage.getItem(cacheKey);

        let lat, lon;

        if (cached) {
            [lat, lon] = JSON.parse(cached);
        } else {
            // Nominatim: renvoie lat/lon en JSON [web:445]
            const url = "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" + encodeURIComponent(query);
            const res = await fetchJSON(url);
            if (!res.length) continue;

            lat = Number(res[0].lat);
            lon = Number(res[0].lon);
            localStorage.setItem(cacheKey, JSON.stringify([lat, lon]));

            // 1 requête/sec recommandé pour éviter d’être limité [web:268]
            await sleep(1000);
        }

        const dates = datesLocations[locKey] || []; // dates pour CE lieu [attached_file:2]

        const marker = L.circleMarker([lat, lon], {
                radius: 6,
                color: "#fff",
                weight: 2,
                fillColor: "#00e5ff",
                fillOpacity: 0.8
            }).addTo(map)
            .bindPopup("<b>" + query + "</b><br>" + (dates.length ? dates.join("<br>") : "Aucune date")); // popup Leaflet [web:272]

        layers.push(marker);
    }

    if (layers.length) {
        map.fitBounds(L.featureGroup(layers).getBounds(), { padding: [20, 20] }); // fitBounds Leaflet [web:272]
    }
})();