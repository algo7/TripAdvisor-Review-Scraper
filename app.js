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
import { csvToJSON, fileExists, combine, reviewJSONToCsv } from './libs/utils.js';
import { browserInstance } from './libs/browser.js'

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


        const [rawData, browser] = await Promise.all([
            // Convert the csv to json
            csvToJSON(dataSourceHotel),
            // Get a browser instance
            getBrowserInstance()
        ])


        console.log(chalk.bold.yellow(`Scraping ${chalk.magenta(rawData.length)} Hotels`));




        // Extract review info and file name of each individual hotel
        const reviewInfo = await Promise.all(
            rawData.map(async (item, index) => {
                // Extract resto info
                const { webUrl: hotelUrl, name: hotelName, id: hotelId, } = item;

                // Start the scraping process
                const finalData = await hotelScraper(hotelUrl, hotelName, hotelId, index, browser);
                const { fileName, } = finalData;
                delete finalData.fileName;
                return { finalData, fileName, };
            })
        );


        await Promise.all(
            reviewInfo.map(async ({ finalData, fileName }) => {
                // Write the review of each individual hotel to files
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
            closeBrowserInstance(),
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


        // Extract review info and file name of each individual resto
        const reviewInfo = await Promise.all(
            rawData.map(async (item, index) => {
                // Extract resto info
                const { webUrl: restoUrl, name: restoName, id: restoId, } = item;
                // Start the scraping process
                const finalData = await restoScraper(restoUrl, restoName,
                    restoId, index, browserInstance);
                const { fileName, } = finalData;
                delete finalData.fileName;
                return { finalData, fileName, };
            })
        );

        await Promise.all(
            reviewInfo.map(async ({ finalData, fileName }) => {
                // Write the review of each individual resto to files
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
            // browserInstance.closeBrowser(),
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






// (async () => {


//     for (let index = 0; index < 20; index++) {
//         const page = await browserInstance.getNewPage()
//         const pageInUse = browserInstance.getInUsePageCount()
//         console.log(`In Use Page Count: ${pageInUse}`)
//         await page.goto('https://www.tripadvisor.com')
//         // Opened page count
//         const pageCount = await browserInstance.countPage()
//         console.log(`Page Count: ${pageCount}`)

//         const pageAva = browserInstance.getAvailablePageCount()
//         console.log(`Available Page Count: ${pageAva}`)

//         browserInstance.handBack(page)

//     }

//     const pageInUse = browserInstance.getInUsePageCount()
//     console.log(`In Use Page Count: ${pageInUse}`)

//     const pageAva = browserInstance.getAvailablePageCount()
//     console.log(`Available Page Count: ${pageAva}`)



//     await browserInstance.closeBrowser()
// })().catch(err => console.log(err))
