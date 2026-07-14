# GitHub Secrets Setup for Styx Registry Deployment

Configure automated deployments to your Hetzner cx23 server.

---

## Required Secrets

Add these 4 secrets to your GitHub repository:

| Secret | Value | From |
|--------|-------|------|
| `STYX_DEPLOY_HOST` | Hetzner server IP | Your server IP (e.g., 1.2.3.4) |
| `STYX_DEPLOY_USER` | SSH user | `styx` (set by provisioning script) |
| `STYX_DEPLOY_KEY` | SSH private key | `~/.ssh/id_rsa` (or your deployment key) |
| `STYX_REGISTRY_DOMAIN` | Domain | Your registry domain (e.g., registry.styx.sh) |

---

## Step 1: Add Secrets to GitHub

1. Go to **GitHub Repository → Settings → Secrets and variables → Actions**
2. Click **New repository secret**
3. Add each secret below

---

## Step 2: STYX_DEPLOY_HOST

```
Secret name: STYX_DEPLOY_HOST
Value: <your-hetzner-ip>
```

Get your IP from Hetzner console.

---

## Step 3: STYX_DEPLOY_USER

```
Secret name: STYX_DEPLOY_USER
Value: styx
```

This is the username created during provisioning.

---

## Step 4: STYX_DEPLOY_KEY

This is your SSH private key for deployment.

```bash
# Copy your private key to clipboard
cat ~/.ssh/id_rsa | pbcopy   # macOS
cat ~/.ssh/id_rsa | xclip -i # Linux
```

Then:
```
Secret name: STYX_DEPLOY_KEY
Value: <paste-your-private-key>
```

**Format:**
```
-----BEGIN OPENSSH PRIVATE KEY-----
MIIFDjBABgkqhkiG9w0BBQcwDQYIKoZIhvcNAQEFBQADggEPA...
[... rest of key ...]
-----END OPENSSH PRIVATE KEY-----
```

---

## Step 5: STYX_REGISTRY_DOMAIN

```
Secret name: STYX_REGISTRY_DOMAIN
Value: registry.styx.sh
```

Replace with your actual domain.

---

## Verify Secrets

Go to **Settings → Secrets and variables → Actions**

You should see all 4 secrets listed:
- ✅ STYX_DEPLOY_HOST
- ✅ STYX_DEPLOY_USER
- ✅ STYX_DEPLOY_KEY
- ✅ STYX_REGISTRY_DOMAIN

---

## Test Deployment

Push a commit to main to trigger the workflow:

```bash
echo "test" >> .trigger
git add .trigger
git commit -m "test: trigger deployment"
git push origin main
```

Monitor at: **GitHub → Actions → deploy-registry**

---

## Troubleshooting

### "Permission denied" on deploy

Check SSH key permissions:
```bash
chmod 600 ~/.ssh/id_rsa
chmod 644 ~/.ssh/id_rsa.pub
```

Verify the key is in `authorized_keys` on the server:
```bash
ssh styx@<server_ip> "cat ~/.ssh/authorized_keys | grep $(cat ~/.ssh/id_rsa.pub)"
```

### Workflow still failing

Verify secrets are correct:
```bash
# Test SSH manually
ssh -i ~/.ssh/id_rsa styx@<server_ip> "echo OK"
```

If that works but GitHub Actions fails, double-check the secret values match exactly (no extra spaces).

### "Host key verification failed"

The workflow needs to accept the Hetzner server's host key on first connection.

This is handled automatically in the GitHub Actions workflow via `StrictHostKeyChecking=no`.

If you see this error, check the workflow file:
```yaml
ssh -o StrictHostKeyChecking=no -i /tmp/deploy_key styx@${{ secrets.STYX_DEPLOY_HOST }}
```

---

## Security Notes

- ✅ Never commit secrets to the repository
- ✅ Rotate SSH keys periodically
- ✅ Use separate keys for different services
- ✅ Restrict GitHub secret access to necessary workflows
- ✅ Monitor GitHub Actions logs for failed deployments

---

## Next Steps

1. [Deploy to Hetzner](DEPLOY.md)
2. [Local development testing](LOCAL_DEVELOPMENT.md)
3. Push code and watch automatic deployments at registry.styx.sh
