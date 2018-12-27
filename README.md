# Manhattan Server

The server for the game Blueprint by Manhattan, using Docker and MySQL.

## Local Deployment Instructions

### Database setup

With a local MySQL server running, from the root directory execute the database creation script:

`> mysql -u DATABASE_USERNAME -p < /database/create.sql`

Where `DATABASE_USERNAME` will depend on your local MySQL server credentials.

### Configuration files

If necessary, edit the configuration file in the `authenticate`, `inventory` and `resource` directories. The default values are:

* `"port": 8000`
* `"dbUsername": "root"`
* `"dbPassword": ""`
* `"dbHost": "host.docker.internal"`
* `"dbName": "blueprint"`

This configuration:
* Opens port 8000 of the Docker container
* Assumes credentials exist for the local MySQL database with a username of "root" and a blank password
* Assumes a MySQL server is hosted locally **not** within a Docker container
* The database name is set to "blueprint".

Note, the configuration file in the `inventory` and `resource` directories currently only specify the port to open. These will be updated to match the configuration above when the respective API calls are implemented.

### Deployment

With the database and configuration files setup and Docker installed, to build the images for each service and deploy the server using Docker swarm, from the root directory type:

`> make`

You can check the services are running with:

`> docker service ls`

To stop the services and take down the Docker swarm, type:

`> make clean`

## Local Testing Instructions

With the database and configuration files setup and Docker installed, to build the images for each service and run the tests, from the root directory type:

`> make test`