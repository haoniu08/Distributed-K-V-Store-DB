# LocalStack Quick Start

## Issue: `docker-compose` command not found

This usually means one of:
1. Docker Compose V2 is installed (use `docker compose` instead of `docker-compose`)
2. Docker Compose is not installed
3. Docker Desktop is not running

## Solution Options

### Option 1: Use Docker Compose V2 (Most Likely)

Modern Docker installations use `docker compose` (with a space) instead of `docker-compose` (with a hyphen).

```bash
# Navigate to project directory
cd /Users/haoniu/Desktop/Distributed-K-V-Store-DB

# Start LocalStack
docker compose -f docker-compose-localstack.yml up -d

# Check status
docker compose -f docker-compose-localstack.yml ps

# View logs
docker compose -f docker-compose-localstack.yml logs -f

# Stop LocalStack
docker compose -f docker-compose-localstack.yml down
```

### Option 2: Install Docker Compose V1

If you prefer the old `docker-compose` command:

**macOS (using Homebrew):**
```bash
brew install docker-compose
```

**Or download directly:**
```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Option 3: Use Docker Run Directly (No Compose)

If you don't want to use Docker Compose at all:

```bash
# Start LocalStack
docker run -d \
  --name localstack \
  -p 4566:4566 \
  -p 4510-4559:4510-4559 \
  -e SERVICES=ecs,ecr,cloudwatch,logs,servicediscovery,elbv2,ec2,iam \
  -e DEBUG=1 \
  -e DATA_DIR=/tmp/localstack/data \
  -e DOCKER_HOST=unix:///var/run/docker.sock \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  localstack/localstack:latest

# Check if it's running
docker ps | grep localstack

# View logs
docker logs -f localstack

# Stop LocalStack
docker stop localstack
docker rm localstack
```

### Option 4: Install Docker Desktop (If Docker Not Installed)

If Docker itself is not installed:

1. **Download Docker Desktop for Mac**: https://www.docker.com/products/docker-desktop/
2. Install and start Docker Desktop
3. Wait for Docker to start (whale icon in menu bar)
4. Then use `docker compose` command

## Verify Docker is Working

```bash
# Check Docker version
docker --version

# Check if Docker daemon is running
docker ps

# If you get "Cannot connect to Docker daemon", start Docker Desktop
```

## Start LocalStack

Once Docker is working, try:

```bash
cd /Users/haoniu/Desktop/Distributed-K-V-Store-DB

# Try V2 first (most common)
docker compose -f docker-compose-localstack.yml up -d

# If that doesn't work, try V1
docker-compose -f docker-compose-localstack.yml up -d

# Or use direct docker run (see Option 3 above)
```

## Test LocalStack

After starting, wait 10-30 seconds, then:

```bash
# Check health
curl http://localhost:4566/_localstack/health

# Or run the test script
./test-localstack.sh
```

## Troubleshooting

### "Cannot connect to Docker daemon"
- Start Docker Desktop
- Wait for it to fully start (check menu bar icon)

### "Port already in use"
- Stop any existing LocalStack: `docker stop localstack`
- Or change the port in docker-compose file

### "Permission denied"
- Make sure Docker Desktop is running
- You may need to add your user to docker group (usually not needed on macOS)

## Next Steps

Once LocalStack is running:
1. Configure AWS CLI to use LocalStack endpoint
2. Test your KV store with LocalStack
3. Or test directly with Docker Compose (you already have `docker-compose-leaderless.yml`)



