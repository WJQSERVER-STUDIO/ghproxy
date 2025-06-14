package weakcache

import (
	"container/list"
	"sync"
	"time"
	"weak" // Go 1.24 引入的 weak 包
)

// DefaultExpiration 默认过期时间，这里设置为 15 分钟。
// 这是一个导出的常量，方便用户使用包时引用默认值。
const DefaultExpiration = 5 * time.Minute

// cleanupInterval 是后台清理 Go routine 的扫描间隔，这里设置为 5 分钟。
// 这是一个内部常量，不导出。
const cleanupInterval = 2 * time.Minute

// cacheEntry 缓存项的内部结构。不导出。
type cacheEntry[T any] struct {
	Value      T
	Expiration time.Time
	key        string // 存储key，方便在list.Element中引用
}

// Cache 是一个基于 weak.Pointer, 带有过期和大小上限 (FIFO) 的泛型缓存。
// 这是一个导出的类型。
type Cache[T any] struct {
	mu sync.RWMutex

	// 修正：缓存存储：key -> weak.Pointer 到 cacheEntry 结构体 (而不是指向结构体的指针)
	// weak.Make(*cacheEntry[T]) 返回 weak.Pointer[cacheEntry[T]]
	data map[string]weak.Pointer[cacheEntry[T]]

	// FIFO 链表：存储 key 的 list.Element
	// 链表头部是最近放入的，尾部是最早放入的（最老的）
	fifoList *list.List
	// FIFO 元素的映射：key -> *list.Element
	fifoMap map[string]*list.Element

	defaultExpiration time.Duration
	maxSize           int // 缓存最大容量，0 表示无限制

	stopCleanup chan struct{}
	wg          sync.WaitGroup // 用于等待清理 Go routine 退出
}

// NewCache 创建一个新的缓存实例。
// expiration: 新添加项的默认过期时间。如果为 0，则使用 DefaultExpiration。
// maxSize: 缓存的最大容量，0 表示无限制。当达到上限时，采用 FIFO 策略淘汰。
// 这是一个导出的构造函数。
func NewCache[T any](expiration time.Duration, maxSize int) *Cache[T] {
	if expiration <= 0 {
		expiration = DefaultExpiration
	}

	c := &Cache[T]{
		// 修正：初始化 map，值类型已修正
		data:              make(map[string]weak.Pointer[cacheEntry[T]]),
		fifoList:          list.New(),
		fifoMap:           make(map[string]*list.Element),
		defaultExpiration: expiration,
		maxSize:           maxSize,
		stopCleanup:       make(chan struct{}),
	}
	// 启动后台清理 Go routine
	c.wg.Add(1)
	go c.cleanupLoop()
	return c
}

// Put 将值放入缓存。如果 key 已存在，会更新其值和过期时间。
// 这是导出的方法。
func (c *Cache[T]) Put(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiration := now.Add(c.defaultExpiration)

	// 如果 key 已经存在，更新其值和过期时间。
	// 在 FIFO 策略中， Put 更新不改变其在链表中的位置，除非旧的 entry 已经被 GC。
	if elem, ok := c.fifoMap[key]; ok {
		// 从 data map 中获取弱引用，wp 的类型现在是 weak.Pointer[cacheEntry[T]]
		if wp, dataOk := c.data[key]; dataOk {
			// wp.Value() 返回 *cacheEntry[T]， entry 的类型现在是 *cacheEntry[T]
			entry := wp.Value()
			if entry != nil {
				// 旧的 cacheEntry 仍在内存中，直接更新
				entry.Value = value
				entry.Expiration = expiration
				// 在严格 FIFO 中，更新不移动位置
				return
			}
			// 如果 weak.Pointer.Value() 为 nil，说明之前的 cacheEntry 已经被 GC 了
			// 此时需要创建一个新的 entry，并将其从旧位置移除，再重新添加
			c.fifoList.Remove(elem)
			delete(c.fifoMap, key)
		} else {
			c.fifoList.Remove(elem)
			delete(c.fifoMap, key)
		}
	}

	// 新建缓存项 (注意这里是结构体值，而不是指针)
	// weak.Make 接收的是指针 *T
	entry := &cacheEntry[T]{ // 创建结构体指针
		Value:      value,
		Expiration: expiration,
		key:        key, // 存储 key
	}

	// 将新的 *cacheEntry[T] 包装成 weak.Pointer[cacheEntry[T]] 存入 data map
	// weak.Make(entry) 现在返回 weak.Pointer[cacheEntry[T]]，类型匹配 data map 的值类型
	c.data[key] = weak.Make(entry)

	// 添加到 FIFO 链表头部 (最近放入/更新的在头部)
	// PushFront 返回新的 list.Element
	c.fifoMap[key] = c.fifoList.PushFront(key)

	// 检查大小上限并进行淘汰 (淘汰尾部的最老项)
	c.evictIfNeeded()
}

