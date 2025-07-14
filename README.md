# rotox

[![Go Documentation](https://pkg.go.dev/badge/github.com/isacskoglund/rotox.svg)](https://pkg.go.dev/github.com/isacskoglund/rotox)
![Test](https://github.com/isacskoglund/rotox/actions/workflows/ci.yml/badge.svg)

**rotox** is a distributed rotating proxy designed to make a large number of public IP addresses available through a single server. The proxy is particularly well suited for running on serverless, scale-to-zero platforms such as Google Cloud Run, allowing users to leverage the large number of public IP addresses available on such platforms, often at negligible cost.

## Architecture

**rotox** consists of two main parts, a **hub** and one or more **probes**. These are separate processes and typically run on different machines. When a client communicates with a target on the internet, the traffic is routed first through the hub, then through one of the many probes and finally to the target. The hub distributes the connections across the pool of probes, allowing the client to make repeated connections to the same target using many different public IPs.

![Architecture](docs/architecture.svg)

## Serverless Compatibility

The hub and the probes communicate over gRPC, which is based on the HTTP/2 protocol. This makes the probes compatible with serverless, scale-to-zero environments such as Google Cloud Run. On such platforms, one can deploy a large number of probes — each potentially offering a unique public IP — without incurring costs during idle time. More precisely, each probe is only active and billed while it is relaying traffic, as opposed to a virtual machine which remains active and incur costs regardless of usage.

However, most common proxy protocols (e.g., HTTP CONNECT, SOCKS5) are not supported in serverless environments like Google Cloud Run due to networking restrictions (often imposed by intermediary layer 7 load balancers). As a result, the hub cannot function properly in such environments. Nor can traditional proxy software. This limitation is one of the core motivations for the hub-probe architecture: probes alone cannot function as proxies in these environments, but they can relay traffic once coordinated by the hub.

The hub must therefore run in an environment that allows clients to initiate raw TCP connections — this can be a VM or any environment that allows unrestricted inbound traffic. Importantly, only one hub is needed no matter the number of probes.

## Client Compatibility

The **rotox** hub currently supports two protocols:

-   **HTTP plaintext requests:** The client sends a standard HTTP request to the proxy, which forwards it unmodified to the target. This mode supports only HTTP — not HTTPS.
-   **HTTP CONNECT requests:** The client sends a CONNECT request to the proxy, specifying the hostname or address of the target. The proxy then establishes a TCP connection to the target and relays traffic bidirectionally. This enables the client and target to establish a secure TLS session, and their communication is no longer limited to the HTTP protocol.

Support for SOCKS5 is planned (see [Roadmap](#roadmap-non-committal)) but not yet implemented.

## Quick Start

The simplest way to get up and running with **rotox** is to run a hub and a single probe locally.

### Using Docker

#### Step 1: Create a `docker-compose.yml` file

```yaml
version: "3.8"

services:
    hub:
        image: ghcr.io/isacskoglund/rotox-hub:latest
        environment:
            CONFIG_FILE: /etc/rotox/config.yml
            HTTP_PROXY_SECRET: proxysecret
            PROBES_SECRET: probesecret
        volumes:
            - ./hub-config.yml:/etc/rotox/config.yml:ro
        ports:
            - "8000:8000" # HTTP proxy

    probe:
        image: ghcr.io/isacskoglund/rotox-probe:latest
        environment:
            LOG_LEVEL: debug
            LOG_FORMAT: json
            PORT: 8000
            SECRET: probesecret
```

> ⚠️ Make sure the secrets match between `hub` and `probe`. The `hub` authenticates itself to the `probe` using the shared `SECRET`.

#### Step 2: Create a hub config file (`hub-config.yml`)

```yaml
log_level: debug
log_format: json

proxies:
    http:
        port: 8000
        secret_env: HTTP_PROXY_SECRET

probes:
    - secret_env: PROBES_SECRET
      require_tls: false
      hosts:
          - probe:8000
```

> 📄 This config is mounted into the hub container and loaded at startup.

#### Step 3: Start the stack

```bash
docker-compose up
```

The hub will start and listen for HTTP proxy requests on port `8000`, relaying incoming traffic through the probe.

### Step 4: Test the proxy with curl

Use curl to send an HTTP request through the rotox proxy (hub running on localhost:8080):

```bash
curl -x http://localhost:8000 http://example.com
```

This sends a plaintext HTTP request via the rotating proxy. The hub will route it through one of the probes.

For HTTPS sites, test the HTTP CONNECT support:

```bash
curl -x http://localhost:8000 -L https://example.com
```

This uses CONNECT to tunnel TLS traffic through the proxy.

## Configuration

### Hub

The hub configuration is loaded from a YAML file specified through an environment variable:

```bash
CONFIG_FILE=/path/to/config.yml
```

Below is an example config file:

```yaml
log_level: info
log_format: json

proxies:
    http:
        port: 8080
        secret_env: HTTP_PROXY_SECRET

telemetry:
    port: 9000
    secret: secret_value

probes:
    - secret_env: PROBES_SECRET_1
      require_tls: false
      hosts:
          - localhost:8000
          - host.docker.internal:8000

    - secret_env: PROBES_SECRET_2
      require_tls: true
      hosts:
          - 10.0.0.1:8000
          - 10.0.0.2:8000
```

**Fields:**

-   `log_level`: Log verbosity (`debug`, `info`, `warn`, `error`).

-   `log_format`: Log output format (`json` or `text`).

-   `proxies`: Defines proxy listeners. Currently supports only `http`.

    -   `port`: Port for proxy client connections.
    -   `secret_env`: Environment variable name holding the proxy secret.

-   `telemetry` (optional): Telemetry server configuration.

    -   `port`: Telemetry server port.
    -   `secret`: Secret for telemetry access.  
        _Telemetry is not yet a fully implemented feature. It is recommended that this field is omitted, leaving the feature disabled._

-   `probes`: List of probe groups. Each group includes:
    -   `secret_env`: Environment variable name for the probe secret.
    -   `require_tls`: Whether TLS is required (`true` or `false`).
    -   `hosts`: List of one or more probe host addresses.

---

### Probe

The probe uses environment variables for configuration:

-   `LOG_LEVEL`: Log verbosity (`debug`, `info`, `warn`, `error`).

-   `LOG_FORMAT`: Log format (`json` or `text`).

-   `PORT`: Port where the probe listens for hub connections.

-   `SECRET`: Secret string used to authenticate the probe with the hub.

You can run multiple probes, each with different `SECRET` and `PORT` values.

## Roadmap (non-committal)

-   SOCKS5 Support
-   Hub telemetry API
-   Web UI for monitoring
-   IP usage statistics

## Disclaimer

This is a hobby project. No guarantees are made regarding stability, performance, or ongoing development. Contributions welcome.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
