// Dependencies
const csvtojsonV2 = require("csvtojson");
const { writeFileSync } = require('fs');

const csvToJSON = async (csvFilePath) => {
    try {
        const parsedJson = await csvtojsonV2().fromFile(csvFilePath)

        const processed = parsedJson.map(resto => {
            return {
                name: resto.name,
                webUrl: resto.webUrl,
            }
        });

        writeFileSync(`${csvFilePath}.json`, JSON.stringify(processed, null, 2));

        return processed

    } catch (err) {
        throw err
    }

};

// csvToJSON('../resto.csv').then(json => console.log(json));