Steps to run:

1. Build the container images for the two applications - logs generator and logs reader: 
  - `docker build . -t logs_generator:1.10 -f generator.Dockerfile`
  - `docker build . -t logs_reader:1.10 -f reader.Dockerfile`

2. Import the images into k3d (or push to dockerhub): 
  - `k3d image import logs_generator:1.10 -c <CLUSTER-NAME>`
  - `k3d image import logs_reader:1.10 -c <CLUSTER-NAME>`

3. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`
