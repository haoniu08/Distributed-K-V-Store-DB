# AWS Deployment Requirements for Vector Clock Leaderless Model

## Overview

To deploy the 5-node leaderless KV store with vector clocks on AWS, you need the following AWS services and Terraform modules.

## Required AWS Services

### 1. **Compute: ECS Fargate** ⭐ Core Service
**Purpose**: Run 5 containerized leaderless nodes

**Requirements**:
- ECS Cluster (1 cluster)
- ECS Service (1 service with 5 tasks, one per node)
- Task Definition (container configuration)
- Fargate launch type (serverless, no EC2 management)

**Configuration**:
- 5 tasks (one for each node: node1-node5)
- Each task runs the leaderless container
- CPU: 0.25 vCPU per task (256 CPU units)
- Memory: 512 MB per task
- Health checks on `/health` endpoint

**Terraform Module**: `modules/ecs/`

---

### 2. **Container Registry: ECR** ⭐ Core Service
**Purpose**: Store Docker images for leaderless nodes

**Requirements**:
- ECR Repository (1 repository)
- Image push/pull permissions
- Lifecycle policies (optional, for cleanup)

**Configuration**:
- Repository name: `distributed-kv-store-leaderless`
- Image tags: `latest`, `vector-clock-v1`
- Cross-region replication (optional)

**Terraform Module**: `modules/ecr/` (or inline resource)

---

### 3. **Service Discovery: Cloud Map** ⭐ Critical for Leaderless
**Purpose**: Allow nodes to discover each other dynamically

**Requirements**:
- Cloud Map Namespace (1 private DNS namespace)
- Cloud Map Service (1 service for leaderless nodes)
- Service instances (5 instances, one per node)

**Configuration**:
- Namespace: `kv-store.local` (or custom domain)
- Service name: `leaderless-nodes`
- Service instances:
  - `node1.kv-store.local:8080`
  - `node2.kv-store.local:8080`
  - `node3.kv-store.local:8080`
  - `node4.kv-store.local:8080`
  - `node5.kv-store.local:8080`
- Health checks: HTTP on port 8080

**Why Critical**: 
- Nodes need to know all other node addresses
- Dynamic IPs require service discovery
- Enables `--all-node-addrs` configuration

**Terraform Module**: `modules/service_discovery/`

---

### 4. **Networking: VPC & Networking** ⭐ Core Infrastructure
**Purpose**: Network isolation and connectivity

**Requirements**:
- VPC (1 VPC)
- Subnets (2+ subnets across AZs for high availability)
- Internet Gateway (for ALB, optional)
- NAT Gateway (for ECS tasks to pull images, optional if using VPC endpoints)
- Security Groups (2 groups: ALB, ECS tasks)

**Configuration**:
- VPC CIDR: `10.0.0.0/16`
- Public subnets: `10.0.1.0/24`, `10.0.2.0/24` (for ALB)
- Private subnets: `10.0.10.0/24`, `10.0.11.0/24` (for ECS tasks)
- Security Group Rules:
  - ALB: Allow inbound 80/443 from internet
  - ECS Tasks: Allow inbound 8080 from ALB, allow all from same security group (inter-node communication)

**Terraform Module**: `modules/network/`

---

### 5. **Load Balancing: ALB (Application Load Balancer)** ⭐ Recommended
**Purpose**: Distribute client requests across all 5 nodes

**Requirements**:
- Application Load Balancer (1 ALB)
- Target Group (1 target group)
- Listener (HTTP/HTTPS on port 80/443)
- Health checks

**Configuration**:
- Scheme: Internet-facing (or internal)
- Type: Application Load Balancer
- Target group: All 5 ECS tasks
- Health check: `/health` endpoint
- Load balancing algorithm: Round-robin (or least connections)

**Why Recommended**:
- Clients can use single endpoint
- Automatic health checking
- Request distribution across all nodes
- SSL termination (optional)

**Terraform Module**: `modules/alb/` (or inline)

---

### 6. **Monitoring: CloudWatch** ⭐ Recommended
**Purpose**: Logs, metrics, and alarms

**Requirements**:
- CloudWatch Log Group (1 log group)
- CloudWatch Metrics (ECS task metrics)
- CloudWatch Alarms (optional, for alerts)

**Configuration**:
- Log group: `/ecs/leaderless-kv-store`
- Log retention: 7 days (or as needed)
- Metrics: CPU, memory, request count, error rate
- Alarms: High error rate, task failures

**Terraform Module**: `modules/cloudwatch/` (or inline)

---

### 7. **IAM: Roles and Policies** ⭐ Required
**Purpose**: Permissions for ECS tasks and services

**Requirements**:
- ECS Task Execution Role (for pulling images, writing logs)
- ECS Task Role (for application permissions, if needed)
- ECS Service Role (for service management)

**Configuration**:
- Task Execution Role:
  - `ecr:GetAuthorizationToken`
  - `ecr:BatchCheckLayerAvailability`
  - `ecr:GetDownloadUrlForLayer`
  - `ecr:BatchGetImage`
  - `logs:CreateLogStream`
  - `logs:PutLogEvents`
- Task Role: (minimal, add as needed)
- Service Role: (auto-created by ECS)

**Terraform Module**: `modules/iam/` (or inline)

---

## Terraform Module Structure

```
terraform/
├── main.tf                    # Main configuration
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── modules/
│   ├── network/              # VPC, subnets, security groups
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── ecr/                  # Container registry
│   │   ├── main.tf
│   │   └── outputs.tf
│   ├── ecs/                  # ECS cluster, service, tasks
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── service_discovery/    # Cloud Map
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── alb/                  # Application Load Balancer
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── cloudwatch/           # Logs and metrics
│       ├── main.tf
│       └── outputs.tf
└── README.md
```

