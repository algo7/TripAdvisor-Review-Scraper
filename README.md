# TripAdvisor-Review-Scraper
Scrape TripAdvisor Reviews

## Docker
1. Create a folder called `reviews` and a folder called `source` in the root directory of the project.
2. The `reviews` folder will contain the scraped reviews.
3. Place the source file in the `source` folder.
   1. The source file is a CSV file containing a list of hotels/restaurants to scrape.
   2. Examples of the source file are provided in the `examples` folder.
   3. The source file for hotels should be named `hotels.csv` and the source file for restaurants should be named `restos.csv`.
4. Edit the `SCRAPE_MODE` (RESTO for restaurants, HOTEL for hotel) variable in the `docker-compose-prod.yml` file to scrape either restaurant or hotel reviews.
5. Run `docker-compose -f docker-compose-prod.yml up` to start the container.
6. Once the scraping process is finished, check the `reviews` folder for the results.
7. Please remember to empty the `reviews` folder before running the scraper again.

## Known Issues
1. The hotel scraper works for English reviews only.
2. The restaurant scraper will scrape all the reviews (you can't choose the language).