
# AdminAPI_WebUI Plugin for Traefik

<img style="float: left; width: 50px; padding-right: 10px;" src="docs/img/logo.png"> Extend Traefik’s functionality with a built-in **Administrative API and Web UI** for managing dynamic configuration files.

This plugin integrates into Traefik as a middleware component, providing an intuitive web interface and REST API to administer dynamic configurations loaded via the file provider.

---

## Setup and Usage

### 1. Configure the Main `traefik.yml` File

To ensure proper plugin operation, configure the file provider to monitor the directory containing your dynamic configuration files:

```yaml
providers:
  providersThrottleDuration: 20s
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
  file:
    directory: /tmp/ # Path to your dynamic configuration files
    watch: true  # Automatically track file changes

experimental:
  plugins:
    traefik_plugin_AdminAPI_WebUI: # Plugin name
      moduleName: "github.com/Miromani4/traefik-plugin-AdminAPI_WebUI"
      version: "v1.3.0" # Use the latest version
```

### 2. Dynamic Configuration Setup

In your dynamic configuration file (YAML), define the middleware and attach it to a dedicated router:

```yaml
http:
  routers:
    admin-ui-router:
      rule: "Host(`admin.localhost`)"
      service: noop@internal  # Required placeholder service
      middlewares:
        - traefik_plugin_AdminAPI_WebUI
      tls:
        {}  # Optional: enable TLS if needed

  middlewares:
    traefik_plugin_AdminAPI_WebUI:
      plugin:
        traefik_plugin_AdminAPI_WebUI:
          root: "/tmp/"  # Path to configuration directory
```

#### The `root` Parameter

Specifies the root directory where Traefik monitors dynamic configuration files (`.yml` or `.yaml`). All operations via the web interface are sandboxed within this directory to enforce security and prevent unauthorized filesystem access.

---

## Local Testing

A `docker-compose.yml` file is provided in the repository for quick local deployment and validation:

```bash
docker-compose up -d
```

Once successfully launched, access the web interface at:  
👉 [http://admin.localhost](http://admin.localhost)

---

## Key Benefits

- **Seamless Integration** — Operates as a standard Traefik middleware without architectural disruption.
- **Centralized Control** — Unified access to manage dynamic configurations via UI or API.
- **Security by Design** — File system access strictly confined via the `root` parameter.
- **GitOps Ready** — Fully compatible with declarative infrastructure management workflows.

---

This plugin is ideal for DevOps and platform engineering teams seeking to automate and simplify microservices infrastructure management through a unified, self-service interface.

---

## Screenshots & Features

### Interface Overview

![Screenshot 1](https://github.com/Miromani4/traefik-plugin-AdminAPI_WebUI/blob/main/docs%2Fimg%2Fadmin_panel.png)

#### YAML Structure Validation & Error Highlighting

![Screenshot 2](https://github.com/Miromani4/traefik-plugin-AdminAPI_WebUI/blob/main/docs%2Fimg%2Fadmin_panel_error.png)

#### Keyboard Shortcuts

`Ctrl + /` — Toggle comment/uncomment for selected line(s)

---

> 💬 **Have questions or feature requests?**  
> Please open an issue on GitHub — we welcome your feedback and contributions!

---
