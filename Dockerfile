# base image
FROM node:slim AS base
LABEL "org.opencontainers.image.source" = "https://github.com/algo7/TripAdvisor-Review-Scraper"

# Set the working directory
WORKDIR /puppeteer

# Install libs for Puppeteer
RUN apt update && apt upgrade -y && apt --no-install-recommends install -y libnss3 libxss1 libasound2 libatk-bridge2.0-0 libgtk-3-0 libdrm-dev libgbm-dev && mkdir reviews source

# Copy the pkg json files
COPY package.json package-lock.json ./

## 2nd Stage: install dependencies
FROM base as dependencies

RUN npm ci --omit=dev

## 3rd Stage: build the final image
FROM base as release

# Copy the dependencies from the dependencies stage
COPY --from=dependencies /puppeteer/node_modules ./node_modules

# Copy rest of the files [from local to the image]
COPY . .

# Drop privilege
USER node

ENV FORCE_COLOR=1 

CMD ["npm","run","start:prod"]
