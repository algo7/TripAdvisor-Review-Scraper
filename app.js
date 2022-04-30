// Dependencies
const path = require('path');
const { existsSync, mkdirSync, writeFileSync, } = require('fs');

// Custom Modules
const hotelScraper = require('./scrapers/hotel');
const restoScraper = require('./scrapers/resto');

// Global variables
const dataDir = path.join(__dirname, './data/');

// Environment variables
const { SCRAP_MODE, } = process.env;

// Check if the data directory exists, otherwise create it
if (!existsSync(dataDir)) {
    try {
        mkdirSync(dataDir);
    } catch (err) {
        console.error(err);
        process.exit(1);
    }
}

if (SCRAP_MODE === 'hotel') {
    hotelScraper().then(csv => {
        // Write the CSV to a file
        writeFileSync('./data/review.csv', csv);
    }).catch(err => console.log(err));
} else {
    restoScraper.start();

}

hotelScraper('https://www.tripadvisor.com/Hotel_Review-g188107-d11761198-Reviews-Hotel_des_Patients-Lausanne_Canton_of_Vaud.html').catch(err => console.log(err));

