<h1 align="center">Dioptra</h1>
<h3 align="center">One command to browse your remote Postgres database. No exposed ports, no config GUIs.</h3>

> dioptra /diˈop.tra/ — ancient Greek: an instrument for measuring angles by sight.
## Why?
I decided to use a local postgresql server in my own VPS instead of having an external one, or using supabase. But at the same time I wanted to easily see if there were new users to my app, or how they were using it without having to log into my VPS, use the CLI and query for the table. I wanted it fast and easy, without exposing any port. So thats why I built this: ONE command line and you get direct access to your database in your browser, allowing you to read and write anything on it.

## How it works
```
dioptra connect --host vps.example.com --ssh-user root --db mydb
```

Opens a browser with a clean dashboard to browse tables and run queries. Press `Ctrl+C` to tear everything down.

```
Your Machine                                VPS
┌──────────────────────────────┐   ┌─────────────────┐
│  Browser ◄──► Dioptra Server │   │                 │
│              (embedded UI)   │   │  PostgreSQL     │
│                    │         │   │  localhost:5432 │
│              SSH Tunnel ─────┼──►│       ▲         │
│              :random → 5432  │   │       └── SSH   │
└──────────────────────────────┘   └─────────────────┘
```

1. Establishes an SSH tunnel (reads `~/.ssh/config`, tries agent/keys/password)
2. Connects to Postgres through the tunnel
3. Starts a local web server with the embedded dashboard
4. Opens your browser
5. `Ctrl+C` closes everything cleanly

## Install

### go install

```bash
go install github.com/tomassar/dioptra/cmd/dioptra@latest
```

Requires Go 1.26+. The binary is self-contained, no Node.js or other runtime needed after install.

### Install script (no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/tomassar/dioptra/main/install.sh | sh
```

Detects your OS and architecture, downloads the right binary from GitHub Releases, and puts it in `/usr/local/bin`.

### GitHub Releases

Download the binary for your platform from [github.com/tomassar/dioptra/releases](https://github.com/tomassar/dioptra/releases), extract the archive, and move the binary somewhere on your `$PATH`:

```bash
tar -xzf dioptra_darwin_arm64.tar.gz
mv dioptra /usr/local/bin/
```

### From source

```bash
git clone https://github.com/tomassar/dioptra.git
cd dioptra
make build
```

Requires Go 1.26+ and Node.js 18+.

## Usage

### Quick connect

```bash
# Minimal, prompts for DB password
dioptra connect --host my-vps.com --ssh-user root --db mydb

# With all options
dioptra connect \
  --host my-vps.com \
  --ssh-user deploy \
  --ssh-port 2222 \
  --db mydb \
  --db-user postgres \
  --db-password secret

# Enable write queries (read-only by default)
dioptra connect --host my-vps.com --db mydb --write
```

### Save connections

```bash
# Save a connection
dioptra save prod \
  --host my-vps.com \
  --ssh-user root \
  --db production \
  --ssh-password mysshpass \
  --db-password mydbpass

# List saved connections
dioptra list

# Connect by name (uses saved credentials)
dioptra connect prod

# Override saved values on the fly
dioptra connect prod --db-user admin

# Remove
dioptra remove prod
```

Config is stored at `~/.config/dioptra/config.toml`.

### Flags

| Flag | Description |
|------|-------------|
| `--host` | SSH host (or alias from `~/.ssh/config`) |
| `--ssh-user` | SSH user (default: current user) |
| `--ssh-port` | SSH port (default: 22) |
| `--ssh-key` | Path to SSH private key |
| `--ssh-password` | SSH password (or set `SSHPASS`) |
| `--db` | Database name |
| `--db-user` | Database user (default: postgres) |
| `--db-password` | Database password (or set `PGPASSWORD`) |
| `--db-host` | Remote DB host (default: localhost) |
| `--db-port` | Remote DB port (default: 5432) |
| `--write` | Allow write queries |
| `--no-browser` | Don't auto-open browser |

### SSH

Dioptra reads your `~/.ssh/config` automatically. If you have:

```
Host myvps
    HostName 1.2.3.4
    User deploy
    IdentityFile ~/.ssh/deploy_key
```

Then `dioptra connect --host myvps --db mydb` just works.

Auth methods are tried in order: SSH agent → key files → passphrase prompt → password prompt.

## Dashboard

- **Browse**: Click any table in the sidebar to see its data with pagination
- **Query**: Write SQL and run it with `Cmd+Enter` (results capped at 1,000 rows)
- **Read-only by default**: Use `--write` flag to enable mutations

## License

MIT
