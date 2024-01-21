# TripAdvisor-Review-Scraper
A simple scraper for TripAdvisor (Hotel, Restaurant, Airline) reviews.

[![Build & Push [Container Provisioner]](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/ci_container_provisioner.yml/badge.svg?branch=main)](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/ci_container_provisioner.yml)

[![Build & Push [Scraper]](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/ci_scraper.yml/badge.svg?branch=main)](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/ci_scraper.yml)

[![Build & Push [VPN Worker]](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/ci_vpn_worker.yml/badge.svg)](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/ci_vpn_worker.yml)

[![CodeQL](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/codeql.yml/badge.svg?branch=main)](https://github.com/algo7/TripAdvisor-Review-Scraper/actions/workflows/codeql.yml)

# [Current Issues](https://github.com/algo7/TripAdvisor-Review-Scraper/issues)

## Table of Contents

- [TripAdvisor-Review-Scraper](#tripadvisor-review-scraper)
- [Current Issues](#current-issues)
  - [Table of Contents](#table-of-contents)
  - [How to Install Docker:](#how-to-install-docker)
  - [Run Using Docker Compose](#run-using-docker-compose)
  - [Run Using Docker CLI](#run-using-docker-cli)
  - [Known Issues](#known-issues)
- [Container Provisioner](#container-provisioner)
- [Proxy Pool](#proxy-pool)

## How to Install Docker:
1. [Windows](https://docs.docker.com/desktop/windows/install/)
2. [Mac](https://docs.docker.com/desktop/mac/install/)
3. [Linux](https://docs.docker.com/engine/install/ubuntu/)

## Run Using Docker Compose
1. Download the repository.
2. Create a folder called `reviews` and a folder called `source` in the `scraper` directory of the project.
3. The `reviews` folder will contain the scraped reviews.
4. Place the source file in the `source` folder.
   - The source file is a CSV file containing a list of hotels/restaurants to scrape.
   - Examples of the source file are provided in the `examples` folder.
   - The source file for hotels should be named `hotels.csv` and the source file for restaurants should be named `restos.csv`.
5. Edit the `SCRAPE_MODE` (RESTO for restaurants, HOTEL for hotel) variable in the `docker-compose.yml` file to scrape either restaurant or hotel reviews.
6. Edit the `CONCURRENCY` variable in the `docker-compose.yml` file to set the number of concurrent requests.
   - A high concurrency number might cause the program to hang depending on the internet connection and the resource availability of your computer.
7. Edit the `LANGUAGE` variable in the `docker-compose.yml` file to the language of the reviews you want to scrape.
   - This option is only supported RESTO mode.
   - Available options are `fr` and `en` which will actaully scrape all the reviews.
8. Run `docker-compose up` to start the container.
9. Once the scraping process is finished, check the `reviews` folder for the results.
10. Samples of the results are included in the `samples` folder.
11. Please remember to empty the `reviews` folder before running the scraper again.

## Run Using Docker CLI 
1. Download the repository.
2. Replace the `-e SCRAP_MODE`, `-e CONCURRENCY`, `-e LANGUAGE` with custom values.
3. Run `docker run --mount type=bind,src="$(pwd)"/reviews,target=/puppeteer/reviews --mount type=bind,src="$(pwd)"/source,target=/puppeteer/source -e SCRAPE_MODE=HOTEL -e CONCURRENCY=5 -e LANGUAGE=en ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest` in the terminal at the root directory of the project.


## Known Issues
1. The hotel scraper works for English reviews only.
2. The restaurant scraper can only scrap english reivews or french reviews.
3. The hotel scraper uses date of review instead of date of stay as the date because the date of stay is not always available.

# Container Provisioner
# Proxy Pool