# qb

Just a simple SQL query builder for PostgreSQL.

## Get it

```console
go get go.step.sm/qb
```

## Usage

```go
package users

import "go.step.sm/qb"

var selectUser, insertUser, updateUser, deleteUser string
var selectUserByEmail string

func init() {
    q := qb.NewQueryBuilder("users", []string{
        "id", "name", "email",
        "created_at", "updated_at", "deleted_at",
    })
    selectUser, insertUser, updateUser, deleteUser = q.Queries()
    selectUserByEmail = q.SelectBy("email")
}
```
