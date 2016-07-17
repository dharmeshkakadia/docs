package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/docs/generate/extract"

	_ "github.com/cockroachdb/pq"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	rand.Seed(time.Now().UnixNano())
	addr := "/go/src/github.com/cockroachdb/cockroach/sql/parser/sql.y"
	bnf, err := extract.GenerateBNF(addr)
	if err != nil {
		return err
	}
	g, err := extract.ParseGrammar(bytes.NewReader(bnf))
	if err != nil {
		return err
	}
	prods := g["stmt"]
	var top extract.Sequences
	var doneLock sync.Mutex
	var done = map[string]bool{}
	for _, p := range prods {
		top = append(top, p.(extract.Sequence))
	}
	saw := func(s string) bool {
		doneLock.Lock()
		r := done[s]
		if !r {
			done[s] = true
		}
		doneLock.Unlock()
		return r
	}
	complete := func(db *sql.DB, s extract.Sequence) {
		var buf bytes.Buffer
		for i, e := range s {
			if i > 0 {
				buf.WriteByte(' ')
			}
			var r string
			switch e := e.(type) {
			case extract.Literal:
				switch e {
				case "SCONST":
					r = "'string'"
				case "ICONST":
					r = "123"
				case "FCONST":
					r = "456.789"
				case "IDENT":
					r = "ident"
				default:
					r = string(e)
				}
			default:
				panic(fmt.Errorf("bad type: %T %v", e, e))
			}
			buf.WriteString(r)
		}
		sql := buf.String()
		/*
			var p parser.Parser
			_, err := p.Parse(sql, parser.Traditional)
			if err != nil {
				//fmt.Println("ERR", err)
			} else {
				//fmt.Println(stmts)
			}
		*/
		//*
		if _, err := db.Exec("ROLLBACK"); err != nil {
			//fmt.Println(err)
		}
		if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS name;"); err != nil {
			//fmt.Println(err)
		}
		//*/
		for _, c := range []string{
			"REVOKE",
			"GRANT",
			"ALTER",
			"DROP",
		} {
			if strings.Contains(sql, c) {
				return
			}
		}
		//fmt.Println(sql)
		rows, err := db.Query(sql)
		if err != nil {
			//fmt.Println("	ERR", err)
			if strings.Contains(err.Error(), "connection") {
				fmt.Println("ERR", err)
				os.Exit(1)
			}
		} else {
			rows.Close()
		}
	}
	worker := func() {
		db, err := sql.Open("postgres", "postgresql://root@localhost:26257/name?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		for {
			cur := top[rand.Intn(len(top))]
			depth := 0
			for {
				next := g.EnumerateSequence(cur)
				if len(next) == 0 {
					break
				} else if len(next) == len(cur) && cur.String() == next.String() {
					if !saw(cur.String()) {
						complete(db, cur)
					}
					break
				}
				depth++
				if depth == 15 {
					break
				}
				cur = next
			}
		}
	}
	for i := 0; i < runtime.GOMAXPROCS(0)*2; i++ {
		go worker()
	}
	select {}
	return nil
}
