# calc_micro_final:
HTTP REST Api Distributed Calculator in Go

# How to start:
You can simply run it by
```
go run cmd/orchestrator/main.go
go run cmd/agent/main.go
```
or build
```
go build -o OUTPUT_ORCHESTRATOR_BINARY cmd/orchestrator/main.go
go build -o OUTPUT_AGENT_BINARY cmd/agent/main.go
```

All settings read from .env file placed near binary or if run through 'go run' - in project root dir
Database file for sqlite will be created automatically near binary or in project root if started as 'go run'

Just rename .env.example to .env (with default values)
By default HTTP port will be 1234, GRPC - 5000


# Default ENV params
```
APP_ENV=production
APP_LISTENING_ADDRESS=localhost
APP_HTTP_LISTEN_PORT=1234
APP_GRPC_LISTEN_PORT=5000

APP_JWT_SECRET_KEY="oqjyriwt7rcihw7tn"
APP_JWT_REFRESH_SECRET_KEY="mscriytiwetcnurtw"
APP_JWT_EXPIRATION_HOURS=720

APP_DB_HOST=localhost
APP_DB_PORT=33060
APP_DB_USER=root
APP_DB_NAME=./golang.sqlite
APP_DB_PASSWORD=123
APP_DB_TYPE=sqlite

#calculation delay in ms
TIME_ADDITION_MS=1000
TIME_SUBTRACTION_MS=1000
TIME_MULTIPLICATIONS_MS=1000
TIME_DIVISIONS_MS=1000
# seconds before allow task to distribute to another worker
TIME_TASK_IN_PROGRESS_REDISTRIBUTE=60

# how many calculation workers to start
CLIENT_COMPUTING_POWER=2
CLIENT_GRPC_ARRT=localhost
CLIENT_GRPC_PORT=5000
```

For sqlite DB only APP_DB_TYPE & APP_DB_NAME are used


# Examples:
   ## api/register
   ### Wrong BODY
   Expect code 500 and {
  "message": "code=400, message=Syntax error: offset=43, error=invalid character '\\n' in string literal, internal=invalid character '\\n' in string literal"
}
   ```http
   POST http://localhost/api/register
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "123
   }
   ```
   ### Fail to register. (For example user already exists)
   Expect code 422 and {"message": "Unprocessable Entity"}
   ```http
   POST http://localhost/api/register
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "123"
   }
   ```
   ### Validation error.
   Expect code 422 and {"Password too short"}
   ```http
   POST http://localhost/api/register
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "12"
   }
   ```
   ### OK.
   Expect code 201
   ```http
   POST http://localhost/api/register
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "123"
   }
   ```
   ## api/login
   ### Wrong BODY
   Expect code 500 and {
  "message": "code=400, message=Syntax error: offset=43, error=invalid character '\\n' in string literal, internal=invalid character '\\n' in string literal"
}
   ```http
   POST http://localhost/api/login
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "123
   }
   ```
   ### Fail to register. (For example wrong password)
   Expect code 403 and {"message": "Incorrect credentials"}
   ```http
   POST http://localhost/api/login
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "123"
   }
   ```
   ### Validation error.
   Expect code 422 and {"Password too short"}
   ```http
   POST http://localhost/api/register
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "12"
   }
   ```
   ### OK.
   Expect code 200 and {
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MiwiYWRtaW4iOmZhbHNlLCJleHAiOjE3NDg0NDAyODN9.nug9qRh8lVFSL-s_32eAU1hnIUCKRwJd2K46Gs9EdUY",
  "expires_at": "2025-05-28T15:51:23.572426+02:00"
}
   ```http
   POST http://localhost/api/login
   Content-Type: application/json

   {
      "email": "a@c.com",
      "password": "123"
   }
   ```
   ## api/v1/calculate
   ### Wrong HTTP Method.
   Expect code 500 and {"error": "Invalid request"}
   ```http
   GET http://localhost/api/v1/calculate
   ```
   ### OK Expression.
   Expect code 201 and {"id": "0A2DDEF9-F67C-6899-5F72-25639EEBD08F"}
   ```http
   POST http://localhost/api/v1/calculate
   Content-Type: application/json

   {
     "expression": "2+2*2"
   }
   ```
   ### Empty or Incorrect expression.
   Expect code 422 and {"error": "Invalid request body"}
   ```http
   POST http://localhost/api/v1/calculate
   Content-Type: application/json
   {
     "expression": ""
   }
   ```
   ## api/v1/expressions
   ### OK Expression.
   Expect code 200 and response
   ```json
   {
       "expressions": [
           {
            "id": "04F88C13-4BD9-B8A4-E1B5-6C9C4A72FCED",
            "expression": "2-2/3",
            "status": "completed",
            "result": 1.3333333333333335,
            "created_at": null,
            "updated_at": null
            },
            {
            "id": "43FFE612-D8DD-38E3-32B7-2A0911871F4F",
            "expression": "2-2/3 +1/3",
            "status": "completed",
            "result": 1.6666666666666667,
            "created_at": null,
            "updated_at": null
            }
       ]
   }
   ```

   ```http
   GET http://localhost/api/v1/expressions
   ```
   ## api/v1/expressions/{ID}
   ### OK Expression.
   Expect code 200 and response
   ```json
       {
        "id": "04F88C13-4BD9-B8A4-E1B5-6C9C4A72FCED",
        "expression": "2-2/3",
        "status": "completed",
        "result": 1.3333333333333335,
        "created_at": null,
        "updated_at": null
        }
   ```

   ```http
   GET http://localhost/api/v1/expressions
   ```

You can do a simple test with curl like
```
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```