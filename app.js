// Dependencies
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import fs, { mkdirSync, } from 'fs';
const { promises: { writeFile, }, } = fs;
import chalk from 'chalk';
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Custom Modules 
import hotelScraper from './scrapers/hotel.js';
import restoScraper from './scrapers/resto.js';
import {
    csvToJSON, fileExists, combine,
    reviewJSONToCsv, dataProcessor
} from './libs/utils.js';
import { browserInstance } from './libs/browser.js'

// Data path
const dataDir = join(__dirname, './reviews/');
const sourceDir = join(__dirname, './source/');

// Environment variables
let { SCRAPE_MODE, CONCURRENCY } = process.env;
CONCURRENCY = parseInt(CONCURRENCY);
if (!CONCURRENCY) CONCURRENCY = 2;


console.log(chalk.bold.blue(`The Scraper is Running in ${chalk.bold.magenta(SCRAPE_MODE)} Mode`));
console.log(chalk.bold.blue(`Concurrency Setting ${chalk.bold.magenta(CONCURRENCY || 2)}`));

// Check if the required directories exist, otherwise create them
if (!fileExists(dataDir)) mkdirSync(dataDir);
if (!fileExists(sourceDir)) mkdirSync(sourceDir);

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


        const [rawData] = await Promise.all([
            // Convert the csv to json
            csvToJSON(dataSourceHotel),
            // Get a browser instance
            browserInstance.launch()
        ])


        console.log(chalk.bold.yellow(`Scraping ${chalk.magenta(rawData.length)} Hotels`));

        // Array to hold the processed data
        const reviewInfo = []

        // Array to hold the promises to be processed
        let processQueue = []

        // Extract review info and file name of each individual hotel
        for (let index = 0; index < rawData.length; index++) {

            if (processQueue.length > CONCURRENCY) {
                const finalData = await dataProcessor(processQueue)
                reviewInfo.push(finalData);
                processQueue = []
            }

            // Extract hotel info
            const item = rawData[index];
            const { webUrl: hotelUrl, name: hotelName, id: hotelId, } = item;

            processQueue.push(hotelScraper(hotelUrl, hotelName, hotelId, index, browserInstance))
        }

        // Resolve processes the left over in the process queue
        const finalData = await dataProcessor(processQueue)
        reviewInfo.push(finalData);

        // Write the review of each individual hotel to files
        await Promise.all(
            reviewInfo
                .flat()
                .map(async ({ finalData, fileName }) => {
                    const dataToWrite = JSON.stringify(finalData, null, 2);
                    await writeFile(`${dataDir}${fileName}.json`, dataToWrite);
                })
        );

        // Combine all the reviews into an array of objects
        const combinedData = combine(SCRAPE_MODE, dataDir);

        // Write the combined JSON data to file
        const jsonData = JSON.stringify(combinedData, null, 2);

        // Convert the combined JSON data to csv
        const csvData = reviewJSONToCsv(combinedData);


        await Promise.all([
            writeFile(`${dataDir}All.json`, jsonData),
            writeFile(`${dataDir}All.csv`, csvData),
            // Close the browser instance
            browserInstance.closeBrowser(),
        ])

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

        const [rawData] = await Promise.all([
            // Convert the csv to json
            csvToJSON(dataSourceResto),
            // Initiate a browser instance
            browserInstance.launch()
        ])

        console.log(chalk.bold.yellow(`Scraping ${chalk.magenta(rawData.length)} Restaurants`));

        // Array to hold the processed data
        const reviewInfo = []

        // Array to hold the promises to be processed
        let processQueue = []

        for (let index = 0; index < rawData.length; index++) {

            if (processQueue.length > CONCURRENCY) {
                const finalData = await dataProcessor(processQueue)
                reviewInfo.push(finalData);
                processQueue = []
            }

            // Extract resto info
            const item = rawData[index];
            const { webUrl: restoUrl, name: restoName, id: restoId, } = item;

            processQueue.push(restoScraper(restoUrl, restoName,
                restoId, index, browserInstance))
        }

        // Resolve processes the left over in the process queue
        const finalData = await dataProcessor(processQueue)
        reviewInfo.push(finalData);

        // Write the review of each individual resto to files
        await Promise.all(
            reviewInfo
                .flat()
                .map(async ({ finalData, fileName }) => {
                    const dataToWrite = JSON.stringify(finalData, null, 2);
                    await writeFile(`${dataDir}${fileName}.json`, dataToWrite);
                })
        );

        // Combine all the reviews into an array of objects
        const combinedData = combine(SCRAPE_MODE, dataDir);

        // Write the combined JSON data to file
        const jsonData = JSON.stringify(combinedData, null, 2);

        // Convert the combined JSON data to csv
        const csvData = reviewJSONToCsv(combinedData);

        // Write the combined data to files and close the browser instance
        await Promise.all([
            writeFile(`${dataDir}All.json`, jsonData),
            writeFile(`${dataDir}All.csv`, csvData),
            // Close the browser instance
            browserInstance.closeBrowser(),
        ])

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


// Report Logic
setInterval(async () => {
    // Get tab stats
    const report = await browserInstance.reportTabStats()
    // Extract heartbeat info
    const { heartbeat: { inUse, openedPage } } = report;
    // Log the tab stats
    console.log(report);
    // Exit if no tab is in use and browser is being loaded
    if (inUse === 0 && typeof (openedPage) !== 'string') process.exit(0);
}, 5000);