// Dependencies
import { Chalk } from 'chalk';

// Environment variables
let { IS_PROVISIONER } = process.env;

// Set the color level of the chalk instance
// 1 = basic color support (16 colors)
let colorLevel = 1;

// If the scraper is being called by the container provisioner, set the color level to 0
if (IS_PROVISIONER) {
    colorLevel = 0;
}
const customChalk = new Chalk({ level: colorLevel });



/**
 * Extract review page url
 * @param {String} hotelUrl - The url of the hotel page 
 * @param {Number} position - The index of the hotel page in the list
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object | Error>} - The object containing the review page urls and the total review count
 */
const extractAllReviewPageUrls = async (hotelUrl, position, browser) => {
    try {
        // Open a new page
        const page = await browser.getNewPage()
        await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3419.0 Safari/537.36');

        // Navigate to the hotel page
        await page.goto(hotelUrl);

        // Wait for the content to load
        await page.waitForSelector('body');

        // Determin current URL
        const currentURL = page.url();

        console.log(`${customChalk.bold.white.dim('Gathering Info: ')}${currentURL.split('-')[4]} ${position}`);

        /**
         * In browser code:
         * Extract the review page url
        */
        const getReviewPageUrls = await page.evaluate(() => {

            // All review count (English Only)
            let langFilterValue = 0;
            document.querySelectorAll('[for*=LanguageFilter_]').forEach((el, index) => {
                if (el.innerText.includes('English')) langFilterValue = index;
            });

            const totalReviewCount = parseInt(document
                .querySelector(`[for= LanguageFilter_${langFilterValue}]`)
                .innerText.split('(')[1]
                .split(')')[0]
                .replace(',', ''));


            // Default review page count
            let noReviewPages = totalReviewCount / 10;

            // Calculate the last review page
            if (totalReviewCount % 10 !== 0) {
                noReviewPages = ((totalReviewCount - totalReviewCount % 10) / 10) + 1;
            }

            // Get the url of the 2nd page of review. The 1st page is the input link
            let url = false;

            // If there is more than 1 review page
            if (document.getElementsByClassName('pageNum').length > 0) {
                url = document.getElementsByClassName('pageNum')[1].href;
            }

            return { noReviewPages, url, totalReviewCount, };

        });

        // Destructure function outputs
        let { noReviewPages, url, totalReviewCount, } = getReviewPageUrls;

        // Array to hold all the review urls
        const reviewPageUrls = [];

        // If there is more than 1 review page, create the review page url base on the rule below
        if (url) {
            let counter = 0;
            // Replace the url page count till the last page
            while (counter < noReviewPages - 1) {
                counter++;
                url = url.replace(/-or[0-9]*/g, `-or${counter * 10}`);
                reviewPageUrls.push(url);
            }
        }

        // Add the first page url
        reviewPageUrls.unshift(hotelUrl);

        // Information for logging
        const data = {
            count: totalReviewCount,
            pageCount: reviewPageUrls.length,
            urls: reviewPageUrls,
        };

        // Hand back the page so it's available again
        browser.handBack(page);

        return data;

    } catch (err) {
        throw err;
    }
};

/**
 * Scrape the page
 * @param {Number} totalReviewCount - The total review count
 * @param {Array<String>} reviewPageUrls - The review page urls
 * @param {Number} [position] - The index of the hotel page in the list
 * @param {String} hotelName - The name of the hotel
 * @param {String} [hotelId] - The id of the hotel
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object| Error>} - THe final data
 */
const scrape = async (totalReviewCount, reviewPageUrls, position, hotelName, hotelId, browser) => {
    try {

        // Open a new page
        const page = await browser.getNewPage()
        await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3419.0 Safari/537.36');

        // Array to hold the review info
        const allReviews = [];

        for (let index = 0; index < reviewPageUrls.length; index++) {

            // Navigate to the page below
            await page.goto(reviewPageUrls[index], { waitUntil: 'networkidle2', });

            // Wait for the content to load
            await page.waitForSelector('body');

            // Determin current URL
            const currentURL = page.url();

            // Progress Report
            console.log({
                'Scraping': currentURL,
                'Pages Left': reviewPageUrls.length - 1 - index,
                'Progress': `${Math.round(((index + 1) / reviewPageUrls.length * 100), 1)}% `,
            });

            // In browser code
            // Extract comments title
            const commentTitle = await page.evaluate(async () => {

                // Extract a tags
                const commentTitleBlocks = document.getElementsByClassName('Qwuub')

                // Array to store the comment titles
                const titles = [];

                // Higher order functions don't work in the browser
                for (let index = 0; index < commentTitleBlocks.length; index++) {
                    titles.push(commentTitleBlocks[index].children[0].innerText);
                }

                return titles;
            });

            // Extract comments text
            const commentContent = await page.evaluate(async () => {

                const commentContentBlocks = document.getElementsByTagName('q');

                // Array use to store the comments
                const comments = [];

                for (let index = 0; index < commentContentBlocks.length; index++) {
                    comments.push(commentContentBlocks[index].children[0].innerText);
                }

                return comments;
            });

            // Format (for CSV processing) the reviews so each review of each page is in an object
            const formatted = commentContent.map((comment, index) => {
                return {
                    title: commentTitle[index],
                    content: comment,
                };
            });

            // Push the formmated review to the  array
            allReviews.push(formatted);

        }

        // Hand back the page so it's available again
        browser.handBack(page);

        // Convert 2D array to 1D
        const reviewFlattened = allReviews.flat();

        // Data structure to be written to file
        const finalData = {
            hotelName,
            hotelId,
            count: totalReviewCount,
            actualCount: reviewFlattened.length,
            position,
            allReviews: reviewFlattened,
            fileName: `${position}_${reviewPageUrls[0].split('-')[4]}`,
        };

        return finalData;

    } catch (err) {
        throw err;
    }
};

/**
 * Start the scraping process
 * @param {String} hotelUrl - The url of the hotel page 
 * @param {String} hotelName - The name of the hotel
 * @param {String} [hotelId] - The id of the hotel
 * @param {Number} [position] - The index of the hotel page in the list
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object | Error>} - The final data
 */
const start = async (hotelUrl, hotelName, hotelId, position, browser) => {
    try {
        const { urls, count, } = await extractAllReviewPageUrls(hotelUrl, position, browser);

        const results = await scrape(count, urls, position, hotelName, hotelId, browser);

        return results;

    } catch (err) {
        throw err;
    }
};

export default start;
