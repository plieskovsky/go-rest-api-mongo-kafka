# Users REST API Documentation

Users REST API provides all the standard CRUD operations of the User entities.

## User creation
### Request
User is created by HTTP POST request on path `/v1/users` with a json body with schema
```json
{
   "first_name":"John",
   "last_name":"Wick",
   "nickname":"johnnywicky",
   "password":"securepwd",
   "email":"johnnywicky@gmail.com",
   "country":"UK"
}
```
All the fields are required.

### Response
- `201 Created` if creation was successful. The response body is a JSON encoded data of the created user
  ```json
  {
   "id":"10e4feb6-40f9-11ef-a3eb-0242ac170004",
   "first_name":"John",
   "last_name":"Wick",
   "nickname":"johnnywicky",
   "password":"securepwd",
   "email":"johnnywicky@gmail.com",
   "country":"UK",
   "created_at":"2024-07-13T09:19:54.625Z",
   "updated_at":"2024-07-13T09:19:54.625Z"
  }
  ```
- `400 Bad Request` if the request data is incorrect. The response body has error details in form of JSON e.g. `{"error":"email is invalid"}`
- `500 Internal Server Error` in case of server failures

### Curl example
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"first_name":"John","last_name":"Wick","nickname":"johnnywicky","password":"securepwd","email":"johnnywicky@gmail.com","country":"UK"}' \
  localhost:8080/v1/users -v
```

## User update
### Request
User is updated by HTTP PUT request on path `/v1/users/<userID>` with a json body with schema
```json
{
  "id":"<userID>",
  "first_name":"Johnn",
  "last_name":"Wickk",
  "nickname":"johnnywickyy",
  "password":"securepwdd",
  "email":"johnnywicky@gmail.comm",
  "country":"UKK"
}
```
All the fields are required.

### Response
- `204 No Content` if update was successful
- `400 Bad Request` if the request data is incorrect. The response body has error details in form of JSON e.g. `{"error":"email is invalid"}`
- `403 Not Found` if the user with given ID wasn't found
- `500 Internal Server Error` in case of server failures

### Curl example
```bash
curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"id":"10e4feb6-40f9-11ef-a3eb-0242ac170004", "first_name":"Johnn","last_name":"Wickk","nickname":"johnnywickyy","password":"securepwdd","email":"johnnywicky@gmail.comm","country":"UKK"}' \
  localhost:8080/v1/users/10e4feb6-40f9-11ef-a3eb-0242ac170004 -v
```

## User delete
### Request
User is deleted by HTTP DELETE request on path `/v1/users/<userID>`

### Response
- `204 No Content` if update was successful
- `400 Bad Request` if the request data is incorrect. The response body has error details in form of JSON e.g. `{"error":"incorrect ID"}`
- `403 Not Found` if the user with given ID wasn't found
- `500 Internal Server Error` in case of server failures

### Curl example
```bash
curl  --request DELETE localhost:8080/v1/users/10e4feb6-40f9-11ef-a3eb-0242ac170004 -v
```

## User retrieval
### Request
User is retrieved by HTTP GET request on path `/v1/users/<userID>`

### Response
- `200 OK` if user was found. The response body is a JSON encoded data of the user
  ```json
  {
   "id":"10e4feb6-40f9-11ef-a3eb-0242ac170004",
   "first_name":"John",
   "last_name":"Wick",
   "nickname":"johnnywicky",
   "password":"securepwd",
   "email":"johnnywicky@gmail.com",
   "country":"UK",
   "created_at":"2024-07-13T09:19:54.625Z",
   "updated_at":"2024-07-13T09:19:54.625Z"
  }
  ```
- `403 Not Found` if the user with given ID wasn't found
- `400 Bad Request` if the request data is incorrect. The response body has error details in form of JSON e.g. `{"error":"incorrect ID"}`
- `500 Internal Server Error` in case of server failures

### Curl example
```bash
 curl  --request GET localhost:8080/v1/users/10e4feb6-40f9-11ef-a3eb-0242ac170004 -v
```

## Multiple users retrieval
### Request
Users are retrieved by HTTP GET request on path `/v1/users` with query parameters defining the sorting, pagination and filtering.

Pagination is controlled by `pageSize` and `page` query parameters. Both have to be a positive integer if defined.
If not provided `pageSize` defaults to `20` and `page` to 0.

Sorting is controlled by `sortBy` query parameter. The format of the parameter value is `field.sortType` e.g. `sortBy=first_name.asc`.
Supported sort types are `asc` and `desc`. Supported sort fields are:
 - last_name
 - first_name
 - nickname
 - password
 - email
 - country
 - created_at
 - updated_at

Filtering is controlled by query parameter in format `field=value` e.g. `country=UK`. The filter is searching for the exact matches.
Supported filter fields are:
- last_name
- first_name
- nickname
- email
- country

### Response
- `200 OK` with a list of users that match the criteria. Returns empty list in case of no match e.g. `[]`
  ```json
  [
   {
      "id":"10e4feb6-40f9-11ef-a3eb-0242ac170004",
      "first_name":"Andrea",
      "last_name":"Ananas",
      "nickname":"any",
      "password":"annaspwd",
      "email":"ann@gmail.com",
      "country":"UK",
      "created_at":"2024-07-12T13:06:34.465Z",
      "updated_at":"2024-07-12T13:06:34.465Z"
   },
   {
      "id":"b79d4ce5-40f9-11ef-a3eb-0242ac170004",
      "first_name":"john",
      "last_name":"wick",
      "nickname":"johnnywicky",
      "password":"securepwd",
      "email":"johnnywicky@gmail.com",
      "country":"UK",
      "created_at":"2024-07-12T12:22:36.734Z",
      "updated_at":"2024-07-12T12:22:36.734Z"
   }
  ]
  ```
- `400 Bad Request` if the request data is incorrect. The response body has error details in form of JSON e.g. `{"error":"unsupported sorting field"}`
- `500 Internal Server Error` in case of server failures
### Curl example
```bash
curl  --request GET -v "localhost:8080/v1/users?pageSize=2&page=1&sortBy=first_name.asc&country=UK"
```
