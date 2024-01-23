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
  - [Requirements](#requirements)
    - [How to Install Docker:](#how-to-install-docker)
  - [Project Layout](#project-layout)
    - [Scraper](#scraper)
    - [Container Provisioner](#container-provisioner)
    - [Proxy Pool](#proxy-pool)

## Requirements
1. Go +v1.21
2. Make [Optional]
3. Docker [Optional]
4. Docker Compose [Optional]
5. Node.js +18 [Optional. Only required if you want to use the scraper written in Node.js, which is deprecated.]

### How to Install Docker:
1. [Windows](https://docs.docker.com/desktop/windows/install/)
2. [Mac](https://docs.docker.com/desktop/mac/install/)
3. [Linux](https://docs.docker.com/engine/install/ubuntu/)

## Project Layout
### Scraper 
There are 2 scrapers available:
1. [Scraper](https://github.com/algo7/TripAdvisor-Review-Scraper/tree/main/scraper) written in Go
2. [Scraper](https://github.com/algo7/TripAdvisor-Review-Scraper/tree/main/scraperjs) written in Node.js [Deprecated]

The scraper written in Go is preferred because it calls the API directly and is much faster than the scraper written in Node.js which goes the traditional way of parsing HTML. The instructions of how to use them are located in their separate folders.


### Container Provisioner
Automates the process of provisioning containers for the scraper.

Please read more about the container provisioner [here](https://github.com/algo7/TripAdvisor-Review-Scraper/tree/main/container_provisioner)

### Proxy Pool
Provides a pool of proxies for the scraper to use.

Please read more about the proxy pool [here](https://github.com/algo7/TripAdvisor-Review-Scraper/tree/main/proxy_pool)