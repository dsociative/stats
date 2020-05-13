package handler

import (
	"context"
	"database/sql"
	"go.uber.org/atomic"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	re        = regexp.MustCompile("^[0-9]+,[a-z]+$")
	insertSql = `INSERT INTO items (id, label, total) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE total = total + ?;`

	closeMaxDelay = time.Second * 10
)

type Cache struct {
	db       *sql.DB
	stmtIns  *sql.Stmt
	cache    sync.Map
	ticker   *time.Ticker
	doneChan chan bool
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewCache(db *sql.DB, ttl time.Duration) *Cache {
	stmtIns, err := db.Prepare(insertSql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Cache{
		db:       db,
		stmtIns:  stmtIns,
		cache:    sync.Map{},
		ticker:   time.NewTicker(ttl),
		ctx:      ctx,
		cancel:   cancel,
		doneChan: make(chan bool, 1),
	}
}

func (c *Cache) Close() {
	c.cancel()
	select {
	case <-c.doneChan:
	case <-time.After(closeMaxDelay):
	}
	c.stmtIns.Close()
	c.db.Close()
}

func (c *Cache) Validate(key string) bool {
	if len(key) > 255 {
		return false
	}
	return re.MatchString(key)
}

func (c *Cache) Incr(key string) {
	value, _ := c.cache.LoadOrStore(key, &atomic.Int64{})
	store := value.(*atomic.Int64)
	store.Add(1)
}

func (c *Cache) splitKey(key string) (id int, label string, result bool) {
	s := strings.Split(key, ",")
	if len(s) != 2 {
		return
	}
	var err error
	id, err = strconv.Atoi(s[0])
	if err == nil {
		label = s[1]
		result = true
	}
	return
}

func (c *Cache) tick() {
	c.cache.Range(func(key, value interface{}) bool {
		s := key.(string)
		store := value.(*atomic.Int64)
		old := store.Swap(0)

		id, label, ok := c.splitKey(s)
		if !ok {
			c.cache.Delete(key)
			return false
		}

		if _, err := c.stmtIns.Exec(id, label, old, old); err != nil {
			log.Println(err)
			store.Add(old)
			return false
		}
		return true
	})
}

func (c *Cache) Loop() {
	for {
		select {
		case <-c.ctx.Done():
			c.tick()
			c.doneChan <- true
			return
		case <-c.ticker.C:
			c.tick()
		}
	}
	return
}
