# Manhattan Server

The server for the game Blueprint by Manhattan, using Go, Docker and MySQL.

## Local Deployment Instructions

### Database setup

With a local MySQL server running, from the root directory execute the database creation script:

`> mysql -u DATABASE_USERNAME -p < database/create.sql`

Where `DATABASE_USERNAME` will depend on your local MySQL server credentials, by default it will be `root`.

Additionally, to preload the database with developer accounts for adding and removing resources, a second script will need to be run. Contact myself, @smithwjv, for the `dev.sql` script, and place it in the database directory, running it as before:

`> mysql -u DATABASE_USERNAME -p < database/dev.sql`

This script also adds test accounts for each account type: `developer`, `lecturer` and `player`.

### Configuration files

If necessary, edit the configuration file, `conf.json`, in the `authenticate`, `inventory`, `resource` and `progress` directories. The default values are:

* `"port": 8000`
* `"dbUsername": "root"`
* `"dbPassword": ""`
* `"dbHost": "host.docker.internal"`
* `"dbName": "blueprint"`

This configuration:
* Opens port 8000 of the Docker container
* Assumes credentials exist for the local MySQL database with a username of "root" and a blank password
* Assumes a MySQL server is hosted locally **not** within a Docker container, note this is OS specific
* The database name is set to "blueprint"

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

## Item Schema

The item schema JSON, found in `progress/serve/`, can be visualised as an item tree:

(If not clear, the jpeg can be found in `schemas/images/`)

![Item tree image has not loaded.](schemas/images/progression-tree.jpeg "Item tree")

Where item types are:

| Item type |                  Meaning                  |
|:---------:|:-----------------------------------------:|
|     1     | Primary resource                          |
|     2     | Machine crafted from blueprint            |
|     3     | Material/component crafted from machinery |
|     4     | Component crafted from blueprint          |
|     5     | Intangible                                |

Though note, the communication beacon is not actually in the item schema, since it will be hardcoded into the desktop client. Also, currently all items aside from intangibles are placeable.

### Quick Reference

For quick item ID to item name reference:

| Item ID |          Name         |
|:-------:|:---------------------:|
|    1    |          Wood         |
|    2    |         Stone         |
|    3    |          Clay         |
|    4    |        Iron ore       |
|    5    |       Copper ore      |
|    6    |         Rubber        |
|    7    |        Diamond        |
|    8    |          Sand         |
|    9    |       Silica ore      |
|    10   |         Quartz        |
|    11   |        Furnace        |
|    12   |        Charcoal       |
|    13   |         Steel         |
|    14   |         Copper        |
|    15   |          Iron         |
|    16   |         Glass         |
|    17   |     Silicon wafer     |
|    18   |       Fibreglass      |
|    19   |      Machine base     |
|    20   |      Wire drawer      |
|    21   |    Uninsulated wire   |
|    22   |     Insulated wire    |
|    23   |      Copper coil      |
|    24   |         Motor         |
|    25   |      Solar panel      |
|    26   |         Welder        |
|    27   |  Satellite dish frame |
|    28   |     Satellite dish    |
|    29   |    Circuit printer    |
|    30   | Printed circuit board |
|    31   |  Transmitter receiver |
|    32   |      Electricity      |
