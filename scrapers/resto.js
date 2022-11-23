// Dependencies
import chalk from 'chalk';


/**
 * Extract the review page urls, total review count, and total review page count
 * @param {String} restoUrl - The url of the restaurant page
 * @param {Number} position - The index of the restaurant page in the list
 * @param {String} language - The language of the reviews that you wantto scrape
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object | Error>} - The object containing the review count, page count, and the review page urls
 */
const extractAllReviewPageUrls = async (restoUrl, position, language, browser) => {
    try {

        // Open a new page
        const page = await browser.getNewPage()
        await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3419.0 Safari/537.36');

        // Navigate to the resto page
        await page.goto(restoUrl);

        // Wait for the content to load
        await page.waitForSelector('body');

        const reviewExists = await page.evaluate(() => {
            // if (document.querySelector('[id=filters_detail_language_filterLang_ALL]')) return true
            if (document.querySelector('[id=filters_detail_language_filterLang_fr]')) return true

            return false
        })

        if (!reviewExists) return browser.handBack(page);


        // Select all language
        // await page.click('[id=filters_detail_language_filterLang_ALL]');
        await page.click('[id=filters_detail_language_filterLang_fr]');


        await page.waitForTimeout(1000);

        // Determin current URL
        const currentURL = page.url();

        console.log(`${chalk.bold.white.dim('Gathering Info: ')}${currentURL.split('-')[4]} ${position}`);

        /**
         * In browser code:
         * Extract the review page url
         */
        const getReviewPageUrls = await page.evaluate(() => {

            // Get the total review count
            // const totalReviewCount = parseInt(document
            //     .getElementsByClassName('reviews_header_count')[0]
            //     .innerText.split('(')[1]
            //     .split(')')[0]
            //     .replace(',', ''));
            const reviewEelement = document.getElementsByClassName('count')
            let totalReviewCount = 0
            for (let index = 0; index < reviewEelement.length; index++) {
                if (reviewEelement[index].parentElement.innerText.split('(')[0].split(' ')[0] === 'French') {
                    totalReviewCount = parseInt(document.getElementsByClassName('count')[index].innerText.split('(')[1].split(')')[0])
                }
            }

            // Default review page count
            let noReviewPages = totalReviewCount / 15;

            // Calculate the last review page
            if (totalReviewCount % 15 !== 0) {
                noReviewPages = ((totalReviewCount - totalReviewCount % 15) / 15) + 1;
            }

            // Get the url of the 2nd page of review. The 1st page is the input link
            let url = false;

            // If there is more than 1 review page
            if (document.getElementsByClassName('pageNum').length > 0) {
                url = document.getElementsByClassName('pageNum')[1].href;
            }

            return {
                noReviewPages,
                url,
                totalReviewCount,
            };
        });

        // Destructure function outputs
        let { noReviewPages, url, totalReviewCount, } = getReviewPageUrls;

        // Array to hold all the review page urls
        const reviewPageUrls = [];

        // If there is more than 1 review page, create the review page url base on the rule below
        if (url) {

            let counter = 0;
            // Replace the url page count till the last page
            while (counter < noReviewPages - 1) {
                counter++;
                url = url.replace(/-or[0-9]*/g, `-or${counter * 15}`);
                reviewPageUrls.push(url);
            }
        }

        // Add the first page url
        reviewPageUrls.unshift(restoUrl);

        // Information for logging
        const data = {
            count: totalReviewCount,
            pageCount: reviewPageUrls.length,
            urls: reviewPageUrls,
        };
        console.log(data)
        // Hand back the page so it's available again
        browser.handBack(page);

        return data;

    } catch (err) {
        throw err;
    }
};

/**
 * Extract the reviews and write to a JSON file
 * @param {Number} totalReviewCount - The total review count
 * @param {Array<String>} reviewPageUrls - The array containing review page urls
 * @param {Number} position - The index of the restaurant page in the list
 * @param {String} restoName - The name of the restaurant
 * @param {String} restoId - The id of the restaurant
 * @param {String} language - The language of the reviews that you wantto scrape
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object | Error>} - The final data
 */
