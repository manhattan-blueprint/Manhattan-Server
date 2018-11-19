# Manhattan Server

The server for the game Blueprint by Manhattan, using Docker.

## Deployment Instructions

With Docker installed, first build the image for each service using the build script, `build.sh`, since the docker-compose file assumes the images already exist. Then initialize the current node, i.e. computer, as the swarm manager:

`docker swarm init`

Next, in the root directory, run all the services as an app with:

`docker stack deploy -c docker-compose.yml blueprint`

You can check they are running with:

`docker service ls`

To take down the app, enter:

`docker stack rm blueprint`

To take down the swarm, enter:

`docker swarm leave --force`
