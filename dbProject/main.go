package main

import (
	"context"
	"dbProject/MyHttp"
	"dbProject/SqlCommand"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var(
	uniDB = "postgresql://habitov:habitov@95.217.232.188:7777/habitov"
	localDB = "postgresql://maui:maui@192.168.0.12:5432/postgres"
)

func main() {
	//handler := MyHttp.MyRequest()
	conn, err := SqlCommand.Connection(localDB)
	if err != nil{
		fmt.Println("faild to connect db: " + err.Error())
		return
	}

	router := MyHttp.MyRequest(MyHttp.Route{ DB: conn })
	http.ListenAndServe(":3000", router)

	//
	//fmt.Printf("Hello")

	//fmt.Println("п1ыртик")
}

func HandlerRequest(w http.ResponseWriter, r *http.Request){
	request(w, "Salim")
}

func request(writer io.Writer, name string){
	fmt.Fprintf(writer, "Hello, %s", name)
}

func checkRights() {
	ctxwr, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctxwr, "Server=95.217.232.188;Port=7777;Username=habitov;Password=habitov")
	if err != nil {
		fmt.Println(os.Stderr, "Unable to connect to db: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var right string
	err = conn.QueryRow(context.Background(), "select rights from is_rights where user_id = 1").Scan(&right)

	if err != nil {
		fmt.Println(os.Stderr, "err")
	}

	fmt.Println(right)
}
