// Dependencies
const puppeteer = require('puppeteer');
const { writeFileSync, readFileSync, } = require('fs');

// Command line args
const myArgs = process.argv.slice(2);

if (!myArgs[3]) {
    console.log('Missing URL')
    process.exit(1);
}


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

        // Set Cookies
        const cookies = readFileSync('cookies.json', 'utf8');
        const deserializedCookies = JSON.parse(cookies);
        await page.setCookie(...deserializedCookies);

        // Go to the home page
        // Navigate to the page below
        await page.goto(myArgs[3], {
            waitUntil: 'networkidle0',
        });

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
            await page.waitForTimeout(2000);
        }


        // In browser code
        const commentTitle = await page.evaluate(async () => {

            // Extract a tags
            const commentBlocks = document.getElementsByClassName('fCitC')

            // Array to store the urls on the page
            const titles = [];

            // Higher order functions don't work in the browser
            for (let index = 0; index < commentBlocks.length; index++) {
                urlList.push(commentBlocks[index].children[0].innerHTML);
            }

            return titles;
        });




        // Write the data to a json file
        writeFileSync('x.csv', JSON.stringify(data));

        // Close the browser
        await browser.close();

    } catch (err) {
        throw err;
    }
};