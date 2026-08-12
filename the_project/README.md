Steps to run:

1. Build the container image: `docker build . -t the_project:1.8`

2. Import the image into k3d (or push to dockerhub): `k3d image import the_project:1.8 -c <CLUSTER-NAME>`

3. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`
