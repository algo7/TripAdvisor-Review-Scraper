// Dependencies
import fs, { readFileSync, readdirSync } from 'fs';
const { promises: { access, }, } = fs;
import { parse } from 'json2csv';
import csvtojsonV2 from 'csvtojson';

/**
 * Check if the given file exists
 * @param {String} filePath 
 * @returns {Promise<Boolean>}
 */
const fileExists = async (filePath) => {
    try {
        await access(filePath);
        return true;
    } catch (err) {
        return false;
    }
};

/**
 * Combine all JSON files in the data directory into a JSON array of object
 * @param {String} scrapeMode - Resturant or hotel
 * @param {String} dataDir - The data directory
 * @returns {Array<Object>} - The combined JSON array of review objects
 */
const combine = (scrapeMode, dataDir) => {
    try {
        // Read all files in the data directory
        const allFiles = readdirSync(dataDir);

        const extracted = allFiles
            // Filter out JSON files
            .filter(fileName => fileName.includes('.json'))
            // Load each file and extract the information
            .map(fileName => {
                const fileContent = JSON.parse(readFileSync(`${dataDir}${fileName}`));

                if (scrapeMode === 'RESTO') {
                    const { restoName, restoId, position, allReviews, } = fileContent;
                    return { restoName, restoId, position, allReviews, };
                }
                const { hotelName, hotelId, position, allReviews, } = fileContent;
                return { hotelName, hotelId, position, allReviews, };
            })
            // Sort the extracted data by the index
            .sort((a, b) => a.position - b.position)
            // Append the name, id, and index to each review
            .map(item => {
                if (scrapeMode === 'RESTO') {
                    const { restoName, restoId, position, allReviews, } = item;
                    return allReviews.map(review => {
                        review.restoName = restoName;
                        review.restoId = restoId;
                        review.position = position;
                        return review;
                    });
                }

                const { hotelName, hotelId, position, allReviews, } = item;
                return allReviews.map(review => {
                    review.hotelName = hotelName;
                    review.hotelId = hotelId;
                    review.position = position;
                    return review;
                });
            })
            .flat()
            // Rearrange the data for converting to CSV
            .map(review => {
                if (scrapeMode === 'RESTO') {
                    const { restoName, restoId, rating, dateOfVist,
                        ratingDate, title, content, } = review;
                    return {
                        restoName, restoId, title, content,
                        rating, dateOfVist, ratingDate,
                    };
                }
                const { hotelName, hotelId, title, content, } = review;
                return { hotelName, hotelId, title, content, };
            });

        return extracted;

    } catch (err) {
        throw err;
    }
};

/**
 * Convert JSON input to CSV
 * @param {Array<Object>} jsonInput - The JSON array of review and restaurant objects
 * @returns {String} - The CSV string
 */
const reviewJSONToCsv = (jsonInput) => {
    try {

        const fields = Object.keys(jsonInput[0]);
        const opts = { fields, };

        // Convert JSON to CSV
        const csv = parse(jsonInput, opts);

        return csv;

    } catch (err) {
        throw err;
    }
};

/**
 * Extract the name, url, and id of the resto from a csv file
 * @param {String} csvFilePath - The location of the csv file 
 * @param {String} scrapeMode - Resturant or hotel
 * @returns {Promise<Object | Error>} - The parsed json object or error message
 */
const csvToJSON = async (csvFilePath, scrapeMode) => {
    try {

        // Read the csv file
        const parsedJson = await csvtojsonV2().fromFile(csvFilePath);

        // Restaurant
        if (scrapeMode === 'RESTO') {
            return parsedJson.map(resto => {
                return { name: resto.name, webUrl: resto.webUrl, id: resto.id, };
            });
        }

        // Hotel
        return parsedJson.map(hotel => {
            return { name: hotel.name, webUrl: hotel.webUrl, id: hotel.id, };
        });

    } catch (err) {
        throw err;
    }

};

export { fileExists, combine, reviewJSONToCsv, csvToJSON };


// const cookiesAvailable = await fileExists('./data/cookies.json');

// if (!cookiesAvailable) {

//     // Navigate to the page below
//     await page.goto(myArgs[0] || process.env.URL);

//     // Log the cookies
//     const cookies = await page.cookies();
//     const cookieJson = JSON.stringify(cookies);
//     writeFileSync('./data/cookies.json', cookieJson);

//     // Close the browser
//     await browser.close();

//     // Exit the process
//     return await extractAllReviewPageUrls();
// }

// // Set Cookies
// const cookies = readFileSync('./data/cookies.json', 'utf8');
// const deserializedCookies = JSON.parse(cookies);
// await page.setCookie(...deserializedCookies);