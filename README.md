# VulnNotes CTF Platform

An educational web application with intentional vulnerabilities for hands-on security training. Covers OWASP Top 10 attack classes.

## Requirements

- Docker
- Docker Compose

## Quick Start

```bash
git clone <repo-url>
cd test-your-might

cp .env.example .env
# Edit .env if needed

docker-compose up -d
```

Add to `/etc/hosts` (Linux/Mac) or `C:\Windows\System32\drivers\etc\hosts` (Windows):

```
127.0.0.1 ctf.local api.ctf.local dashboard.ctf.local evil.ctf.local
```

| Service   | URL                                  |
|-----------|--------------------------------------|
| App       | https://ctf.local           |
| API       | https://api.ctf.local       |
| Dashboard | https://dashboard.ctf.local |
| Evil Site | https://evil.ctf.local      |

## Configuration

```bash
# .env
DOMAIN=ctf.local   # or IP / your domain
HOST_PORT=80
JWT_SECRET=secret123        # change in production
TEAM_COUNT=6                # 2–22 teams
```

Default credentials: `alice / password123`, `bob / qwerty`

## Database Management

Team databases are stored in `data/` and persist across restarts. Progress tracking is isolated in `data/progress.db` — even if a team resets their DB, their dashboard score is preserved.

```bash
# Reset one team's data (users + notes only)
rm data/scorpion.db && docker-compose restart backend

# Reset one team's progress on the dashboard
sqlite3 data/progress.db "DELETE FROM completed_tasks WHERE team='scorpion';"

# Full reset (all teams)
rm data/*.db && docker-compose restart backend
```

## Instructor Dashboard

Connect to `https://dashboard.ctf.local` to monitor team progress in real time via WebSocket.

## Rebuild / Logs

```bash
docker-compose up -d --build   # rebuild after code changes
docker-compose logs -f         # all services
docker-compose logs -f backend # backend only
docker-compose down            # stop
```

## Warning

This application contains intentional security vulnerabilities. Do not expose it to the internet without proper network isolation. For classroom / CTF use only.

## License

MIT
