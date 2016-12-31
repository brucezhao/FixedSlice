# FixedSlice
元素地址固定的Slice，不会因为append而改变。  
Slice在元素插入时，如果容量已经满了，会重新拷贝一份到新内存上，所有元素的地址全部变了，这个特性在有时候是比较麻烦的。比如我在写并发服务器的时候用到了类似下面这样的代码：
```go
	s1 := []int{1, 2, 3}
	pS1 := &s1[0]
	s1 = append(s1, 4, 5)
	*pS1 = 7
	fmt.Println(s1)
```
预期值是[7 2 3 4 5]，但打印出来的是[1 2 3 4 5]，这是因为append(s1, 4, 5)已经导致了内存改变，原来pS1指向的内存区域已经不是s1[0]的地址了。  在写并发服务器的时候，由于可能会有其它的goroutine使用append，所以想取地址时，必须每次加锁后重新调用pS1 = &s1[0]，麻烦不说，效率也低，  所以我写了元素地址固定不变的Slice包，原理是当slice容易满时，新建一个slcie，而不是扩展它。  
测试代码:
```go
func main() {
	fs := fixedslice.New(3)
	fs.Append(20)
	fs.Append("test")
	fs.Append(true)

	fmt.Println(fs)                 //输出[20 test true]
	var pfs *interface{} = fs.At(1) //取第二个元素的地址

	fs.Append(50)
	fs.Append(60)   //增加两个元素后，已经超过了初始容量的大小
	fmt.Println(fs) //输出[20 test true 50 60]
	*pfs = 10       //通过指针修改第二个元素的值

	fmt.Println(fs) //输出[20 10 true 50 60]，可见*pfs=10确实修改了元素的值，
					//也证明了元素的地址并没有因为容量增加瑞改变了地址

	fs1 := fixedslice.New(3)
	fs1.Copy(fs) //拷贝

	fmt.Println(fs1) //输出[20 10 true 50 60]
	pfs = fs1.At(2)
	*pfs = 30        //通过指针修改第三个元素的值
	fmt.Println(fs)  //输出[20 10 true 50 60]
	fmt.Println(fs1) //输出[20 10 30 50 60]，可见fs的值并没有变，证明fs1.Copy是深拷贝
}
```
