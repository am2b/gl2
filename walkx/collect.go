package walkx

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "errors"
)

// 符号链接不是常规文件,也不是目录
type ReadDirItem struct {
    Name      string `json:"name"`
    IsDir     bool   `json:"isdir"`
    IsSymlink bool   `json:"issymlink"`
    FullPath  string `json:"fullpath"`
}

// 非递归
// 包含隐藏
// 不解析符号链接
// 收集子目录,文件,符号链接
// 参数:
// root:根目录
// 返回值:
// 切片,其元素是结构体:ReadDirItem
// 错误
func ReadDir(root string) ([]ReadDirItem, error) {
    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }

    //entries:[]os.DirEntry
    entries, err := os.ReadDir(root)
    if err != nil {
        return nil, err
    }

    items := make([]ReadDirItem, 0, len(entries))

    //entry:os.DirEntry
    for _, entry := range entries {
        item := ReadDirItem{
            Name:     entry.Name(),
            FullPath: filepath.Join(root, entry.Name()),
        }

        mode := entry.Type()

        if mode.IsDir() {
            item.IsDir = true
        }
        if mode&os.ModeSymlink != 0 {
            item.IsSymlink = true
        }

        items = append(items, item)
    }

    return items, nil
}

type WalkDirItem struct {
    Name      string `json:"name"`
    IsSymlink bool   `json:"issymlink"`
    FullPath  string `json:"fullpath"`
}

// 递归
// 包含隐藏
// 不解析符号链接
// 收集文件,符号链接
// 参数:
// root:根目录
// 返回值:
// 切片,其元素是结构体:WalkDirItem
// 错误
func WalkDir(root string) ([]WalkDirItem, error) {
    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }

    var items []WalkDirItem

    //回调函数
    //path:当前正在访问的文件或目录的绝对路径
    //d:当前节点的信息
    //err:如果在访问该文件或目录过程中发生了错误(例如权限不足),该参数将包含该错误,否则为nil
    //每次访问到某个条目(目录,文件或符号链接),都会调用回调函数
    err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // 跳过根目录本身(只收集根目录下的子项)
        if path == root {
            //返回nil:已经处理完这个节点(即使什么也没做),请继续处理下一个节点
            //如果回调返回其他非nil错误,filepath.WalkDir会立即停止遍历
            return nil
        }

        //想要跳过某个项时,返回filepath.SkipDir即可
        //if d.IsDir() && d.Name() == ".git" {
        //     return filepath.SkipDir
        // }

        //不收集目录
        mode := d.Type()

        if mode.IsDir() {
            return nil
        }
        items = append(items, WalkDirItem{
            Name:      d.Name(),
            IsSymlink: mode&fs.ModeSymlink != 0,
            FullPath:  path,
        })

        // 返回nil继续遍历下一个条目
        return nil
        //end 回调函数
    })

    if err != nil {
        return nil, err
    }

    return items, nil
}

type WalkDirFollowSymlinksItem struct {
    Name     string `json:"name"`
    FullPath string `json:"fullpath"`
}

