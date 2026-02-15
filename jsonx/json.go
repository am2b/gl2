package jsonx

import (
    "os"
    "encoding/json"
)

// 将任意对象序列化成JSON格式的字节,然后写入文件中
// any:是一个类型约束(interface{}的别名),表示类型T可以是任何类型
// v:An object of any type(可以成功序列化成JSON的数据结构)
func WriteJSON(fileName string, v any) error {
    data, err := json.Marshal(v)
    if err != nil {
        return err
    }

    tmp := fileName + ".tmp"

    if err := os.WriteFile(tmp, data, 0o644); err != nil {
        return err
    }

    if err := os.Rename(tmp, fileName); err != nil {
        return err
    }

    return nil
}

func WriteJSONIndent(fileName string, v any) error {
    //"":前缀字符串(Prefix),在这里设置为空字符串,表示JSON的每一行开头没有额外的字符
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return err
    }

    tmp := fileName + ".tmp"

    if err := os.WriteFile(tmp, data, 0o644); err != nil {
        return err
    }

    return os.Rename(tmp, fileName)
}

// 从指定的JSON文件中读取内容,然后将其解析成一个T类型的对象
// 调用方法:v,err := ReadJSON[T](fileName)
func ReadJSON[T any](fileName string) (T, error) {
    var zero T

    data, err := os.ReadFile(fileName)
    if err != nil {
        return zero, err
    }

    var v T
    if err := json.Unmarshal(data, &v); err != nil {
        return zero, err
    }

    return v, nil
}
