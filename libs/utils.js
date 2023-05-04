// Dependencies
import { readFileSync, readdirSync, accessSync } from 'fs';
import { parse } from 'json2csv';
import csvtojsonV2 from 'csvtojson';

/**
 * Check if the given file exists
 * @param {String} filePath 
 * @returns {<Boolean>}
 */
const fileExists = (filePath) => {
    try {
        accessSync(filePath);
        return true;
    } catch (err) {
        return false;
    }
};

/**
 * Convert month string to number
 * @param {String} monthString - The month string
 * @returns {Number} - The month number
 */
const monthStringToNumber = (monthString) => {
    switch (monthString) {
        case 'January':
            return 1;
        case 'February':
            return 2;
        case 'March':
            return 3;
        case 'April':
            return 4;
        case 'May':
            return 5;
        case 'June':
            return 6;
        case 'July':
            return 7;
        case 'August':
            return 8;
        case 'September':
            return 9;
        case 'October':
            return 10;
        case 'November':
            return 11;
        case 'Jan':
            return 1;
        case 'Feb':
            return 2;
        case 'Mar':
            return 3;
        case 'Apr':
            return 4;
        case 'Jun':
            return 6;
        case 'Jul':
            return 7;
        case 'Aug':
            return 8;
        case 'Sep':
            return 9;
        case 'Oct':
            return 10;
        case 'Nov':
            return 11;
        default:
            return 12;
    }
};

/**
 * Convert comment rating string (the class name of the rating element) to number (the actual rating)
 * @param {String} ratingString - The class name of the rating element
 * @returns {Number} - The actual rating
 */
const commentRatingStringToNumber = (ratingString) => {

    switch (ratingString) {
        case "bubble_10":
            return 1;
        case "bubble_15":
            return 1.5;
        case "bubble_20":
            return 2;
        case "bubble_25":
            return 2.5;
        case "bubble_30":
            return 3;
        case "bubble_35":
            return 3.5;
        case "bubble_40":
            return 4;
        case "bubble_45":
            return 4.5;
        case "bubble_50":
            return 5;
        default:
            return 5;
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
            .filter(fileName => fileName.includes('.json') && fileName !== 'All.json')
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
            // Sort the extracted data by the index so all reviews of the same hotel are together
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
                const { hotelName, hotelId, title, content, month, year, rating } = review;

                // Check if the hotel ID is supplied
                if (!hotelId) return { hotelName, title, content, month, year, rating };
                return { hotelName, hotelId, title, content, month, year, rating };

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

/**
 * Take in an array of review promises, resolve it then reshape the review object
 * @param {Array<Promise<Object>>} arrayToBeProcessed 
 * @returns {Promise<Array<Object> | Error>}
 */
const dataProcessor = async (arrayToBeProcessed) => {
    try {
        const toBeProcessed = await Promise.all(arrayToBeProcessed);
        const processed = toBeProcessed.map(data => {
            const { fileName, } = data;
            delete data.fileName;
            return { finalData: data, fileName }
        })
        return processed
    } catch (err) {
        throw err;
    }
}

/**
 * Block all bs, and keep html only
 * @param {puppeteer.Browser.Page} page - The puppeteer page object
 */
const noBs = async (page) => {
    try {
        // Enable request interception
        await page.setRequestInterception(true);

        // Block all images
        page.on('request', (req) => {
            if (req.resourceType() === 'image' ||
                req.resourceType() === 'stylesheet'
                || req.resourceType() === 'font') return req.abort();
            return req.continue();
        });
    } catch (err) {
        throw err
    }
};

export { fileExists, combine, reviewJSONToCsv, csvToJSON, dataProcessor, noBs, monthStringToNumber, commentRatingStringToNumber };


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