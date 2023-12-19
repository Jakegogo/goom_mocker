// Package mocker 定义了 mock 的外层用户使用 API 定义,
// 包括函数、方法、接口、未导出函数(或方法的)的 Mocker 的实现。
// 当前文件实现了 Mocker 接口各实现类的构造链创建，以便通过链式构造一个 Mocker 对象。
package mocker

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/tencent/goom/internal/iface"
	"github.com/tencent/goom/internal/logger"
)

// Builder Mock 构建器, 负责创建一个链式构造器.
type Builder struct {
	pkgName string
	mockers map[interface{}]Mocker
}

// Pkg 指定包名，当前包无需指定
// 对于跨包目录的私有函数的 mock 通常都是因为代码设计可能有问题, 此功能会在未来版本中移除
// 后续仅支持同包下的未导出方法的 mock
// Deprecated: 对于跨包目录的私有函数的 mock 通常都是因为代码设计可能有问题
func (b *Builder) Pkg(name string) *Builder {
	b.pkgName = name
	return b
}

// PkgName 返回包名
// Deprecated: 对于跨包目录的私有函数的 mock 通常都是因为代码设计可能有问题
func (b *Builder) PkgName() string {
	return b.pkgName
}

// Create 创建 Mock 构建器
// 非线程安全的,不能在多协程中并发地 mock 或 reset 同一个函数
func Create() *Builder {
	// callerDeps 当前的调用栈栈层次
	const callerDeps = 2
	return &Builder{
		pkgName: currentPkg(callerDeps),
		mockers: make(map[interface{}]Mocker, 30),
	}
}

// Interface 指定接口类型的变量定义
// iFace 必须是指针类型, 比如 i 为 interface 类型变量, iFace 传递&i
func (b *Builder) Interface(iFace interface{}) *CachedInterfaceMocker {
	mKey := reflect.TypeOf(iFace).String()
	if mocker, ok := b.mockers[mKey]; ok && !mocker.Canceled() {
		b.reset2CurPkg()
		return mocker.(*CachedInterfaceMocker)
	}

	// 创建 InterfaceMocker
	// context 和 interface 类型绑定
	mocker := NewDefaultInterfaceMocker(b.pkgName, iFace, iface.NewContext())
	cachedMocker := NewCachedInterfaceMocker(mocker)
	b.cache(mKey, cachedMocker)
	b.reset2CurPkg()
	return cachedMocker
}

// cache 添加到缓存
func (b *Builder) cache(mKey interface{}, cachedMocker Mocker) {
	b.mockers[mKey] = cachedMocker
}

// Struct 指定结构体名称
// 比如需要 mock 结构体函数 (*conn).Write(b []byte)，则 name="conn"
func (b *Builder) Struct(obj interface{}) *CachedMethodMocker {
	mKey := reflect.ValueOf(obj).Type().String()
	if mocker, ok := b.mockers[mKey]; ok && !mocker.Canceled() {
		b.reset2CurPkg()
		return mocker.(*CachedMethodMocker)
	}

	mocker := NewMethodMocker(b.pkgName, obj)
	cachedMocker := NewCachedMethodMocker(mocker)
	b.cache(mKey, cachedMocker)
	b.reset2CurPkg()
	return cachedMocker
}

// Func 指定函数定义
// funcDef 函数，比如 foo
// 方法的 mock, 比如 &Struct{}.method
func (b *Builder) Func(obj interface{}) *DefMocker {
	var key = runtime.FuncForPC(reflect.ValueOf(obj).Pointer()).Name()
	if mocker, ok := b.mockers[key]; ok && !mocker.Canceled() {
		b.reset2CurPkg()
		return mocker.(*DefMocker)
	}

	mocker := NewDefMocker(b.pkgName, obj)
	b.cache(key, mocker)
	b.reset2CurPkg()
	return mocker
}

// ExportStruct 导出私有结构体
// 比如需要 mock 结构体函数 (*conn).Write(b []byte)，则 name="conn"
func (b *Builder) ExportStruct(name string) *CachedUnexportedMethodMocker {
	if mocker, ok := b.mockers[b.pkgName+"_"+name]; ok && !mocker.Canceled() {
		b.reset2CurPkg()
		return mocker.(*CachedUnexportedMethodMocker)
	}

	structName := name
	if strings.Contains(name, "*") {
		structName = fmt.Sprintf("(%s)", name)
	}

	mocker := NewUnexportedMethodMocker(b.pkgName, structName)
	cachedMocker := NewCachedUnexportedMethodMocker(mocker)
	b.cache(b.pkgName+"_"+name, cachedMocker)
	b.reset2CurPkg()
	return cachedMocker
}

// ExportFunc 导出私有函数
// 比如需要 mock 函数 foo()， 则 name="pkg_name.foo"
// 比如需要 mock 方法, pkg_name.(*struct_name).method_name
// name string foo 或者(*struct_name).method_name
func (b *Builder) ExportFunc(name string) *UnexportedFuncMocker {
	if name == "" {
		panic("func name is empty")
	}

	if mocker, ok := b.mockers[b.pkgName+"_"+name]; ok && !mocker.Canceled() {
		b.reset2CurPkg()
		return mocker.(*UnexportedFuncMocker)
	}

	mocker := NewUnexportedFuncMocker(b.pkgName, name)
	b.cache(b.pkgName+"_"+name, mocker)
	b.reset2CurPkg()
	return mocker
}

// Var 变量 mock, target 类型必须传递指针类型
func (b *Builder) Var(target interface{}) VarMock {
	cacheKey := fmt.Sprintf("var_%d", reflect.ValueOf(target).Pointer())
	if mocker, ok := b.mockers[cacheKey]; ok && !mocker.Canceled() {
		return mocker.(VarMock)
	}

	mocker := NewVarMocker(target)
	b.cache(cacheKey, mocker)
	return mocker
}

// Reset 取消当前 builder 的所有 Mock
func (b *Builder) Reset() *Builder {
	for _, mocker := range b.mockers {
		mocker.Cancel()
		// callerDeps 当前的调用栈栈层次
		const callerDeps = 5
		logger.Consolefc(logger.DebugLevel, "mockers [%s] resets.", logger.Caller(callerDeps), mocker.String())
	}
	return b
}

// reset2CurPkg 设置回当前的包
func (b *Builder) reset2CurPkg() {
	b.pkgName = currentPackage()
}

// currentPackage 获取当前调用的包路径
func currentPackage() string {
	// callerDeps 当前的调用栈栈层次
	const callerDeps = 4
	return currentPkg(callerDeps)
}

// currentPkg 获取调用者的包路径
func currentPkg(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	callerName := runtime.FuncForPC(pc).Name()

	if i := strings.Index(callerName, ".("); i > -1 {
		return callerName[:i]
	}

	if i := strings.LastIndex(callerName, "/"); i > -1 {
		realIndex := strings.Index(callerName[i:len(callerName)-1], ".")
		return callerName[:realIndex+i]
	}

	realIndex := strings.Index(callerName, ".")
	return callerName[:realIndex]
}

// OpenDebug 开启 debug 模式
// 也可以通过添加环境变量开启 debug: GOOM_DEBUG=true
// 1.可以查看 apply 和 reset 的状态日志
// 2.查看 mock 调用日志
func OpenDebug() {
	logger.OpenDebug()
}

// CloseDebug 关闭 debug 模式
func CloseDebug() {
	logger.CloseDebug()
}

// OpenTrace 打开日志跟踪
func OpenTrace() {
	logger.OpenTrace()
}

// CloseTrace 关闭日志跟踪
func CloseTrace() {
	logger.CloseTrace()
}
