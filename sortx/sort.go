package sort

import (
    "golang.org/x/text/collate"
    "golang.org/x/text/language"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "unicode"
)

//====================排序辅助函数=====================
/*
判断一个分块字符串s是否是数字块
因为正则表达式保证了,如果一个块以数字开头(\d+),那么这个块中的所有字符都必须是数字
*/
func isDigit(s string) bool {
    //字符串是否非空,并且第一个字符是否在'0'到'9'的范围内
    return len(s) > 0 && s[0] >= '0' && s[0] <= '9'
}

/*
检查块中的所有字符(rune)是否是汉字
注意:中文标点的Unicode码值不在汉字范围内
*/
func isChinese(s string) bool {
    for _, r := range s {
        //检查r是否属于Unicode的"Han"字符集(即汉字)
        //r < 0x4E00 || r > 0x9FFF(最常用的汉字范围0x4E00到0x9FFF,约2万多个)
        if !unicode.Is(unicode.Han, r) {
            return false
        }
    }
    return true
}

//=======================数值排序=====================
/*
作用:把一个字符串分割成数字块和非数字块交替的序列,例如,"file10.txt"会被分割成["file", "10", ".txt"]
\d+:匹配一个或多个数字(数字块)
|:或
\D+:匹配一个或多个非数字字符(文本块)
*/
var reChunkForNaturalLess = regexp.MustCompile(`(\d+|\D+)`)

/*
自然排序
稳定排序
数字按数值比较,注意:"file-07"和"file-7"被视为是相等的,它们在排序后的相对顺序将与它们在原始切片中的顺序保持一致(除非07,7后面还有不同的内容)
将字符串分解成数字块和非数字块,然后对数字块进行数值比较,而不是按字符逐个比较,这解决了传统字典序排序中"file10"排在"file2"之前的错误
中文,英文,符号按正常Unicode排序

注意:在比较的过程中,这个函数会被多次调用

参数:
a,b:要比较的两个字符串
ignoreCase:可忽略大小写
*/
func NaturalLess(a, b string, ignoreCase bool) bool {
    if ignoreCase {
        a, b = strings.ToLower(a), strings.ToLower(b)
    }

    //将输入字符串a和b分解成数字块和非数字块的交替序列
    aChunks := reChunkForNaturalLess.FindAllString(a, -1)
    bChunks := reChunkForNaturalLess.FindAllString(b, -1)

    //迭代比较aChunks和bChunks中对应位置的块,直到其中一个序列结束
    for i := 0; i < len(aChunks) && i < len(bChunks); i++ {
        ca, cb := aChunks[i], bChunks[i]

        //如果当前块ca和cb都是数字块:
        if isDigit(ca) && isDigit(cb) {
            //strings.TrimLeft(ca, "0"):处理前导0,确保"007"和"7"被视为相等的数值,然后使用strconv.Atoi将其转换为整数
            na, _ := strconv.Atoi(strings.TrimLeft(ca, "0"))
            nb, _ := strconv.Atoi(strings.TrimLeft(cb, "0"))
            //如果数值不相等
            if na != nb {
                //数值小的排在前面(使"2"排在"10"之前)
                return na < nb
            }
            //如果数值相等(例如"01"和"1"),则continue,比较下一个块
            continue
        }

        //如果不是两个数字块的比较(文本块之间的比较,或一个是数字块一个是文本块),则直接按字符串字典序比较
        if ca != cb {
            return ca < cb
        }
    }

    //如果循环结束,意味着两个字符串在所有公共的块上都相等
    //那么就认为较短的那个字符串(分块数量少)是小的,应该排在前面,例如,"file2"应该排在"file2.txt"的前面
    return len(aChunks) < len(bChunks)
}

func NaturalSort(slice []string, ignoreCase bool) {
    //稳定排序:保持相等元素的原始相对顺序
    sort.SliceStable(slice, func(i, j int) bool {
        return NaturalLess(slice[i], slice[j], ignoreCase)
    })
}

//==================数值和拼音排序===========================
/*
作用:把一个字符串分成3种类型的块:数字,中文,其它
\p{Han}+:匹配一个或多个汉字,\p{Han}是一个Unicode属性,代表所有汉字(包括简体,繁体,日文汉字等)
\d+:匹配一个或多个数字
[^\p{Han}\d]+:匹配一个或多个既不是汉字也不是数字的任何其他字符(如英文字母,标点符号,特殊符号等)
*/
var reChunkForNaturalLessWithPinyin = regexp.MustCompile(`([\p{Han}]+|\d+|[^\p{Han}\d]+)`)

