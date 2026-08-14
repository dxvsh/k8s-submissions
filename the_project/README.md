Steps to run:

1. Build the container image: `docker build . -t the_project:1.12`

2. Import the image into k3d (or push to dockerhub): `k3d image import the_project:1.12 -c <CLUSTER-NAME>`

3. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`

4. Access the todo-app on `http://localhost:8081`
