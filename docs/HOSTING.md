# Some notes on hosting

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
caddy validate --config /path/to/Caddyfile
```

* If the command succeeds, it will exit with no output and a status code of 0, indicating a valid configuration.
* If there are any syntax errors or structural issues, it will print an error message to the console and exit with a non-zero status code.
* If your file is named just Caddyfile and is in your current directory, you can simply run `caddy validate` without the --config flag. 

### Using caddy adapt

* The caddy adapt command is another way to check your Caddyfile. It converts the Caddyfile into Caddy's native JSON format. This process will catch most syntax errors.
* To use it:

```bash
caddy adapt --config /path/to/Caddyfile
```
* If successful, it will output the resulting JSON configuration to standard output (stdout).
* If it fails, it means there is a syntax error in your Caddyfile.
  * Both commands are excellent ways to perform a "dry run" of your configuration and ensure it is correct before deploying it to a production environment. For more details, consult the Caddy Documentation on the command-line interface.
  * Validating Caddyfile Syntax
  * You can validate the syntax of a Caddyfile using the command line with the caddy validate or caddy adapt commands.
If there are any syntax errors or structural issues, it will print an error message to the console and exit with a non-zero status code.
If your file is named just Caddyfile and is in your current directory, you can simply run caddy validate without the --config flag. 
Using caddy adapt
The caddy adapt command is another way to check your Caddyfile. It converts the Caddyfile into Caddy's native JSON format. This process will catch most syntax errors. 
To use it:
bash
caddy adapt --config /path/to/Caddyfile
Use code with caution.

If successful, it will output the resulting JSON configuration to standard output (stdout).
If it fails, it means there is a syntax error in your Caddyfile. 
Both commands are excellent ways to perform a "dry run" of your configuration and ensure it is correct before deploying it to a production environment. For more details, consult the Caddy Documentation on the command-line interface. 


Performance & Operational Mistakes
Not enabling keepalive connections: By default, a new connection is often opened for every request to the backend server, which is inefficient. Enabling keepalive connections to upstream servers reuses existing connections, significantly improving performance.
Default DNS resolution settings: Many proxies (like Nginx) resolve domain names only once at startup. If your backend server's IP address changes, the proxy will keep sending traffic to the old IP unless you manually reload the configuration or use the resolver directive for dynamic resolution.
Improper timeout settings: Timeouts (e.g., proxy_read_timeout) are often left at their defaults. If a backend server is slow to respond, users may encounter 504 Gateway Timeout errors. These settings should be adjusted to match the expected behavior and load of the backend application.
Inadequate resource limits: Failing to increase system resource limits, such as the maximum number of open file descriptors, can lead to service disruptions under heavy load. 
Functional Configuration Mistakes
Not forwarding the correct client IP/Host headers: Backend applications often need the original client's IP address for logging, analytics, or security purposes. Without correctly setting headers like X-Real-IP or X-Forwarded-For, the backend will only see the proxy's IP address.
Mishandling redirects (absolute URLs): If the backend application sends an absolute URL in a redirect (e.g., in a 302 response) that is different from the external URL the client used, the client may be redirected to an internal, inaccessible URL.
Improper WebSocket proxying: WebSocket connections require specific headers (Upgrade and Connection) to be handled correctly by the proxy; otherwise, real-time communication features may fail. 
Security Mistakes
Assuming a reverse proxy is a WAF: A basic reverse proxy provides an abstraction layer but does not automatically provide Web Application Firewall (WAF), Intrusion Prevention System (IPS), or Intrusion Detection System (IDS) features. Additional security measures or a dedicated WAF are needed for robust protection.
Exposing internal services or paths: Misconfigured routing rules, virtual hosts, or aliases can inadvertently expose internal applications or sensitive files to the public internet.
Leaving default credentials/settings: Not changing default configurations or credentials for the proxy management interface or underlying systems leaves them vulnerable to attackers who can use publicly known default values.
Outdated software: Failing to apply security patches to the reverse proxy software itself is a critical mistake that leaves known vulnerabilities open for exploitation.
Host Header manipulation vulnerabilities: If not properly configured, an attacker can manipulate the Host header to access unintended backend servers or exploit misconfigured virtual hosts. 



