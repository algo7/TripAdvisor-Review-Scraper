# Base image
FROM node:slim AS BASE
LABEL "org.opencontainers.image.source" = "https://github.com/algo7/TripAdvisor-Review-Scraper"

# Set the working directory
WORKDIR /puppeteer

# Install libs for Puppeteer
RUN apt update && apt upgrade -y && apt --no-install-recommends install -y libnss3 libxss1 libasound2 libatk-bridge2.0-0 libgtk-3-0 libdrm-dev libgbm-dev && npm install -g npm@8.8.0

# Copy the pkg json files
COPY package.json package-lock.json ./

## 2nd Stage: install dependencies
FROM BASE as DEPENDENCIES

RUN npm ci --omit=dev

## 3rd Stage: build the final image
FROM BASE as RELEASE

# Copy the dependencies from the DEPENDENCIES stage
COPY --from=DEPENDENCIES /puppeteer/node_modules ./node_modules

# Change permission for the chromium binary
RUN chmod -R o+rwx node_modules/puppeteer/.local-chromium

# Copy rest of the files [from local to the image]
COPY . .

# Drop privilege
USER node

CMD ["npm","run","start:prod"]
