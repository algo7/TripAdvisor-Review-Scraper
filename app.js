// Dependencies
const path = require('path');
const { existsSync, mkdirSync, } = require('fs');
const { writeFile, } = require('fs/promises');

// Custom Modules
const hotelScraper = require('./scrapers/hotel');
const restoScraper = require('./scrapers/resto');
const { restoCsvToJSON, fileExists, } = require('./utils');

// Data path
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

// Data source
const dataSource = path.join(__dirname, './data/resto.csv');


const hotelScraperInit = async () => {
    try {
        const csv = await hotelScraper();
        await writeFile(`/${dataDir}reviews.csv`, csv);
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
        const sourceFileAvailable = await fileExists(dataSource);
        if (!sourceFileAvailable) {
            throw Error('Source file does not exist');
        }

        // Convert the csv to json
        const rawData = await restoCsvToJSON(dataSource);
        console.log(`Scraping ${rawData.length} restaurants`);

        await Promise.all(
            rawData.map(async (item, index) => {
                // Extract resto info
                const { webUrl: restoUrl, name: restoName, id: restoId, } = item;
                // Start the scraping process
                const finalData = await restoScraper(restoUrl, restoName, restoId, index);
                const { fileName, } = finalData;
                delete finalData.fileName;
                // Write the data to file
                const dataToWrite = JSON.stringify(finalData, null, 2);
                await writeFile(`${dataDir}${fileName}.json`, dataToWrite);
            })
        );

        return 'Scraping Done';

    } catch (err) {
        throw err;
    }
};

/**
 * The init function
 * @returns {Promise<String | Error>} - The done message or error message
 */
const init = async () => {
    try {

        switch (SCRAP_MODE) {
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
    .then(msg => console.log(msg))
    .catch(err => {
        console.log(err);
        process.exit(1);
    });
