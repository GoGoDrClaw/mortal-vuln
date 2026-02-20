FROM caddy:2-alpine

COPY Caddyfile /etc/caddy/Caddyfile

EXPOSE 80 8081

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
