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
	fs := FixedSlice.NewFixedSlice(3)
	fs.Append(20)
	fs.Append("test")
	fs.Append(true)

	fs.Append(50)
	fs.Append(60)

	for i := 0; i < fs.Count(); i++ {
		fmt.Println(i, "=", *fs.At(i))
	}

	fs1 := FixedSlice.NewFixedSlice(3)
	fs1.Copy(fs)
	for i := 0; i < fs1.Count(); i++ {
		fmt.Println(i, "=", *fs1.At(i))
	}

	var i interface{} = *fs.At(0)
	var ii int = i.(int)
	fmt.Println("ii=", ii)
}
```
