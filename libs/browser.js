// Dependencies
import puppeteer from 'puppeteer';

/**
 * Create a browser instance
 */
class Browser {

    constructor() {
        // Puppeteer configs
        this.config = {
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
        // The browser instance
        this.browser = null

        // Number of pages in use
        this.pageInUse = []

        // Number of pages available
        this.pageIdle = []
    }

    /**
     * Private method to launch the browser
     * @returns {Promise<puppeteer.Browser>}
     */
    async launch() {
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
            newPage.setDefaultNavigationTimeout(0)

            this.pageInUse.push(newPage)
            return newPage
        }

        /**
        * Return a new page if the amount of opened page is less than 4
        * It's written as 5 as there is one unsable blank page by default
        */
        const openedPage = await this.#countPage()

        if (openedPage < 5) {
            const newPage = await this.browser.newPage()
            newPage.setDefaultNavigationTimeout(0)
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

    }

    /**
     * Put the page back to the available page array
     * @param {puppeteer.Browser.Page} page A page instance
     * @returns {Undefined}
     */
    handBack(page) {
        this.pageInUse.shift()
        this.pageIdle.push(page)
    }

    /**
    * Report Tab Stats
    * @returns {Object} 
    */
    reportTabStats() {
        return {
            Idle: this.#getAvailablePageCount(),
            inUse: this.#getInUsePageCount(),
        }
    }

    /**
    * Close the browser instance
    * @returns {Promise<Undefined>}
    */
    async closeBrowser() {
        await this.browser.close()
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
    * @returns {Promise<Number>}
    */
    async #countPage() {
        const pages = await this.browser.pages()
        return pages.length
    }
}

// Initialize a new browser instance
const browserInstance = new Browser();

export { browserInstance };

