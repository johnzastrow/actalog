# Some notes on hosting

## Deploying ActaLog
ActaLog can be hosted in various environments, including local servers, cloud platforms, and containerized setups. Below are some general guidelines for deploying ActaLog effectively.

### Deployment Considerations
1. **Environment**: Choose an appropriate environment based on your needs. ActaLog can run on Linux, Windows, or macOS. For production environments, Linux servers are often preferred for their stability and performance.
2. **Dependencies**: Ensure that all necessary dependencies are installed. ActaLog requires Node.js and a database (e.g., PostgreSQL, MySQL) to function correctly.
3. **Configuration**: Configure ActaLog according to your requirements. This includes setting up
4. database connections, environment variables, and any other application-specific settings.
5. **Security**: Implement security best practices, such as using HTTPS, setting up firewalls, and regularly updating software to protect against vulnerabilities.
6. **Scaling**: Consider how you will scale ActaLog as your user base grows. This may involve load balancing, database optimization, and using container orchestration tools like Kubernetes.
7. **Monitoring**: Set up monitoring and logging to keep track of application performance and errors. Tools like Prometheus, Grafana, or ELK Stack can be useful for this purpose.
8. **Backup**: Regularly back up your database and application data to prevent data loss in case of failures.
9. **Documentation**: Refer to the official ActaLog documentation for detailed deployment instructions and best practices.
10. **Support**: Join the ActaLog community or seek professional support if you encounter issues during deployment.
11. By following these guidelines, you can ensure a smooth deployment of ActaLog that meets your operational needs.

### Deploying with Docker
ActaLog can be easily deployed using Docker, which simplifies the setup process and ensures consistency across different environments. Below are the steps to deploy ActaLog using Docker:
1. **Install Docker**: Ensure that Docker is installed on your server. You can follow the official Docker installation guide for your operating system.
2. **Pull the ActaLog Docker Image**: Use the following command to pull the latest ActaLog Docker image from Docker Hub:
   ```bash
    docker pull actalog/actalog:latest
    ```
3. **Create a Docker Network**: Create a Docker network to allow communication between the ActaLog container and the database container:
4.   ```bash
    docker network create actalog-network
    ```
5. **Run the Database Container**: Start a database container (e.g., PostgreSQL) and connect it to the Docker network:
6.  ```bash
    docker run -d --name actalog-db --network actalog-network -e POSTGRES_USER=actalog -e POSTGRES_PASSWORD=yourpassword -e POSTGRES_DB=actalog_db postgres:latest
    ```
7. **Run the ActaLog Container**: Start the ActaLog container, linking it to the database container:
8.  ```bash
    docker run -d --name actalog --network actalog-network -p 3000:3000 -e DB_HOST=actalog-db -e DB_USER=actalog -e DB_PASSWORD=yourpassword -e DB_NAME=actalog_db actalog/actalog:latest
    ```
9. **Access ActaLog**: Once the containers are running, you can access ActaLog by navigating to `http://your-server-ip:3000` in your web browser.
10. **Persist Data**: To ensure that your database data persists across container restarts,
    consider using Docker volumes to store the database data outside of the container.
11. **Monitor and Manage**: Use Docker commands to monitor and manage your ActaLog and database containers as needed.


### Deploying with Docker Compose
Using Docker Compose simplifies the deployment process by allowing you to define and manage multi-container applications with a single configuration file. Below are the steps to deploy ActaLog using Docker Compose:
1. **Install Docker and Docker Compose**: Ensure that both Docker and Docker Compose are installed
2. on your server. You can follow the official installation guides for your operating system.
3. **Create a Docker Compose File**: Create a `docker-compose.yml` file with the following content:

    --- fix ---

### Deploying from Source
To deploy ActaLog from source, follow these steps:

   **See SETUP.md for initial setup instructions.**

    
