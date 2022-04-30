// Dependencies
const { promises: { access, }, } = require('fs');
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

const restoJsonsToCsv = () => {
    try {
        const fields = ['title', 'content'];
        const opts = { fields, };


    } catch (err) {
        throw err;
    }
};

module.exports = { fileExists, };