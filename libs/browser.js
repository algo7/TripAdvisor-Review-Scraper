// Dependencies
import puppeteer from 'puppeteer';

const config = {
    timeout: 0,
    headless: true,
    devtools: false,
    defaultViewport: {
        width: 1920,
        height: 1080,
    },
    args: [
        '--disable-gpu',
        '--disable-dev-shm-usage',
        '--disable-setuid-sandbox',
        '--no-sandbox'
    ],

};

let instance = null;

/**
 * Get a browser instance
 * @returns {Prmoise<Object | Error>} - Browser instance
 */
const getBrowserInstance = async () => {
    try {

        if (!instance) instance = await puppeteer.launch(config);

        return instance;
    } catch (err) {
        throw err;
    }
};

/**
 * Close a browser instance
 * @returns {Prmoise<undefined | Error>} - Browser instance
 */
const closeBrowserInstance = async () => {
    try {

        if (!instance) {
            throw Error('Now Browser instance has been launched');
        }

        await instance.close();

    } catch (err) {
        throw err;
    }
};

export { getBrowserInstance, closeBrowserInstance };

