Steps to run:

1. Build the container image: `docker build . -t logoutput:1.7`

2. Import the image into k3d (or push to dockerhub): `k3d image import logoutput:1.7 -c <CLUSTER-NAME>`

3. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`
