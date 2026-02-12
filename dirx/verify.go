package dir

import (
    "errors"
    "os"
    "path/filepath"
)

var (
    ErrDirNotExist = errors.New("directory does not exist")
    ErrNotDir      = errors.New("path is not a directory")
)

// 验证目录是否存在且是否为目录
// 参数:
// dir:被验证的目录路径
// 返回值:
// string:目录的绝对路径
// error:错误
func VerifyDir(dir string) (string, error) {
    //转换为绝对路径
    dir, err := filepath.Abs(dir)
    if err != nil {
        return "", err
    }

    //clean
    //dir = filepath.Clean(dir)

    //stat
    info, err := os.Stat(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return "", ErrDirNotExist
        }
        return "", err
    }

    if !info.IsDir() {
        return "", ErrNotDir
    }

    return dir, nil
}

// 同VerifyDir,只是当发生错误的时候,直接panic(err)
func MustVerifyDir(dir string) string {
    d, err := VerifyDir(dir)
    if err != nil {
        panic(err)
    }
    return d
}
