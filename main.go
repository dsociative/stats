package main

import (
	"flag"
	"github.com/dsociative/stats/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "net/http/pprof"
)

var (
	dsn       = flag.String("db", "root:secret@/stats", "mysql dsn")
	ttl       = flag.Duration("ttl", time.Second*30, "cache ttl")
	addr      = flag.String("addr", ":8089", "bind addr")
	pprof     = flag.Bool("pprof", false, "enables pprof http handler")
	pprofAddr = flag.String("pprof_addr", "localhost:8787", "addr for pprof handler")
)

func main() {
	flag.Parse()
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		log.Fatalln(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalln(err)
	}

	c := handler.NewCache(db, *ttl)
	go c.Loop()
	defer c.Close()
	h := handler.NewHandler(c)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigs
		c.Close()
		os.Exit(0)
	}()

	if *pprof {
		go func() {
			http.ListenAndServe(*pprofAddr, nil)
		}()
	}

	if err := http.ListenAndServe(*addr, h); err != nil {
		log.Panic(err)
	}
}
