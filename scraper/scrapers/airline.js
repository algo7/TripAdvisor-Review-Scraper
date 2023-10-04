import { Chalk } from 'chalk';

// Environment variables
let { IS_PROVISIONER } = process.env;

let colorLevel = 1;

if (IS_PROVISIONER) {
    colorLevel = 0;
}
const customChalk = new Chalk({ level: colorLevel });


/**
 * Extract review page url
 * @param {String} airlineUrl - The url of the airline page 
 * @param {Number} position - The index of the airline page in the list
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object | Error>} - The object containing the review page urls and the total review count
 */
const extractAllReviewPageUrls = async (hotelUrl, position, browser) => {
    try {

        // Open a new page 
        const page = await browser.getNewPage()

        // Navigate to the hotel page
        await page.goto(airlineUrl);

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
            let noReviewPages = totalReviewCount / 5;

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