### Reverse Proxy Setup
Setting up a reverse proxy is essential for securely hosting ActaLog in a production environment. A reverse proxy acts as an intermediary between clients and the ActaLog application, providing benefits such as load balancing, SSL termination, and improved security.
Here are the general steps to set up a reverse proxy for ActaLog:
    1.  **Choose a Reverse Proxy**: Select a reverse proxy server such as Nginx, Apache, or Caddy based on your preferences and requirements.
    2.  **Install the Reverse Proxy**: Install the chosen reverse proxy server on
    3.  your server following the official installation instructions.
    4.  **Configure the Reverse Proxy**: Set up the reverse proxy to forward
    5.  requests to the ActaLog application. This typically involves creating a configuration file that specifies the server name, port, and proxy settings.
    6.  **Enable HTTPS**: For secure communication, configure SSL/TLS certificates
    7.  for your reverse proxy. You can use Let's Encrypt for free SSL certificates.
    8.  **Test the Configuration**: After setting up the reverse proxy, test
    9.  the configuration to ensure that requests are correctly forwarded to ActaLog and that HTTPS is functioning properly.
    10. **Monitor and Maintain**: Regularly monitor the reverse proxy logs and
    


## Example Reverse Proxy Configuration

When hosting ActaLog or similar applications behind a reverse proxy, it's important to set up your DNS and proxy server correctly to ensure smooth operation and accessibility.

Given a zone file such as the following from Linode:

```
; mydomainname.site [XXXXXXX]
$TTL 86400
@  IN  SOA  ns1.linode.com. myemail.gmail.com. 2021000021 14400 14400 1209600 86400
@    NS  ns1.linode.com.
@    NS  ns2.linode.com.
@    NS  ns3.linode.com.
@    NS  ns4.linode.com.
@    NS  ns5.linode.com.
@      MX  10  mail.mydomainname.site.
@      A  the.public.ip.address
al    30  A  the.public.ip.address  ; example of a subdomain for ActaLog. This would run internally on port 3000 but mapped to 80/443 externally
apo    30  A  the.public.ip.address ; example of a subdomain for another container perhaps running on the same server and port 9443, but mapped to 443 externally
mail      A  the.public.ip.address
recipe      A  the.public.ip.address ; example of a subdomain for a recipe app
www      AAAA  2600:3c02::f03c:95ff:feda:027e ; example of an IPv6 address for www
```

I like to use Caddy as a proxy server because it has automatic HTTPS via Let's Encrypt and is easy to configure.

Here is an example Caddyfile for the above zone file:

```
mydomainname.site, www.mydomainname.site {  
    reverse_proxy localhost:8080  # assuming ActaLog is running on port 8080 internally
}
al.mydomainname.site {
    reverse_proxy localhost:3000  # ActaLog running on port 3000 internally
        log {
                output file /var/log/caddy/al.access.log {
                        roll_size 1MB # Create new file when size exceeds 10MB
                        roll_keep 5 # Keep at most 5 rolled files
                        #            roll_keep_days 14 # Delete files older than 14 days
                }
        }
}
apo.mydomainname.site {
    reverse_proxy localhost:9443  # another app running on port 9443 internally
}
recipe.mydomainname.site {
    reverse_proxy localhost:5000  # recipe app running on port 5000 internally
}
```

* Make sure to adjust the internal ports according to where your applications are running. With this setup, Caddy will handle incoming requests and route them to the appropriate internal service based on the subdomain.
* Remember to open the necessary ports (80 and 443) on your server's firewall to allow HTTP and HTTPS traffic.
* Also, ensure that your DNS settings are correctly pointing to your server's public IP address for each of the subdomains you wish to use.

## Validating Caddyfile Syntax

### Using caddy validate

You can validate the syntax of a Caddyfile using the command line with the `caddy validate` or `caddy adapt` commands.

The caddy validate command loads and provisions the configuration, checking for any errors that might arise during the loading and provisioning stages, without actually starting the server.

To use it:

```bash
caddy validate --config /path/to/Caddyfile # often /etc/caddy/Caddyfile
```