---

## Key Configuration Details

### ECS Task Definition

```hcl
resource "aws_ecs_task_definition" "leaderless" {
  family                   = "leaderless-kv-store"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256    # 0.25 vCPU
  memory                   = 512    # 512 MB
  
  container_definitions = jsonencode([{
    name  = "leaderless-node"
    image = "${aws_ecr_repository.leaderless.repository_url}:latest"
    
    environment = [
      {
        name  = "NODE_ID"
        value = "node1"  # Will be different for each task
      },
      {
        name  = "ALL_NODE_ADDRS"
        value = "node1.kv-store.local:8080,node2.kv-store.local:8080,node3.kv-store.local:8080,node4.kv-store.local:8080,node5.kv-store.local:8080"
      },
      {
        name  = "PORT"
        value = "8080"
      }
    ]
    
    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]
    
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.leaderless.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
    
    healthCheck = {
      command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 60
    }
  }])
}
```

### Service Discovery Configuration

```hcl
resource "aws_service_discovery_service" "leaderless" {
  name = "leaderless-nodes"
  
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.kv_store.id
    
    dns_records {
      ttl  = 10
      type = "A"
    }
    
    routing_policy = "MULTIVALUE"
  }
  
  health_check_grace_period_seconds = 30
}
```

### Security Group Rules

```hcl
# Security group for ECS tasks
resource "aws_security_group_rule" "ecs_inter_node" {
  type                     = "ingress"
  from_port                = 8080
  to_port                  = 8080
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.ecs_tasks.id
  security_group_id        = aws_security_group.ecs_tasks.id
  description              = "Allow inter-node communication"
}

# Security group for ALB
resource "aws_security_group_rule" "alb_to_ecs" {
  type                     = "egress"
  from_port                = 8080
  to_port                  = 8080
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.alb.id
  security_group_id        = aws_security_group.ecs_tasks.id
  description              = "Allow ALB to reach ECS tasks"
}
```

---

## Deployment Steps

1. **Build and Push Docker Image**:
   ```bash
   docker build -f Dockerfile.leaderless -t leaderless-kv-store .
   aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin <account>.dkr.ecr.us-west-2.amazonaws.com
   docker tag leaderless-kv-store:latest <account>.dkr.ecr.us-west-2.amazonaws.com/leaderless-kv-store:latest
   docker push <account>.dkr.ecr.us-west-2.amazonaws.com/leaderless-kv-store:latest
   ```

2. **Initialize Terraform**:
   ```bash
   cd terraform
   terraform init
   ```

3. **Plan Deployment**:
   ```bash
   terraform plan -var="aws_region=us-west-2"
   ```

4. **Deploy**:
   ```bash
   terraform apply -var="aws_region=us-west-2"
   ```

5. **Verify**:
   ```bash
   # Get ALB endpoint
   terraform output alb_dns_name
   
   # Test health endpoint
   curl http://$(terraform output -raw alb_dns_name)/health
   ```

---

## Cost Estimation (Approximate)

| Service | Monthly Cost (5 nodes) |
|---------|------------------------|
| **ECS Fargate** | ~$15-30 (0.25 vCPU, 512MB per task) |
| **ECR** | ~$1-2 (storage + transfer) |
| **Cloud Map** | ~$0.50 (5 service instances) |
| **ALB** | ~$20-25 (base + LCU) |
| **VPC** | Free (NAT Gateway optional, ~$32 if used) |
| **CloudWatch** | ~$5-10 (logs + metrics) |
| **Data Transfer** | Variable (depends on usage) |
| **Total** | **~$40-70/month** (without NAT Gateway) |

---

## Optional Services

### 1. **Route 53** (Optional)
- Custom domain name
- DNS management
- Health checks

### 2. **ACM (Certificate Manager)** (Optional)
- SSL/TLS certificates for HTTPS
- Free certificates for ALB

### 3. **Auto Scaling** (Optional)
- Scale nodes based on CPU/memory
- Not typically needed for fixed 5-node cluster

### 4. **Secrets Manager** (Optional)
- Store sensitive configuration
- Not needed for current implementation

---

## Key Differences from Local Deployment

1. **Service Discovery**: Cloud Map instead of static IPs
2. **Dynamic IPs**: ECS tasks get dynamic IPs, need service discovery
3. **Network Isolation**: VPC with private subnets
4. **Load Balancing**: ALB for client access
5. **Container Registry**: ECR instead of local Docker
6. **Monitoring**: CloudWatch instead of local logs

---

## Critical Configuration Notes

1. **Node Discovery**: 
   - Use Cloud Map service names: `node1.kv-store.local:8080`
   - Configure `--all-node-addrs` with Cloud Map DNS names

2. **Inter-Node Communication**:
   - Security groups must allow traffic between ECS tasks
   - Use `self = true` in security group rules

3. **Health Checks**:
   - ECS service health checks
   - ALB target group health checks
   - Cloud Map health checks (optional)

4. **Vector Clock Node IDs**:
   - Use consistent node IDs: `node1`, `node2`, etc.
   - Pass via environment variables
   - Must match Cloud Map service instance names

---

## Summary

**Minimum Required Services**:
1. ✅ ECS Fargate (compute)
2. ✅ ECR (container registry)
3. ✅ Cloud Map (service discovery) - **Critical for leaderless**
4. ✅ VPC & Networking (infrastructure)
5. ✅ IAM (permissions)

**Recommended Services**:
6. ✅ ALB (load balancing)
7. ✅ CloudWatch (monitoring)

**Total**: 5-7 AWS services, organized into 5-6 Terraform modules.


