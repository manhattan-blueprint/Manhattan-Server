# Blueprint Server API

The base URL for the schema is `http://foo.com`<br>
All endpoints must start with `/api/v1`<br>
No POST or URL parameters can be blank<br>
All requests, aside from Authentication, must contain the access token as a header
`Authorization: Bearer <token>`, where each token is a 64 character string

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
                "latitude": 123.456, 
                "longitude": 123.678
            }  
        },
        {
            "item_id": 2, 
            "location": {
                "latitude": 123.467, 
                "longitude": 123.688
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
    "error":"Unauthorized auth token is invalid"
}
```