// 递归
// 包含隐藏
// 解析符号链接
// 收集文件
// 参数:
// root:根目录
// 返回值:
// 切片,其元素是结构体:WalkDirEvalSymlinksItem
// 错误
func WalkDirFollowSymlinks(root string) ([]WalkDirFollowSymlinksItem, error) {
    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }

    // visited:记录已访问的路径,防止符号链接循环
    visited := make(map[string]struct{})
    // 初始根目录加入已访问集合，避免符号链接指向根目录导致循环
    visited[root] = struct{}{}

    var results []WalkDirFollowSymlinksItem

    var walk func(string) error
    walk = func(current string) error {
        //读取当前目录下的所有条目(包含隐藏文件,os.ReadDir默认返回所有条目)
        entries, err := os.ReadDir(current)
        if err != nil {
            return err
        }

        for _, entry := range entries {
            fullPath := filepath.Join(current, entry.Name())
            //获取条目类型(文件/目录/符号链接等)
            mode := entry.Type()

            // 1,处理符号链接
            if mode&fs.ModeSymlink != 0 {
                // 解析符号链接的真实目标路径
                target, err := filepath.EvalSymlinks(fullPath)
                if err != nil {
                    return err
                }

                target, err = filepath.Abs(target)
                if err != nil {
                    return err
                }

                // 检查目标路径是否已访问,防止循环
                if _, seen := visited[target]; seen {
                    continue
                }
                // 标记目标路径为已访问
                visited[target] = struct{}{}

                // 获取目标路径的文件信息
                info, err := os.Stat(target)
                if err != nil {
                    return err
                }

                // 若目标是目录:递归遍历该目录
                if info.IsDir() {
                    // 进入符号链接目录
                    if err := walk(target); err != nil {
                        return err
                    }
                } else {
                    // 若目标是文件:收集文件信息
                    results = append(results, WalkDirFollowSymlinksItem{
                        //目标文件的名称
                        Name:     info.Name(),
                        FullPath: target,
                    })
                }

                //符号链接处理完毕,跳过后续逻辑
                continue
            } //end 检查符号链接

            // 2,处理目录(非符号链接)
            if mode.IsDir() {
                //检查当前目录路径是否已访问(防止循环)
                if _, seen := visited[fullPath]; seen {
                    continue
                }
                visited[fullPath] = struct{}{}
                // 递归遍历子目录
                if err := walk(fullPath); err != nil {
                    return err
                }
                continue
            }

            //3,处理普通文件(非符号链接,非目录)
            results = append(results, WalkDirFollowSymlinksItem{
                Name:     entry.Name(),
                FullPath: fullPath,
            })
        }

        return nil
    }

    //启动递归遍历(从根目录开始)
    if err := walk(root); err != nil {
        return nil, err
    }

    return results, nil
}

type ReadDirItemExt struct {
    Name     string `json:"name"`
    FullPath string `json:"fullpath"`
}

// 非递归
// 包含隐藏
// 不解析符号链接
// 收集指定扩展名的文件(不包含目录与符号链接),扩展名大小写不敏感
// 参数:
// root:根目录
// exts:允许的文件扩展名(不区分大小写,如:[]string{".jpg",".png"})(扩展名可以不带.)
// 返回值:
// 切片,元素类型为ReadDirItemExt
// 错误
func ReadDirExt(root string, exts []string) ([]ReadDirItemExt, error) {
    if len(exts) == 0 {
        return nil, fmt.Errorf("extensions list cannot be empty")
    }

    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }

    info, err := os.Stat(root)
    if err != nil {
        return nil, err
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("not a directory: %s", root)
    }

    // 预处理扩展名
    extSet := make(map[string]struct{}, len(exts))
    for _, e := range exts {
        e = strings.ToLower(e)
        if !strings.HasPrefix(e, ".") {
            e = "." + e
        }
        extSet[e] = struct{}{}
    }

    entries, err := os.ReadDir(root)
    if err != nil {
        return nil, err
    }

    results := make([]ReadDirItemExt, 0, len(entries))

    for _, entry := range entries {
        // 排除目录和符号链接
        if entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
            continue
        }

        name := entry.Name()

        // 检查后缀
        //filepath.Ext(name):提取扩展名
        ext := strings.ToLower(filepath.Ext(name))
        if _, ok := extSet[ext]; !ok {
            continue
        }

        results = append(results, ReadDirItemExt{
            Name:     name,
            FullPath: filepath.Join(root, name),
        })
    }

    return results, nil
}

