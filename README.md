# LETS LIVE

## ABOUT
This is a project about livestreaming techonologies and solutions related to livestreaming.
The project aims to create a functioning livestreaming website from a to z like Twitch.

## PORTS
8000: The main api
1935: The RTMP default port (ingest)
8889: The web server port (use to get the index.m3u8 and stream.m3u8 files)
5000: Web UI
8888: The port to get .ts files (This port uses nginx as a reverse proxy to get file from the IPFS network)
4001: Our bootstrap node port (allows other nodes outside the network to connect in)

## TECHNOLOGIES AND TOOLS
- Golang
- FFMpeg
- CephFS
- IPFS
- Prometheus and Grafana
- Docker and K8s
