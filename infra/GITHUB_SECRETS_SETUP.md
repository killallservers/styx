# GitHub Secrets Setup Guide

Complete step-by-step instructions for configuring the 4 required secrets for production deployment automation.

## Quick Summary

You need to add these 4 secrets to your GitHub repository:

| Secret Name | Value | Where it Comes From |
|---|---|---|
| `PRODUCTION_SERVER_IP` | Server IP address | Your Hetzner server IP |
| `PRODUCTION_SSH_KEY` | Private SSH key | `~/.ssh/hetzner/<SERVER-ID>/root` |
| `DEPLOY_SSH_PUBLIC_KEY` | Public SSH key | `~/.ssh/hetzner/<SERVER-ID>/deploy.pub` |
| `PRODUCTION_DOMAIN` | Domain name | Your domain (e.g., `allora.example.com`) |

---

## Prerequisites

Before adding secrets to GitHub, you must have SSH keys generated locally:

```bash
# Generate root key (for GitHub Actions bootstrap)
ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/root -C "admin@allora.style"

# Generate deploy key (for developers/deploy scripts)
ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/deploy -C "admin@allora.style"

# Verify both files exist
ls -la ~/.ssh/hetzner/<SERVER-ID>/{root,deploy,deploy.pub}
```

These keys should match the ones you already added to the server via `ssh-copy-id`.

---

## Step 1: Get Your Server IP

Your Hetzner server IP is shown in the Hetzner console or in your provisioning email.

Example: `123.45.67.89`

---

## Step 2: Copy the Root SSH Private Key

This is the private key GitHub Actions uses to bootstrap the server (SSH as root).

**Location:** `~/.ssh/hetzner/<SERVER-ID>/root`

**How to get it:**

```bash
# Display the key (copy entire output)
cat ~/.ssh/hetzner/<SERVER-ID>/root
```

**Expected format:**
```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUtY25vbm
... (many lines) ...
-----END OPENSSH PRIVATE KEY-----
```

**⚠️ Important:** This is your private key. Handle it securely. GitHub keeps it encrypted.

---

## Step 3: Copy the Deploy SSH Public Key

This is the public key added to the deploy user on the server (developers use this for normal operations).

**Location:** `~/.ssh/hetzner/<SERVER-ID>/deploy.pub`

**How to get it:**

```bash
# Display the key (copy entire output)
cat ~/.ssh/hetzner/<SERVER-ID>/deploy.pub
```

**Expected format:**
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIF... admin@allora.style
```

**⚠️ Note:** This is the `.pub` (public) file, not the private key. Public keys are safe to share.

---

## Step 4: Decide on Your Production Domain

Choose your production domain. Examples:
- `allora.style`
- `discover.allora.style`
- `api.allora.style`

**⚠️ Important Format:** Domain name only, **no `https://` or `http://` prefix**.

**Examples:**
- ✅ `allora.style`
- ❌ `https://allora.style` (wrong)
- ✅ `my-domain.com`
- ❌ `http://my-domain.com` (wrong)

---

## Step 5: Add Secrets to GitHub

### Via GitHub Web UI (Easiest)

1. Go to your repository on GitHub
2. Click **Settings** (top right)
3. Click **Secrets and variables** → **Actions** (left sidebar)
4. Click **New repository secret** (green button)
5. Add each secret below (one at a time)

### Secret #1: PRODUCTION_SERVER_IP

- **Name:** `PRODUCTION_SERVER_IP`
- **Value:** Your server IP (e.g., `123.45.67.89`)
- Click **Add secret**

### Secret #2: PRODUCTION_SSH_KEY

- **Name:** `PRODUCTION_SSH_KEY`
- **Value:** Contents of `~/.ssh/hetzner/<SERVER-ID>/root` (entire private key)
- Click **Add secret**

### Secret #3: DEPLOY_SSH_PUBLIC_KEY

