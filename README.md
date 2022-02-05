#  ELC - orchestrator of development environments

With ELC you can:
* start a couple of docker-compose projects with one command
* define dependencies across docker-compose projects
* use one docker-compose template for several services
* describe sets of services for different cases (testing, development, monitoring)
* use containerized development tools

## How to use

Make a workspace config file, which contains:

**global variables**
```yaml
name: ensi
variables:
  NETWORK: ensi
  BASE_DOMAIN: ensi.127.0.0.1.nip.io
```
**docker compose templates**
```yaml
templates:
  - name: php80
    path: {{.WorkspacePath}}/templates/php8
    compose_file: {{.WorkspacePath}}/templates/php8/docker-compose.yml
    variables:
      BASE_IMAGE: php:8.0-fpm-alpine
      APP_IMAGE: php80:latest
      NGINX_IMAGE: nginx:1.19-alpine
```

**service definitions**
```yaml
services: 
  - name: api
    extends: php80
    path: {{.WorkspacePath}}/apps/api
    dependencies:
      - database
```

Register workspace in elc:
```bash
$ elc workspace add ensi /path/to/workspace/
$ elc workspace select ensi
```

Start some services:

```bash
$ elc start api
```

Invoke some tool

```bash
$ cd /path/to/service/directory
$ elc composer install
```