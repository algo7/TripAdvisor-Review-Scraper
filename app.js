// Dependencies
const puppeteer = require('puppeteer');
const { writeFileSync, readFileSync, promises: { access } } = require('fs');

// Command line args
const myArgs = process.argv.slice(2);

if (!myArgs[0]) {
    console.log('Missing URL')
    process.exit(1);
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

// Scrape the page
const scrap = async () => {
    try {

        // Launch the browser
        const browser = await puppeteer.launch({
            headless: false,
            devtools: false,
            defaultViewport: {
                width: 1920,
                height: 1080,
            },
        });

        // Open a new page
        const page = await browser.newPage();

        const cookiesAvailable = await fileExists('./cookies.json');

        if (!cookiesAvailable) {


            // Navigate to the page below
            await page.goto(myArgs[0]);

            // Navigate to the page below
            await page.goto(myArgs[0], {
                waitUntil: 'networkidle0',
            });

            // Log the cookies
            const cookies = await page.cookies();
            const cookieJson = JSON.stringify(cookies);
            writeFileSync('cookies.json', cookieJson);

            // Close the browser
            return await browser.close();
        }

        // Set Cookies
        const cookies = readFileSync('cookies.json', 'utf8');
        const deserializedCookies = JSON.parse(cookies);
        await page.setCookie(...deserializedCookies);

        // Navigate to the page below
        await page.goto(myArgs[0]);

        await page.waitForTimeout(1000);

        // Determin current URL
        const currentURL = page.url();

        console.log(`Scraping: ${currentURL}`);



        // In browser code

        // Determine if the page is scrolled to the bottom
        let scrollToBottom = await page.evaluate(() => window.innerHeight + window.scrollY >= document.body.offsetHeight);

        // Scroll to the bottom
        while (!scrollToBottom) {

            scrollToBottom = await page.evaluate(() => window.innerHeight + window.scrollY >= document.body.offsetHeight);
            await page.mouse.wheel({ deltaY: 3000, });
        }


        // In browser code

        // Extract comments title
        const commentTitle = await page.evaluate(async () => {

            // Extract a tags
            const commentTitleBlocks = document.getElementsByClassName('fCitC')

            // Array to store the urls on the page
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

            const comments = []

            for (let index = 0; index < commentContentBlocks.length; index++) {
                comments.push(commentContentBlocks[index].children[0].innerText)

            }

            return comments
        })


        // Write the data to a json file
        // writeFileSync('x.csv', JSON.stringify(data));

        // Close the browser
        // await browser.close();

    } catch (err) {
        throw err;
    }
};

scrap().catch(err => console.error(err));