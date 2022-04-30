// Dependencies
const csvtojsonV2 = require('csvtojson');
const { writeFileSync, } = require('fs');

/**
 * Extract the name, url, and id of the resto from a csv file
 * @param {String} csvFilePath - The location of the csv file 
 * @returns {Promise<Object | Error>} - The parsed json object or error message
 */
const restoCsvToJSON = async (csvFilePath) => {
    try {

        // Read the csv file
        const parsedJson = await csvtojsonV2().fromFile(csvFilePath);

        // Extract the fields
        const processed = parsedJson.map(resto => {
            return {
                name: resto.name,
                webUrl: resto.webUrl,
                id: resto.id,
            };
        });

        // Write to JSON file
        writeFileSync(`${csvFilePath}.json`, JSON.stringify(processed, null, 2));

        return processed;

    } catch (err) {
        throw err;
    }

};


module.exports = { restoCsvToJSON, };

// restoCsvToJSON('../resto.csv').then(json => console.log(json));