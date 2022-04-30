// Dependencies
const path = require('path');
const { existsSync, mkdirSync, writeFileSync, } = require('fs');

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
        writeFileSync(`/${dataDir}reviews.csv`, csv);
    } catch (err) {
        throw err;
    }
};

const restoScraperInit = async () => {
    try {
        const sourceFileAvailable = await fileExists(dataSource);

        if (!sourceFileAvailable) {
            throw Error('Source file does not exist');
        }

        const rawData = await restoCsvToJSON(dataSource);

        await Promise.all(
            rawData.map(async (item, index) => {

                const { webUrl: restoUrl, name: restoName, id: restoId, } = item;

                // Logging
                console.log('Now Is', [index], restoUrl);

                const finalData = await restoScraper(restoUrl, restoName, restoId, index);
                const { fileName, } = finalData;
                delete finalData.fileName;

                // Write to file
                writeFileSync(`${dataDir}${fileName}.json`,
                    JSON.stringify(finalData, null, 2));
            })
        );

    } catch (err) {
        throw err;
    }
};

const init = async () => {
    try {

        switch (SCRAP_MODE) {
            case 'HOTEL': await hotelScraperInit();
                break;
            case 'RESTO': await restoScraperInit();
                break;
            default:
                throw Error('Invalid Scrap Mode');
        }

    } catch (err) {
        throw err;
    }
};

init().catch(err => {
    console.log(err);
    process.exit(1);
});
