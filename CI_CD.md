# CI/CD Pipeline Documentation

## Overview

The Payment System uses **GitHub Actions** for continuous integration and continuous deployment (CI/CD). The pipeline automatically runs tests, builds Docker images, checks code quality, and scans for security vulnerabilities on every push and pull request.

```
Push to GitHub
    ↓
Tests (Unit, Race Detection, Coverage)
    ↓
Linting & Code Quality
    ↓
Security Scanning
    ↓
Build Docker Images
    ↓
Deploy (on main branch only)
    ↓
Create Release
```

---

## Workflows

### 1. CI/CD Pipeline (`ci-cd.yml`)

Main pipeline that runs on every push and pull request.

**Triggers**:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches
- Changes in service directories or workflow files

**Jobs**:

#### Test Job
- Runs for all 3 services in parallel
- Tests: `go test -v -race -coverprofile=coverage.out ./...`
- Uploads coverage to Codecov
- **Failure**: Blocks subsequent jobs

#### Lint Job
- Runs `golangci-lint` on each service
- Checks code quality and style
- **Failure**: Blocks merge and deployment

#### Security Job
- Scans dependencies for vulnerabilities
- Uses Trivy security scanner
- Uploads SARIF report to GitHub Security tab

#### Integration Test Job
- Spins up PostgreSQL and Kafka
- Runs integration tests across all services
- Verifies service-to-service communication

#### Build Job
- Builds Docker images for all services
- Pushes to GitHub Container Registry (ghcr.io)
- Only runs on push events (not PRs)
- Tags: `latest` and `<commit-sha>`

#### Deploy Job
- Only runs on `main` branch after all checks pass
- Creates GitHub release with deployment info
- Provides instructions for deployment

### 2. Manual Deploy Workflow (`deploy.yml`)

Manual deployment workflow for on-demand deployments.

**Triggers**:
- Manual trigger via GitHub Actions UI (`workflow_dispatch`)
- Choose environment: `staging` or `production`

**Steps**:
1. Build Docker images
2. Push to container registry
3. Provide deployment instructions

**Usage**:
```
GitHub → Actions → Deploy to Docker Hub → Run workflow → Select environment
```

### 3. Health Check Workflow (`health-check.yml`)

Periodic health monitoring of the codebase.

**Triggers**:
- Scheduled: Every 6 hours (cron: `0 */6 * * *`)
- Manual trigger available

**Checks**:
- Dependency verification (`go mod verify`)
- Security vulnerability scanning (`govulncheck`)
- Creates health status artifact
- Retained for 30 days

### 4. Code Quality Workflow (`code-quality.yml`)

Detailed code quality and security analysis.

**Triggers**:
- Pull requests to `main` or `develop`
- Push to `main` or `develop`

**Checks**:
- Code formatting (`gofmt`)
- Import checking
- `go vet` analysis
- Race condition detection
- Code coverage (minimum 60%)
- Secret scanning (TruffleHog)
- Static analysis (Gosec)

**Failure Threshold**:
- Code coverage < 60%: ❌ Blocks merge
- Formatting issues: ❌ Blocks merge
- Vet errors: ❌ Blocks merge

---

## Status Badges

Add these to your README.md for visibility:

```markdown
![CI/CD](https://github.com/<owner>/<repo>/workflows/CI%2FCD%20Pipeline/badge.svg?branch=main)
![Code Quality](https://github.com/<owner>/<repo>/workflows/Code%20Quality/badge.svg)
![Health Check](https://github.com/<owner>/<repo>/workflows/Health%20Check/badge.svg)
```

---

## Docker Images

### Registry

Images are pushed to GitHub Container Registry:
```
ghcr.io/<username>/payment-system-<service>:latest
ghcr.io/<username>/payment-system-<service>:<sha>
```

### Pulling Images

```bash
# Pull latest version
docker pull ghcr.io/<username>/payment-system-user-service:latest

# Pull specific version
docker pull ghcr.io/<username>/payment-system-user-service:<commit-sha>

# Log in if using private registry
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
```

### Updating docker-compose.yml

After images are built and pushed:

