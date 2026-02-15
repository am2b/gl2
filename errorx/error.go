package errorx

import (
    "fmt"
    "os"
    "errors"
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
