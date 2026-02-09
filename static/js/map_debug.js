const el = document.getElementById("map");
const artistId = Number(el.dataset.artistId);

console.log("=== DEBUGGING MAP ===");
console.log("Artist ID:", artistId);

const map = L.map("map").setView([20, 0], 2);
L.tileLayer("https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png", {
    attribution: "© OpenStreetMap contributors · © CARTO",
    maxZoom: 19
}).addTo(map);

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
    console.log("Fetching:", url);
    const r = await fetch(url);
    if (!r.ok) throw new Error(url + " => HTTP " + r.status);
    const data = await r.json();
    console.log("Response:", data);
    return data;
}

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

(async() => {
    try {
        console.log("1. Fetching locations...");
        const locPayload = await fetchJSON("https://groupietrackers.herokuapp.com/api/locations");
        
        console.log("2. Looking for artist ID", artistId, "in locations");
        const locEntry = locPayload.index.find(x => Number(x.id) === artistId);
        console.log("Found locations entry:", locEntry);

        console.log("3. Fetching relations...");
        const relPayload = await fetchJSON("https://groupietrackers.herokuapp.com/api/relation");
        
        const relEntry = relPayload.index.find(x => Number(x.id) === artistId);
        console.log("Found relation entry:", relEntry);
        
        const datesLocations = (relEntry && relEntry.datesLocations) ? relEntry.datesLocations : {};

        if (!locEntry || !locEntry.locations || !locEntry.locations.length) {
            console.error("❌ No locations found for this artist!");
            return;
        }

        console.log("4. Processing", locEntry.locations.length, "locations:", locEntry.locations);

        const layers = [];
        let processedCount = 0;

        for (const locKey of locEntry.locations) {
            console.log(`\n--- Processing location ${processedCount + 1}: "${locKey}" ---`);
            
            const query = normalizeLocationKey(locKey);
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

            const dates = datesLocations[locKey] || [];
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
            processedCount++;
            
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