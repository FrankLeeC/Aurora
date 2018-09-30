# Aurora



[TOC]

## Config

读取配置文件。读取当前路径下的app.conf文件。

可指定多种运行环境，指定RunMode为当前环境，即可切换运行环境。

未指定运行环境的配置项在全局域生效。

```properties
RunMode=dev   # config.GetRunMode()

# ----------------------------------------------------global
foo=bar   # config.GetString("foo")
abc=123   # config.GetInt("abc")

# -------------------------------------------------dev
[dev]
port=9090   # config.GetString("port")
abc=234  # config.GetInt("abc")  会覆盖global>abc
fzz=eval(15*24+90/19-7)  # config.GetEval("fzz") 可自动计算，仅支持简单的四则运算，不包含括号运算
fxx=eval(5/3*3)  # config.GetEval("fxx") 输出为 5， 乘除一起时优先乘法

[[mysql]]
defaultPagesize=5   # config.GetInt("mysql>defaultPagesize")
[[[source1]]]
uri=dev:123456@tcp(127.0.0.1:3306)/aurora?charset=utf8&loc=Local  # config.GetString("mysql>source1")
useSomething=true  # config.GetBool("mysql>source1>useSomething") It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.

[[[source2]]]
uri=dev:123456@tcp(192.168.173.11:3306)/aurora?charset=utf8&loc=Local  # config.GetString("mysql>source2>uri")

[[other]]
...


# --------------------------------------------------------test
[test]
port=9099
fzz=eval(10*24+90/19-7)
fxx=eval(5/3*3)  # 输出为 5， 乘除一起时优先乘法

[[mysql]]
defaultPagesize=10
[[[source1]]]
uri=test:123456@tcp(127.0.0.1:3306)/aurora?charset=utf8&loc=Local
useSomething=False  # It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.

[[[source2]]]
uri=test:123456@tcp(192.168.173.11:3306)/aurora?charset=utf8&loc=Local

[[other]]
...

# ---------------------------------------------------------------prod
[prod]
port=80
fzz=eval(5*24+90/19-7)
fxx=eval(5/3*3)  # 输出为 5， 乘除一起时优先乘法

[[mysql]]
defaultPagesize=20
[[[source1]]]
uri=prod:123456@tcp(127.0.0.1:3306)/aurora?charset=utf8&loc=Local
useSomething=1  # It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.

[[[source2]]]
uri=prod:123456@tcp(192.168.173.11:3306)/aurora?charset=utf8&loc=Local


[[other]]
...

```

***

## JobQueue

用于控制并发量的任务队列

```Go
jq := job.NewJobQueue(maxWorker, rate)  // maxPoolSize = rate * maxWorker
for i := 0; i < N; i++ {
    jq.Submit(task)
}

type task struct {}

func(t *task) Work() {}
```

当前等待执行任务数量达到```maxPoolSize```时，```Submit()```会阻塞。

当```JobQueue```关闭后，```Submit()```不会提交任务。

```jq.Close()```方法不是协程安全的！

***

## Log

日志模块

日志等级：Trace < Info < Warn < Error < Fatal

```Go
const (

	// Trace trace level
	Trace = iota

	// Info info level
	Info

	// Warn warn level
	Warn

	// Error error level
	Error

	// Fatal fatal level
	Fatal
)

const (

	// Day 2006-01-02
	Day = 1 << iota
	// Time 15:04:05
	Time
	// Lfile full path of file
	Lfile
	// Sfile file name only
	Sfile
	// Std Day|Time|Sfile
	Std
)

type LoggerOption struct {
	Level    int  // 级别
	MaxLine  int  // 最大行
	MaxSize  int  // 最大文件大小（字节数）
	Mode     int  // 格式
	Compress bool  // 是否压缩
	LeftDay  int  // 如果压缩，保留近LeftDay天不压缩
}

logger := log.NewLogger("./logs/app.log", &log.LoggerOption{...})  // 绝对路径和相对路径都可以
logger := log.NewLogger("./logs/app2.log", &log.LoggerOption{
    Level: log.Warn,
    MaxLine: 5000,
    MaxSize: 999999,
    Mode: log.Time|log.Lfile,
    Compress: true,
    LeftDay: 3,
})
```

***

## ORM

orm工具。

```SQL
CREATE TABLE T_TEST(
    A INTEGER PRIMARY KEY AUTO_INCREMENT,
    B VARCHAR(100),
    C TIMESTAMP
);
```

```Go
type Test struct {
    Fa int `column:"A"`
    Fb string `column:"B"`
    Fc time.Time `column:"C"`
}
```



单条记录插入

```GO
-- 两个横线表示注释，注释必须独占一行，不可以在SQL后!
--- [Insert]
INSERT INTO T_TEST (B, C) VALUES (#item.Fb#, #item.Fc#)

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
t := Test{...}
m := orm.M{s"item": t}
count, err := orm.InitGQL().Use("Insert").M(m).Insert().Result()
lastID, err := orm.InitGQL().Use("Insert").M(m).Insert().ReturnLastID().Result()
```

批量插入