* If the command succeeds, it will exit with no output and a status code of 0, indicating a valid configuration.
* If there are any syntax errors or structural issues, it will print an error message to the console and exit with a non-zero status code.
* If your file is named just Caddyfile and is in your current directory, you can simply run `caddy validate` without the --config flag.

### Using caddy adapt

* The caddy adapt command is another way to check your Caddyfile. It converts the Caddyfile into Caddy's native JSON format. This process will catch most syntax errors.
* To use it:

```bash
caddy adapt --config /path/to/Caddyfile # often /etc/caddy/Caddyfile
```

* If successful, it will output the resulting JSON configuration to standard output (stdout).
* If it fails, it means there is a syntax error in your Caddyfile.
* Both commands are excellent ways to perform a "dry run" of your configuration and ensure it is correct before deploying it to a production environment. For more details, consult the Caddy Documentation on the command-line interface.

## Common Reverse Proxy Configuration Mistakes

When configuring a reverse proxy for applications like ActaLog, several common mistakes can lead to performance issues, functional problems, or security vulnerabilities. Here are some of the most frequent errors to watch out for:

### Performance & Operational Mistakes

* **Not enabling keepalive connections**: By default, a new connection is often opened for every request to the backend server, which is inefficient. Enabling keepalive connections to upstream servers reuses existing connections, significantly improving performance.

HTTP keep-alive is enabled by default in Caddy, but you can configure its behavior in the reverse_proxy block of your Caddyfile or JSON config. You can customize settings like the idle timeout and the interval for probing connections.

#### Caddyfile

```caddyfile
your_site {
    reverse_proxy localhost:8080 {
        health_uri /health
        health_interval 1m
        health_timeout 5s
    }
    # To adjust the HTTP Keep-Alive settings, add a transport block.
    # This is a basic example. The full configuration options are in the Caddy JSON structure.
    transport http {
        keep_alive {
            # Optional: Enable/disable keep-alive (default is true)
            # enabled true 

            # Optional: Set probe interval (default 30s)
            # probe_interval 60s

            # Optional: Set idle timeout (default 5 minutes)
            # idle_timeout 5m
        }
    }
}
```

#### JSON config

```json
{
  "apps": {
    "http": {
      "servers": {
        "my_server": {
          "listen": [":443"],
          "routes": [
            {
              "handle": [
                {
                  "handler": "reverse_proxy",
                  "upstreams": [{"address": "localhost:8080"}],
                  "transport": {
                    "http": {
                      "keep_alive": {
                        "enabled": true,
                        "probe_interval": "30s",
                        "idle_timeout": "5m"
                      }
                    }
                  }
                }
              ]
            }
          ]
        }
      }
    }
  }
}
```

* **Default DNS resolution settings**: Many proxies (like Nginx) resolve domain names only once at startup. If your backend server's IP address changes, the proxy will keep sending traffic to the old IP unless you manually reload the configuration or use the resolver directive for dynamic resolution.
* **Improper timeout settings**: Timeouts (e.g., proxy_read_timeout) are often left at their defaults. If a backend server is slow to respond, users may encounter 504 Gateway Timeout errors. These settings should be adjusted to match the expected behavior and load of the backend application.

You can adjust Caddy's timeout settings in the Caddyfile or JSON config by using the `timeouts` directive, which can set a default for all timeouts or be configured individually for read, header, write, and idle times. For a reverse_proxy in Caddyfile, you can use the `response_header_timeout` and ``transport timeouts`. Timeouts can be set in duration formats like 30s, 1m, or 5m, or set to 0 or none to disable them.

#### In the Caddyfile

**Global timeouts**

Use the timeouts global option to set a default for all timeouts across your Caddyfile.
caddyfile

**Set all timeouts to 1 minute**
timeouts 1m

**Set specific timeouts**

```
timeouts {
    read 30s
    write 20s
    idle 5m
}
```

**Per-site timeouts**

