package path

import (
    "os"
    "path/filepath"
)

// 解析给定路径的符号链接,返回它最终指向的真实绝对路径,如果路径不是符号链接,也返回绝对路径
func ResolveSymlink(path string) (string, error) {
    //获取文件信息(不跟随符号链接)
    info, err := os.Lstat(path)
    if err != nil {
        return "", err
    }

    var resolved string

    if info.Mode()&os.ModeSymlink != 0 {
        resolved, err = filepath.EvalSymlinks(path)
        if err != nil {
            return "", err
        }
    } else {
        resolved = path
    }

    abs, err := filepath.Abs(resolved)
    if err != nil {
        return "", err
    }

    //clean
    //abs = filepath.Clean(abs)

    return abs, nil
}
