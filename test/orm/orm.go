package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/FrankLeeC/Aurora/orm"

	_ "github.com/go-sql-driver/mysql"
)

type Test struct {
	Fa int       `column:"A"`
	Fb string    `column:"B"`
	Fc time.Time `column:"C"`
}

func init() {
	orm.SHOWSQL = true
	err := orm.RegisterDataSource("default", "root:xxx@tcp(127.0.0.1:3306)/aurora?loc=Local", nil)
	if err != nil {
		log.Printf("register error: %s\n", err.Error())
	}
}

func main() {
	// Insert()
	// BatchInsert()
	// SelectOne()
	// SelectIfNotNil()
	// SelectIf()
	// SelectWhere()
	// DeleteCondition()
	// UpdateCondition()
	UpdateSet()
	// Transaction()
}

func Insert() {
	t := Test{
		Fb: "123",
		Fc: time.Now(),
	}
	m := orm.M{"item": t}
	count, err := orm.InitGQL().Use("Insert").M(m).Insert().Result()
	// lastID, err := orm.InitGQL().Use("Insert").M(m).Insert().ReturnLastID().Result()
	log.Printf("count: %d, err: %v\n", count, err)
}

func BatchInsert() {
	N := 10
	arr := make([]Test, 0, N)
	for i := 1; i < N; i++ {
		arr = append(arr, Test{Fb: strconv.Itoa(i), Fc: time.Now()})
	}
	m := orm.M{"values": arr}
	count, err := orm.InitGQL().Use("BatchInsert").M(m).Insert().Result() // 默认使用default数据源
	log.Printf("count: %d, err: %v\n", count, err)
}

func SelectOne() {
	var rs Test
	count, err := orm.InitGQL().Use("SelectOne").One(&rs)
	log.Printf("count: %d, err: %v, rs: %v\n", count, err, rs)
}

func SelectIfNotNil() {
	var rs []Test
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-05-09 22:38:15", time.Local)
	m := orm.M{"time": t}
	// m := orm.M{}
	count, err := orm.InitGQL().Use("SelectIfNotNil").M(m).All(&rs)
	log.Printf("count: %d, err: %v, rs: %v\n", count, err, rs)
}

func SelectIf() {
	var rs []Test
	m := orm.M{"f": true, "id": -1}
	// m := orm.M{"f": false}
	count, err := orm.InitGQL().Use("SelectIf").M(m).All(&rs)
	log.Printf("count: %d, err: %v, rs: %v\n", count, err, rs)
}

func SelectWhere() {
	var rs []Test
	// m := orm.M{"id": 1, "b": "123"}
	// m := orm.M{"id": 7, "b": "3"}
	// m := orm.M{"id": 2}
	m := orm.M{"id": 7}
	count, err := orm.InitGQL().Use("SelectWhere").M(m).All(&rs)
	log.Printf("count: %d, err: %v, rs: %v\n", count, err, rs)
}

func DeleteCondition() {
	m := orm.M{"values": [...]int{1, 2, 3, 4, 5}}
	count, err := orm.InitGQL().Use("DeleteCondition").M(m).Delete().Result()
	log.Printf("count: %d, err: %v\n", count, err)
}

func UpdateCondition() {
	// m := orm.M{"values": [...]int{2, 3, 4, 5, 6}, "b": "222", "c": time.Now(), "a": 0}
	// m := orm.M{"values": [...]int{2, 3, 4, 5, 6}, "b": "999", "a": 0}
	// m := orm.M{"values": [...]int{2, 3, 4, 5, 6}, "b": "222", "d": "999", "a": 0}
	// m := orm.M{"values": [...]int{2, 3, 4, 5, 6}, "b": "999", "d": "222", "a": 2}
	m := orm.M{"values": [...]int{2, 3, 4, 5, 6}, "b": "999", "d": "222", "a": 10, "c": time.Now()}
	count, err := orm.InitGQL().Use("UpdateCondition").M(m).Update().Result()
	log.Printf("count: %d, err: %v\n", count, err)
}

func UpdateSet() {
	// m := orm.M{}
	// m := orm.M{"c": time.Now()}
	m := orm.M{"c": time.Now(), "a": 6}
	count, err := orm.InitGQL().Use("UpdateSet").M(m).Update().Result()
	log.Printf("count: %d, err: %v\n", count, err)
}

func Transaction() {
	tran := orm.InitTran("default")
	t := Test{
		Fa: 100,
		Fb: "101",
		Fc: time.Now(),
	}
	m := orm.M{"item": t}
	var i, i2 int64
	// orm.InitGQL().Use("TranInsert").M(m).Insert().Result()
	tran.Insert("TranInsert", m, &i, nil)
	m = orm.M{"a": 100, "b": "1111"}
	tran.Update("TranUpdate", m, &i2, nil)
	err := tran.Commit()
	fmt.Println(err)
	fmt.Print("transaction error: ", tran.Error())
}
