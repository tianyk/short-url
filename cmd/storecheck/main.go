package main

import (
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tianyk/short-url/proto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: storecheck <leveldb dir>")
		os.Exit(1)
	}
	db, err := leveldb.OpenFile(os.Args[1], nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer db.Close()

	counter := ""
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := string(iter.Key())
		if key == "URL-ID" {
			counter = string(iter.Value())
			continue
		}
		if len(key) < 2 || key[:2] != "s-" {
			fmt.Printf("OTHER key=%q value=%q\n", key, iter.Value())
			continue
		}
		msg := new(proto.ShortUrlMessage)
		ok := "OK"
		if err := msg.Unmarshal(iter.Value()); err != nil {
			ok = "MALFORMED"
		}
		fmt.Printf("s- key=%s %s expire=%d long=%q\n", key, ok, msg.Expire, msg.LongUrl)
	}
	iter.Release()
	fmt.Println("---")
	fmt.Printf("URL-ID counter = %q\n", counter)
}
