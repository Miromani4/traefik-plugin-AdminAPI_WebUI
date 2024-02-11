# AdminAPI_WebUI

add Admin API & Web UI for Traefik configured dynamic files!

This is a plugin for [Traefik](https://traefik.io) to add a **Admin API & Web UI for Traefik** as a middleware.

## Usage

### Configuration

Here is an example of a file provider dynamic configuration (given here in
YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration

http:
  routers:
    my-waeb-router:
      rule: host(`admin.localhost`)
      service: noop@internal # required
      middlewares:
        - traefik_plugin_AdminAPI_WebUI

  middlewares:
    traefik_plugin_AdminAPI_WebUI:
      plugin:
        traefik_plugin_AdminAPI_WebUI:
          root: "/tmp/"
```

#### `root`

The `root` parameter is the root directory of dynamic configuration files.

### Local test

There is a `docker-compose.yml` file to test the plugin locally:

```bash
docker-compose up -d
```

Then, you can go to [http://waeb.localhost](http://waeb.localhost) to see the
result.
