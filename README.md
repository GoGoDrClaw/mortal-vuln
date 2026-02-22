# VulnNotes CTF Platform

An educational web application with intentional vulnerabilities for hands-on security training. Covers OWASP Top 10 attack classes.

## Requirements

- Docker
- Docker Compose (v2)

## Quick Start

```bash
git clone <repo-url>
cd test-your-might

cp .env.example .env
# Edit .env if needed (at minimum change JWT_SECRET and POSTGRES_PASSWORD)

docker compose up -d
```

Add to `/etc/hosts` (Linux/Mac) or `C:\Windows\System32\drivers\etc\hosts` (Windows):

```
127.0.0.1 ctf.local api.ctf.local dashboard.ctf.local
```

| Service   | URL                                   |
|-----------|---------------------------------------|
| App       | https://ctf.local                     |
| API       | https://api.ctf.local                 |
| Dashboard | https://dashboard.ctf.local           |

## Configuration

```env
DOMAIN=ctf.local            # or IP / your domain
JWT_SECRET=secret123        # change before running!
POSTGRES_PASSWORD=ctfpassword
RATE_LIMIT_RPS=20           # requests/sec per IP (DDoS protection)
RATE_LIMIT_BURST=40
```

## How it works

Each player creates their own session:

1. Open the app → **NEW GAME**: choose a fighter, enter a nickname
2. A **save code** is shown (format `XXXX-XXXX`) — write it down
3. On another device or after cookie loss → **CONTINUE**: enter your save code

Each session has an isolated SQLite database for CTF tasks (SQL injection stays contained). Progress (completed tasks, score) is persisted in PostgreSQL.

## Default CTF credentials

Each player's database is seeded with three users. Finding their passwords is part of the challenge.

## Data

- **PostgreSQL**: player sessions and task completions (persists across restarts)
- **SQLite** (`data/sessions/<uuid>.db`): per-player CTF database (users, notes)

```bash
# Reset one player's CTF database via the app UI (⟳ RESET REALM button)
# or via API:
curl -X POST https://api.ctf.local/api/reset -b "session_id=<uuid>"

# Wipe all sessions and progress (full reset)
docker compose down -v   # removes postgres_data volume
rm -rf data/sessions/
docker compose up -d
```

## Instructor Dashboard

Connect to `https://dashboard.ctf.local` to monitor player progress in real time via WebSocket. Shows nickname, character, score, and live task feed.

## Rebuild / Logs

```bash
docker compose up -d --build   # rebuild after code changes
docker compose logs -f         # all services
docker compose logs -f backend # backend only
docker compose down            # stop
docker compose down -v         # stop + remove volumes (full wipe)
```

## Warning

This application contains intentional security vulnerabilities. Do not expose it to the internet without proper network isolation. For classroom / CTF use only.

## License

MIT
