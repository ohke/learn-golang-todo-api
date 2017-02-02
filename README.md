This is a repository of ToDo Web API server developed for Golang study.

## Preparatin
Create 2 DynamoDB tables on your AWS account.

| Table name | Partition key | Sort key |
| --: | :-- | :-- |
| todo-user-table | Id | (none) |
| todo-todo-table | Id | UserId |

## Build
1. Clone or download this repository.
2. Edit ``config.json`` file for your AWS environment.
3. Get below packages.

  ```
  >go get github.com/aws/aws-sdk-go
  >go get github.com/google/uuid
  >go get gopkg.in/gin-gonic/gin.v1
  ```
4. ``go build app.go`` 

## Run
```
>.\app.exe
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /register                 --> main.main.func1 (3 handlers)
[GIN-debug] POST   /login                    --> main.main.func2 (3 handlers)
[GIN-debug] POST   /logout                   --> main.main.func3 (3 handlers)
[GIN-debug] POST   /todo                     --> main.main.func4 (3 handlers)
[GIN-debug] GET    /todo/:id                 --> main.main.func5 (3 handlers)
[GIN-debug] Listening and serving HTTP on :8090
```
