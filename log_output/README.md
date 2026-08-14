Steps to run:

1. Create the persistent volume and a correspoding claim to it:
  - `kubectl apply -f ../persistent-volume-definitions/persistentvolume.yaml`
  - `kubectl apply -f ../persistent-volume-definitions/persistentvolumeclaim.yaml`

2. Build and start the ping-pong application container (instructions in `../ping_pong/README.md`)

3. Build the container images for the two applications - logs generator and logs reader: 
  - `docker build . -t logs_generator:2.1 -f generator.Dockerfile`
  - `docker build . -t logs_reader:2.1 -f reader.Dockerfile`

4. Import the images into k3d (or push to dockerhub): 
  - `k3d image import logs_generator:2.1 -c <CLUSTER-NAME>`
  - `k3d image import logs_reader:2.1 -c <CLUSTER-NAME>`

5. Create the Deployment, Service and Ingress resources: `kubectl apply -f manifests/`

6. Going to the `/pingpong` endpoint, takes you to the pingpong application's counter which increments on each request. Hitting `/pings` gives you the current ping count.

7. Going to `/status` now shows you the UUID string, timestamp, and the pingpong counter.
