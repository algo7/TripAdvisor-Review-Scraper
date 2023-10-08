// Dependencies
import puppeteer from 'puppeteer-extra'
import AdblockerPlugin from 'puppeteer-extra-plugin-adblocker';
import blockResourcesPlugin from 'puppeteer-extra-plugin-block-resources';
import randUserAgent from "rand-user-agent";
import useProxy from 'puppeteer-page-proxy';

import { EventEmitter } from 'events';

// Environments variables
let { CONCURRENCY } = process.env;
CONCURRENCY = parseInt(CONCURRENCY) + 1;
if (!CONCURRENCY) CONCURRENCY = 3;

/**
 * Create a browser instance
 */
class Browser extends EventEmitter {

    constructor() {

        // Call the parent constructor
        super();

        // Puppeteer configs
        this.config = {
            headless: false,
            devtools: true,
            defaultViewport: {
                width: 1280,
                height: 1024,
            },
            args: [
                '--disable-gpu',
                '--disable-web-security',
                '--disable-dev-shm-usage',
                '--disable-setuid-sandbox',
                '--no-sandbox',
                '--autoplay-policy=user-gesture-required',
                '--disable-background-networking',
                '--disable-background-timer-throttling',
                '--disable-backgrounding-occluded-windows',
                '--disable-breakpad',
                '--disable-client-side-phishing-detection',
                '--disable-component-update',
                '--disable-default-apps',
                '--disable-domain-reliability',
                '--disable-extensions',
                '--disable-features=AudioServiceOutOfProcess',
                '--disable-hang-monitor',
                '--disable-ipc-flooding-protection',
                '--disable-notifications',
                '--disable-offer-store-unmasked-wallet-cards',
                '--disable-popup-blocking',
                '--disable-print-preview',
                '--disable-prompt-on-repost',
                '--disable-renderer-backgrounding',
                '--disable-setuid-sandbox',
                '--disable-speech-api',
                '--disable-sync',
                '--hide-scrollbars',
                '--ignore-gpu-blacklist',
                '--metrics-recording-only',
                '--mute-audio',
                '--no-default-browser-check',
                '--no-first-run',
                '--no-pings',
                '--no-zygote',
                '--password-store=basic',
                '--use-gl=swiftshader',
                '--use-mock-keychain',
                '--disable-gl-drawing-for-tests',
                '-bwsi',
                '--disable-canvas-aa',
                '--disable-2d-canvas-clip-aa',
                '--disable-accelerated-2d-canvas',
                '--disable-infobars',
                '--ignore-certificate-errors',
            ],

        };
        // The browser instance
        this.browser = null

        // Number of pages in use
        this.pageInUse = []

        // Number of pages available
        this.pageIdle = []
    }

    /**
     * Method to launch the browser
     * @returns {Promise<puppeteer.Browser>}
     */
    async launch() {
        puppeteer
            .use(blockResourcesPlugin({ blockedTypes: new Set(['stylesheet', 'image', 'font', 'media', 'other']) }))
            .use(AdblockerPlugin({ blockTrackers: true }))
        this.browser = await puppeteer.launch(this.config)
        return this.browser
    }

    /**
     * Open a new page or get a page from the available pages
     * @returns {Promise<puppeteer.Browser.Page>}
     */
    async getNewPage() {

        // Return a new page if not browser hasn't been launched
        if (!this.browser) {
            this.browser = await this.launch()
            const newPage = await this.browser.newPage()
            await newPage.setDefaultTimeout(8 * 10000);
            await newPage.setDefaultNavigationTimeout(8 * 10000)
            // Generate a random user agent
            await newPage.setUserAgent(randUserAgent("desktop"))
            await useProxy(newPage, 'http://127.0.0.1:8888')
            this.pageInUse.push(newPage)

            return newPage
        }

        /**
        * Return a new page if the amount of opened page is less than 2
        * It's written as 3 as there is one unsable blank page by default
        */
        const openedPage = await this.#countPage()

        if (openedPage < CONCURRENCY) {
            const newPage = await this.browser.newPage()
            await newPage.setDefaultTimeout(8 * 10000);
            await newPage.setDefaultNavigationTimeout(8 * 10000)
            await newPage.setUserAgent(randUserAgent("desktop"))
            await useProxy(newPage, 'http://127.0.0.1:8888')
            this.pageInUse.push(newPage)
            return newPage
        }

        // If there are pages available
        if (this.pageIdle.length > 0) {
            // Get the first page in the available page array
            const page = this.pageIdle.shift()
            // Push the page into the in use page array
            this.pageInUse.push(page)
            return page
        }


        // Otherwise, wait until a page becomes available.
        await new Promise(resolve => this.once('pageIdle', resolve));
        return this.getNewPage();
    }

    /**
     * Put the page back to the available page array
     * @param {puppeteer.Browser.Page} page A page instance
     * @returns {Undefined}
     */
    handBack(page) {
        // Find the page in the in use page array
        const pageIndex = this.pageInUse.indexOf(page);

        // If the page is found
        if (pageIndex > -1) {
            this.pageInUse.splice(pageIndex, 1);
            this.pageIdle.push(page);
        }
    }

    /**
    * Close the browser instance
    * @returns {Promise<Undefined>}
    */
    async closeBrowser() {
        await this.browser.close()
    }

    /**
    * Report Tab Stats
    * @returns {Promise<Object>} 
    */
    async reportTabStats() {

        const openedPage = await this.#countPage()
        // If the countPage function is able to get the page opened
        if (openedPage) return {
            heartbeat: {
                Idle: this.#getAvailablePageCount(),
                inUse: this.#getInUsePageCount(),
                openedPage,
            }
        }

        // Pages are still being loaded
        return {
            heartbeat: {
                Idle: this.#getAvailablePageCount(),
                inUse: this.#getInUsePageCount(),
                openedPage: 'Browser is still loading',
            }
        }

    }

    // Private methods
    /**
     * Return the number of pages available
     * @returns {Number}
     */
    #getAvailablePageCount() {
        return this.pageIdle.length
    }

    /**
     * Return the number of pages in use
     * @returns {Number}
     */
    #getInUsePageCount() {
        return this.pageInUse.length
    }

    /**
    * Count the number of pages started by the browser.
    * Not so reliable when a lot pages are opened in a short time
    * @returns {Promise<Number> | Undefined}
    */
    async #countPage() {
        if (this.browser) {
            const pages = await this.browser.pages()
            return pages.length
        }

        return undefined
    }
}

// Initialize a new browser instance
const browserInstance = new Browser();

export { browserInstance };

