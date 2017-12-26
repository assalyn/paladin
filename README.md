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
