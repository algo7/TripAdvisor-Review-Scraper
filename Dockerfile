FROM node:slim
WORKDIR /puppeteer
RUN apt update && apt upgrade -y && apt-get install -y libnss3 libxss1 libasound2 libatk-bridge2.0-0 libgtk-3-0 libdrm-dev libgbm-dev
COPY *.json app.js ./
RUN npm ci
RUN chmod -R o+rwx node_modules/puppeteer/.local-chromium
USER node
CMD ["npm","run","start:prod"]