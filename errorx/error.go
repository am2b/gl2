package errorx

import (
    "errors"
    "fmt"
    "os"
)

// 包装错误
// 用于:基础设施层,业务层
func Wrap(op string, err error) error {
    return fmt.Errorf("%s: %w", op, err)
}

// 生成错误
// 用于:基础设施层,业务层
func New(op string) error {
    return errors.New(op)
}

// 用于:展示层
func Fatal(err error) {
    if err == nil {
        return
    }

    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
}

// 底层:只负责构造错误
// 顶层:决定如何展示错误
