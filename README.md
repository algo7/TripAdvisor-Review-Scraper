# TripAdvisor-Review-Scraper
## Scrape TripAdvisor Reviews

## How to Install Docker:
1. [Windows](https://docs.docker.com/desktop/windows/install/)
2. [Mac](https://docs.docker.com/desktop/mac/install/)
3. [Linux](https://docs.docker.com/engine/install/ubuntu/)

## Docker
1. Create a folder called `reviews` and a folder called `source` in the root directory of the project.
2. The `reviews` folder will contain the scraped reviews.
3. Place the source file in the `source` folder.
   - The source file is a CSV file containing a list of hotels/restaurants to scrape.
   - Examples of the source file are provided in the `examples` folder.
   - The source file for hotels should be named `hotels.csv` and the source file for restaurants should be named `restos.csv`.
4. Edit the `SCRAPE_MODE` (RESTO for restaurants, HOTEL for hotel) variable in the `docker-compose-prod.yml` file to scrape either restaurant or hotel reviews.
5. Edit the `CONCURRENCY` variable in the `docker-compose-prod.yml` file to set the number of concurrent requests.
   - A high concurrency number might cause the program to hang depending on the internet connection and the resource availability of your computer.
6. Run `docker-compose -f docker-compose-prod.yml up` to start the container.
7. Once the scraping process is finished, check the `reviews` folder for the results.
8. Samples of the results are included in the `samples` folder.
9. Please remember to empty the `reviews` folder before running the scraper again.

## Docker CLI 
1. Replace the `-e SCRAP_MODE` and `-e CONCURRENCY` with custom values.
2. `docker run --mount type=bind,src="$(pwd)"/reviews,target=/puppeteer/reviews --mount type=bind,src="$(pwd)"/source,target=/puppeteer/source -e SCRAPE_MODE=HOTEL -e CONCURRENCY=5 ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest`

## Known Issues
1. The hotel scraper works for English reviews only.
2. The restaurant scraper will scrape all the reviews (you can't choose the language).

## If you are lazy
1. Go to the `builds` folder and you will find 3 files:
   - setup-windows-amd64.exe => For Windows Users
   - setup-darwin-amd64.bin => For Mac Users
   - setup-linux-amd64.bin => For Linux Users
2. Run the executable corresponding to your OS to automate the setup.
3. You still have to install Docker though :)