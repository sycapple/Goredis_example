/*
	用go语言搭建Redis库
*/

package main

import (
	"fmt"  //基本库
	"sync" //用于单例模式
	"time" //用于设计过期时间
)

type Database struct {
	// 在这里定义数据库内结构
	data       map[string]string    //数据：类型都为字符串的键值对
	expiration map[string]time.Time //过期时间
	mu         sync.Mutex           //异步信号
}

var instance *Database
var once sync.Once //Once是只执行一次动作的对象

/*
	单例模式的核心：
	如果没有创建，就创建一个新的数据库
	如果已经创建了，就返回创建的数据库
	可以使用OnceDo方法,当且仅当第一次被调用时才执行函数
*/

func GetInstance() *Database {
	//有且仅有一次的初始化
	once.Do(func() {
		instance = &Database{
			data:       make(map[string]string),
			expiration: make(map[string]time.Time),
		}
	})
	return instance
}

/*
	基本的操作函数
	是对数据库指针指向的地址操作
	所造成的影响是“全局的”
*/

// 存储键值对
func (db *Database) Set(key, value string, expiration int) {
	db.mu.Lock()
	defer db.mu.Unlock() //Defer用于在函数退出前执行“解锁”语句

	db.data[key] = value //赋值操作
	if expiration > 0 {
		db.expiration[key] = time.Now().Add(time.Duration(expiration) * time.Second) //从创建时间起，以秒计算了过期时间的持续时间
	}
}

// 检索键值对，返回可能存在的value以及表征“是否存在该键值对”的布尔代数
func (db *Database) Get(key string) (string, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 如果键不存在，ok 的值为 false，v的值为该类型的零值
	value, ok := db.data[key]
	if !ok {
		return "", false
	}

	//检查是否过期
	expiration, exists := db.expiration[key]
	if exists && time.Now().After(expiration) { //如果设置过有效期且有效期已过
		delete(db.data, key) //删除键值对
		delete(db.expiration, key)
		return "", false
	}

	return value, true
}

// 删除键值对
func (db *Database) Delete(key string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.data, key)
	delete(db.expiration, key)
}

// 查看是否存在键值对
func (db *Database) Exists(key string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, ok := db.data[key] //用_表示用于丢弃变量的值
	return ok
}

// 返回数据库中存储的所有键的列表
func (db *Database) Keys() []string {
	db.mu.Lock()
	defer db.mu.Unlock()

	keys := make([]string, 0, len(db.data))
	for key := range db.data {
		keys = append(keys, key) //将所有的key添加到一个字符串中返回
	}
	return keys
}

func (d *Database) SetExpiration(key string, expiration int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 检查键是否存在
	_, exists := d.data[key]
	if !exists {
		fmt.Println("Key not found.")
		return
	}

	// 如果 expiration 为 0，则表示没有过期时间，直接返回
	if expiration == 0 {
		return
	}

	// 设置键的过期时间
	if _, ok := d.expiration[key]; !ok {
		d.expiration[key] = time.Now().Add(time.Duration(expiration) * time.Second)
		fmt.Printf("Expiration time set for key '%s': %d seconds\n", key, expiration)
	} else {
		fmt.Printf("Key '%s' already has an expiration time set.\n", key)
	}
}

func main() { //开始进入主函数
	db := GetInstance()

	db.Set("xiaoming", "175", 0)
	db.Set("zhangsan", "156", 10)
	db.Set("lisi", "180", 0)
	db.Set("lwangwu", "188", 0)

	for {
		fmt.Println("1. Set key-value pair")
		fmt.Println("2. Get value by key")
		fmt.Println("3. Delete key-value pair")
		fmt.Println("4. Check if key exists")
		fmt.Println("5. List all keys")
		fmt.Println("6. Add an expiration time for key-value pairs that already exist")
		fmt.Println("7. Exit")

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scanln(&choice)
		fmt.Print()

		switch choice {
		case 1: //第一类情况：SET
			var key, value string
			var expiration int
			fmt.Print("Enter your key: ")
			fmt.Scanln(&key)
			fmt.Print("Enter your value: ")
			fmt.Scanln(&value)
			fmt.Print("Enter expiration (0 for no expiration): ")
			fmt.Scanln(&expiration)
			db.Set(key, value, expiration)
			fmt.Printf("Your key-value { %s - %s } pair set.", key, value)

		case 2: //第二类情况：GET
			var key string
			fmt.Print("Enter key: ")
			fmt.Scanln(&key)
			value, exists := db.Get(key)
			if exists {
				fmt.Println("Value:", value)
			} else {
				fmt.Printf("Key %s not found.", key)
			}

		case 3: //第三类情况：删除
			var key string
			fmt.Print("Enter key you want to delete: ")
			fmt.Scanln(&key)
			db.Delete(key)
			fmt.Printf("Key %s deleted.", key)

		case 4: //第四类情况：检查
			var key string
			fmt.Print("Enter key to check: ")
			fmt.Scanln(&key)
			exists := db.Exists(key)
			if exists {
				fmt.Printf("Key %s exists.", key)
			} else {
				fmt.Printf("Key %s not found.", key)
			}

		case 5: //第五类情况：列出所有的keys
			keys := db.Keys()
			fmt.Println("Keys:", keys)

		case 6: //第六类情况：添加
			var key string
			var expiration int
			fmt.Print("Enter key to set expiration: ")
			fmt.Scanln(&key)
			fmt.Print("Enter expiration time in seconds (0 for no expiration): ")
			fmt.Scanln(&expiration)
			db.SetExpiration(key, expiration)
			fmt.Println("Expiration time set for the key.")

		case 7: //第七类情况：退出
			fmt.Println("Exiting.")
			return

		default:
			fmt.Println("Invalid choice. Please enter a valid option.")
		}
	}
}
