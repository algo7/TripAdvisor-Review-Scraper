// Dependencies
const path = require('path');
const { existsSync, mkdirSync, } = require('fs');

// Custom Modules
const hotelScraper = require('./scrapers/hotel');
const restoScraper = require('./scrapers/resto');

// Global variables
const dataDir = path.join(__dirname, './data/');

// Check if the data directory exists, otherwise create it
if (!existsSync(dataDir)) {
    try {
        mkdirSync(dataDir);
    } catch (err) {
        console.error(err);
        process.exit(1);
    }
}