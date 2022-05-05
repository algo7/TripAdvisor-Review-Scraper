# TripAdvisor-Review-Scraper
## Scrape TripAdvisor Reviews

## How to Install Docker:
1. [Windows](https://docs.docker.com/desktop/windows/install/)
2. [Mac](https://docs.docker.com/desktop/mac/install/)
3. [Linux](https://docs.docker.com/engine/install/ubuntu/)

## Docker
1. Download the repository.
2. Create a folder called `reviews` and a folder called `source` in the root directory of the project.
3. The `reviews` folder will contain the scraped reviews.
4. Place the source file in the `source` folder.
   - The source file is a CSV file containing a list of hotels/restaurants to scrape.
   - Examples of the source file are provided in the `examples` folder.
   - The source file for hotels should be named `hotels.csv` and the source file for restaurants should be named `restos.csv`.
5. Edit the `SCRAPE_MODE` (RESTO for restaurants, HOTEL for hotel) variable in the `docker-compose-prod.yml` file to scrape either restaurant or hotel reviews.
6. Edit the `CONCURRENCY` variable in the `docker-compose-prod.yml` file to set the number of concurrent requests.
   - A high concurrency number might cause the program to hang depending on the internet connection and the resource availability of your computer.
7. Run `docker-compose -f docker-compose-prod.yml up` to start the container.
8. Once the scraping process is finished, check the `reviews` folder for the results.
9. Samples of the results are included in the `samples` folder.
10. Please remember to empty the `reviews` folder before running the scraper again.

## Docker CLI 
1. Download the repository.
2. Replace the `-e SCRAP_MODE` and `-e CONCURRENCY` with custom values.
3. Run `docker run --mount type=bind,src="$(pwd)"/reviews,target=/puppeteer/reviews --mount type=bind,src="$(pwd)"/source,target=/puppeteer/source -e SCRAPE_MODE=HOTEL -e CONCURRENCY=5 ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest` in the terminal at the root directory of the project.

## If you are lazy
1. Download the repository.
2. Go to the `builds` folder and you will find 3 files:
   - setup-windows-amd64.exe => For Windows Users
   - setup-darwin-amd64.bin => For Mac Users
   - setup-linux-amd64.bin => For Linux Users
3. Run the executable corresponding to your OS to automate the setup.
4. You still have to install Docker though :)

## If you are really lazy
1. Go to the [release](https://github.com/algo7/TripAdvisor-Review-Scraper/releases) page and donwload the latest setup binaries for your operating system.
   - setup-windows-amd64.exe => For Windows Users
   - setup-darwin-amd64.bin => For Mac Users
   - setup-linux-amd64.bin => For Linux Users
2. Run it.
3. But there is still a catch: you still have to install Docker first :)

## Notes:
- The Docker image is 1.77GB. If it's your first time running the scraper, the setup program might take some time depending on your internet connection.
- Below are the outputs indicating that the image is being downloaded.
```bash
Pulling scraper (ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest)...
latest: Pulling from algo7/tripadvisor-review-scraper/scrap
```

## Known Issues
1. The hotel scraper works for English reviews only.
2. The restaurant scraper will scrape all the reviews (you can't choose the language).
