// Dependencies
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import fs, { mkdirSync } from 'fs';
const { promises: { writeFile, }, } = fs;
import { Chalk } from 'chalk';
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
let { SCRAPE_MODE, CONCURRENCY, LANGUAGE,
    HOTEL_NAME, HOTEL_URL,
    IS_PROVISIONER // If the scraper is being called by the container provisioner
} = process.env;

CONCURRENCY = parseInt(CONCURRENCY);
if (!CONCURRENCY) CONCURRENCY = 2;
if (!LANGUAGE || LANGUAGE !== 'fr') LANGUAGE = 'en';
if (!SCRAPE_MODE) SCRAPE_MODE = 'HOTEL';

// Set the color level of the chalk instance
// 1 = basic color support (16 colors)
let colorLevel = 1;

// If the scraper is being called by the container provisioner, set the color level to 0
if (IS_PROVISIONER) {
    colorLevel = 0;
}
const customChalk = new Chalk({ level: colorLevel });



console.log(customChalk.bold.blue(`The Scraper is Running in ${customChalk.bold.magenta(SCRAPE_MODE)} Mode`));
console.log(customChalk.bold.blue(`Concurrency Setting ${customChalk.bold.magenta(CONCURRENCY || 2)}`));
console.log(customChalk.bold.blue(`Review Language ${customChalk.bold.magenta(LANGUAGE)}`));

// Check if the required directories exist, otherwise create them
if (!fileExists(dataDir)) mkdirSync(dataDir);
if (!fileExists(sourceDir)) mkdirSync(sourceDir);


// Data source
const dataSourceResto = join(__dirname, './source/restos.csv');
const dataSourceHotel = join(__dirname, './source/hotels.csv');
const dataSourceAirline = join(__dirname, './source/airlines.csv');

/**
 * Scrape the hotel pages
 * @returns {Promise<String | Error>} - The done message or error message
 */
const hotelScraperInit = async () => {
    try {

        // Get a browser instance
        await browserInstance.launch()

        // Get the raw data from env variables or csv file
        let rawData = [
            {
                name: HOTEL_NAME,
                webUrl: HOTEL_URL,
            }
        ]

        // If the env variables are not set, get the data from the csv file
        if (!rawData[0].name) {

            // Check if the source file exists
            const sourceFileAvailable = fileExists(dataSourceHotel);
            if (!sourceFileAvailable) {
                throw Error('Source file does not exist');
            }

            // Convert the csv to json
            rawData = await csvToJSON(dataSourceHotel)
        };

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

        // Resolve processes left over in the process queue
        const finalData = await dataProcessor(processQueue)
        reviewInfo.push(finalData);

        // Write the review of each individual hotel to json files
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


        // If the scraper is being called by the container provisioner, then export the csv only
        if (IS_PROVISIONER) {
            await Promise.all([
                writeFile(`${dataDir}All.csv`, csvData),
                // Close the browser instance
                browserInstance.closeBrowser(),
            ])

            fs.readdirSync(dataDir)
                .forEach(file => {
                    console.log(file);
                });

            console.log('Scraping Done');
            process.exit(0);

        }

        // Otherwise, export both the csv and json
        await Promise.all([
            writeFile(`${dataDir}All.json`, jsonData),
            writeFile(`${dataDir}All.csv`, csvData),
            // Close the browser instance
            browserInstance.closeBrowser(),
        ])

        console.log('Scraping Done');
        process.exit(0);

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
        const sourceFileAvailable = fileExists(dataSourceResto);
        if (!sourceFileAvailable) {
            throw Error('Source file does not exist');
        }

        const [rawData] = await Promise.all([
            // Convert the csv to json
            csvToJSON(dataSourceResto),
            // Initiate a browser instance
            browserInstance.launch()
        ])

        console.log(customChalk.bold.yellow(`Scraping ${customChalk.magenta(rawData.length)} Restaurants`));

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
                restoId, index, LANGUAGE, browserInstance))
        }

        // Resolve processes left over in the process queue
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
 * Scrape the resto pages
 * @returns {Promise<String | Error>} - The done message or error message
 */
const airlineScraperInit = async () => {
    try {

        // Check if the source file exists
        const sourceFileAvailable = fileExists(dataSourceAirline);
        if (!sourceFileAvailable) {
            throw Error('Source file does not exist');
        }

        const [rawData] = await Promise.all([
            // Convert the csv to json
            csvToJSON(dataSourceAirline),
            // Initiate a browser instance
            browserInstance.launch()
        ])

        console.log(customChalk.bold.yellow(`Scraping ${customChalk.magenta(rawData.length)} Restaurants`));

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
                restoId, index, LANGUAGE, browserInstance))
        }

        // Resolve processes left over in the process queue
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
    .then(msg => console.log(customChalk.bold.green(msg)))
    .catch(err => {
        console.log(err);
        process.exit(1);
    });


// Report Logic
setInterval(async () => {

    // Standalone mode
    if (!IS_PROVISIONER) {
        // Get tab stats
        const report = await browserInstance.reportTabStats()
        // Extract heartbeat info
        const { heartbeat: { inUse, openedPage } } = report;
        // Log the tab stats
        console.log(report);
        // Exit if no tab is in use and browser is being loaded
        if (inUse === 0 && typeof (openedPage) !== 'string') process.exit(0);
    }


}, 5000);