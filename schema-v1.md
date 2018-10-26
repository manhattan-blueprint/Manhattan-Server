# Blueprint Server API

The base URL for the schema is `http://foo.com` <br>
All endpoints must start with `/api/v1`<br>
All requests, aside from Authentication, must contain the access token as a header
`Authorization: Bearer <token>`

# Authentication
`/authenticate` (POST) <br>
**Description**: Validate an existing user and get access tokens 

**Request Contents**:

Parameter | Type | Description
---|---|---
username | String | The username
password | String | Password (encrypted???)

**Response**: <br>
Code 200:
```json
{
    "access_token": "abcdefgh",
    "refresh_token": "abddeefd"
}
```
Code 401:
```json
{
    "error": "The credentials provided do not match any user"
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
    "access_token": "abcdefgh",
    "refresh_token": "abddeefd"
}
```
Code 401:
```json
{
    "error": "The refresh token provided does not match any user"
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
    "error": "Unauthorized auth token is invalid"
}
```

---
`/inventory` (POST) <br>
**Description**: Add item to inventory

**Request Contents**:

Parameter | Type | Description
---|---|---
item_id | Int | The item collected
quantity | Int | The quantity of item collected

**Response**: <br>
Code 200:
```json
{}
```
Code 400:
```json
{
    "error": "An item with this id does not exist"
}
```

Code 401:
```json
{
    "error": "Unauthorized auth token is invalid"
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
    "error": "Unauthorized auth token is invalid"
}
```

# Resources
`/resources` (GET) <br>
**Description**: Get resources within a radius
**Response**: <br>
Code 200:
```json
{
    "items": [
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
    "error": "Location provided is invalid"
}
```

Code 401:
```json
{
    "error": "Unauthorized auth token is invalid"
}
```
