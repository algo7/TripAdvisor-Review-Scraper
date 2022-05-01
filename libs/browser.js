// Dependencies
import puppeteer from 'puppeteer';

let instance = null;

/**
 * Get a browser instance
 * @returns {Prmoise<Object | Error>} - Browser instance
 */
const getBrowserInstance = async () => {
    try {
        if (!instance)
            instance = await puppeteer.launch({
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
            });
        return instance;
    } catch (err) {
        throw err;
    }
};

export default getBrowserInstance;