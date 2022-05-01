// Dependencies
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import fs, { mkdirSync } from 'fs';
const { promises: { writeFile, }, } = fs;
import chalk from 'chalk';
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Custom Modules
import hotelScraper from './scrapers/hotel.js';
import restoScraper from './scrapers/resto.js';
import { csvToJSON, fileExists, combine, reviewJSONToCsv } from './libs/utils.js';
import { getBrowserInstance, closeBrowserInstance } from './libs/browser.js';

// Data path
const dataDir = join(__dirname, './reviews/');
const sourceDir = join(__dirname, './source/');

// Environment variables
const { SCRAPE_MODE, } = process.env;

console.log(chalk.bold.blue(`The Scraper is Running in ${chalk.bold.magenta(SCRAPE_MODE)} Mode`));

// Check if the required directories exist, otherwise create them
if (!fileExists(dataDir)) {
    mkdirSync(dataDir);
}

if (!fileExists(sourceDir)) {
    mkdirSync(sourceDir);
}

// Data source
const dataSourceResto = join(__dirname, './source/restos.csv');
const dataSourceHotel = join(__dirname, './source/hotels.csv');

/**
 * Scrape the hotel pages
 * @returns {Promise<String | Error>} - The done message or error message
 */
const hotelScraperInit = async () => {
    try {
        // Check if the source file exists
        const sourceFileAvailable = await fileExists(dataSourceHotel);
        if (!sourceFileAvailable) {
            throw Error('Source file does not exist');
        }

        // Convert the csv to json
        const rawData = await csvToJSON(dataSourceHotel);
        console.log(chalk.bold.yellow(`Scraping ${chalk.magenta(rawData.length)} Hotels`));

        // Get a browser instance
        const browser = await getBrowserInstance();

        await Promise.all(
            rawData.map(async (item, index) => {
                // Extract resto info
                const { webUrl: hotelUrl, name: hotelName, id: hotelId, } = item;

                // Start the scraping process
                const finalData = await hotelScraper(hotelUrl, hotelName, hotelId, index, browser);
                const { fileName, } = finalData;
                delete finalData.fileName;

                // Write the data to file
                const dataToWrite = JSON.stringify(finalData, null, 2);
                await writeFile(`${dataDir}${fileName}.json`, dataToWrite);
            })
        );

        // Combine all the reviews into an array of objects
        const combinedData = combine(SCRAPE_MODE, dataDir);

        // Write the combined JSON data to file
        const dataToWrite = JSON.stringify(combinedData, null, 2);
        await writeFile(`${dataDir}All.json`, dataToWrite);

        // Convert the combined JSON data to csv
        const csvData = reviewJSONToCsv(combinedData);
        await writeFile(`${dataDir}All.csv`, csvData);

        // Close the browser instance
        await closeBrowserInstance();

        return 'Scraping Done';

    } catch (err) {
        throw err;
    }
};

/**
 * Scrape the resto pages
 * @returns {Promise<String | Error>} - The done message or error message
 */
const restoScraperInit = async () => {
    try {

        // Check if the source file exists
        const sourceFileAvailable = await fileExists(dataSourceResto);
        if (!sourceFileAvailable) {
            throw Error('Source file does not exist');
        }

        // Convert the csv to json
        const rawData = await csvToJSON(dataSourceResto);
        console.log(chalk.bold.yellow(`Scraping ${chalk.magenta(rawData.length)} Restaurants`));

        // Get a browser instance
        const browser = await getBrowserInstance();

        await Promise.all(
            rawData.map(async (item, index) => {
                // Extract resto info
                const { webUrl: restoUrl, name: restoName, id: restoId, } = item;
                // Start the scraping process
                const finalData = await restoScraper(restoUrl, restoName, restoId, index, browser);
                const { fileName, } = finalData;
                delete finalData.fileName;
                // Write the data to file
                const dataToWrite = JSON.stringify(finalData, null, 2);
                await writeFile(`${dataDir}${fileName}.json`, dataToWrite);
            })
        );

        // Combine all the reviews into an array of objects
        const combinedData = combine(SCRAPE_MODE, dataDir);

        // Write the combined JSON data to file
        const dataToWrite = JSON.stringify(combinedData, null, 2);
        await writeFile(`${dataDir}All.json`, dataToWrite);

        // Convert the combined JSON data to csv
        const csvData = reviewJSONToCsv(combinedData);
        await writeFile(`${dataDir}All.csv`, csvData);

        // Close the browser instance
        await closeBrowserInstance();

        return 'Scraping Done';

    } catch (err) {
        throw err;
    }
};

/**
 * The main init function
 * @returns {Promise<String | Error>} - The done message or error message
 */
const init = async () => {
    try {

        switch (SCRAPE_MODE) {
            case 'HOTEL': return await hotelScraperInit();
            case 'RESTO': return await restoScraperInit();
            default: throw Error('Invalid Scrap Mode');
        }

    } catch (err) {
        throw err;
    }
};

// Start the program
init()
    .then(msg => console.log(chalk.bold.green(msg)))
    .catch(err => {
        console.log(err);
        process.exit(1);
    });

