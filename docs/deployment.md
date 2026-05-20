# Deployment

## Local
- start Rust kernel
- start `vnodex-node`
- ensure the kernel URL is reachable

## Production
- keep node data outside the repo
- disable debug routes
- enable metrics
- persist identity keys
- run behind a reverse proxy if exposing HTTP publicly
