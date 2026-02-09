#!/bin/bash

echo "🚀 Démarrage du serveur..."
./server &
SERVER_PID=$!

# Attendre que le serveur démarre
sleep 2

echo "🧪 Test des filtres..."

echo "📊 1. Test filtre année de création:"
curl -s "http://localhost:8081/filters?creation_year_min=1990&creation_year_max=2000" > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Filtre année de création - OK"
else
    echo "❌ Filtre année de création - ERREUR"
fi

echo "👥 2. Test filtre nombre de membres:"
curl -s "http://localhost:8081/filters?member_count=4" > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Filtre nombre de membres - OK"
else
    echo "❌ Filtre nombre de membres - ERREUR"
fi

echo "💿 3. Test filtre premier album:"
curl -s "http://localhost:8081/filters?album_year_min=1980&album_year_max=1990" > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Filtre premier album - OK"
else
    echo "❌ Filtre premier album - ERREUR"
fi

echo "🌍 4. Test filtre lieux:"
curl -s "http://localhost:8081/filters?location=london" > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Filtre lieux - OK"
else
    echo "❌ Filtre lieux - ERREUR"
fi

echo "🔍 5. Test recherche combinée:"
curl -s "http://localhost:8081/filters?q=queen&creation_year_min=1970&member_count=4" > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ Recherche combinée - OK"
else
    echo "❌ Recherche combinée - ERREUR"
fi

echo "🛑 Arrêt du serveur..."
kill $SERVER_PID

echo "✅ Tests terminés!"
