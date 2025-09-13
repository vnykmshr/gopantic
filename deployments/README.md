# 🚀 Gopantic Enterprise Deployment Guide

This directory contains enterprise-grade deployment templates and infrastructure-as-code configurations for deploying Gopantic-based applications at scale.

## 📁 Directory Structure

```
deployments/
├── docker/                 # Docker configurations
│   └── Dockerfile.production   # Production-ready Dockerfile
├── kubernetes/             # Kubernetes manifests
│   ├── gopantic-app.yaml      # Main application deployment
│   └── redis-cluster.yaml     # Redis caching cluster
├── helm/                   # Helm charts
│   └── gopantic-app/          # Complete Helm chart
│       ├── Chart.yaml         # Chart metadata
│       └── values.yaml        # Default values
├── terraform/              # Infrastructure as Code
│   └── aws-infrastructure.tf  # AWS infrastructure setup
└── README.md               # This file
```

## 🐳 Docker Deployment

### Production Docker Build

The production Dockerfile uses multi-stage builds for optimal security and size:

```bash
# Build the Docker image
docker build -f deployments/docker/Dockerfile.production -t your-app:latest .

# Run with environment variables
docker run -p 8080:8080 \
  -e GOPANTIC_STREAM_WORKERS=10 \
  -e GOPANTIC_CACHE_ENABLED=true \
  -e GOPANTIC_REDIS_ADDRESS=redis:6379 \
  your-app:latest
```

### Key Features:
- **Scratch-based**: Minimal attack surface (< 20MB final image)
- **Non-root user**: Runs as user 65534 for security
- **Health checks**: Built-in health monitoring
- **Multi-architecture**: Supports AMD64 and ARM64

## ☸️ Kubernetes Deployment

### Quick Start

1. **Deploy Redis cluster:**
```bash
kubectl apply -f deployments/kubernetes/redis-cluster.yaml
```

2. **Deploy the application:**
```bash
kubectl apply -f deployments/kubernetes/gopantic-app.yaml
```

3. **Verify deployment:**
```bash
kubectl get pods -l app=gopantic-app
kubectl get svc gopantic-app-service
```

### Production Considerations

The Kubernetes manifests include:
- **High Availability**: 3 replicas with anti-affinity rules
- **Security**: Pod Security Standards, NetworkPolicies, non-root containers
- **Monitoring**: Prometheus metrics integration
- **Resource Management**: CPU/memory requests and limits
- **Health Checks**: Liveness and readiness probes
- **Graceful Shutdown**: Proper termination handling

## 🎯 Helm Deployment

### Installation

1. **Add the Helm chart:**
```bash
helm install gopantic-app ./deployments/helm/gopantic-app \
  --set image.repository=your-registry/gopantic-app \
  --set image.tag=1.0.0 \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=your-app.example.com
```

2. **Customize configuration:**
```bash
# Copy and modify values
cp deployments/helm/gopantic-app/values.yaml my-values.yaml
# Edit my-values.yaml for your environment
helm upgrade gopantic-app ./deployments/helm/gopantic-app -f my-values.yaml
```

### Helm Features:
- **Configurable**: 50+ configuration options
- **Secure by Default**: Security best practices enabled
- **Monitoring Ready**: Prometheus integration included
- **Auto-scaling**: HPA support with CPU/memory metrics
- **Production Ready**: PodDisruptionBudget, resource limits

## 🏗️ Terraform Infrastructure (AWS)

### Prerequisites

1. **Install Terraform** (>= 1.0)
2. **Configure AWS credentials**
3. **Set up S3 bucket for state** (recommended)

### Deployment

```bash
cd deployments/terraform

# Initialize Terraform
terraform init

# Plan the infrastructure
terraform plan \
  -var="environment=prod" \
  -var="app_name=my-gopantic-app" \
  -var="aws_region=us-west-2"

# Apply the infrastructure
terraform apply
```

### Infrastructure Components:

#### Networking:
- **VPC**: Dedicated network with public/private subnets
- **NAT Gateway**: High-availability internet access for private subnets
- **Security Groups**: Least-privilege network access control

#### Compute:
- **Application Load Balancer**: High-availability traffic distribution
- **Target Groups**: Health-checked application instances
- **Auto Scaling**: Dynamic capacity based on demand

#### Caching:
- **ElastiCache Redis**: Multi-AZ Redis cluster with encryption
- **Parameter Groups**: Optimized configuration for caching workloads
- **Subnet Groups**: Secure private network placement

#### Outputs:
- Load Balancer DNS name for traffic routing
- VPC and subnet IDs for application deployment
- Redis endpoint for application configuration

