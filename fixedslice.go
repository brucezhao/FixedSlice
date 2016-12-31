// FixedSlice project FixedSlice.go
// 作者：赵亦平
// 日期：2016年12月30日
/*
元素地址固定的Slice，不会因为append而改变。
Slice在元素插入时，如果容量已经满了，会重新拷贝一份到新内存上，所有元素的地址全部变了，这个特性在有时候是比较麻烦的。
比如我在写并发服务器的时候用到了类似下面这样的代码：

    s1 := []int{1, 2, 3}
    pS1 := &s1[0]
    s1 = append(s1, 4, 5)
    *pS1 = 7
    fmt.Println(s1)

预期值是[7 2 3 4 5]，但打印出来的是[1 2 3 4 5]，这是因为append(s1, 4, 5)已经导致了内存改变，原来pS1指向的内存区
域已经不是s1[0]的地址了。 在写并发服务器的时候，由于可能会有其它的goroutine使用append，所以想取地址时，必须每次加
锁后重新调用pS1 = &s1[0]，麻烦不说，效率也低， 所以我写了元素地址固定不变的Slice包，原理是当slice容易满时，新建一
个slcie，而不是扩展它。
*/
package fixedslice

import (
	"fmt"
)

type FixedSlice struct {
	initCapacity int            //initCapacity值决定了每个slice的容量
	Datas        []interface{}  //维护slice的slice
	currentSlice *[]interface{} //当前的slice的指针，添加的的元素进入该slice
	currentIndex int            //当前slice的索引值
	count        int            //元素总数
}

//构造函数
func New(initCapacity int) *FixedSlice {
	var fs FixedSlice

	fs.initCapacity = initCapacity
	fs.Datas = make([]interface{}, 0)

	sl := make([]interface{}, 0, initCapacity)
	fs.Datas = append(fs.Datas, &sl)
	fs.currentSlice = &sl
	fs.currentIndex = 0

	return &fs
}

func (fs *FixedSlice) String() string {
	var sRet string = "["
	var p *interface{}

	for i := 0; i < fs.count; i++ {
		p = fs.At(i)
		sRet += fmt.Sprint(*p)
		if i != fs.count-1 {
			sRet += " "
		}
	}
	sRet += "]"

	return sRet
}

//添加元素，参数为interface{}，意味着可以添加任何类型的元素
func (fs *FixedSlice) Append(v interface{}) *FixedSlice {
	//如果当前slice的容量已满，则新建一个
	if len(*fs.currentSlice) == cap(*fs.currentSlice) {
		sl := make([]interface{}, 0, fs.initCapacity)
		fs.Datas = append(fs.Datas, &sl)
		fs.currentSlice = &sl
		fs.currentIndex++
	}
	*fs.currentSlice = append(*fs.currentSlice, v)
	fs.count++
	return fs
}

//返回元素的总数
func (fs *FixedSlice) Count() int {
	return fs.count
}

//根据索引取值，注意此处取出的是*interface{}，使用时需要进行类型转换，示例：
/*
	var i interface{} = *fs.At(0)
	var ii int = i.(int)
	fmt.Println("ii=", ii)
*/
func (fs *FixedSlice) At(index int) *interface{} {
	//首先确定是哪个slice
	i1 := index / fs.initCapacity
	if i1 > fs.currentIndex {
		return nil
	}

	//再在找到的slice中取值
	i2 := index % fs.initCapacity
	if (i1 == fs.currentIndex) && (i2 >= len(*fs.currentSlice)) {
		return nil
	}

	var s interface{} = fs.Datas[i1]
	var s1 *[]interface{} = s.(*[]interface{})

	return &(*s1)[i2]
}

//从src中拷贝内容，是深度拷贝
func (fs *FixedSlice) Copy(src *FixedSlice) *FixedSlice {
	//原内容清空
	fs.Datas = nil
	fs.initCapacity = src.initCapacity

	var pss *[]interface{}
	var s interface{}
	//将src中的内容一个个地copy过来
	for i := 0; i <= src.currentIndex; i++ {
		s = src.Datas[i]
		pss = s.(*[]interface{})

		//ss必须在循环里头声明，否则所有的slice的内容都是一样的
		var ss []interface{} = make([]interface{}, fs.initCapacity)
		copy(ss, *pss)
		fs.Datas = append(fs.Datas, &ss)

		if i == src.currentIndex-1 {
			fs.currentSlice = &ss
		}
	}
	fs.currentIndex = src.currentIndex
	fs.count = src.count

	return fs
}
