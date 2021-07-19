package service

import (
    "log"
    "math/rand"
    "os"
    "os/signal"
    "path"
    "regexp"
    "strconv"
    "sync"
    "syscall"

    "github.com/pkg/errors"
    "github.com/syndtr/goleveldb/leveldb"
    leveldbErrors "github.com/syndtr/goleveldb/leveldb/errors"
    "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
    store  *leveldb.DB
    dbLock sync.Mutex
)

var UrlIdRegexp = regexp.MustCompile("[^0-9a-z]+$")

// 随机偏移
var randomOffset = [100]int64{
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
    4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43,
}

var (
    readOpt  = &opt.ReadOptions{DontFillCache: false}
    writeOpt = &opt.WriteOptions{Sync: false, NoWriteMerge: false}
)

func getCacheKey(urlId string) string {
    return "s-" + urlId
}

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
        if err == leveldbErrors.ErrNotFound {
            currentIdBytes = InitUrlId
        } else {
            return 0, errors.Wrap(err, "Get last urlId error")
        }
    }

    currentId, err := strconv.ParseInt(string(currentIdBytes), 10, 64)
    if err != nil {
        return 0, errors.WithStack(err)
    }

    offset := randomOffset[rand.Intn(len(randomOffset))]
    nextId := currentId + offset
    err = store.Put(UrlIdCounterKey, []byte(strconv.FormatInt(nextId, 10)), writeOpt)
    if err != nil {
        return 0, errors.WithStack(err)
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
    log.Printf("[%s <=> %s]", urlId, url)
    err = store.Put([]byte(getCacheKey(urlId)), []byte(url), writeOpt)

    return urlId, errors.Wrap(err, "Set shortUrl error")
}

// FindLongUrl 查询短地址对应的长地址
func FindLongUrl(urlId string) (string, error) {
    longUrlBytes, err := store.Get([]byte(getCacheKey(urlId)), readOpt)
    if err != nil {
        return "", errors.Wrap(err, "Get longUrl error")
    }

    return string(longUrlBytes), nil
}
