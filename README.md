# aio-gw
[EXPERIMENTAL]: All-in-one Network Gateway for Malware analysis. currently at Alpha stage.


HELP NEEDED: if you're keen to contribute to `aio-gw`, ping me! Lots to be done :)

## High level design
```mermaid
flowchart TD
subgraph containers
    subgraph TLS Decryption
    PolarProxy
    PolarProxy -.-> tun2socks
    tun2socks -.-> |raw packets|dump2poip
    dump2poip
    end
    Ingress --> |socks:1080| PolarProxy
    tun2socks --> |socks:1080| C(Egrees Glider)
    PolarProxy -.-> |pcap over ip|Arkime
    dump2poip -.-> |pcap over ip|Arkime
    Arkime -.-> |metadata| Elasticsearch
    C -->|SOCKS:9050| Tor
    C -->|DNS:9053| Tor
    Tor --> Internet
end

subgraph Monitoring
    computer ==> |8005/TCP|Arkime
    computer ==> |8081/TCP|PolarProxy
end

subgraph Sandbox
    Z(any PC/container\nrunning tun2socks or socks5 client) --> |SOCKS:1080|Ingress
end
```
## Requirements

A clean VM with docker and either `docker-compose` or `docker compose` installed. minimum 2GB of RAM is required. 

## Installation

- set your desired environment as per the `environment` file
- run `docker compose --env-file ./environment up -d`

within a few moments, you should have all containers up and running and port `1080/TCP` listening on incomming SOCKS5 requests to be terminated/captured.