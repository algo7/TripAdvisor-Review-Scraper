// Dependencies
const puppeteer = require('puppeteer');

/**
 * Extract the review page urls, total review count, and total review page count
 * @param {String} restoUrl - The url of the restaurant page
 * @returns {Promise<Object | Error>} - The object containing the review count, page count, and the review page urls
 */
const extractAllReviewPageUrls = async (restoUrl) => {
    try {

        // Launch the browser
        const browser = await puppeteer.launch({
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

        // Open a new page
        const page = await browser.newPage();

        // Navigate to the resto page
        await page.goto(restoUrl);

        // Wait for the content to load
        await page.waitForSelector('body');

        // Select all language
        await page.click('[id=filters_detail_language_filterLang_ALL]');

        await page.waitForTimeout(1000);

        // Expand the reviews
        await page.click('.taLnk.ulBlueLinks');

        // Wait for the reviews to load
        await page.waitForFunction('document.querySelector("body").innerText.includes("Show less")');

        // Extract the review page url
        const getReviewPageUrls = await page.evaluate(() => {

            // Get the total review count
            const totalReviewCount = parseInt(document
                .getElementsByClassName('reviews_header_count')[0]
                .innerText.split('(')[1]
                .split(')')[0]
                .replace(',', ''));

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
        // console.log(data);

        await browser.close();

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
 * @returns {Promise<Object | Error>} - The final data
 */
const scrap = async (totalReviewCount, reviewPageUrls, position, restoName, restoId) => {
    try {

        // Launch the browser
        const browser = await puppeteer.launch({
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

        // Open a new page
        const page = await browser.newPage();

        // Array to hold all the reviews 
        const allReviews = [];

        // Loop through all the review pages and extract the reviews
        for (let index = 0; index < reviewPageUrls.length; index++) {

            // Navigate to each review page
            await page.goto(reviewPageUrls[index], { waitUntil: 'networkidle2', });

            // Wait for the content to load
            await page.waitForSelector('body');

            // Select all language
            await page.click('[id=filters_detail_language_filterLang_ALL]');

            await page.waitForTimeout(1000);

            // Determine current URL
            const currentURL = page.url();
            console.log(`Scraping: ${currentURL} | ${reviewPageUrls.length - 1 - index} Pages Left`);

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
            allReviews.push(reviews);

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

        await browser.close();

        return finalData;

    } catch (err) {
        throw err;
    }
};

/**
 * Start the scrapping process
 * @param {String} restoUrl - The url of the restaurant page 
 * @param {String} restoName - The name of the restaurant
 * @param {String} restoId - The id of the restaurant
 * @param {Number} position - The index of the restaurant page in the list
 * @returns {Promise<Object | Error>} - The final data
 */
const start = async (restoUrl, restoName, restoId, position) => {
    try {

        const { urls, count, } = await extractAllReviewPageUrls(restoUrl);

        const finalData = await scrap(count, urls, position, restoName, restoId);

        return finalData;

    } catch (err) {
        throw err;
    }
};


module.exports = start;

// (async () => {
//     // Loop through the list of restaurants
//     for (let index = 0; index < items.length; index++) {

//         // Get the restaurant url
//         const restoUrl = items[index].webUrl;

//         // Get the restaurant name
//         const restoName = items[index].name;

//         // Get the restaurant id
//         const restoId = items[index].id;

//         // Logging
//         console.log('Now Is', [index], restoUrl);

//         // Start the scrapping process
//         const isDone = await start(restoUrl, restoName, restoId, index);

//         console.log(isDone);
//     }
// })().catch(err => console.log(err));


// extractAllReviewPageUrls('https://www.tripadvisor.com/Restaurant_Review-g652156-d17621567-Reviews-Kalasin-Bulle_La_Gruyere_Canton_of_Fribourg.html').then(x => console.log(x)).catch(err => console.log(err));
// const a = require('./data/Kalasin.json');
// a.allReviews.forEach(x => console.log(x.title));

// The review count array based in the "Traveler rating info"
//   const reviewCount = [];

// // Extract the review count for each rating
// document.getElementsByClassName('choices')[0].querySelectorAll('.row_num').forEach(el => reviewCount.push(el.innerText));

// // Sum them to get the total review count
// const totalReviewCount = reviewCount.map(count => parseInt(count)).reduce((a, b) => a + b);