// Get 从缓存中获取值。返回获取到的值和是否存在/是否有效。
// 这是导出的方法。
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock() // 先读锁
	// 从 data map 中获取弱引用，wp 的类型现在是 weak.Pointer[cacheEntry[T]]
	wp, ok := c.data[key]
	c.mu.RUnlock() // 立即释放读锁，如果需要写操作（removeEntry）可以获得锁

	var zero T // 零值

	if !ok {
		return zero, false
	}

	// 尝试获取实际的 cacheEntry 指针
	// wp.Value() 返回 *cacheEntry[T]， entry 的类型现在是 *cacheEntry[T]
	entry := wp.Value()

	if entry == nil {
		// 对象已被GC回收，需要清理此弱引用
		c.removeEntry(key) // 内部会加写锁
		return zero, false
	}

	// 检查过期时间 (通过 entry 指针访问字段)
	if time.Now().After(entry.Expiration) {
		// 逻辑上已过期
		c.removeEntry(key) // 内部会加写锁
		return zero, false
	}

	// 在 FIFO 缓存中，Get 操作不改变项在链表中的位置
	return entry.Value, true // 通过 entry 指针访问值字段
}

// removeEntry 从缓存中移除项。
// 这个方法是内部使用的，不导出。需要被调用者确保持有写锁，或者内部自己加锁。
// 考虑到 Get 和 cleanupLoop 可能会调用，让其内部自己加锁更安全。
func (c *Cache[T]) removeEntry(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 从 data map 中删除
	delete(c.data, key)

	// 从 FIFO 链表和 fifoMap 中删除
	if elem, ok := c.fifoMap[key]; ok {
		c.fifoList.Remove(elem)
		delete(c.fifoMap, key)
	}
}

// evictIfNeeded 检查是否需要淘汰最老（FIFO 链表尾部）的项。
// 这个方法是内部使用的，不导出。必须在持有写锁的情况下调用。
func (c *Cache[T]) evictIfNeeded() {
	if c.maxSize > 0 && c.fifoList.Len() > c.maxSize {
		// 淘汰 FIFO 链表尾部的元素 (最老的)
		oldest := c.fifoList.Back()
		if oldest != nil {
			keyToEvict := oldest.Value.(string) // 链表元素存储的是 key
			c.fifoList.Remove(oldest)
			delete(c.fifoMap, keyToEvict)
			delete(c.data, keyToEvict) // 移除弱引用
		}
	}
}

// Size 返回当前缓存中的弱引用项数量。
// 注意：这个数量可能包含已被 GC 回收但尚未清理的项。
// 这是一个导出的方法。
func (c *Cache[T]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// cleanupLoop 后台清理 Go routine。不导出。
func (c *Cache[T]) cleanupLoop() {
	defer c.wg.Done()
	// 使用内部常量 cleanupInterval
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpiredAndGCed()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpiredAndGCed 扫描并清理已过期或已被 GC 回收的项。不导出。
func (c *Cache[T]) cleanupExpiredAndGCed() {
	c.mu.Lock() // 清理时需要写锁
	defer c.mu.Unlock()

	now := time.Now()
	keysToRemove := make([]string, 0, len(c.data)) // 预估容量

	// 遍历 data map 查找需要清理的键
	for key, wp := range c.data {
		// wp 的类型是 weak.Pointer[cacheEntry[T]]
		// wp.Value() 返回 *cacheEntry[T]， entry 的类型是 *cacheEntry[T]
		entry := wp.Value() // 尝试获取强引用

		if entry == nil {
			// 已被 GC 回收
			keysToRemove = append(keysToRemove, key)
		} else if now.After(entry.Expiration) {
			// 逻辑过期 (通过 entry 指针访问字段)
			keysToRemove = append(keysToRemove, key)
		}
	}

	// 执行删除操作
	for _, key := range keysToRemove {
		// 从 data map 中删除
		delete(c.data, key)
		// 从 FIFO 链表和 fifoMap 中删除
		// 需要再次检查 fifoMap，因为在持有锁期间，evictIfNeeded 可能已经移除了这个 key
		if elem, ok := c.fifoMap[key]; ok {
			c.fifoList.Remove(elem)
			delete(c.fifoMap, key)
		}
	}
}

// StopCleanup 停止后台清理 Go routine。
// 这是一个导出的方法。
func (c *Cache[T]) StopCleanup() {
	close(c.stopCleanup)
	c.wg.Wait() // 等待 Go routine 退出
}
