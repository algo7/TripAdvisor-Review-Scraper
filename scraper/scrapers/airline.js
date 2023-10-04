import { Chalk } from 'chalk';
import { monthStringToNumber, commentRatingStringToNumber } from '../libs/utils.js';

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
 * @param {puppeteer.Browser} browser - A browser instance
 * @returns {Promise<Object | Error>} - The object containing the review page urls and the total review count
 */
const extractAllReviewPageUrls = async (airlineUrl, position, browser) => {
    try {

        // Open a new page 
        const page = await browser.getNewPage()

        // Navigate to the airline page
        await page.goto(airlineUrl);

        // Wait for the content to load
        await page.waitForSelector('body');

        // Wait for the page to load
        await page.waitForTimeout(1000 * 5);

        // Determin current URL
        const currentURL = page.url();

        console.log(`${customChalk.bold.white.dim('Gathering Info: ')}${currentURL.split('-')[4] || "Main Page"} ${position}`);

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
            if (totalReviewCount % 5 !== 0) {
                noReviewPages = ((totalReviewCount - totalReviewCount % 5) / 5) + 1;
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
                url = url.replace(/-or[0-9]*/g, `-or${counter * 5}`);
                reviewPageUrls.push(url);
            }
        }

        // Add the first page url
        reviewPageUrls.unshift(airlineUrl);

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
 * Scrape the page
 * @param {Number} totalReviewCount - The total review count
 * @param {Array<String>} reviewPageUrls - The review page urls
 * @param {Number} [position] - The index of the airline page in the list
 * @param {String} airlineName - The name of the airline
 * @param {String} [airlineId] - The id of the airline
 * @param {puppeteer.Browser} browser - A browser instance
 * @returns {Promise<Object| Error>} - THe final data
 */
const scrape = async (totalReviewCount, reviewPageUrls, position, airlineName, airlineId, browser) => {
    try {

        // Array to hold the review info
        const allReviews = [];

        for (let index = 0; index < reviewPageUrls.length; index++) {
            // Open a new page
            const page = await browser.getNewPage()

            // Navigate to the page below
            await page.goto(reviewPageUrls[index], { waitUntil: 'networkidle2', });

            // Wait for the content to load
            await page.waitForSelector('body');

            const reviewExpandable = await page.evaluate(() => {
                if (document.querySelector('[data-test-target="expand-review"]')) return true
                return false
            })

            if (reviewExpandable) {

                // Expand the reviews
                await page.click("[data-test-target='expand-review'] > :first-child");

                // Wait for the reviews to load
                await page.waitForFunction('document.querySelector("body").innerText.includes("Read less")');
            }

            // Determine current URL
            const currentURL = page.url();

            // Progress Report
            if (!IS_PROVISIONER) {
                console.log({
                    'Scraping': currentURL,
                    'Pages Left': reviewPageUrls.length - 1 - index,
                    'Progress': `${Math.round(((index + 1) / reviewPageUrls.length * 100), 1)}%`,
                });
            } else {
                console.log("Scraping: ", currentURL)
                console.log("Pages Left: ", `${reviewPageUrls.length - 1 - index}`)
                console.log("Progress: ", `${Math.round(((index + 1) / reviewPageUrls.length * 100), 1)}%`)
            }


            // In browser code
            // Extract comments title
            const commentTitle = await page.evaluate(async () => {

                // Extract a tags
                const commentTitleBlocks = document.querySelectorAll('[data-test-target="review-title"]');

                // Array to store the comment titles
                const titles = [];

                // Higher order functions don't work in the browser
                for (let index = 0; index < commentTitleBlocks.length; index++) {
                    titles.push(commentTitleBlocks[index].children[0].innerText);
                }

                return titles;
            });

            // Extract comment rating
            const commentRating = await page.evaluate(async () => {
                const commentRatingBlocks = document.querySelectorAll('[data-test-target="review-rating"]');
                const ratings = [];

                for (let index = 0; index < commentRatingBlocks.length; index++) {
                    ratings.push(commentRatingBlocks[index].children[0].classList[1]);
                }

                return ratings;
            });

            // Extract date of stay
            const commentDateOfReview = await page.evaluate(async () => {

                // const commentDateOfStayBlocks = document.getElementsByClassName('teHYY')
                const commentDateBlocks = document.getElementsByClassName("cRVSd")

                // const datesOfStay = [];
                const datesOfReview = [];


                for (let index = 0; index < commentDateBlocks.length; index++) {

                    // Split the date of comment text block into an array
                    const reviewDate = commentDateBlocks[index].children[0].innerText.split('review').pop().split(' ')

                    let isYesterday = reviewDate[1] === "Yesterday";
                    // In case the review was posted yesterday, the array will only have the length of 2
                    let isCurrentMonth = reviewDate[2]?.length != 4;

                    let month = undefined;
                    let year = undefined;

                    if (isYesterday) {
                        // If the review was posted "Yesterday"
                        const currentTime = new Date()
                        month = currentTime.getMonth();
                        year = currentTime.getFullYear();
                    } else if (isCurrentMonth) {
                        // If the review date is in the current month (["","Oct,"1"])
                        const currentTime = new Date()
                        month = reviewDate[1];
                        year = currentTime.getFullYear();
                    } else {
                        // If the review date is > 1 month old (["","Oct,"2020"])
                        month = reviewDate[1];
                        year = reviewDate[2];
                    }

                    datesOfReview.push({
                        month,
                        year
                    });

                }

                return datesOfReview;
            });

            // Extract comments text
            const commentContent = await page.evaluate(async () => {

                const commentContentBlocks = document.getElementsByClassName('QewHA H4 _a');

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
                    month: monthStringToNumber(commentDateOfReview[index].month),
                    year: commentDateOfReview[index].year,
                    rating: commentRatingStringToNumber(commentRating[index]),
                };
            });

            // Push the formmated review to the  array
            allReviews.push(formatted);


            // Hand back the page so it's available again
            browser.handBack(page);
        }


        // Convert 2D array to 1D
        const reviewFlattened = allReviews.flat();

        // Data structure to be written to file
        const finalData = {
            airlineName,
            airlineId,
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
 * @param {String} airlineUrl - The url of the airline page 
 * @param {String} airlineName - The name of the airline
 * @param {String} [airlineId] - The id of the airline
 * @param {Number} [position] - The index of the airline page in the list
 * @param {puppeteer.Browser} browser - A browser instance
 * @returns {Promise<Object | Error>} - The final data
 */
const start = async (airlineUrl, airlineName, airlineId, position, browser) => {
    try {
        const { urls, count, } = await extractAllReviewPageUrls(airlineUrl, position, browser);

        const results = await scrape(count, urls, position, airlineName, airlineId, browser);

        return results;

    } catch (err) {
        throw err;
    }
};

export default start;