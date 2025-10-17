Kubernetes manifests for the e-commerce microservices

Quick start (minikube/kind):

1. Build images locally and load them into the cluster (minikube example):

   # from repo root
   docker build -t product-service:latest ./services/product-service
   docker build -t user-service:latest ./services/user-service
   docker build -t cart-service:latest ./services/cart-service
   docker build -t order-service:latest ./services/order-service
   docker build -t payment-service:latest ./services/payment-service

   # If using minikube:
   minikube image load product-service:latest
   minikube image load user-service:latest
   minikube image load cart-service:latest
   minikube image load order-service:latest
   minikube image load payment-service:latest

2. Apply manifests:
   kubectl apply -f k8s/namespace.yaml
   kubectl apply -f k8s/services.yaml
   kubectl apply -f k8s/ingress.yaml

Notes:
- These manifests are minimal and intended for local testing. For production you should add ConfigMaps, Secrets, resource limits, liveness/readiness probes, and RBAC.
- The Ingress uses the nginx ingress controller by default; install an ingress controller in your cluster before applying.
