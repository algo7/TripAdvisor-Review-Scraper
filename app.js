// Dependencies
const puppeteer = require('puppeteer');
const { writeFileSync, readFileSync, promises: { access }, existsSync, mkdirSync } = require('fs');
const { parse } = require('json2csv');


// Data Directory
const dataDir = './data';

// Global vars for csv parser
const fields = ['title', 'content'];
const opts = { fields };

// Command line args
const myArgs = process.argv.slice(2);

// Check if the url is missing
if (!myArgs[0] && !process.env.URL) {
    console.log('Missing URL')
    process.exit(1);
}

// Check if the data directory exists, otherwise create it
if (!existsSync(dataDir)) {
    try {
        mkdirSync(dataDir);
    } catch (err) {
        console.error(err)
        process.exit(1);
    }
}

/**
 * Check if the given file exists
 * @param {String} filePath 
 * @returns {Promise<Boolean>}
 */
const fileExists = async (filePath) => {
    try {
        await access(filePath)
        return true
    } catch {
        return false
    }
}


/**
 * Scrape the page
 * @param {Array<String>} urlList 
 * @returns {Promise<Undefined | Error>}
 */
const scrap = async (urlList) => {
    try {

        if (!urlList || urlList.length === 0) {
            throw new Error('No url to scrape');
        }

        // Launch the browser
        const browser = await puppeteer.launch({
            headless: true,
            devtools: false,
            defaultViewport: {
                width: 1920,
                height: 1080,
            },
            args: [
                "--disable-gpu",
                "--disable-dev-shm-usage",
                "--disable-setuid-sandbox",
                "--no-sandbox",
            ]
        });

        // Open a new page
        const page = await browser.newPage();

        const cookiesAvailable = await fileExists('./data/cookies.json');

        if (!cookiesAvailable) {

            // Navigate to the page below
            await page.goto(myArgs[0] || process.env.URL);

            // Log the cookies
            const cookies = await page.cookies();
            const cookieJson = JSON.stringify(cookies);
            writeFileSync('./data/cookies.json', cookieJson);

            // Close the browser
            return await browser.close();
        }

        // Set Cookies
        const cookies = readFileSync('./data/cookies.json', 'utf8');
        const deserializedCookies = JSON.parse(cookies);
        await page.setCookie(...deserializedCookies);

        // Navigate to the page below
        await page.goto(myArgs[0] || process.env.URL);

        await page.waitForTimeout(1000);

        // Add 1st review page to the urlList
        urlList.unshift(myArgs[0] || process.env.URL);

        // Array to hold the review info
        const reviewInfo = []

        for (let index = 0; index < urlList.length; index++) {
            // Navigate to the page below
            await page.goto(urlList[index]);
            await page.waitForTimeout(3000);

            // Determin current URL
            const currentURL = page.url();

            console.log(`Scraping: ${currentURL} | ${urlList.length - 1 - index} Pages Left`);

            // In browser code
            // Extract comments title
            const commentTitle = await page.evaluate(async () => {

                // Extract a tags
                const commentTitleBlocks = document.getElementsByClassName('fCitC')

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

                const commentContentBlocks = document.getElementsByTagName('q')

                // Array use to store the comments
                const comments = []

                for (let index = 0; index < commentContentBlocks.length; index++) {
                    comments.push(commentContentBlocks[index].children[0].innerText)
                }

                return comments
            })

            // Format (for CSV processing) the reviews so each review of each page is in an object
            const formatted = commentContent.map((comment, index) => {
                return {
                    title: commentTitle[index],
                    content: comment
                }
            })

            // Push the formmated review to the  array
            reviewInfo.push(formatted);

        }

        // Close the browser
        await browser.close();

        // Convert 2D array to 1D
        return reviewInfo.flat();

    } catch (err) {
        throw err;
    }
};

/**
 * Extract review page url
 * @returns {Promise<Undefined | Error>}
 */
const extractAllReviewPageUrls = async () => {
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
                "--disable-gpu",
                "--disable-dev-shm-usage",
                "--disable-setuid-sandbox",
                "--no-sandbox",
            ]
        });

        // Open a new page
        const page = await browser.newPage();

        const cookiesAvailable = await fileExists('./data/cookies.json');

        if (!cookiesAvailable) {

            // Navigate to the page below
            await page.goto(myArgs[0] || process.env.URL);

            // Log the cookies
            const cookies = await page.cookies();
            const cookieJson = JSON.stringify(cookies);
            writeFileSync('./data/cookies.json', cookieJson);

            // Close the browser
            await browser.close();

            // Exit the process
            return await extractAllReviewPageUrls();
        }

        // Set Cookies
        const cookies = readFileSync('./data/cookies.json', 'utf8');
        const deserializedCookies = JSON.parse(cookies);
        await page.setCookie(...deserializedCookies);

        // Navigate to the page below
        await page.goto(myArgs[0] || process.env.URL);

        await page.waitForTimeout(5000);

        // Determin current URL
        const currentURL = page.url();

        console.log(`Gathering Info: ${currentURL}`);

        // In browser code
        const reviewPageUrls = await page.evaluate(() => {

            // All review count
            // let totalReviewCount = parseInt(document.querySelectorAll("a[href='#REVIEWS']")[1].innerText.split('\n')[1].split(' ')[0].replace(',', ''))

            // English review count
            let totalReviewCount = null

            // For restaurant reviews
            if (document.getElementsByClassName('ui_radio dQNlC')[1]) {
                totalReviewCount = parseInt(document.getElementsByClassName('ui_radio dQNlC')[1].innerText.split('(')[1].split(')')[0].replace(',', ''))
            }
            // For hotel reviews
            totalReviewCount = parseInt(document.getElementsByClassName("filterLabel")[18].innerText.split('(')[1].split(')')[0].replace(',', ''))




            // Calculate the last review page
            totalReviewCount = (totalReviewCount - totalReviewCount % 5) / 5

            // Get the url format
            const url = document.getElementsByClassName('pageNum')[1].href

            return { totalReviewCount, url }

        })

        // Destructure function outputs
        let { totalReviewCount, url } = reviewPageUrls;

        // Array to hold all the review urls
        const allUrls = []

        let counter = 0
        // Replace the url page count till the last page
        while (counter < totalReviewCount) {
            counter++
            url = url.replace(/-or[0-9]*/g, `-or${counter * 5}`)
            allUrls.push(url)
        }


        // JSON structure
        const data = {
            count: allUrls.length * 5,
            pageCount: allUrls.length,
            urls: allUrls
        };

        // Write the data to a json file
        writeFileSync('./data/reviewUrl.json', JSON.stringify(data));

        // Close the browser
        await browser.close();

        return allUrls

    } catch (err) {
        throw err
    }
}


const start = async () => {
    try {
        // Extract review page urls
        const allReviewsUrl = await extractAllReviewPageUrls();

        // Scrape the review page
        const results = await scrap(allReviewsUrl);

        // Convert JSON to CSV
        const csv = parse(results, opts);

        // Write the CSV to a file
        writeFileSync('./data/review.csv', csv);

        // Exit the process
        console.log('Done');

        process.exit(0);

    } catch (err) {
        throw err;
    }
}
start().catch(err => console.error(err));
