## Khan - Pod Connection Tracking Metrics Exporter

Khan captures connection tracking snapshots on Pods, and Nodes and exposes them as  prometheus metrics. Note that the metrics don't constitute realtime connection info, only snapshots that are polled with a default period of 30s.

The use case for this application is for tracking down pods/services that are leaking connections or finding an unknown client that is overloading a server.

This application is composed of a 'controller' that runs as a deployment. The controller is mainly an API for the node agents to retrieve mappings of IP-to-pod for IPs found in the conntrack table. The 'agent' runs as a daemonset on each node and captures the conntrack table and converts it to a set of prometheus metrics.

## Running in Docker Compose
  - Create a file called .env in the root of your project, which will not be checked in
    - Run `make setup` to generate the file with empty values
        - Your .env file should look like:
          ```bash
          IMAGE_NAME=controller_app_container
          MY_ENV_USER=
          MY_ENV_TOKEN=
          MY_ENV_PRESET=DMELAB
          ```
    - #### Additional Details
      ```bash
      # we admit that this process is too manual at the moment, but since these details need to be secret and not checked in, here we are...
      # create an image name that will be used by docker-compose to name your image
      IMAGE_NAME=
      # your BitBucket username. Should be your attuid
      MY_ENV_USER=
      # create an access token on BitBucket with read access, and paste that token as this value
      MY_ENV_TOKEN
      # set your client preset
      MY_ENV_PRESET=
      ```
  - Run the following command:
    ```bash
    make docker-build && make docker-up
    ```
  - That's it!
