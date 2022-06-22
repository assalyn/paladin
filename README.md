# paladin
这个项目是受以前的策划数据解析程序启发而开发的新一代策划数据解析工具。

# 设计核心
xlsx为核心，代码全解耦，以策划友好性为第一位

# 需求
1. enum提取，能将策划表中的常量转换成数字。
2. 可选的合表/不合表。比如 装备下面有子类：刀，剑，鞋。从策划角度，他们最好能放在xlsx的一张工作簿的不同子表中，但是从程序的角度，他们是最好能合表的。
3. 最好能根据xlsx生成前端(c#)/后端(go)的桩文件，即get函数（这个有点难？）。
4. 可以指定是否水平读取，比如全局配置表和装备表的模式完全不同，要能支持.
5. 要能直接根据xlsx生成数据结构，最多一个config.toml作为全局配置文件，不能每个文件一个配置文件。
6. 像编译系统一样，单次尽可能多尽可能详细的报出错误。
7. 参数展开，可将{0}, {1}这样的参数替换成参数表中内容。
8. 多语言支持，水平扩展，生成到客户端文件夹locale/cn, en, kr这些文件夹下。客户端根据设置加载语言选项

# 初步设计
1. 以xlsx为数据核心，以json为数据导出格式
2. 可适配自动生成不同语言的桩代码，避免因改动xlsx导致的差错。(代码生成）
3. 一张工作簿对应一个json文件. 一个格式，格式对不上报错

# 局限性
1. 不支持slice和map的多级结构. 后面只能接简单数据结构或单层struct
2. 多层struct是允许的，但是不允许多层struct的member分散出现，也不允许struct内部再跟slice或map
3. 做slice和map时，要求使用的名称为单一字符串如[awake]或驼峰字符串如[awakeMaterial]否则读取会失败
4. enum.xlsx枚举表特殊，没有复杂结构, 因此第四行不需要留白，可存储数据

# 表达式
* \<struct> 尖括号用于表达子struct
* [slice] 方括号用于表达slice
* {map} 大括号用于表达map
* \- 横杠用于略去解析（多用于备注，说明表单内容，但不输出）
* 同一slice/map的不同member，使用[slice]#1 字段描述，比如slice类型有type和value两个字段，则这两个字段的desc栏都填[slice]#1, 第二个value的desc栏填[slice]#2

# 代码生成规则
* 使用字段名的全小写输出为json字段，使用此字段的驼峰式作为字段类名. 比如spinType, 输出的json文件类型为"spintype"，输出的类成员名为Spintype

# 配置文件配置项
* workbook   string     // 工作簿
* sheet      []string   // 子表
* horizontal bool       // 是否水平解析
* output     []string   // 输出类型选项 json, cs, go. 默认全输出
* type       string     // 表类型 server_only; client_local_read
* duplicate  bool       // 重复结构, 设置为true, 避免重复解析出现的校验错误
* enums      []EnumItem // 枚举替换
