# Proxy Pool
Proxy Pool is a docker image that runs both HTTP and SOCKS5 Proxies over OpenVPN (config to be provided by the user via docker bind mounts). `sockd`, `squid`, and `openvpn` client are managed by `supervisord` in the container. The service integrates with the Container Provisioner to provide a pool of proxies for the scraper to use. The container provisioner uses `docker-compose labels` to distinguish between different proxies. At this moment, the container provisioner only supports connecting to the Proxy Pool using HTTP proxies. Each service in the `docker-compose.yml` file represents a single proxy in the pool. The `docker-compose.yml` file for the proxy pool is located in the `proxy_pool` folder.

The Proxy Pool service can also be used directly with the scraper. Just make sure that the `PROXY_ADDRESS` environment variable is in the `docker-compose.yml` file for the scraper.

## Running the Proxy Pool
1. Pull the latest scraper Docker image
```bash
docker pull ghcr.io/algo7/tripadvisor-review-scraper/vpn_worker:latest
```
2. Create a docker-compose.yml file containing the configurations for each proxy (see the docker-compose.yml provided in the proxy_pool folder).
3. Place the OpenVPN config file of each proxy in the corresponding bind mount folder speicified in the docker-compose.yml file.
4. Run `docker-compose up` to start the container.