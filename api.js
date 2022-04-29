
// Dependencies
const puppeteer = require('puppeteer');
const { writeFileSync, } = require('fs');
const path = require('path');

// Global variables
const dataPath = path.join(__dirname, './data/');

const scrap = async (restoUrl) => {
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

        const page = await browser.newPage();
        await page.goto(restoUrl);
        await page.waitForSelector('body');

        // Select all language
        await page.click('[id=filters_detail_language_filterLang_ALL]');

        // Wait till all the review shows
        await page.waitForTimeout(1000);

        // Expand the reviews
        await page.click('.taLnk.ulBlueLinks');

        // Wait for the reviews to load
        await page.waitForFunction('document.querySelector("body").innerText.includes("Show less")');

        // Extract the review page url
        const reviewPageUrls = await page.evaluate(() => {

            const totalReviewCount = parseInt(document.getElementsByClassName('reviews_header_count')[0].innerText.split('(')[1].split(')')[0].replace(',', ''));

            let noReviewPages = totalReviewCount / 15;
            // Calculate the last review page
            if (totalReviewCount % 15 !== 0) {
                noReviewPages = ((totalReviewCount - totalReviewCount % 15) / 15) + 1;
            }

            // Get the url of the 2nd page of review. The 1st page is the input link
            const url = document.getElementsByClassName('pageNum')[1].href;

            return {
                noReviewPages,
                url,
                totalReviewCount,
            };


        });

        // Destructure function outputs
        let { noReviewPages, url, totalReviewCount, } = reviewPageUrls;

        // Array to hold all the review urls
        const allUrls = [];

        let counter = 0;
        // Replace the url page count till the last page
        while (counter < noReviewPages - 1) {
            counter++;
            url = url.replace(/-or[0-9]*/g, `-or${counter * 15}`);
            allUrls.push(url);
        }

        // Add the first page url
        allUrls.unshift(restoUrl);

        // Information for loggin
        const data = {
            count: totalReviewCount,
            pageCount: allUrls.length,
            urls: allUrls,
        };
        console.log(data);

        // Array to hold all the reviews 
        const allReviews = [];
        for (let index = 0; index < allUrls.length; index++) {

            await page.goto(allUrls[index]);
            await page.waitForSelector('body');
            // Select all language
            await page.click('[id=filters_detail_language_filterLang_ALL]');

            // Wait till all the review shows
            await page.waitForTimeout(1000);

            // Expand the reviews
            await page.click('.taLnk.ulBlueLinks');

            // Wait for the reviews to load
            await page.waitForFunction('document.querySelector("body").innerText.includes("Show less")');
            // Determin current URL
            const currentURL = page.url();
            console.log(`Scraping: ${currentURL} | ${allUrls.length - 1 - index} Pages Left`);

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

                    /* Part 4 */

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
            count: totalReviewCount,
            allReviews,
        };

        // Write to file
        writeFileSync(`./${dataPath}${restoUrl.split('-')[4]}.json`, JSON.stringify(finalData, null, 2));

        return 'Done';
    } catch (err) {
        throw err;
    }
};

scrap('https://www.tripadvisor.com/Restaurant_Review-g652156-d17621567-Reviews-Kalasin-Bulle_La_Gruyere_Canton_of_Fribourg.html').then(x => console.log(x)).catch(err => console.log(err));


  // The review count array based in the "Traveler rating info"
//   const reviewCount = [];

        // // Extract the review count for each rating
        // document.getElementsByClassName('choices')[0].querySelectorAll('.row_num').forEach(el => reviewCount.push(el.innerText));

        // // Sum them to get the total review count
        // const totalReviewCount = reviewCount.map(count => parseInt(count)).reduce((a, b) => a + b);