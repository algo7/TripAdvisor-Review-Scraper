# TripAdvisor-Review-Scraper
A simple scraper for TripAdvisor reviews.

## Table of Contents

- [TripAdvisor-Review-Scraper](#tripadvisor-review-scraper)
  - [Table of Contents](#table-of-contents)
  - [How to Install Docker:](#how-to-install-docker)
  - [Run Using Docker Compose](#run-using-docker-compose)
  - [Run Using Docker CLI](#run-using-docker-cli)
  - [Known Issues](#known-issues)
- [Container Provisioner](#container-provisioner)
  - [Pull the latest scraper Docker image](#pull-the-latest-scraper-docker-image)
  - [Credentials Configuration](#credentials-configuration)
    - [R2 Bucket Credentials](#r2-bucket-credentials)
    - [R2 Bucket URL](#r2-bucket-url)
  - [Run the container provisioner](#run-the-container-provisioner)
  - [Visit the UI](#visit-the-ui)
  - [Live Demo](#live-demo)

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
Container Provisioner is a tool written in [Go](https://go.dev/) that provides a UI for the users to interact with the scraper. It uses [Docker API](https://docs.docker.com/engine/api/) to provision the containers and run the scraper. The UI is written in raw HTML and JavaScript while the backend web framwork is [Fiber](https://docs.gofiber.io/).

The scraped reviews will be uploaded to [Cloudflare R2 Buckets](https://www.cloudflare.com/lp/pg-r2/) for storing. R2 is S3-Compatible; therefore, technically, one can also use AWS S3 for storing the scraped reviews.

## Pull the latest scraper Docker image
```bash
docker pull ghcr.io/algo7/tripadvisor-review-scraper/scraper:latest
```
## Credentials Configuration
### R2 Bucket Credentials
You will need to create a folder called `credentials` in the `container_provisioner` directory of the project. The `credentials` folder will contain the credentials for the R2 bucket. The credentials file should be named `creds.json` and should be in the following format:
```json
{
    "bucketName": "<R2_Bucket_Name>",
    "accountId": "<Cloudflare_Account_Id>",
    "accessKeyId": "<R2_Bucket_AccessKey_ID>",
    "accessKeySecret": "<R2_Bucket_AccessKey_Secret>"
}
```
### R2 Bucket URL
You will also have to set the `R2_URL` environment variable in the `docker-compose.yml` file to the URL of the R2 bucket. The URL should end with a `/`.

## Run the container provisioner
The `docker-compose.yml` for the provisioner is located in the `container_provisioner` folder.

## Visit the UI
The UI is accessible at `http://localhost:3000`.

## Live Demo
A live demo of the container provisioner is available at [https://algo7.tools](https://algo7.tools).