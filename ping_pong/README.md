Steps to run:

1. Build the container image: `docker build . -t ping_pong:1.11`

2. Import the image into k3d (or push to dockerhub): `k3d image import ping_pong:1.11 -c <CLUSTER-NAME>`

3. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`

4. Note that the ingress for this application is shared with the log-output application as mentioned in the instructions and can be found at: `log_output/manifests/ingress.yaml`
