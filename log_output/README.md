Steps to run:

1. Build the container image: `docker build . -t logoutput:1.1`

2. Import the image into k3d (or push to dockerhub): `k3d image import logoutput:1.1 -c <CLUSTER-NAME>`

3. Create Deployment: `kubectl create deployment logoutput --image=logoutput:1.1`