```yaml
services:
  user-service:
    image: ghcr.io/<username>/payment-system-user-service:latest
    
  payment-service:
    image: ghcr.io/<username>/payment-system-payment-service:latest
    
  notification-service:
    image: ghcr.io/<username>/payment-system-notification-service:latest
```

Then deploy:
```bash
docker-compose up -d
```

---

## GitHub Secrets & Environments

### Required Secrets

No additional secrets are required! The pipeline uses `GITHUB_TOKEN` which is automatically provided.

### Optional: Docker Hub Push

To push to Docker Hub instead of GitHub Container Registry:

1. Create Docker Hub account
2. Create access token: Docker Hub → Account Settings → Security → New Access Token
3. Add to GitHub secrets:
   - Settings → Secrets and variables → Actions → New repository secret
   - Name: `DOCKERHUB_USERNAME`
   - Value: Your Docker Hub username
   - Name: `DOCKERHUB_TOKEN`
   - Value: Your Docker Hub access token

Then modify workflow to:
```yaml
- name: Log in to Docker Hub
  uses: docker/login-action@v2
  with:
    username: ${{ secrets.DOCKERHUB_USERNAME }}
    password: ${{ secrets.DOCKERHUB_TOKEN }}
```

### Environment-specific Configuration

GitHub allows environment-specific secrets:

1. Settings → Environments → New environment
2. Create `staging` and `production` environments
3. Add environment-specific secrets (API keys, database URLs, etc.)

```yaml
jobs:
  deploy:
    environment: ${{ inputs.environment }}  # Uses env secrets
```

---

## Viewing Workflow Results

### GitHub UI

1. Push code to GitHub
2. Go to **Actions** tab
3. See workflow run in progress
4. Click to see detailed logs for each job

### Status Checks on Pull Request

Workflows run automatically on PR and block merge if failing:
- ✅ All checks pass → Merge enabled
- ❌ Any check fails → Merge blocked

### Downloading Artifacts

```bash
# Health check report
GitHub → Actions → Health Check (latest run) → Download artifacts

# Coverage reports
GitHub → Actions → CI/CD Pipeline → Download coverage artifact
```

---

## Best Practices

### 1. Commit Messages

Use conventional commits to link to jobs:

```
feat: add new API endpoint
ci: update workflow timeout

# Prefix types:
# feat: new feature
# fix: bug fix
# ci: CI/CD changes
# docs: documentation
# test: test additions
# chore: maintenance
```

### 2. PR Requirements

Enable branch protection on `main`:
- Settings → Branches → Add rule → `main`
- ✅ Require status checks to pass
- ✅ Require code reviews
- ✅ Dismiss stale PR approvals
- ✅ Require branches to be up to date

### 3. Test Before Pushing

Run tests locally before pushing:

```bash
# User Service
cd user-service
go test -race -v ./...
golangci-lint run

# Payment Service
cd payment-service
go test -race -v ./...
golangci-lint run

# Notification Service
cd notification-service
go test -race -v ./...
golangci-lint run
```

### 4. Coverage Targets

Current minimum: **60%**

To check locally:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Troubleshooting

### Build Fails: "go mod download"

**Cause**: Dependency download timeout

**Solution**: 
- Check internet connection
- Increase timeout in workflow
- Run locally: `go mod download && go mod verify`

### Lint Fails: Formatting

**Cause**: Code not formatted

**Solution**:
```bash
gofmt -s -w ./...
go vet ./...
golangci-lint run --fix
```

### Test Fails: Race Condition

**Cause**: Concurrent access issue

**Solution**:
- Use sync.Mutex for shared data
- Avoid global variables
- Use channels for communication

### Docker Build Fails: Layer Size

**Cause**: Image too large

**Solution**:
- Use multi-stage builds
- Move to Alpine base images
- Remove build artifacts

### Coverage Below Threshold

**Cause**: Tests don't cover all code paths

**Solution**:
```bash
# See which lines aren't covered
go tool cover -html=coverage.out

# Add tests for uncovered code
# Then update threshold if reasonable
```

### Images Not Pushing to Registry

**Cause**: Authentication issues

