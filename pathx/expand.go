package pathx

import (
    "os"
    "path/filepath"
    "strings"
)

// 替换路径中的~和环境变量,然后clean路径,并将路径转换为绝对路径
// 参数:
// path:表示路径的字符串,如果为空:""的话,则表示当前目录:.
// 返回值:
// string:绝对路径
// error:错误
func ExpandPath(path string) (string, error) {
    //展开环境变量
    path = os.ExpandEnv(path)

    //展开开头的~或~/
    if path == "~" || strings.HasPrefix(path, "~/") {
        home, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }

        if path == "~" {
            path = home
        } else {
            path = filepath.Join(home, path[2:])
        }
    }

    //clean
    //Clean不理解shell语义,Clean是纯路径算法,不理解~,不理解$VAR
    //filepath.Clean("") == "."
    //clean := filepath.Clean(path)

    //转为绝对路径
    abs, err := filepath.Abs(path)
    if err != nil {
        return "", err
    }

    return abs, nil
}

// 同ExpandPath,只是当发生错误的时候,直接panic(err)
func MustExpandPath(path string) string {
    path, err := ExpandPath(path)
    if err != nil {
        panic(err)
    }
    return path
}
