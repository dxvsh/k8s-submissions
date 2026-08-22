Steps to run:

1. Build the container image: `docker build . -t ping_pong:2.7`

2. Import the image into k3d (or push to dockerhub): `k3d image import ping_pong:2.7 -c <CLUSTER-NAME>`

3. Decrypt and apply the secret for postgres connection: `sops decrypt secret.enc.yaml | kubectl apply -f -`

4. Create the Deployment, Service and StatefulSet resources: `kubectl apply -f manifests/`

5. Note that the ingress for this application is shared with the log-output application as mentioned in the instructions and can be found at: `log_output/manifests/ingress.yaml`
