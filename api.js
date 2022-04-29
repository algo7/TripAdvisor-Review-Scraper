
const puppeteer = require('puppeteer');

const scrap = async (restoUrl) => {

    // Launch the browser
    const browser = await puppeteer.launch({
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




    const reviews = await page.evaluate(() => {
        // document.getElementById('filters_detail_language_filterLang_ALL').click()
        const results = [];
        const reviewCount = [];
        document.getElementsByClassName('choices')[0].querySelectorAll('.row_num').forEach(el => reviewCount.push(el.innerText));
        return reviewCount;
        // const items = document.body.querySelectorAll('.review-container');
        // items.forEach(item => {

        //     /* Get and format Rating */
        //     let ratingElement = item.querySelector('.ui_bubble_rating').getAttribute('class');
        //     let integer = ratingElement.replace(/[^0-9]/g, '');
        //     let parsedRating = parseInt(integer) / 10;

        //     /* Get and format date of Visit */
        //     let dateOfVisitElement = item.querySelector('.prw_rup.prw_reviews_stay_date_hsx').innerText;
        //     let parsedDateOfVisit = dateOfVisitElement.replace('Date of visit:', '').trim();

        //     /* Part 4 */

        //     results.push({
        //         rating: parsedRating,
        //         dateOfVisit: parsedDateOfVisit,
        //         ratingDate: item.querySelector('.ratingDate').getAttribute('title'),
        //         title: item.querySelector('.noQuotes').innerText,
        //         content: item.querySelector('.partial_entry').innerText,

        //     });

        // });
        // return results;
    });
    console.log(reviews);
};

scrap('https://www.tripadvisor.com/Restaurant_Review-g652156-d7285961-Reviews-L_Indus_Bar-Bulle_La_Gruyere_Canton_of_Fribourg.html');