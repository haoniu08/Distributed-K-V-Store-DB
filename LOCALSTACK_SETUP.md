# LocalStack Setup Guide

## What is LocalStack?

LocalStack is a local AWS emulator that runs AWS services on your laptop. It's perfect for:
- Testing AWS applications locally
- Developing without AWS account/credit card
- Avoiding cloud costs during development
- Faster development cycles

**Student Plan**: Free access with GitHub Student Developer Pack - [LocalStack for Students](https://www.localstack.cloud/localstack-for-students)

## Setup Options

### Option 1: LocalStack Community Edition (Free, Open Source)

**Install via Docker:**
```bash
# Pull LocalStack image
docker pull localstack/localstack

# Run LocalStack
docker run -d \
  --name localstack \
  -p 4566:4566 \
  -p 4510-4559:4510-4559 \
  -e SERVICES=ecs,ecr,cloudwatch,logs,servicediscovery,elbv2,ec2 \
  -e DEBUG=1 \
  -e DATA_DIR=/tmp/localstack/data \
  localstack/localstack
```

**Or use Docker Compose:**
```yaml
version: '3.8'
services:
  localstack:
    container_name: localstack
    image: localstack/localstack
    ports:
      - "4566:4566"
      - "4510-4559:4510-4559"
    environment:
      - SERVICES=ecs,ecr,cloudwatch,logs,servicediscovery,elbv2,ec2
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
    volumes:
      - "./localstack-data:/var/lib/localstack"
      - "/var/run/docker.sock:/var/run/docker.sock"
```

### Option 2: LocalStack Pro (Student Plan - Free with GitHub Student Pack)

1. **Verify GitHub Student Developer Pack**:
   - Go to GitHub Settings → Billing & Licensing → Education Benefits
   - Ensure you have active education benefits

2. **Sign up for LocalStack**:
   - Visit: https://www.localstack.cloud/localstack-for-students
   - Click "Link your GitHub"
   - Follow the signup process

3. **Install LocalStack CLI**:
   ```bash
   pip install localstack
   ```

4. **Authenticate**:
   ```bash
   localstack login
   ```

5. **Start LocalStack**:
   ```bash
   localstack start
   ```

## Configure AWS CLI for LocalStack

```bash
# Configure AWS CLI to use LocalStack endpoint
aws configure set aws_access_key_id test
aws configure set aws_secret_access_key test
aws configure set default.region us-west-2

# Or use environment variables
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export AWS_ENDPOINT_URL=http://localhost:4566
```

## Test LocalStack

```bash
# Test S3 (example)
aws --endpoint-url=http://localhost:4566 s3 ls

# Test ECS
aws --endpoint-url=http://localhost:4566 ecs list-clusters

# Test CloudWatch
aws --endpoint-url=http://localhost:4566 logs describe-log-groups
```

## Using LocalStack with Terraform

### Option 1: Environment Variables

```bash
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export TF_VAR_aws_endpoint=http://localhost:4566
```

### Option 2: Terraform Provider Configuration

Update `provider.tf`:
```hcl
provider "aws" {
  region                      = var.aws_region
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  endpoints {
    ecs                = "http://localhost:4566"
    ecr                = "http://localhost:4566"
    cloudwatch         = "http://localhost:4566"
    logs               = "http://localhost:4566"
    servicediscovery   = "http://localhost:4566"
    elbv2              = "http://localhost:4566"
    ec2                = "http://localhost:4566"
  }
}
```

## Services Available in LocalStack

### Community Edition (Free):
- S3, Lambda, DynamoDB, API Gateway, SQS, SNS, IAM, CloudFormation
- **Limited**: ECS, ECR (basic support)

### Pro/Student Edition:
- Full ECS, ECR support
- Cloud Map (Service Discovery)
- Application Load Balancer
- CloudWatch Logs
- And many more...

## For Your KV Store Project

LocalStack can help you test:
1. **ECS**: Deploy your leaderless nodes
2. **ECR**: Store Docker images
3. **Cloud Map**: Service discovery (if using Pro/Student)
4. **ALB**: Load balancing
5. **CloudWatch**: Logging

## Next Steps

1. **Install LocalStack** (choose Community or Student plan)
2. **Start LocalStack**
3. **Configure AWS CLI** to use LocalStack endpoint
4. **Test basic services** to verify it's working
5. **Adapt Terraform** to use LocalStack endpoints
6. **Deploy your KV store** locally

## Resources

- [LocalStack Documentation](https://docs.localstack.cloud/)
- [LocalStack for Students](https://www.localstack.cloud/localstack-for-students)
- [LocalStack GitHub](https://github.com/localstack/localstack)



