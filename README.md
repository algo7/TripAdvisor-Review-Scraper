# TripAdvisor-Review-Scraper
Scrape TripAdvisor Reviews

## Docker
1. Create a folder called `reviews` in the root directory of the project.
2. Edit the `URL` variable in the `docker-compose-prod.yml` file to point to the page that you want to scrape.
3. Run `docker-compose -f docker-compose-prod.yml up` to start the container.