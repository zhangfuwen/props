# props: Go (golang) library for handling Java-style property files

This library provides compatibility with Java property files for Go.

There are two main types provided:
* `Properties` - read and write property files in Java format
* `Expander` - replaces property references wrapped by '${}' at runtime (as 
found in Ant/Log4J/JSP EL/Spring)

The full Java property file format (including all comment types, line 
continuations, key-value separators, unicode escapes, etc.) is supported.


## 新特性
github.com/rickar/props 支持文件的读写但有两个不足:
1. 无顺序，读一文件，修改一些值后写入（另一个）文件，文件中各属性的顺序全乱了。本质原因是用map[string]string存储的
2. 不记录注释，读完之后再写入（另一个）文件，注释全没了

我尝试把map[string]string改用order_map代替。并记录了注释，但是还没有时间测试。
原作者的test都是过的，所以原有功能应该没有打破

发现问题可以给我的提issue，一定要写详细些。
直接修复了然后提pull request更好。

你也可以尝试用用我的github.com/zhangfuwen/property, 这个库进一步把数据直接解析到结构体里.

