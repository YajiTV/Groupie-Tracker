// Récupération des données depuis le template Go
const el = document.getElementById("map");
const artistId = Number(el.dataset.artistId);

console.log("=== MAP WITH SERVER DATA ===");
console.log("Artist ID:", artistId);

// Récupération des données locations depuis le template (injecté par Go)
const locationsData = window.artistLocations || [];
console.log("Locations data from server:", locationsData);

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

    if (country === "usa") country = "united states";
    if (country === "uk") country = "united kingdom";

    const result = (cityOrRegion + ", " + country).trim();
    console.log(`Normalize: "${key}" -> "${result}"`);
    return result;
}

async function fetchJSON(url) {
    console.log("Geocoding:", url);
    const r = await fetch(url);
    if (!r.ok) throw new Error(url + " => HTTP " + r.status);
    return r.json();
}

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

(async() => {
    try {
        if (!locationsData || !locationsData.length) {
            console.error("❌ No locations data found!");
            return;
        }

        console.log(`4. Processing ${locationsData.length} locations:`, locationsData);

        const layers = [];

        for (let i = 0; i < locationsData.length; i++) {
            const location = locationsData[i];
            console.log(`\n--- Processing location ${i + 1}: "${location.name}" ---`);
            
            const query = normalizeLocationKey(location.name);
            const cacheKey = "geo:" + query.toLowerCase();
            let cached = localStorage.getItem(cacheKey);

            let lat, lon;

            if (cached) {
                [lat, lon] = JSON.parse(cached);
                console.log(`✅ Using cached coordinates: ${lat}, ${lon}`);
            } else {
                console.log(`🌍 Geocoding: ${query}`);
                const url = "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" + encodeURIComponent(query);
                
                try {
                    const res = await fetchJSON(url);
                    if (!res.length) {
                        console.warn(`❌ No results for: ${query}`);
                        continue;
                    }

                    lat = Number(res[0].lat);
                    lon = Number(res[0].lon);
                    console.log(`✅ Geocoded: ${lat}, ${lon}`);
                    
                    localStorage.setItem(cacheKey, JSON.stringify([lat, lon]));
                    await sleep(1000);
                } catch (error) {
                    console.error(`❌ Error geocoding ${query}:`, error);
                    continue;
                }
            }

            const dates = location.dates || [];
            console.log(`📅 Dates for this location:`, dates);

            console.log(`📍 Creating marker at [${lat}, ${lon}]`);
            const marker = L.circleMarker([lat, lon], {
                radius: 8,
                color: "#fff",
                weight: 2,
                fillColor: "#00e5ff",
                fillOpacity: 0.8
            }).addTo(map)
            .bindPopup("<b>" + query + "</b><br>" + (dates.length ? dates.join("<br>") : "Aucune date"));

            layers.push(marker);
            
            console.log(`✅ Marker created! Total markers: ${layers.length}`);
        }

        console.log(`\n🎯 SUMMARY: Created ${layers.length} markers`);

        if (layers.length) {
            map.fitBounds(L.featureGroup(layers).getBounds(), { padding: [20, 20] });
            console.log("✅ Map bounds adjusted to fit all markers");
        } else {
            console.error("❌ No markers created!");
        }

    } catch (error) {
        console.error("💥 FATAL ERROR:", error);
    }
})();