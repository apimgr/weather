#!/usr/bin/env node

const axios = require('axios');
const fs = require('fs');
const path = require('path');

async function downloadData() {
  const dataDir = path.join(__dirname, '../src/public/api');

  // Ensure directory exists
  if (!fs.existsSync(dataDir)) {
    fs.mkdirSync(dataDir, { recursive: true });
  }

  console.log('📥 Downloading location data...');

  try {
    // Download countries data
    console.log('☕ Downloading countries data...');
    const countriesResponse = await axios.get('https://github.com/apimgr/countries/raw/refs/heads/main/countries.json', {
      timeout: 30000
    });

    const countriesPath = path.join(dataDir, 'countries.json');
    fs.writeFileSync(countriesPath, JSON.stringify(countriesResponse.data, null, 2));
    console.log(`✅ Downloaded ${countriesResponse.data.length} countries to ${countriesPath}`);

    // Download cities data
    console.log('☕ Downloading cities data (this may take a moment)...');
    const citiesResponse = await axios.get('https://github.com/apimgr/citylist/raw/refs/heads/master/api/city.list.json', {
      timeout: 60000 // Longer timeout for large file
    });

    const citiesPath = path.join(dataDir, 'cities.json');
    fs.writeFileSync(citiesPath, JSON.stringify(citiesResponse.data, null, 2));
    console.log(`✅ Downloaded ${citiesResponse.data.length} cities to ${citiesPath}`);

    // Create metadata file
    const metadata = {
      downloaded: new Date().toISOString(),
      sources: {
        countries: 'https://github.com/apimgr/countries/raw/refs/heads/main/countries.json',
        cities: 'https://github.com/apimgr/citylist/raw/refs/heads/master/api/city.list.json'
      },
      counts: {
        countries: countriesResponse.data.length,
        cities: citiesResponse.data.length
      }
    };

    const metadataPath = path.join(dataDir, 'metadata.json');
    fs.writeFileSync(metadataPath, JSON.stringify(metadata, null, 2));
    console.log(`📊 Created metadata file: ${metadataPath}`);

    console.log('🎉 Data download complete!');
    console.log(`📍 Total: ${metadata.counts.countries} countries, ${metadata.counts.cities} cities`);

  } catch (error) {
    console.error('❌ Download failed:', error.message);
    process.exit(1);
  }
}

// Run if called directly
if (require.main === module) {
  downloadData();
}

module.exports = downloadData;