const scrape = async (totalReviewCount, reviewPageUrls, position, restoName, restoId, language, browser) => {
    try {

        // Open a new page
        const page = await browser.getNewPage()
        await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3419.0 Safari/537.36');

        // Array to hold all the reviews 
        const allReviews = [];

        // Loop through all the review pages and extract the reviews
        for (let index = 0; index < reviewPageUrls.length; index++) {

            // Navigate to each review page
            await page.goto(reviewPageUrls[index], { waitUntil: 'networkidle2', });

            // Wait for the content to load
            await page.waitForSelector('body');

            // Select all language
            // await page.click('[id=filters_detail_language_filterLang_ALL]');
            await page.click('[id=filters_detail_language_filterLang_fr]');

            await page.waitForTimeout(1000);

            const reviewExpandable = await page.evaluate(() => {
                if (document.querySelector('.taLnk.ulBlueLinks')) return true
                return false
            })

            if (reviewExpandable) {

                // Expand the reviews
                await page.click('.taLnk.ulBlueLinks');

                // Wait for the reviews to load
                await page.waitForFunction('document.querySelector("body").innerText.includes("Show less")');
            }



            // Determine current URL
            const currentURL = page.url();

            // Progress Report
            console.log({
                'Scraping': currentURL,
                'Pages Left': reviewPageUrls.length - 1 - index,
                'Progress': `${Math.round(((index + 1) / reviewPageUrls.length * 100), 1)}%`,
            });

            const reviews = await page.evaluate(() => {

                const results = [];

                const items = document.body.querySelectorAll('.review-container');
                items.forEach(item => {

                    /* Get and format Rating */
                    let ratingElement = item.querySelector('.ui_bubble_rating').getAttribute('class');
                    let integer = ratingElement.replace(/[^0-9]/g, '');
                    let parsedRating = parseInt(integer) / 10;

                    /* Get and format date of Visit */
                    let dateOfVisitElement = item.querySelector('.prw_rup.prw_reviews_stay_date_hsx').innerText;
                    let parsedDateOfVisit = dateOfVisitElement.replace('Date of visit:', '').trim();

                    // Push the review to the result array
                    results.push({
                        rating: parsedRating,
                        dateOfVisit: parsedDateOfVisit,
                        ratingDate: item.querySelector('.ratingDate').getAttribute('title'),
                        title: item.querySelector('.noQuotes').innerText,
                        content: item.querySelector('.partial_entry').innerText,
                    });

                });
                return results;

            });

            // Push the reviews to the array
            allReviews.push(...reviews);

        }

        // Data structure to be written to file
        const finalData = {
            restoName,
            restoId,
            count: totalReviewCount,
            actualCount: allReviews.length,
            position,
            allReviews,
            fileName: `${position}_${reviewPageUrls[0].split('-')[4]}`,
        };

        // Hand back the page so it's available again
        browser.handBack(page);

        return finalData;

    } catch (err) {
        throw err;
    }
};

/**
 * Start the scraping process
 * @param {String} restoUrl - The url of the restaurant page 
 * @param {String} restoName - The name of the restaurant
 * @param {String} restoId - The id of the restaurant
 * @param {Number} position - The index of the restaurant page in the list
 * @param {String} language - The language of the reviews to scrape
 * @param {Object} browser - A browser instance
 * @returns {Promise<Object | Error>} - The final data
 */
const start = async (restoUrl, restoName, restoId, position, language, browser) => {
    try {

        const extracted = await extractAllReviewPageUrls(restoUrl, position, language, browser);

        // If the resto has no reviews
        if (!extracted) return {
            restoName,
            restoId,
            count: 0,
            actualCount: 0,
            position,
            allReviews: [{
                rating: 0,
                dateOfVisit: 0,
                ratingDate: 0,
                title: 0,
                content: 0,
            }],
            fileName: `${position}_${restoUrl.split('-')[4]}`,
        };

        const { urls, count, } = extracted

        const results = await scrape(count, urls, position, restoName, restoId, language, browser);

        return results;

    } catch (err) {
        throw err;
    }
};


export default start;