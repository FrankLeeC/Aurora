-- 两个横线表示注释，注释必须独占一行，不可以在SQL后!
--- [Insert]
INSERT INTO T_TEST (B, C) VALUES (#item.Fb#, #item.Fc#)



--- [BatchInsert]
INSERT INTO T_TEST (B, C)
VALUES
--- range [values, item]
(#item.Fb#, #item.Fc#)
--- endrange



--- [SelectOne]
SELECT * FROM T_TEST LIMIT 0, 1



--- [SelectIfNotNil]
SELECT * FROM T_TEST
WHERE B = '123'
-- 如果time字段不为空
--- ifnotnil [time]
AND C = #time#
--- endif



--- [SelectIf]
SELECT * FROM T_TEST
WHERE B = '123'
-- 运算符支持 ==, !=, >, >=, <, <=; 数据类型支持bool, int, string, float等,字符串类型不需要写引号。具体的数据类型由变量'f'决定。
--- if [f == true]
AND A > #id#
--- endif



--- [SelectWhere]
SELECT * FROM T_TEST
--- where
--- if [id < 5]
A = #id#
--- endif
--- ifnotnil [b]
and B = #b#
--- endif
--- endwhere

--- [SelectWhere2]
SELECT * FROM T_TEST
--- where
--- ifnotnil [b]
and B = #b#
--- endif
--- if [id < 5]
or (A = #id# and 1 = 1)
--- endif
--- endwhere


--- [DeleteCondition]
DELETE FROM T_TEST
WHERE A IN
--- range [values, item, (, )]
#item#
--- endrange



--- [UpdateCondition]
UPDATE T_TEST SET B = #b#
--- ifnotnil [c]
, C = #c#
--- endif
--- where
--- ifnotnil [d]
AnD B = #d#
--- endif
--- if [a > 9]
oR A > #a#
--- endif
OR A IN
--- range [values, item, (, )]
#item#
--- endrange
--- endwhere

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


--- [TranInsert]
INSERT INTO T_TEST (A, B, C) VALUES (#item.Fa#, #item.Fb#, #item.Fc#)

--- [TranUpdate]
UPDATE T_TEST SET B = #b# WHERE A = #a#
