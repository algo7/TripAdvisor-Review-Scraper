// Dependencies
const { promises: { access, }, readdirSync, } = require('fs');
const { request, } = require('http');
const { parse, } = require('json2csv');

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
 * Convert JSON input to CSV
 * @param {Array<Object>} jsonInput - The JSON array of review and restaurant objects
 * @returns {String} - The CSV string
 */
const restoJsonsToCsv = (jsonInput) => {
    try {
        const fields = ['name', 'id'];
        const opts = { fields, };


    } catch (err) {
        throw err;
    }
};

/**
 * Combine all JSON files in the data directory into a JSON array of object
 * @returns {Array<Object>}
 */
const combine = () => {
    try {
        const allFiles = readdirSync('../data/');

        const extracted = allFiles.map(file => {

            // eslint-disable-next-line global-require
            const fileContent = require(`../data/${file}`);

            const { restoName, restoId, position, allReviews, } = fileContent;
            return {
                restoName,
                restoId,
                position,
                allReviews,

            };
        }).sort((a, b) => a.position - b.position);

        return extracted;

    } catch (err) {
        throw err;
    }
};
module.exports = { fileExists, combine, };
