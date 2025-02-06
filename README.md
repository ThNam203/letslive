# LETS LIVE

## ABOUT
This is a project about livestreaming techonologies and solutions related to livestreaming.
The project aims to create a functioning livestreaming website from a to z like Twitch.

## PORTS
- 8000: The main API endpoints.
- 1935: The RTMP default port (for stream ingestion)  
- 8889: The web server port (use to get the index.m3u8 and stream.m3u8 files)
- 8888: The port to get .ts files (This port uses nginx as a reverse proxy to get file from the IPFS network)  
- 5000: Web UI  
- 4001: IPFS bootstrap node port (allows other nodes outside the network to connect in)
- 5601: Kibana port for logging
- 8002: Kong API gateway management.
- 3000: Grafana for IPFS.

## HOW IT WORKS
- Ingestion: The RTMP is used to get the content of the livestream (through OBS, will work for a built-in in browser)  
- Transcode: From the RTMP, use FFMpeg to generate the HLS files which also has adaptive bitrate streaming (ABS)  
- When files are generated, we have few ideas:  
  * Directly serving files, easy but will put a burden on transcode server and latency will not be fast.
  * Push files into a remote storage (**AWS preferred - not yet**): (I'm using IPFS currently but only for "using IPFS" demo purpose), then rewrite the index.m3u8 file (video players use this files
know where to retrive files to play) pointing to the remote location.
- The storage is also a problem, but we will look into it at another time (CephFS) for distributed file storage?

## TECHNOLOGIES AND TOOLS
- Golang, PostgresQL.
- FFMpeg for transcoding from RTMP to HLS.
- Consul with Kong API Gateway for service discovery and api gateway.
- Storage: IPFS.
- Docker Compose

## ARCHITECTURE
![Architecture](docs/architecture.png)