**Solution**:
```bash
# Check GITHUB_TOKEN permissions
GitHub → Settings → Developer settings → Personal access tokens
# Ensure "write:packages" scope is selected

# Or use:
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin
```

---

## Monitoring & Alerts

### Email Notifications

GitHub automatically emails on workflow failure:
- Workflow fails → Email sent
- Workflow passes after failure → Email sent

### Slack Integration

Add Slack notifications (optional):

1. Create Slack app and webhook
2. Add to GitHub secrets: `SLACK_WEBHOOK`
3. Add to workflow:

```yaml
- name: Notify Slack
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    webhook-url: ${{ secrets.SLACK_WEBHOOK }}
```

### GitHub Issues on Failure

Automatically create issues on failure:

```yaml
- name: Create issue on failure
  if: failure()
  uses: actions/github-script@v6
  with:
    script: |
      github.rest.issues.create({
        owner: context.repo.owner,
        repo: context.repo.repo,
        title: `CI/CD failed: ${context.payload.head_commit.message}`,
        body: `Workflow: ${context.workflow}\nRun: ${context.runId}`
      })
```

---

## Performance Tips

### Caching

Cache Go dependencies to speed up builds:

```yaml
- uses: actions/setup-go@v4
  with:
    go-version: '1.21'
    cache: true  # Automatically cache go mod files
```

### Parallel Matrix

Runs all service tests in parallel (current setup):

```yaml
strategy:
  matrix:
    service: [user-service, payment-service, notification-service]
```

Saves ~60% time vs sequential runs.

### Docker Buildx Cache

Uses GitHub Actions cache for Docker builds:

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

---

## Security Practices

### Secrets Management

**Never commit secrets!**

✅ **Correct**:
```yaml
PASSWORD: ${{ secrets.DATABASE_PASSWORD }}
```

❌ **Incorrect**:
```yaml
PASSWORD: actual_password_here
```

### Dependency Scanning

Runs automatically:
- Security vulnerabilities (`gosec`)
- Supply chain attacks (`TruffleHog`)
- Dependency advisories (`govulncheck`)

### SARIF Reports

Security results uploaded to GitHub Security tab:
- Repository → Security → Code scanning alerts
- See all vulnerabilities in one place

---

## Scheduled Maintenance

### Health Check (Every 6 hours)

```
00:00 → Health check runs
06:00 → Health check runs
12:00 → Health check runs
18:00 → Health check runs
```

### Dependency Updates

Consider adding `dependabot.yml` to auto-update dependencies:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/user-service"
    schedule:
      interval: "weekly"
  - package-ecosystem: "docker"
    directory: "/user-service"
    schedule:
      interval: "weekly"
```

---

## Example: Full Deployment Flow

### On Main Branch Push

```
1. Tests run (user, payment, notification services) ✅
2. Linting checks ✅
3. Security scanning ✅
4. Docker images built for all services ✅
5. Images pushed to ghcr.io with latest tag ✅
6. GitHub release created with deployment info ✅
7. Email notification sent ✅
```

### Manual Staging Deployment

```
1. Go to Actions tab
2. Select "Deploy to Docker Hub"
3. Click "Run workflow"
4. Choose "staging" environment
5. Watch build and push in real-time
6. Pull staging images: docker pull ghcr.io/.../staging
```

### Manual Production Deployment

```
1. Ensure main branch is fully tested
2. Go to Actions → Deploy to Docker Hub
3. Choose "production" environment
4. Images pushed with production tag
5. Update docker-compose.yml with new image tag
6. Run: docker-compose up -d
7. Verify: curl http://localhost:8080/health
```

---

## Next Steps

1. ✅ Push code to GitHub (main branch)
2. ✅ Watch workflows run in Actions tab
3. ✅ Verify all checks pass
4. ✅ Pull Docker images
5. ✅ Update docker-compose.yml
6. ✅ Deploy with `docker-compose up -d`
7. ✅ Monitor health checks

---

## Related Documentation

- [Main README](./README.md) - Project overview
- [API Documentation](./API.md) - API endpoints
- [Docker Setup](./docker-compose.yml) - Container orchestration
