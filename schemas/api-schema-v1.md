# Blueprint Server API

* The base URL for the schema is `http://smithwjv.ddns.net`
* A port must be specified for each service
  * 8000 for authenticate
  * 8001 for inventory
  * 8002 for resources
  * 8003 for progress
* All endpoints must start with `/api/v1`
* No POST or URL parameters can be blank
* All requests, aside from Authentication and item schema, must contain the access token as a header
`Authorization: Bearer <token>`, where each token is a 64 character string
* All errors will be a JSON of the form `"error":"Example error"`
* The item schema is served from the progress service, so uses the 8003 port

# Item Schema

`/item-schema` (GET) <br>
**Description**: Get the item schema JSON

**Response**: <br>
```json
{
    "items":[
        {
            "item_id":1,
            "name":"wood",
            "type":1,
            "blueprint":[],
            "machine_id":null,
            "recipe":[]
        }
    ]
}
```

# Authentication

`/authenticate/register` (POST) <br>
**Description**: Create a new user and get auth tokens and account type

**Request Contents**:

Parameter | Type | Description
---|---|---
username | String | User username
password | String | User password (plaintext, protected by https)

**Response**: <br>
```json
{
    "access":"abcdefgh",
    "refresh":"ijklmnop",
    "account_type":"player"
}
```

`/authenticate` (POST) <br>
**Description**: Validate an existing user and get access tokens and account type

**Request Contents**:

Parameter | Type | Description
---|---|---
username | String | User username
password | String | User password (plaintext, protected by https)

**Response**: <br>
```json
{
    "access":"abcdefgh",
    "refresh":"ijklmnop",
    "account_type":"player"
}
```

---
`/authenticate/refresh` (POST) <br>
**Description**: Fetch a new access token once expired

**Request Contents**:

Parameter | Type | Description
---|---|---
refresh_token | String | The previous refresh token

**Response**: <br>
```json
{
    "access":"abcdefgh",
    "refresh":"ijklmnop",
    "account_type":"player"
}
```

# Inventory
`/inventory` (GET) <br>
**Description**: Fetch inventory for user, only returns items they have, not all possible

**Response**: <br>
```json
{
    "items":[
        {"item_id": 0, "quantity": 1},
        {"item_id": 1, "quantity": 3},
        {"item_id": 2, "quantity": 300},
    ] 
}
```

---
`/inventory` (POST) <br>
**Description**: Add item(s) to inventory

**Request Contents**:

Parameter | Type | Description
---|---|---
items | List | List of item_id, quantity pairs to add

Where each list element has the following contents:

Parameter | Type | Description
---|---|---
item_id  | Int | The item to add (1 - 16 inclusive)
quantity | Int | Quantity of the item to add (1 or greater)

**Response**: <br>
```json
{}
```

---
`/inventory` (DELETE) <br>
**Description**: Delete all inventory items for user

**Response**: <br>
```json
{}
```

# Resources
`/resources` (GET) <br>
**Description**: Get resources within a radius

**URL Parameters**:

Parameter | Type | Description
---|---|---
lat  | Float | Latitude coordinate
long | Float | Longitude coordinate

**Response**: <br>
```json
{
    "spawns":[
        {
            "item_id":1, 
            "location": {
                "latitude":50.12345678, 
                "longitude":-2.61234567
            },
            "quantity":3
        },
        {
            "item_id":2, 
            "location": {
                "latitude":50.87654321, 
                "longitude":-2.67654321
            },
            "quantity":5
        }
    ]
}
```

---
`/resources` (POST) <br>
**Description**: Add resource(s), from a developer account

**Request Contents**:

Parameter | Type | Description
---|---|---
spawns | List | List of item_id, location pairs to add

Where each list element has the following contents:

Parameter | Type | Description
---|---|---
item_id  | Int | The item to add (1 - 16 inclusive)
location | Object | The location of the item to add
quantity | Int | The quantity to add (1 or greater)

Where the location object has the following contents:

Parameter | Type | Description
---|---|---
Latitude  | Float | Latitude coordinate
Longitude | Float | Longitude coordinate

Example:

```json
{
    "spawns":[
        {
            "item_id":1, 
            "location": {
                "latitude":50.12345678, 
                "longitude":-2.61234567
            },
            "quantity":3
        },
        {
            "item_id":2, 
            "location": {
                "latitude":50.87654321, 
                "longitude":-2.67654321
            },
            "quantity":5
        }
    ]
}
```

**Response**: <br>
```json
{}
```

---
`/resources` (DELETE) <br>
**Description**: Remove resource(s), from a developer account

**Request Contents**:

Parameter | Type | Description
---|---|---
spawns | List | List of item_id, location pairs to remove

Where each list element has the following contents:

Parameter | Type | Description
---|---|---
item_id  | Int | The item to remove (1 - 16 inclusive)
location | Object | The location of the item to remove
quantity | Int | The quantity to remove (1 or greater, must match)

Where the location object has the following contents:

Parameter | Type | Description
---|---|---
Latitude  | Float | Latitude coordinate
Longitude | Float | Longitude coordinate

Example:

```json
{
    "spawns":[
        {
            "item_id": , 
            "location": {
                "latitude":50.12345678, 
                "longitude":-2.61234567
            },
            "quantity":3  
        },
        {
            "item_id":2, 
            "location": {
                "latitude":50.87654321, 
                "longitude":-2.67654321
            },
            "quantity":5  
        }
    ]
}
```

**Response**: <br>
```json
{}
```

# Progress
`/progress` (GET) <br>
**Description**: Fetch progress for user, only returns blueprints that they have completed, not all possible

**Response**: <br>
```json
{
    "blueprints":[
        {"item_id":0},
        {"item_id":1},
        {"item_id":2},
    ]
}
```

---
`/progress` (POST) <br>
**Description**: Add blueprint(s) to progress

**Request Contents**:

Parameter | Type | Description
---|---|---
blueprints | List | List of completed blueprints

Where each list element has the following contents:

Parameter | Type | Description
---|---|---
item_id  | Int | The blueprint item_id to add (1 - 16 inclusive)

**Response**: <br>
```json
{}
```

---
`progress/leaderboard` (GET) <br>
**Description**: Fetch all player progress, i.e. all blueprints completed, from a developer account. Note this is unordered

**Response**: <br>
```json
{
    "leaderboard":[
        {
            "username":"John",
            "item_id":4
        },
        {
            "username":"Leo",
            "item_id":12
        },
        {
            "username":"John",
            "item_id":8
        }
    ]
}
```

---
`progress/desktop-state` (POST) <br>
**Description**: Add desktop state JSON

**Response**: <br>
```json
{}
```

---
`progress/desktop-state` (GET) <br>
**Description**: Get desktop state JSON

**Response**: <br>
```json
{
    "game_state":"..."
}
```