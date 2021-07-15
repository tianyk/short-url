package utils

import "log"

func init()  {
    // 设置日志格式
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}