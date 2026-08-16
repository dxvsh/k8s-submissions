Steps to run:

1. Build the container image: `docker build . -t project_backend:2.2`

2. Import the image into k3d (or push to dockerhub): `k3d image import project_backend:2.2 -c <CLUSTER-NAME>`

3. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`

4. Your project backend is now available to handle requests for fetching todos (`GET localhost:8081/api/todos`) and creating them (`POST localhost:8081/api/todos`)
