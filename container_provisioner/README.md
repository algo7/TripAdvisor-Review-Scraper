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