## ⚙️ Configuration Guide

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GOPANTIC_STREAM_WORKERS` | Number of concurrent workers | `10` | No |
| `GOPANTIC_CACHE_ENABLED` | Enable caching layer | `true` | No |
| `GOPANTIC_CACHE_BACKEND` | Cache backend (memory/redis) | `memory` | No |
| `GOPANTIC_REDIS_ADDRESS` | Redis connection string | `localhost:6379` | If using Redis |
| `GOPANTIC_METRICS_ENABLED` | Enable metrics collection | `true` | No |
| `GOPANTIC_CIRCUIT_BREAKER_THRESHOLD` | Error rate threshold | `0.1` | No |
| `GOPANTIC_RETRY_ATTEMPTS` | Number of retry attempts | `3` | No |

### Performance Tuning

#### Small Workloads (< 1,000 items/min):
```bash
GOPANTIC_STREAM_WORKERS=2
GOPANTIC_CACHE_BACKEND=memory
```

#### Medium Workloads (1,000-50,000 items/min):
```bash
GOPANTIC_STREAM_WORKERS=8
GOPANTIC_CACHE_BACKEND=redis
GOPANTIC_CACHE_TTL=300s
```

#### Large Workloads (> 50,000 items/min):
```bash
GOPANTIC_STREAM_WORKERS=20
GOPANTIC_CACHE_BACKEND=redis
GOPANTIC_STREAM_BUFFER_SIZE=1000
GOPANTIC_BATCH_SIZE=100
```

## 📊 Monitoring and Observability

### Metrics Endpoints

- **Application Metrics**: `http://localhost:8081/metrics`
- **Health Check**: `http://localhost:8080/health`
- **Readiness Check**: `http://localhost:8080/ready`

### Key Metrics to Monitor:

#### Performance Metrics:
- `gopantic_stream_throughput` - Items processed per second
- `gopantic_processing_time_avg` - Average processing time
- `gopantic_worker_utilization` - Worker pool utilization

#### Error Metrics:
- `gopantic_error_rate` - Processing error rate
- `gopantic_circuit_breaker_open` - Circuit breaker status
- `gopantic_retry_count` - Number of retries

#### Cache Metrics:
- `gopantic_cache_hit_rate` - Cache hit ratio
- `gopantic_cache_memory_usage` - Memory consumption
- `gopantic_cache_operations_total` - Cache operations

### Prometheus Integration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'gopantic-app'
    static_configs:
      - targets: ['gopantic-app:8081']
    scrape_interval: 30s
    metrics_path: /metrics
```

### Grafana Dashboard

Key visualizations:
- **Throughput Graph**: Items/second over time
- **Error Rate Graph**: Error percentage with alerting
- **Latency Histogram**: Processing time distribution
- **Worker Utilization**: Resource usage metrics
- **Cache Performance**: Hit rates and memory usage

## 🔒 Security Best Practices

### Container Security:
- ✅ **Non-root execution** (user 65534)
- ✅ **Read-only root filesystem**
- ✅ **Minimal base image** (scratch)
- ✅ **No privilege escalation**
- ✅ **Dropped capabilities**

### Network Security:
- ✅ **NetworkPolicies** for traffic control
- ✅ **TLS encryption** for data in transit
- ✅ **Private subnets** for application layers
- ✅ **Security groups** with least privilege

### Data Security:
- ✅ **Encryption at rest** (Redis/storage)
- ✅ **Secret management** (Kubernetes secrets)
- ✅ **Input validation** (built-in Gopantic features)
- ✅ **Audit logging** for compliance

## 🚨 Troubleshooting

### Common Issues:

#### High Memory Usage:
```bash
# Check worker configuration
kubectl logs -l app=gopantic-app | grep "worker"

# Reduce workers if needed
kubectl set env deployment/gopantic-app GOPANTIC_STREAM_WORKERS=5
```

#### Cache Connection Issues:
```bash
# Verify Redis connectivity
kubectl exec -it deploy/gopantic-app -- nc -zv redis 6379

# Check cache configuration
kubectl get configmap gopantic-config -o yaml
```

#### Performance Issues:
```bash
# Check metrics
curl http://localhost:8081/metrics | grep gopantic

# Scale up if needed
kubectl scale deployment gopantic-app --replicas=5
```

### Log Analysis:
```bash
# Application logs
kubectl logs -l app=gopantic-app --tail=100

# Performance logs
kubectl logs -l app=gopantic-app | grep "throughput\|latency"

# Error logs
kubectl logs -l app=gopantic-app | grep "ERROR\|WARN"
```

## 📈 Scaling Guidelines

### Horizontal Scaling:

#### Metrics-based Scaling:
```yaml
# HPA configuration
targetCPUUtilizationPercentage: 70
targetMemoryUtilizationPercentage: 80
minReplicas: 3
maxReplicas: 20
```

#### Custom Metrics Scaling:
```yaml
# Scale based on queue depth
metrics:
- type: Pods
  pods:
    metric:
      name: gopantic_queue_depth
    target:
      type: AverageValue
      averageValue: "10"
```

### Vertical Scaling:
```yaml
# Resource adjustments
resources:
  requests:
    memory: "256Mi"    # Start here
    cpu: "200m"        # Increase as needed
  limits:
    memory: "1Gi"      # Monitor usage
    cpu: "1000m"       # Scale based on load
```

## 🎯 Production Checklist

### Pre-deployment:
- [ ] **Load testing** completed with expected traffic
- [ ] **Security scanning** of container images
- [ ] **Resource limits** configured appropriately
- [ ] **Monitoring and alerting** configured
- [ ] **Backup and recovery** procedures tested

### Post-deployment:
- [ ] **Health checks** passing consistently
- [ ] **Metrics collection** working
- [ ] **Log aggregation** configured
- [ ] **Performance benchmarks** validated
- [ ] **Scaling policies** tested

## 📞 Support

For deployment issues or questions:
- **Documentation**: Check the main README.md
- **Performance**: Review benchmark results in `/benchmarks`
- **Examples**: Reference implementations in `/examples`
- **Issues**: GitHub issues for bug reports and feature requests

---

**Next Steps**: After deployment, monitor your application using the provided metrics endpoints and consider implementing custom dashboards for your specific use case.