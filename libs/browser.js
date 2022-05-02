// Dependencies
import puppeteer from 'puppeteer';

/**
 * Create a browser instance
 */
class Browser {

    constructor() {
        // Puppeteer configs
        this.config = {
            headless: false,
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
        this.pageAvailable = []

        this.pageOpenRequests = 1
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
  * Count the number of pages started by the browser
  * @returns {Promise<Number>}
  */
    async countPage() {
        const pages = await this.browser.pages()
        return pages.length
    }

    /**
     * Open a new page or get a page from the available pages
     * @returns {Promise<puppeteer.Browser.Page>}
     */
    async getNewPage() {

        // // Return a new page if not browser hasn't been launched
        // if (!this.browser) {
        //     this.browser = await this.launch()
        //     const newPage = await this.browser.newPage()
        //     this.pageInUse.push(newPage)
        //     return newPage
        // }

        /**
        * Return a new page if the amount of opened page is less than 10
        * It's written as 11 as there is one unsable blank page by default
        */

        console.log(`Opened Page: ${await this.countPage()}`)
        this.pageOpenRequests = this.pageOpenRequests + 1
        console.log(`Opened Page: ${this.pageOpenRequests}`)

        if (this.pageOpenRequests < 11) {
            const newPage = await this.browser.newPage()
            this.pageInUse.push(newPage)
            return newPage
        }

        // If there are pages available
        if (this.pageAvailable.length > 0) {
            // Get the first page in the available page array
            const page = this.pageAvailable.shift()
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
        this.pageAvailable.push(page)
    }

    /**
     * Close the browser instance
     * @returns {Promise<Undefined>}
     */
    async closeBrowser() {
        await this.browser.close()
    }

    /**
     * Return the number of pages available
     * @returns {Number}
     */
    getAvailablePageCount() {
        return this.pageAvailable.length
    }

    /**
     * Return the number of pages in use
     * @returns {Number}
     */
    getInUsePageCount() {
        return this.pageInUse.length
    }
}

// Initialize a new browser instance
const browserInstance = new Browser();

export { browserInstance };

