FROM node:slim AS BASE

WORKDIR /puppeteer

RUN apt update && apt upgrade -y && apt --no-install-recommends install -y libnss3 libxss1 libasound2 libatk-bridge2.0-0 libgtk-3-0 libdrm-dev libgbm-dev


COPY package.json package-lock.json ./

## 2nd Stage: install dependencies
FROM BASE as DEPENDENCIES

RUN npm ci --production

## 3rd Stage: build the final image
FROM BASE as RELEASE

# Copy the dependencies from the DEPENDENCIES stage
COPY --from=DEPENDENCIES /puppeteer/node_modules ./node_modules

RUN chmod -R o+rwx node_modules/puppeteer/.local-chromium

# Copy rest of the files [from local to the image]
COPY . .

# Drop privilege
USER node

CMD ["npm","run","start:prod"]
