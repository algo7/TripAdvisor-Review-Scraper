# TripAdvisor-Review-Scraper
A simple scraper for TripAdvisor reviews.

## Table of Contents

- [TripAdvisor-Review-Scraper](#tripadvisor-review-scraper)
  - [Table of Contents](#table-of-contents)
  - [How to Install Docker:](#how-to-install-docker)
  - [Run Using Docker Compose](#run-using-docker-compose)
  - [Run Using Docker CLI](#run-using-docker-cli)
  - [If you are lazy](#if-you-are-lazy)
  - [If you are really lazy](#if-you-are-really-lazy)
  - [Notes:](#notes)
  - [Known Issues](#known-issues)
- [Container Provisioner](#container-provisioner)
  - [Pull the latest scraper Docker image](#pull-the-latest-scraper-docker-image)
  - [Run the container provisioner](#run-the-container-provisioner)
  - [Visit the UI](#visit-the-ui)
  - [Live Demo](#live-demo)

## How to Install Docker:
1. [Windows](https://docs.docker.com/desktop/windows/install/)
2. [Mac](https://docs.docker.com/desktop/mac/install/)
3. [Linux](https://docs.docker.com/engine/install/ubuntu/)

## Run Using Docker Compose
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
7. Edit the `LANGUAGE` variable in the `docker-compose-prod.yml` file to the language of the reviews you want to scrape.
   - This option is only supported RESTO mode.
   - Available options are `fr` and `en` which will actaully scrape all the reviews.
8. Run `docker-compose -f docker-compose-prod.yml up` to start the container.
9. Once the scraping process is finished, check the `reviews` folder for the results.
10. Samples of the results are included in the `samples` folder.
11. Please remember to empty the `reviews` folder before running the scraper again.

## Run Using Docker CLI 
1. Download the repository.
2. Replace the `-e SCRAP_MODE`, `-e CONCURRENCY`, `-e LANGUAGE` with custom values.
3. Run `docker run --mount type=bind,src="$(pwd)"/reviews,target=/puppeteer/reviews --mount type=bind,src="$(pwd)"/source,target=/puppeteer/source -e SCRAPE_MODE=HOTEL -e CONCURRENCY=5 -e LANGUAGE=en ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest` in the terminal at the root directory of the project.

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
- The Docker image size is close to 1GB. If it's your first time running the scraper, the setup program might take some time depending on your internet connection.
- Below are the outputs indicating that the image is being downloaded.
```bash
Pulling scraper (ghcr.io/algo7/tripadvisor-review-scraper/scrap:latest)...
latest: Pulling from algo7/tripadvisor-review-scraper/scrap
```
- Some operating systems, Windows especailly, or Anti-Virus software might block the download. In that case, you can safely ignore the warnings. This project is 100% open source. The [release](https://github.com/algo7/TripAdvisor-Review-Scraper/releases) page also contains links to [Virus Total](https://www.virustotal.com/gui/home/upload) scan results of each binary.

## Known Issues
1. The hotel scraper works for English reviews only.
2. The restaurant scraper can only scrap english reivews or french reviews.

# Container Provisioner
Container Provisioner is a tool written in [Go](https://go.dev/) that provides a UI for the users to interact with the scraper. It uses [Docker API](https://docs.docker.com/engine/api/) to provision the containers and run the scraper. The UI is written in raw HTML and JavaScript while the backend web framwork is [Fiber](https://docs.gofiber.io/).

The scraped reviews will be uploaded to [Cloudflare R2 Buckets](https://www.cloudflare.com/lp/pg-r2/) for storing. R2 is S3-Compatible; therefore, technically, one can also use AWS S3 for storing the scraped reviews.

## Pull the latest scraper Docker image
```bash
docker pull ghcr.io/algo7/tripadvisor-review-scraper/scrape:latest
```
## Run the container provisioner
The `docker-compose.yml` for the provisioner is located in the `container_provisioner` folder.

You will need to create a folder called `credentials` in the root directory of the project. The `credentials` folder will contain the credentials for the R2 bucket. The credentials file should be named `creds.json` and should be in the following format:
```json
{
    "bucketName": "<R2_Bucket_Name>",
    "accountId": "<Cloudflare_Account_Id>",
    "accessKeyId": "<R2_Bucket_AccessKey_ID>",
    "accessKeySecret": "<R2_Bucket_AccessKey_Secret>"
}
```

## Visit the UI
The UI is accessible at `http://localhost:3000`.

## Live Demo
A live demo of the container provisioner is available at [https://algo7.tools](https://algo7.tools).