```Go
--- [BatchInsert]
INSERT INTO T_TEST (B, C)
VALUES
--- range [values, item]
(#item.Fb#, #item.Fc#)
--- endrange


orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
arr := make([]Test, 0, N)
... // append some instance
m := orm.M{"values": arr}
count, err := orm.InitGQL().Use("BatchInsert").M(m).Insert().Result()  // 默认使用default数据源
count, err := orm.InitGQL().UseDatasource("xxx").Use("BatchInsert").M(m).Insert().Result()

lastID, err := orm.InitGQL().Use("BatchInsert").M(m).Insert().ReturnLastID().Result()
```

查询单个

```Go
--- [SelectOne]
SELECT * FROM T_TEST LIMIT 0, 1

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
var rs Test
count, err := orm.InitGQL().Use("SelectAll").One(&rs)
```

条件查询

```go
--- [SelectIfNotNil]
SELECT * FROM T_TEST
WHERE B = '123'
-- 如果time字段不为空
--- ifnotnil [time]
AND C = #time#
--- endif

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
var rs []Test
m := orm.M{"time": time.Time...}
// m := orm.M{}
count, err := orm.InitGQL().Use("SelectIfNotNil").M(m).All(&rs)
```

条件查询

```go
--- [SelectIf]
SELECT * FROM T_TEST
WHERE B = '123'
-- 运算符支持 ==, !=, >, >=, <, <=; 数据类型支持bool, int, string, float等,字符串类型不需要写引号。具体的数据类型由变量'f'决定。
--- if [f == true]
AND A > #id#
--- endif

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
var rs []Test
m := orm.M{"f": true, "id": 3}
// m := orm.M{"f": false}
count, err := orm.InitGQL().Use("SelectIf").M(m).All(&rs)
```

条件查询动态where

```go
--- [SelectWhere]
SELECT * FROM T_TEST
--- where
--- if [id < 5]
A = #id#
--- endif
--- ifnotnil [b]
B = #b#
--- endif
--- endwhere

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
var rs []Test
m := orm.M{"id": 2, "b": "123"}
// m := orm.M{"id": 7, "b": "999"}
// m := orm.M{"id": 2}
// m := orm.M{"id": 7}
count, err := orm.InitGQL().Use("SelectWhere").M(m).All(&rs)
```

删除

```go
--- [DeleteCondition]
DELETE FROM T_TEST
WHERE A IN
--- range [values, item, (, )]
#item#
--- endrange

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
m := orm.M{"values": [...]int{1, 2, 3, 4, 5}}
count, err := orm.InitGQL().Use("DeleteCondition").M(m).Delete().Result()
```

复杂的动态where

```Go
--- [UpdateCondition]
UPDATE T_TEST SET B = #b#
--- ifnotnil [c]
, C = #c#
--- endif
--- where
--- ifnotnil [d]
B = #d#
--- endif
--- if [a > 9]
A > #a#
--- endif
A IN
--- range [values, item, (, )]
#item#
--- endrange
--- endwhere

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
// m := orm.M{"values": [...]int{1, 2, 3, 4, 5}, "b": "222", "c": time.Now(), "a": 0}
// m := orm.M{"values": [...]int{1, 2, 3, 4, 5}, "b": "222", "a": 0}
// m := orm.M{"values": [...]int{1, 2, 3, 4, 5}, "b": "222", "d": "999", "a": 0}
m := orm.M{"values": [...]int{1, 2, 3, 4, 5}, "b": "222", "d": "999", "a": 20}
// m := orm.M{"values": [...]int{1, 2, 3, 4, 5}, "b": "222", "d": "999", "a": 10, "c": time.Now()}
count, err := orm.InitGQL().Use("UpdateCondition").M(m).Update().Result()
```

条件更新

```Go
--- [UpdateSet]
UPDATE T_TEST
--- set
--- ifnotnil [c]
C = #c#
--- endif
--- if [a > 3]
B = '777'
--- endif
--- endset

orm.RegisterDataSource("xxx", uri, nil)  // 第三个参数是*orm.Option 设置连接池大小(默认10)，最大空闲数(默认3)， 最长生命周期(默认5分钟)
orm.SHOWSQL = true  // 在当前目录的AURORA_ORM_LOG下可以查看orm.log，包含sql，参数以及执行时间
// m := orm.M{}
// m := orm.M{"c": time.Now()}
m := orm.M{"c": time.Now(), "a": 6}
count, err := orm.InitGQL().Use("UpdateSet").M(m).Update().Result()
```

注意事项：
***如果使用 --- where，条件前不需要自己添加AND！***
***如果使用 --- set，语句后不需要自己添加逗号！***



事务

```Go
--- [TranInsert]
INSERT INTO T_TEST (A, B, C) VALUES (#item.Fa#, #item.Fb#, #item.Fc#)

--- [TranUpdate]
UPDATE T_TEST SET B = #b# WHERE A = #a#

tran := orm.InitTran("default")
t := Test{
    Fa: 100,
    Fb: "101",
    Fc: time.Now(),
}
m := orm.M{"item": t}
var i, i2 int64
orm.InitGQL().Use("TranInsert").M(m).Insert().Result()
tran.Insert("TranInsert", m, &i, nil)
m = orm.M{"a": 100, "b": "1111"}
tran.Update("TranUpdate", m, &i2, nil)
err := tran.Commit()
fmt.Println(err)
fmt.Print("transaction error: ", tran.Error())
```

