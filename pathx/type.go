package pathx

import (
    "os"
)

type PathKind uint8

const (
    PathUnknown PathKind = iota
    PathFile
    PathDir
    PathSymlink
    PathOther
)

// 判断路径的类型
func PathKindOf(path string) (PathKind, error) {
    //用Lstat获取路径本身的信息(不解析符号链接)
    info, err := os.Lstat(path)
    if err != nil {
        return PathUnknown, err
    }

    mode := info.Mode()

    switch {
    case mode&os.ModeSymlink != 0:
        return PathSymlink, nil
    case mode.IsDir():
        return PathDir, nil
    case mode.IsRegular():
        return PathFile, nil
    default:
        return PathOther, nil
    }
}
