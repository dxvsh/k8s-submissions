Steps to run:

1. Create a persistent volume and a correspoding claim to it:
  - `kubectl apply -f ../persistent-volume-definitions/project-pv.yaml`
  - `kubectl apply -f ../persistent-volume-definitions/project-pvc.yaml`

2. Build the container image: `docker build . -t the_project:2.4`

3. Import the image into k3d (or push to dockerhub): `k3d image import the_project:2.4 -c <CLUSTER-NAME>`

4. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`

5. Build and start the project-backend service. See instructions here: `../project_backend/README.md`

6. Access the todo-app on `http://localhost:8081`
