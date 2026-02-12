package path

import (
    "os"
    "strings"
    "path/filepath"
    "errors"
    "runtime"
)

// 判断路径是否是一个普通文件
func IsRegularFile(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        return false, err
    }
    return info.Mode().IsRegular(), nil
}

// 判断路径是否是一个目录
func IsDir(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        return false, err
    }
    return info.IsDir(), nil
}

// 判断两个path是否一致,没有解析符号连接
func IsEqual(a, b string) (bool, error) {
    pa, err := filepath.Abs(a)
    if err != nil {
        return false, err
    }
    pb, err := filepath.Abs(b)
    if err != nil {
        return false, err
    }

    //pa = filepath.Clean(pa)
    //pb = filepath.Clean(pb)

    if runtime.GOOS == "windows" {
        //EqualFold:不区分大小写的字符串比较
        return strings.EqualFold(pa, pb), nil
    }
    return pa == pb, nil
}

// 判断subPath是否是parentPath的逻辑子路径(无需路径实际存在)
// 没有处理~和环境变量,需要用户自己提前处理
func IsSubPath(parentPath, subPath string) (bool, error) {
	if parentPath == "" || subPath == "" {
		return false, errors.New("empty path")
	}

    //即使路径实际不存在,只要格式合法,filepath.Abs依然能返回基于当前工作目录的绝对路径
	parent, err := filepath.Abs(parentPath)
	if err != nil {
		return false, err
	}
	sub, err := filepath.Abs(subPath)
	if err != nil {
		return false, err
	}

	//parent = filepath.Clean(parent)
	//sub = filepath.Clean(sub)

    //Windows卷标校验
	if filepath.VolumeName(parent) != filepath.VolumeName(sub) {
		return false, nil
	}

    //计算sub相对于parent的相对路径
    //返回的rel是从parent到sub所需的路径
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false, err
	}

	// "." => same path (not sub)
	if rel == "." {
		return false, nil
	}

	// starts with ".." => outside
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return false, nil
	}

	return true, nil
}
