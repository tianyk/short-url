package service

import (
    "os"
    "os/signal"
    "path"
    "strconv"
    "sync"
    "syscall"

    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/errors"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
    store  *leveldb.DB
    dbLock sync.Mutex
)

var (
    readOpt  = &opt.ReadOptions{DontFillCache: false}
    writeOpt = &opt.WriteOptions{Sync: false, NoWriteMerge: false}
)

func init() {
    workspace, err := os.Getwd()
    if err != nil {
        panic(err)
    }

    db, err := leveldb.OpenFile(path.Join(workspace, "short-url-store"), nil)
    if err != nil {
        panic(err)
    }
    store = db

    // 关闭数据库
    sigs := make(chan os.Signal)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        // 接收信号
        <-sigs
        store.Close()
        os.Exit(0)
    }()
}

var (
    // 存放ID计数器的Key
    UrlIdCounterKey = []byte("URL-ID")
    InitUrlId       = []byte("46656")
)

// newId 获取一个新的ID
func newId() (int64, error) {
    dbLock.Lock()
    defer dbLock.Unlock()

    currentIdBytes, err := store.Get(UrlIdCounterKey, readOpt)
    if err != nil {
        if err == errors.ErrNotFound {
            currentIdBytes = InitUrlId
        } else {
            return 0, err
        }
    }

    currentId, err := strconv.ParseInt(string(currentIdBytes), 10, 64)
    if err != nil {
        return 0, err
    }

    nextId := currentId + 1
    err = store.Put(UrlIdCounterKey, []byte(strconv.FormatInt(nextId, 10)), writeOpt)
    if err != nil {
        return 0, err
    }

    return nextId, nil
}

// CreateShortUrl 生成短地址
func CreateShortUrl(url string) (string, error) {
    id, err := newId()
    if err != nil {
        return "", err
    }

    urlId := strconv.FormatInt(id, 36)
    err = store.Put([]byte("s-"+urlId), []byte(url), writeOpt)

    return urlId, err
}

// FindLongUrl 查询短地址对应的长地址
func FindLongUrl(urlId string) (string, error) {
    longUrlBytes, err := store.Get([]byte("s-"+urlId), readOpt)
    if err != nil {
        return "", err
    }

    return string(longUrlBytes), nil
}
