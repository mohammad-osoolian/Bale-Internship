# Message Broker

A high-performance, gRPC-based message broker written in Go, designed for efficient publish-subscribe communication.  
The service is containerized with Docker, deployed on a **custom Kubernetes cluster** (1 master, 2 workers) built from scratch using `kubeadm`, and fully integrated with **Prometheus** and **Grafana** for observability.  
Extensive **load testing** was performed to validate scalability and performance under heavy traffic.

---

## 🚀 Features

- **gRPC API** for low-latency communication
- **Publish/Subscribe** messaging model
- **Multiple storage backends**:  
  - In-memory  
  - PostgreSQL  
  - ScyllaDB
- **Full observability stack**:  
  - **Prometheus** for metrics (latency, throughput, error rates)  
  - **Grafana** dashboards for real-time monitoring  
- **Load testing framework** to simulate high-concurrency clients and measure broker performance
- **Graceful shutdown** with context management
- **Kubernetes-ready** with deployment and service manifests
- **Dockerized** for portability and CI/CD pipelines
