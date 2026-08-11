Steps to run:

1. Build the container image: `docker build . -t the_project:1.5`

2. Import the image into k3d (or push to dockerhub): `k3d image import the_project:1.5 -c <CLUSTER-NAME>`

3. Create Deployment: `kubectl apply -f manifests/deployment.yaml`
