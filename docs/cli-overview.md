# n8r CLI

The n8r CLI is the command-line interface for Injectionator. It authenticates users via OAuth Device Flow and provides access to their Injectionator profile and data.

## Installation

```bash
brew install injectionator/n8r-brew/n8r
```

To upgrade:

```bash
brew update
brew upgrade n8r
```

## Commands

### `n8r`

Displays the branded banner. If not logged in, shows alpha access notice and prompts to authenticate. If logged in, shows available commands.

### `n8r login`

Authenticates with your Injectionator account using OAuth Device Flow:

1. CLI requests a device code from `injectionator.com`
2. You visit `https://injectionator.com/cli-auth` in your browser
3. Enter the code displayed in your terminal
4. CLI receives an access token and stores it locally

Credentials are saved to `~/.n8r/credentials.json`. Tokens expire after 30 days.

### `n8r profile`

Displays your Injectionator profile including:

- Name, email, role, organization
- Interests
- Cohort memberships
- Points and missions progress
- Earned badges

Requires authentication.

### `n8r logout`

Removes stored credentials from `~/.n8r/credentials.json`.

### `n8r status`

Shows current authentication status including token expiry.

### `n8r version`

Prints the CLI version. Also available as `n8r --version` or `n8r -v`.

## Architecture

### CLI (this repo)

Written in Go with no external dependencies (stdlib only). Source layout:

- `cmd/n8r/main.go` — entrypoint and command routing
- `internal/auth/device.go` — device flow HTTP client
- `internal/auth/token.go` — token storage (`~/.n8r/credentials.json`)
- `internal/config/config.go` — base URL, version
- `Formula/n8r.rb` — Homebrew formula

### Website API (injectionator-website repo)

Endpoints on `injectionator.com` that the CLI calls:

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/auth/device/code` | POST | None | Generate device code + user code |
| `/api/auth/device/token` | POST | None | CLI polls for access token |
| `/api/auth/device/authorize` | POST | Clerk session | Browser approves device code |
| `/api/auth/profile` | GET | Bearer token | Returns user profile data |

The `/cli-auth` page is where users enter the code shown by `n8r login`. It requires Clerk authentication.

### Database (Cloudflare D1)

Two tables support the CLI auth flow:

- `device_codes` — stores device code, user code, status (pending/approved/denied/expired), user_id, expires_at
- `user_api_tokens` — stores hashed access tokens, user_id, scopes, expires_at

### Auth Flow

```
CLI                        injectionator.com              Browser
 |                              |                            |
 |-- POST /device/code -------->|                            |
 |<-- device_code, user_code ---|                            |
 |                              |                            |
 |  "Visit .../cli-auth"        |                            |
 |  "Enter code: ABCD-EFGH"    |                            |
 |                              |                            |
 |                              |<--- User visits /cli-auth -|
 |                              |<--- POST /device/authorize |
 |                              |     (user_code + session)  |
 |                              |--- "Authorized" ---------->|
 |                              |                            |
 |-- POST /device/token ------->|                            |
 |<-- access_token -------------|                            |
 |                              |                            |
 |  Token saved to              |                            |
 |  ~/.n8r/credentials.json     |                            |
```

## Releasing

Releases use pinned GitHub Release assets to avoid sha256 mismatch issues.

To cut a new release (e.g., v0.3.0):

```bash
# 1. Build the source tarball (exclude Formula to avoid circular sha256 issues)
git archive --format=tar.gz --prefix=n8r-0.3.0/ HEAD -- cmd internal go.mod Makefile > /tmp/n8r-0.3.0.tar.gz

# 2. Get the sha256
shasum -a 256 /tmp/n8r-0.3.0.tar.gz

# 3. Update Formula/n8r.rb with new version URL and sha256

# 4. Commit and push the formula update
git add Formula/n8r.rb
git commit -m "Bump formula to v0.3.0"
git push origin main

# 5. Tag the commit BEFORE the formula update (so tarball doesn't include the new sha256)
git tag v0.3.0 HEAD~1
git push origin v0.3.0

# 6. Create the GitHub Release with the tarball
gh release create v0.3.0 /tmp/n8r-0.3.0.tar.gz --title "v0.3.0" --notes "Release notes here"
```

Users upgrade with `brew update && brew upgrade n8r`.

## Current Version

v0.2.0 — adds `n8r profile` command.