/*
自然排序
稳定排序
数字按数值比较,注意:"file-07"和"file-7"被视为是相等的,它们在排序后的相对顺序将与它们在原始切片中的顺序保持一致(除非07,7后面还有不同的内容)
中文按拼音
英文,符号按正常Unicode排序

注意:这个"支持拼音"的版本是完全向后兼容的,即:
没有汉字:和普通自然排序效果完全一致
有汉字:自动启用拼音排序逻辑
注意:在比较的过程中,这个函数会被多次调用

参数:
a,b:要比较的两个字符串
ignoreCase:可忽略大小写
*/
func NaturalLessWithPinyin(a, b string, ignoreCase bool) bool {
    // 拼音排序器
    var zhCollator = collate.New(language.Chinese)

    if ignoreCase {
        a, b = strings.ToLower(a), strings.ToLower(b)
    }

    aChunks := reChunkForNaturalLessWithPinyin.FindAllString(a, -1)
    bChunks := reChunkForNaturalLessWithPinyin.FindAllString(b, -1)

    for i := 0; i < len(aChunks) && i < len(bChunks); i++ {
        ca, cb := aChunks[i], bChunks[i]

        //数字块
        if isDigit(ca) && isDigit(cb) {
            na, _ := strconv.Atoi(strings.TrimLeft(ca, "0"))
            nb, _ := strconv.Atoi(strings.TrimLeft(cb, "0"))
            if na != nb {
                return na < nb
            }
            //数值相等(如:"007" vs "7"),则不能确定顺序,必须继续比较下一个块
            continue
        }

        //中文块
        if isChinese(ca) && isChinese(cb) {
            if cmp := zhCollator.CompareString(ca, cb); cmp != 0 {
                //按拼音比较
                return cmp < 0
            }
            //拼音相同,比较下一个块
            continue
        }

        //其他块或混合块(如"abc"vs"def")或(如"123"vs"abc")等
        /*
           比较回退到默认的字典排序:
           "1" (Unicode 49) < "A" (Unicode 65)
           "A" (Unicode 65) < "a" (Unicode 97)
           "a" (Unicode 97) < "中" (Unicode 20013)
           因此,当混合类型比较时,优先级是:数字>英文/符号>中文(即数字排最前)
        */
        if ca != cb {
            return ca < cb
        }
    }

    //前面都相同那么就短的更小
    return len(aChunks) < len(bChunks)
}

func NaturalSortWithPinyin(slice []string, ignoreCase bool) {
    sort.SliceStable(slice, func(i, j int) bool {
        return NaturalLessWithPinyin(slice[i], slice[j], ignoreCase)
    })
}

// 中文标点 -> 英文标点
var punctMap = map[string]string{
    "，":  ",",   // 逗号
    "。":  ".",   // 句号
    "！":  "!",   // 感叹号
    "？":  "?",   // 问号
    "：":  ":",   // 冒号
    "；":  ";",   // 分号
    "‘":  "'",   // 左单引号
    "’":  "'",   // 右单引号
    "“":  "\"",  // 左双引号
    "”":  "\"",  // 右双引号
    "（":  "(",   // 左圆括号
    "）":  ")",   // 右圆括号
    "【":  "[",   // 左方括号
    "】":  "]",   // 右方括号
    "「":  "[",   // 左书名号(竖排引号)
    "」":  "]",   // 右书名号(竖排引号)
    "『":  "[",   // 左双书名号(竖排引号)
    "』":  "]",   // 右双书名号(竖排引号)
    "《":  "<",   // 左书名号(双角号)
    "》":  ">",   // 右书名号(双角号)
    "、":  ",",   // 顿号
    "～":  "~",   // 波浪号
    "　":  " ",   // 全角空格
    "——": "-",   // 破折号
    "…":  "...", // 省略号
}

// 替换字符串中的中文标点为英文标点
func ReplaceChinesePunct(s string) string {
    for zh, en := range punctMap {
        s = strings.ReplaceAll(s, zh, en)
    }
    return s
}
