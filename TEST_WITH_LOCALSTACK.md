# Testing Your KV Store

## ✅ LocalStack Status

LocalStack is running! However, I notice that **ECS, ECR, and Service Discovery** are not in the available services list. This is because:

- **Community Edition**: Has limited ECS/ECR support
- **Pro/Student Edition**: Full ECS/ECR/Service Discovery support

## Option 1: Test Your KV Store Directly (Recommended for Now)

You can test your leaderless KV store **without LocalStack** using Docker Compose:

```bash
# Build and start all 5 nodes
docker compose -f docker-compose-leaderless.yml up -d

# Check if all nodes are running
docker compose -f docker-compose-leaderless.yml ps

# Test the service
curl http://localhost:8080/health

# Write a value
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"hello"}'

# Read the value
curl "http://localhost:8080/get?key=test"

# View logs
docker compose -f docker-compose-leaderless.yml logs -f

# Stop when done
docker compose -f docker-compose-leaderless.yml down
```

## Option 2: Test Basic LocalStack Services

Even without ECS/ECR, you can test other services:

```bash
# Set LocalStack endpoint
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2

# Test S3 (if enabled)
aws --endpoint-url=$AWS_ENDPOINT_URL s3 ls

# Test CloudWatch Logs (available)
aws --endpoint-url=$AWS_ENDPOINT_URL logs describe-log-groups

# Test EC2 (available)
aws --endpoint-url=$AWS_ENDPOINT_URL ec2 describe-instances
```

## Option 3: Get LocalStack Pro/Student Edition

For full ECS/ECR support:

1. **Sign up for Student Plan** (free with GitHub Student Pack):
   - Visit: https://www.localstack.cloud/localstack-for-students
   - Link your GitHub account
   - Install: `pip install localstack`
   - Login: `localstack login`
   - Start: `localstack start`

2. **Or use Community Edition with workarounds**:
   - Test your application directly with Docker Compose
   - Use LocalStack for other services (S3, Lambda, etc.)

## Recommended Next Steps

1. **Test your KV store now** with `docker-compose-leaderless.yml` (no LocalStack needed)
2. **If you need ECS/ECR testing**, get LocalStack Student Plan
3. **For now**, LocalStack is great for testing other AWS services

## Quick Test Commands

```bash
# Test your leaderless KV store (5 nodes)
docker compose -f docker-compose-leaderless.yml up -d

# Wait a few seconds, then test
sleep 5
curl http://localhost:8080/health

# Write and read
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"hello-world"}'

curl "http://localhost:8080/get?key=test"
```



