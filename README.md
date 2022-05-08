# deploy2docker

Deploy to Remote Docker using SSH


### Config
```yml
services:
  - name: nginx
    image: nginx
    ports:
      - 80
    volumes:
      - /var/log/nginx:/var/log/nginx
      - /var/www/html:/var/www/html
    environment:
      - NGINX_PORT=80
    networks:
      - frontend

```

### Deploy

```sh
deploy2docker --remote user@remote --password password
```
