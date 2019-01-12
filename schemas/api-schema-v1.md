# Blueprint Server API

* The base URL for the schema is `http://smithwjv.ddns.net`
* A port must be specified for each service
  * 8000 for authenticate
  * 8001 for inventory
  * 8002 for resources
* All endpoints must start with `/api/v1`
* No POST or URL parameters can be blank
* All requests, aside from Authentication and item schema, must contain the access token as a header
`Authorization: Bearer <token>`, where each token is a 64 character string

# Item Schema

`/item-schema` (GET) <br>
**Description**: Get the item schema JSON

**Response**: <br>
Code 200:
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
**Description**: Create a new user and get auth tokens

**Request Contents**:

Parameter | Type | Description
---|---|---
username | String | User username
password | String | User password (plaintext, protected by https)

**Response**: <br>
Code 200:
```json
{
    "access":"abcdefgh",
    "refresh":"ijklmnop"
}
```
Code 400:
```json
{
    "error":"Invalid username or password"
}
```
or
```json
{
    "error":"Username already exists"
}
```

`/authenticate` (POST) <br>
**Description**: Validate an existing user and get access tokens 

**Request Contents**:

Parameter | Type | Description
---|---|---
username | String | User username
password | String | User password (plaintext, protected by https)

**Response**: <br>
Code 200:
```json
{
    "access":"abcdefgh",
    "refresh":"ijklmnop"
}
```
Code 400:
```json
{
    "error":"Invalid username or password"
}
```
Code 401:
```json
{
    "error":"The credentials provided do not match any user"
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
Code 200:
```json
{
    "access":"abcdefgh",
    "refresh":"ijklmnop"
}
```
Code 400:
```json
{
    "error":"Invalid refresh token"
}
```
Code 401:
```json
{
    "error":"The refresh token provided does not match any user"
}
```

# Inventory
`/inventory` (GET) <br>
**Description**: Fetch inventory for given user associated with access token. Only returns items that they have, not all possible.

**Response**: <br>
Code 200:
```json
{
    "items": [
        {"item_id": 0, "quantity": 1},
        {"item_id": 1, "quantity": 3},
        {"item_id": 2, "quantity": 300},
    ] 
}
```
Code 401:
```json
{
    "error":"The access token provided does not match any user"
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
Code 200:
```json
{}
```
Code 400:
```json
{
    "error":"Invalid item list"
}
```
or
```json
{
    "error":"Empty item list"
}
```
or
```json
{
    "error":"Invalid item ID in list"
}
```
or
```json
{
    "error":"Invalid item quantity in list"
}
```
Code 401:
```json
{
    "error":"The access token provided does not match any user"
}
```

---
`/inventory` (DELETE)<br>
**Description**: Delete all inventory items for user

**Response**: <br>
Code 200:
```json
{}
```

Code 401:
```json
{
    "error":"The access token provided does not match any user"
}
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
Code 200:
```json
{
    "spawns": [
        {
            "item_id": 1, 
            "location": {
                "latitude": 50.12345678, 
                "longitude": -2.61234567
            }  
        },
        {
            "item_id": 2, 
            "location": {
                "latitude": 50.87654321, 
                "longitude": -2.67654321
            }  
        }
    ]
}
```

Code 400: 
```json
{
    "error":"Latitude and longitude parameters are required"
}
```
or
```json
{
    "error":"Could not convert latitude to float"
}
```
or
```json
{
    "error":"Could not convert longitude to float"
}
```
or
```json
{
    "error":"Invalid latitude, must be between -90 and 90"
}
```
or
```json
{
    "error":"Invalid longitude, must be between -180 and 180"
}
```
Code 401:
```json
{
    "error":"The access token provided does not match any user"
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

Where the location object has the following contents:

Parameter | Type | Description
---|---|---
Latitude  | Float | Latitude coordinate
Longitude | Float | Longitude coordinate

Example:

```json
{
    "spawns": [
        {
            "item_id": 1, 
            "location": {
                "latitude": 50.12345678, 
                "longitude": -2.61234567
            }  
        },
        {
            "item_id": 2, 
            "location": {
                "latitude": 50.87654321, 
                "longitude": -2.67654321
            }  
        }
    ]
}
```

**Response**: <br>
Code 200:
```json
{}
```
Code 400:
```json
{
    "error":"Invalid spawn list"
}
```
or
```json
{
    "error":"Empty spawn list"
}
```
or
```json
{
    "error":"Invalid item ID in list"
}
```
or
```json
{
    "error":"Could not convert latitude to float"
}
```
or
```json
{
    "error":"Could not convert longitude to float"
}
```
or
```json
{
    "error":"Invalid latitude, must be between -90 and 90"
}
```
or
```json
{
    "error":"Invalid longitude, must be between -180 and 180"
}
```
Code 401:
```json
{
    "error":"The access token provided does not match any user"
}
```
or
```json
{
    "error":"User must be a developer"
}
```

---
`/resources` (DELETE)<br>
**Description**: Remove resource(s), from a developer account

**Request Contents**:

Parameter | Type | Description
---|---|---
spawns | List | List of item_id, location pairs to add

Where each list element has the following contents:

Parameter | Type | Description
---|---|---
item_id  | Int | The item to add (1 - 16 inclusive)
location | Object | The location of the item to add

Where the location object has the following contents:

Parameter | Type | Description
---|---|---
Latitude  | Float | Latitude coordinate
Longitude | Float | Longitude coordinate

Example:

```json
{
    "spawns": [
        {
            "item_id": 1, 
            "location": {
                "latitude": 50.12345678, 
                "longitude": -2.61234567
            }  
        },
        {
            "item_id": 2, 
            "location": {
                "latitude": 50.87654321, 
                "longitude": -2.67654321
            }  
        }
    ]
}
```

**Response**: <br>
Code 200:
```json
{}
```
Code 400:
```json
{
    "error":"Invalid spawn list"
}
```
or
```json
{
    "error":"Empty spawn list"
}
```
or
```json
{
    "error":"Invalid item ID in list"
}
```
or
```json
{
    "error":"Could not convert latitude to float"
}
```
or
```json
{
    "error":"Could not convert longitude to float"
}
```
or
```json
{
    "error":"Invalid latitude, must be between -90 and 90"
}
```
or
```json
{
    "error":"Invalid longitude, must be between -180 and 180"
}
```
Code 401:
```json
{
    "error":"The access token provided does not match any user"
}
```
or
```json
{
    "error":"User must be a developer"
}
```
