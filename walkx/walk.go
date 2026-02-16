package walkx

import (
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

type WalkOptions struct {
    Recursive      bool     // 是否递归
    IncludeHidden  bool     // 是否包含隐藏文件/目录(以.开头的文件)
    IgnoreSuffixes []string // 忽略的后缀名,如:.log .tmp
    IgnoreDirs     []string // 忽略的目录名，精确匹配
    IgnoreFiles    []string // 忽略的文件名，精确匹配
    IgnoreKeywords []string // 文件名中包含关键字则忽略
}

type UserWalkFunc func(path string, d fs.DirEntry) error

func Walk(root string, opt WalkOptions, fn UserWalkFunc) error {
    if opt.Recursive {
        return walkRecur(root, opt, fn)
    }
    return walkNonRecur(root, opt, fn)
}

func walkNonRecur(root string, opt WalkOptions, fn UserWalkFunc) error {
    root, err := filepath.Abs(root)
    if err != nil {
        return err
    }

    // 将忽略目录转为map提高匹配效率
    ignoreDirsMap := make(map[string]struct{}, len(opt.IgnoreDirs))
    for _, d := range opt.IgnoreDirs {
        ignoreDirsMap[d] = struct{}{}
    }

    // 将忽略文件转为map提高匹配效率
    ignoreFilesMap := make(map[string]struct{}, len(opt.IgnoreFiles))
    for _, d := range opt.IgnoreFiles {
        ignoreFilesMap[d] = struct{}{}
    }

    // entries:[]os.DirEntry
    entries, err := os.ReadDir(root)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        name := entry.Name()
        mode := entry.Type()
        isSymlink := mode&fs.ModeSymlink != 0

        shouldIgnore := false

        switch {
        case mode.IsDir():
            if !opt.IncludeHidden && strings.HasPrefix(name, ".") {
                continue
            }
            if _, ok := ignoreDirsMap[name]; ok {
                continue
            }
        case mode.IsRegular():
            if !opt.IncludeHidden && strings.HasPrefix(name, ".") {
                continue
            }
            if _, ok := ignoreFilesMap[name]; ok {
                continue
            }
            for _, suf := range opt.IgnoreSuffixes {
                if strings.HasSuffix(name, suf) {
                    shouldIgnore = true
                }
            }
            for _, kw := range opt.IgnoreKeywords {
                if strings.Contains(name, kw) {
                    shouldIgnore = true
                }
            }
            if shouldIgnore {
                continue
            }
        case isSymlink:
            continue
        default:
            continue
        }

        // 只有通过了层层过滤后的项,才会最终执行用户的回调函数
        err = fn(filepath.Join(root, name), entry)
        if err != nil {
            return err
        }
    }

    return nil
}

func walkRecur(root string, opt WalkOptions, fn UserWalkFunc) error {
    root, err := filepath.Abs(root)
    if err != nil {
        return err
    }

    // 将忽略目录转为map提高匹配效率
    ignoreDirsMap := make(map[string]struct{}, len(opt.IgnoreDirs))
    for _, d := range opt.IgnoreDirs {
        ignoreDirsMap[d] = struct{}{}
    }

    // 将忽略文件转为map提高匹配效率
    ignoreFilesMap := make(map[string]struct{}, len(opt.IgnoreFiles))
    for _, d := range opt.IgnoreFiles {
        ignoreFilesMap[d] = struct{}{}
    }

    // 回调函数会从root开始,深度优先地遍历目录树
    // 每次访问到某个条目时都会调用该回调函数
    // 在回调函数中:
    // 如果返回了nil,则continue(已经处理完了这个节点(即使什么也没做),请继续处理下一个节点)
    // 如果返回了filepath.SkipDir:
    // 那么,
    // 当path是目录时,会跳过该目录以及其子树(相当于不进入该目录)
    // 如果path不是目录,SkipDir的效果是等同于返回nil
    // 如果返回其他非nil错误,该回调函数会立即停止遍历,并将该错误向上传递
    // 参数:
    // path:当前正在访问的文件或目录的绝对路径
    // d:当前的节点信息
    // err:在试图访问path时发生的错误,否则为nil
    var walkDirCallback func(path string, d fs.DirEntry, err error) error

    // 每次访问到某个条目时,都会调用回调函数
    // 该回调函数负责"过滤"
    walkDirCallback = func(path string, d fs.DirEntry, err error) error {
        // 如果回调函数在尝试读取path时遇到了错误,那么它会把这个错误作为第三个参数传进来
        if err != nil {
            return err
        }

        // 跳过根目录本身(只收集根目录下的子项)
        if path == root {
            return nil
        }

        name := d.Name()
        mode := d.Type()
        isSymlink := mode&fs.ModeSymlink != 0

        switch {
        case mode.IsDir():
            if !opt.IncludeHidden && strings.HasPrefix(name, ".") {
                return filepath.SkipDir
            }
            if _, ok := ignoreDirsMap[name]; ok {
                return filepath.SkipDir
            }
        case mode.IsRegular():
            if !opt.IncludeHidden && strings.HasPrefix(name, ".") {
                return nil
            }
            if _, ok := ignoreFilesMap[name]; ok {
                return nil
            }
            for _, suf := range opt.IgnoreSuffixes {
                if strings.HasSuffix(name, suf) {
                    return nil
                }
            }
            for _, kw := range opt.IgnoreKeywords {
                if strings.Contains(name, kw) {
                    return nil
                }
            }
        case isSymlink:
            // 跳过该符号链接(无论它指向文件还是目录)
            return fs.SkipDir
        default:
            return nil
        }

        // 只有通过了层层过滤后的项,才会最终执行用户的回调函数
        err = fn(path, d)
        if err != nil {
            return err
        }

        return nil
    }

    err = filepath.WalkDir(root, walkDirCallback)
    if err != nil {
        return err
    }

    return nil
}
