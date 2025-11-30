# openvpn-log-api

[![Version](https://img.shields.io/badge/version-1.0.0-green)](https://github.com/cinquecentoandrey/openvpn-log-api/releases)
[![Go Version](https://img.shields.io/badge/Go-1.23.4-00ADD8?logo=go)](https://go.dev/dl/go1.23.4)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![REST API](https://img.shields.io/badge/REST-API-FF6B6B)](https://en.wikipedia.org/wiki/REST)
[![OpenVPN](https://img.shields.io/badge/OpenVPN-Logs-EA7E20)](https://openvpn.net)

> Lightweight REST API service for retrieving OpenVPN Access Server logs by date range

**Features:**
- 📅 Filter logs by date range
- 📊 JSON-formatted output
- 🔐 Token-based authentication
- ⚡ Fast log retrieval via `logdba`
- 🐳 Docker support


To build the binary using Docker, run the following command:
```bash
docker build --target export --output type=local,dest=./bin .
```
The binary will appear in the ./bin directory.

----

If Go is installed on your machine, you can build it using the following command:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o openvpn-log-api main.go
```

### Steps to run the application on a server
0. Upload the compiled binary to the server.
---
1. Copy the binary to the directory:
```bash
cp openvpn-log-api /usr/local/bin/
```
---
2. Create a user ovpnloguser:
```bash
useradd --no-create-home --shell /bin/false ovpnloguser
```
---
3. Create a directory for configuration:
```bash
mkdir /etc/openvpn-log-api/config
```
---
4. Create config.yaml. Copy the contents from the config/config.yaml file in the repository:
```bash 
nano /etc/openvpn-log-api/config/config.yaml
```
---
5. Create a unit file. Copy the contents from the openvpn-log-api.service file in the repository:
```bash
nano /etc/systemd/system/openvpn-log-api.service
```
---
6. Change the owner of the binary:
```bash
chown ovpnloguser:ovpnloguser /usr/local/bin/openvpn-log-api
```
---
7. Grant read permissions for the log.db file to the user:
```bash
setfacl  -m u:ovpnloguser:r /usr/local/openvpn_as/etc/db/log.db
```
---
8. Reload the systemd configuration:
```bash
systemctl daemon-reload
```
---
9. Enable the service to start on boot:
```bash
systemctl enable openvpn-log-api
```
---
10. Start the service:
```bash
systemctl start openvpn-log-api
```
---
11. If you've made changes to the firewall settings, it is necessary to restart OpenVPN.
---

### Configuration params
```yaml
server:
  port: 8080                                                    // Port on which the application accepts incoming HTTP requests.
  host: "0.0.0.0"                                               // IP address or host on which the application listens for connections.

auth:
  api_token: "<token>"                                          // Token used for request authentication, passed by the client in the HTTP request header.
  token_header: "<token-header>"                                // Name of the HTTP header in which the client must pass the authentication token.

logging:
  level: "info"                                                 // Logging level.
  file_path: ""                                                 // Path to the log file. If the value is empty (""), logs are output to stdout.

vpn:
  logdba_path: "/usr/local/openvpn_as/scripts/logdba"           // Path to the `logdba` utility used for working with OpenVPN logs.
  logdb_path: "sqlite:////usr/local/openvpn_as/etc/db/log.db"   // Path to the OpenVPN log database.
  timeout: 20                                                   // Timeout (in seconds) for log read operations from the database.
```

### P.S. 
If a log file path is specified in the config, you need to make ovpnloguser the owner of that file:
```bash
chown ovpnloguser:ovpnloguser <path-to-log-file>
```
