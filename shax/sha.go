package shax

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "os"
)

func Sha256(filePath string) (string, error) {
    // 以只读的方式打开
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    // 如果放在错误处理的前面:那么程序会尝试对一个nil对象注册Close函数(如果os.Open(filePath)返回了错误,那么返回的file指针通常是nil),当函数退出时,defer执行nil.Close(),这会直接引发Panic
    // 只有当err == nil时,才说明文件描述符被成功分配了,此时再调用defer才是安全的
    defer file.Close()

    hash := sha256.New()
    // 创建一个1MB的缓冲区
    // 即使文件很小,这个buffer也只是暂占用1MB的内存,不会影响计算结果
    buf := make([]byte, 1024*1024)

    if _, err := io.CopyBuffer(hash, file, buf); err != nil {
        return "", err
    }

    return hex.EncodeToString(hash.Sum(nil)), nil
}