You can also set timeouts for individual sites.
caddyfile

```
site.com {
    timeouts {
        read_header 1m
        read_body 0 # Disable read_body timeout
    }
}
```

**reverse_proxy timeouts**

For reverse proxy configurations, you can specify additional timeouts within the reverse_proxy block:

* `response_header_timeout`: Sets a timeout for when the backend does not write any response headers.
* `transport`: Sets a timeout for an individual API call to the backend.
*

```caddyfile
reverse_proxy example.com {
    header_up X-Real-IP {remote_host}
    response_header_timeout 10s
    transport 20s
}
```

#### In JSON configuration

**Global timeouts**
You can set global timeouts in the apps.http.servers block.

```json
{
    "apps": {
        "http": {
            "servers": {
                "myserver": {
                    "routes": [
                        {
                            "handle": [
                                {
                                    "handler": "reverse_proxy",
                                    "upstreams": [
                                        {"dial": "localhost:8080"}
                                    ]
                                }
                            ]
                        }
                    ],
                    "timeouts": {
                        "read_header": "1m",
                        "read_body": "1m",
                        "write": "1m",
                        "idle": "5m"
                    }
                }
            }
        }
    }
}
```

**Per-handler timeouts**

Some handlers may support their own timeouts. For example, `reverse_proxy` can take `response_header_timeout` and transport under the response field.

```json
{
    "apps": {
        "http": {
            "servers": {
                "myserver": {
                    "routes": [
                        {
                            "handle": [
                                {
                                    "handler": "reverse_proxy",
                                    "upstreams": [
                                        {"dial": "localhost:8080"}
                                    ],
                                    "response": {
                                        "header_timeout": "10s",
                                        "transport_timeout": "20s"
                                    }
                                }
                            ]
                        }
                    ]
                }
            }
        }
    }
}
```

* **Inadequate resource limits**: Failing to increase system resource limits, such as the maximum number of open file descriptors, can lead to service disruptions under heavy load.

### Functional Configuration Mistakes

* **Not forwarding the correct client IP/Host headers**: Backend applications often need the original client's IP address for logging, analytics, or security purposes. Without correctly setting headers like X-Real-IP or X-Forwarded-For, the backend will only see the proxy's IP address.
* **Mishandling redirects (absolute URLs)**: If the backend application sends an absolute URL in a redirect (e.g., in a 302 response) that is different from the external URL the client used, the client may be redirected to an internal, inaccessible URL.
* **Improper WebSocket proxying**: WebSocket connections require specific headers (Upgrade and Connection) to be handled correctly by the proxy; otherwise, real-time communication features may fail.

### Security Mistakes

* **Assuming a reverse proxy is a WAF**: A basic reverse proxy provides an abstraction layer but does not automatically provide Web Application Firewall (WAF), Intrusion Prevention System (IPS), or Intrusion Detection System (IDS) features. Additional security measures or a dedicated WAF are needed for robust protection.
* **Exposing internal services or paths**: Misconfigured routing rules, virtual hosts, or aliases can inadvertently expose internal applications or sensitive files to the public internet.
* **Leaving default credentials/settings**: Not changing default configurations or credentials for the proxy management interface or underlying systems leaves them vulnerable to attackers who can use publicly known default values.
* **Outdated software**: Failing to apply security patches to the reverse proxy software itself is a critical mistake that leaves known vulnerabilities open for exploitation.
* **Host Header manipulation vulnerabilities**: If not properly configured, an attacker can manipulate the Host header to access unintended backend servers or exploit misconfigured virtual hosts.
* **Improper SSL/TLS configuration**: Misconfiguring SSL/TLS settings can lead to security vulnerabilities, such as man-in-the-middle attacks or insecure data transmission.

## Conclusion

Careful configuration and regular review of your reverse proxy settings are essential to ensure optimal performance, functionality
, and security for applications like ActaLog. Always test your configuration in a staging environment before deploying it to production.