- **Name:** `DEPLOY_SSH_PUBLIC_KEY`
- **Value:** Contents of `~/.ssh/hetzner/<SERVER-ID>/deploy.pub`
- Click **Add secret**

### Secret #4: PRODUCTION_DOMAIN

- **Name:** `PRODUCTION_DOMAIN`
- **Value:** Your domain name (e.g., `allora.style`, no `https://`)
- Click **Add secret**

---

## Step 6: Verify Secrets Were Added

In GitHub, go back to **Settings → Secrets and variables → Actions**.

You should see all 4 secrets listed:
- ✅ `PRODUCTION_SERVER_IP`
- ✅ `PRODUCTION_SSH_KEY`
- ✅ `DEPLOY_SSH_PUBLIC_KEY`
- ✅ `PRODUCTION_DOMAIN`

Each shows a timestamp of when it was added. The actual values are hidden (GitHub doesn't show them).

---

## Step 7: Test the Workflow

Now that secrets are configured, the GitHub Actions workflows should work.

### Test Initialize Server Workflow

1. Go to GitHub → **Actions**
2. Click **Initialize Server** (left sidebar)
3. Click **Run workflow** (blue button)
4. A dropdown appears with no inputs (because we removed them!)
5. Click **Run workflow** (green button at bottom)
6. Watch the logs (takes 30-45 minutes)

**Expected:** The workflow should successfully SSH to your server and run bootstrap.

### Test Deploy Workflow (After Bootstrap)

1. Go to GitHub → **Actions**
2. Click **Deploy to Production**
3. Click **Run workflow**
4. Click **Run workflow** (green button)
5. Watch the logs (takes 5-10 minutes)

**Expected:** The workflow should pull Docker images and deploy to production.

---

## Troubleshooting

### Workflow fails: "SSH key verification failed"

**Cause:** The `PRODUCTION_SSH_KEY` secret is invalid or corrupted.

**Fix:**
1. Verify the key is copied **in full** (from `-----BEGIN` to `-----END`)
2. Check that newlines are preserved
3. Regenerate the key if needed:
   ```bash
   rm ~/.ssh/hetzner/<SERVER-ID>/root*
   ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/root
   ssh-copy-id -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP>
   ```
4. Update the secret in GitHub with the new key

### Workflow fails: "Host key verification failed"

**Cause:** The server IP is unreachable or incorrect.

**Fix:**
1. Verify `PRODUCTION_SERVER_IP` is correct
2. Test connectivity: `ping <SERVER_IP>`
3. Verify the server is running (check Hetzner console)

### Workflow fails: "Permission denied (publickey)"

**Cause:** The root SSH key is not authorized on the server.

**Fix:**
1. SSH manually to verify:
   ```bash
   ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> "echo OK"
   ```
2. If that fails, re-run `ssh-copy-id`:
   ```bash
   ssh-copy-id -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP>
   ```

### The secret shows "Failed to decrypt" in workflow logs

**Cause:** GitHub couldn't access the secret (rare infrastructure issue).

**Fix:**
1. Rotate the secret (delete and re-add it)
2. Contact GitHub support if it persists

---

## Reference

- **GitHub Secrets Docs:** https://docs.github.com/en/actions/security-guides/encrypted-secrets
- **SSH Key Setup:** See `infra/DEPLOY.md` "CRITICAL PREREQUISITE" section
- **Workflows:** `.github/workflows/initialize-server.yml`, `.github/workflows/deploy.yml`

---

## Next Steps

After secrets are configured:

1. ✅ Run **Initialize Server** workflow (one-time, 30-45 min)
2. ✅ Verify bootstrap completion (5 min)
3. ✅ Configure `.env` on server (2 min)
4. ✅ Download models (5-10 min)
5. ✅ Start services manually (5 min)
6. ✅ Push code to main → **Deploy** workflow runs automatically

See `infra/DEPLOY.md` for complete deployment guide.
