package main

import (
	"flag"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/kataras/iris"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
)

type Users struct {
	ID   int64
	Name string `xrom:"varchar(30)"`
	Age  int    `xrom:"int(10)"`
	Sex  string `xrom:"varchar(2)"`
}

type Config struct {
	ServerHost string
}

var config *Config

func parseFlags() {
	config = &Config{}
	flag.StringVar(&config.ServerHost, "p", "0.0.0.0:9800", "default server address")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	parseFlags()

	app := iris.New()

	orm, err := xorm.NewEngine("sqlite3", "./users.db")
	if err != nil {
		fmt.Printf("orm failed to initialized, %s", err)
	}

	iris.RegisterOnInterrupt(func() {
		orm.Close()
	})

	err = orm.Sync2(new(Users))

	if err != nil {
		fmt.Printf("orm failed to initialized Users table, %s", err)
	}

	// Get
	// curl http://localhost:9800
	app.Get("/", func(ctx iris.Context) {
		ctx.WriteString("Hello Iris.")
	})

	// Post
	// curl  http://localhost:9800/users -X POST -d '{"ID":1, "Name":"ben", "Age":25, "Sex":"m"}'
	app.Post("/users", func(ctx iris.Context) {
		var user Users
		if err := ctx.ReadJSON(&user); err != nil {
			ctx.WriteString("Invalid user.")
		}
		orm.Insert(user)
		ctx.Writef("Add successful, %#v\n", user)
	})

	// DELETE
	// curl  http://localhost:9800/users/1 -X DELETE
	app.Delete("/users/{id:int}", func(ctx iris.Context) {
		parmid := ctx.Params().Get("id")
		id, err := strconv.ParseInt(parmid, 10, 64)
		if err != nil {
			ctx.WriteString("id is not valid")
		}

		user := Users{ID: id}
		orm.Delete(user)
		ctx.Writef("user deleted: %#v\n", user)
	})

	// Select
	// curl  http://localhost:9800/users/1
	app.Get("/users/{id:int}", func(ctx iris.Context) {
		parmid := ctx.Params().Get("id")
		id, err := strconv.ParseInt(parmid, 10, 64)
		if err != nil {
			ctx.Writef("Invalid id, %s", err)
		}

		user := Users{ID: id}
		if ok, _ := orm.Get(&user); ok {
			ctx.Writef("user found: %#v\n", user)
		} else {
			ctx.WriteString("user not found.")
		}
	})

	// Update
	// curl  http://localhost:9800/users/1 -X PUT -d '{"Name":"jane", "Age":20, "Sex":"f"}'
	app.Put("/users/{id:int}", func(ctx iris.Context) {
		var user Users
		if err := ctx.ReadJSON(&user); err != nil {
			ctx.WriteString("Invalid user.")
		}

		parmid := ctx.Params().Get("id")
		id, err := strconv.ParseInt(parmid, 10, 64)
		if err != nil {
			ctx.Writef("Invalid id, %s", err)
		}

		affected, err := orm.ID(id).AllCols().Update(&user)
		if err != nil {
			ctx.Writef("user updated failed, %s", err)

		} else {
			ctx.Writef("user updated: %d", affected)

		}
	})

	app.Run(iris.Addr(config.ServerHost), iris.WithoutServerError(iris.ErrServerClosed))
}
