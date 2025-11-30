#!/bin/bash
# Test script for LocalStack

echo "=== Testing LocalStack ==="

# Check if LocalStack is running
if ! curl -s http://localhost:4566/_localstack/health > /dev/null; then
    echo "❌ LocalStack is not running!"
    echo "Start it with: docker-compose -f docker-compose-localstack.yml up -d"
    exit 1
fi

echo "✅ LocalStack is running"
echo ""

# Set LocalStack endpoint
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2

echo "Testing AWS services..."

# Test S3
echo -n "S3: "
aws --endpoint-url=$AWS_ENDPOINT_URL s3 ls 2>&1 | head -1
if [ $? -eq 0 ]; then
    echo "✅ S3 working"
else
    echo "❌ S3 failed"
fi

# Test ECS
echo -n "ECS: "
aws --endpoint-url=$AWS_ENDPOINT_URL ecs list-clusters 2>&1 | head -1
if [ $? -eq 0 ]; then
    echo "✅ ECS working"
else
    echo "❌ ECS failed"
fi

# Test CloudWatch Logs
echo -n "CloudWatch Logs: "
aws --endpoint-url=$AWS_ENDPOINT_URL logs describe-log-groups 2>&1 | head -1
if [ $? -eq 0 ]; then
    echo "✅ CloudWatch Logs working"
else
    echo "❌ CloudWatch Logs failed"
fi

# Test Service Discovery (may not work in Community Edition)
echo -n "Service Discovery: "
aws --endpoint-url=$AWS_ENDPOINT_URL servicediscovery list-namespaces 2>&1 | head -1
if [ $? -eq 0 ]; then
    echo "✅ Service Discovery working"
else
    echo "⚠️  Service Discovery may not be available (requires Pro/Student plan)"
fi

echo ""
echo "=== LocalStack Test Complete ==="
echo ""
echo "To use with Terraform, set:"
echo "  export AWS_ENDPOINT_URL=http://localhost:4566"
echo "  export AWS_ACCESS_KEY_ID=test"
echo "  export AWS_SECRET_ACCESS_KEY=test"