// 递归
// 包含隐藏
// 不解析符号链接
// 收集指定扩展名的文件(不包含目录与符号链接),扩展名大小写不敏感
// 参数:
// root:根目录
// exts:允许的文件扩展名(不区分大小写,如:[]string{".jpg",".png"})(扩展名可以不带.)
// fn:用户回调函数
// 返回值:
// 错误
func WalkDirExt(root string, exts []string, fn func(path string, info fs.FileInfo) error,) error {
    if len(exts) == 0 {
        return errors.New("extensions list cannot be empty")
    }

    root, err := filepath.Abs(root)
    if err != nil {
        return err
    }

    info, err := os.Stat(root)
    if err != nil {
        return err
    }
    if !info.IsDir() {
        return errors.New("not a directory")
    }

    // 预处理扩展名
    extSet := make(map[string]struct{})
    for _, e := range exts {
        e = strings.ToLower(e)
        if !strings.HasPrefix(e, ".") {
            e = "." + e
        }
        extSet[e] = struct{}{}
    }

    // 这里的path就已经是绝对路径了
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
        if walkErr != nil {
            return err
        }

        // 跳过根目录自身
        if path == root {
            return nil
        }

        // 跳过目录
        if d.IsDir() {
            return nil
        }

        // 跳过符号链接
        if d.Type()&fs.ModeSymlink != 0 {
            return nil
        }

        // 检查文件扩展名
        ext := strings.ToLower(filepath.Ext(d.Name()))
        if _, ok := extSet[ext]; !ok {
            return nil
        }

        // 获取完整文件信息
        info, err := d.Info()
        if err != nil {
            return err
        }

        // 注意:对于每一个走到这里的文件,调用一次回调函数,n个文件,就会调用n次该回调函数
        return fn(path, info)
    })
}

type WalkDirExtFollowSymlinksItem struct {
    Name     string `json:"name"`
    FullPath string `json:"fullpath"`
}

// 递归
// 包含隐藏
// 解析符号链接
// 收集指定扩展名的文件(不包含目录与符号链接),扩展名大小写不敏感
// 参数:
// root:根目录
// exts:允许的文件扩展名(不区分大小写,如:[]string{".jpg",".png"})(扩展名可以不带.)
// 返回值:
// 切片,元素类型为WalkDirExtFollowSymlinksItem
// 错误
func WalkDirExtFollowSymlinks(root string, exts []string) ([]WalkDirExtFollowSymlinksItem, error) {
    if len(exts) == 0 {
        return nil, fmt.Errorf("extensions list cannot be empty")
    }

    root, err := filepath.Abs(root)
    if err != nil {
        return nil, err
    }

    info, err := os.Stat(root)
    if err != nil {
        return nil, err
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("not a directory: %s", root)
    }

    // 预处理扩展名
    extSet := make(map[string]struct{}, len(exts))
    for _, e := range exts {
        e = strings.ToLower(e)
        if !strings.HasPrefix(e, ".") {
            e = "." + e
        }
        extSet[e] = struct{}{}
    }

    var results []WalkDirExtFollowSymlinksItem

    visited := make(map[string]struct{})

    var walk func(string) error
    walk = func(current string) error {
        realPath, err := filepath.EvalSymlinks(current)
        if err != nil {
            return err
        }

        realPath, err = filepath.Abs(realPath)
        if err != nil {
            return err
        }

        // 防止循环
        if _, ok := visited[realPath]; ok {
            return nil
        }
        visited[realPath] = struct{}{}

        entries, err := os.ReadDir(realPath)
        if err != nil {
            return err
        }

        for _, entry := range entries {
            fullPath := filepath.Join(realPath, entry.Name())

            mode := entry.Type()

            // 处理符号链接
            if mode&fs.ModeSymlink != 0 {
                target, err := filepath.EvalSymlinks(fullPath)
                if err != nil {
                    return err
                }

                target, err = filepath.Abs(target)
                if err != nil {
                    return err
                }

                targetInfo, err := os.Stat(target)
                if err != nil {
                    return err
                }

                if targetInfo.IsDir() {
                    if err := walk(target); err != nil {
                        return err
                    }
                } else {
                    ext := strings.ToLower(filepath.Ext(targetInfo.Name()))
                    if _, ok := extSet[ext]; ok {
                        results = append(results, WalkDirExtFollowSymlinksItem{
                            Name:     targetInfo.Name(),
                            FullPath: target,
                        })
                    }
                }

                continue
            }

            // 普通目录
            if entry.IsDir() {
                if err := walk(fullPath); err != nil {
                    return err
                }
                continue
            }

            // 普通文件
            ext := strings.ToLower(filepath.Ext(entry.Name()))
            if _, ok := extSet[ext]; !ok {
                continue
            }

            results = append(results, WalkDirExtFollowSymlinksItem{
                Name:     entry.Name(),
                FullPath: fullPath,
            })
        }

        return nil
    }

    // 启动遍历
    if err := walk(root); err != nil {
        return nil, err
    }

    return results, nil
}
