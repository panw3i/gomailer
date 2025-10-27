package gomailer

import (
	"sort"
	"sync"
)

// Handler 定义了单个钩子处理器
// 多个处理器可以共享相同的 ID
// 如果未明确设置 ID，将由 Hook.Bind 自动生成
type Handler[T Resolver] struct {
	// Func 定义要执行的处理器函数
	//
	// 注意：用户需要调用 e.Next() 才能继续执行钩子链
	Func func(T) error

	// Id 处理器的唯一标识符
	//
	// 可以用于后续通过 Hook.Unbind 移除处理器
	//
	// 如果缺失，在将处理器添加到钩子时会自动生成
	Id string

	// Priority 允许更改处理器在钩子中的默认执行优先级
	//
	// 如果为 0，处理器将按注册顺序执行
	// 数字越小，优先级越高（越先执行）
	Priority int
}

// Hook 定义了一个通用的并发安全的事件钩子管理结构
//
// 使用自定义事件时，必须嵌入基础的 Event 类型
//
// 示例:
//
//	type CustomEvent struct {
//		Event
//		SomeField int
//	}
//
//	h := Hook[*CustomEvent]{}
//
//	h.BindFunc(func(e *CustomEvent) error {
//		println(e.SomeField)
//		return e.Next()
//	})
//
//	h.Trigger(&CustomEvent{ SomeField: 123 })
type Hook[T Resolver] struct {
	handlers []*Handler[T]   // 处理器列表
	mu       sync.RWMutex    // 读写互斥锁，保证并发安全
}

// Bind 将提供的处理器注册到当前钩子队列
//
// 如果 handler.Id 为空，会更新为自动生成的值
//
// 如果当前钩子列表中已有 ID 匹配 handler.Id 的处理器，
// 则旧处理器会被新处理器替换
//
// 参数:
//   - handler: 要绑定的处理器
// 返回:
//   - string: 处理器的 ID
func (h *Hook[T]) Bind(handler *Handler[T]) string {
    h.mu.Lock()
    defer h.mu.Unlock()

    // 防御：空处理器或空函数直接忽略
    if handler == nil || handler.Func == nil {
        return ""
    }

    var exists bool

	if handler.Id == "" {
		// 生成新的 ID
		handler.Id = generateHookId()

		// 确保 ID 不重复
	DUPLICATE_CHECK:
		for _, existing := range h.handlers {
			if existing.Id == handler.Id {
				handler.Id = generateHookId()
				goto DUPLICATE_CHECK
			}
		}
	} else {
		// 替换已存在的处理器
		for i, existing := range h.handlers {
			if existing.Id == handler.Id {
				h.handlers[i] = handler
				exists = true
				break
			}
		}
	}

	// 添加新处理器
	if !exists {
		h.handlers = append(h.handlers, handler)
	}

	// 按优先级排序处理器，保持相同优先级项的原始顺序
	sort.SliceStable(h.handlers, func(i, j int) bool {
		return h.handlers[i].Priority < h.handlers[j].Priority
	})

	return handler.Id
}

// BindFunc 类似于 Bind，但只需提供函数即可注册新处理器
//
// 注册的处理器具有默认优先级 0，ID 将自动生成
//
// 如果要注册具有自定义优先级或 ID 的处理器，请使用 Bind 方法
//
// 参数:
//   - fn: 处理器函数
// 返回:
//   - string: 自动生成的处理器 ID
func (h *Hook[T]) BindFunc(fn func(e T) error) string {
	return h.Bind(&Handler[T]{Func: fn})
}

// Unbind 通过 ID 移除一个或多个钩子处理器
//
// 参数:
//   - idsToRemove: 要移除的处理器 ID 列表
func (h *Hook[T]) Unbind(idsToRemove ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, id := range idsToRemove {
		for i := len(h.handlers) - 1; i >= 0; i-- {
			if h.handlers[i].Id == id {
				h.handlers = append(h.handlers[:i], h.handlers[i+1:]...)
				break // 目前在第一次出现时停止，因为我们不允许重复的 ID
			}
		}
	}
}

// UnbindAll 移除所有已注册的处理器
func (h *Hook[T]) UnbindAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers = nil
}

// Length 返回已注册的钩子处理器总数
func (h *Hook[T]) Length() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.handlers)
}

// Trigger 依次执行所有已注册的钩子处理器，将指定的事件作为参数
//
// 可选地，此方法还允许注册额外的一次性处理器函数，
// 这些函数将临时追加到处理器队列中
//
// 注意！每个钩子处理器必须调用 event.Next() 才能继续钩子链的执行
//
// 参数:
//   - event: 要传递给处理器的事件
//   - oneOffHandlerFuncs: 可选的一次性处理器函数列表
// 返回:
//   - error: 如果任何处理器返回错误，则返回该错误
func (h *Hook[T]) Trigger(event T, oneOffHandlerFuncs ...func(T) error) error {
	// 获取所有处理器（包括一次性处理器）
	h.mu.RLock()
	handlers := make([]func(T) error, 0, len(h.handlers)+len(oneOffHandlerFuncs))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.Func)
	}
	handlers = append(handlers, oneOffHandlerFuncs...)
	h.mu.RUnlock()

	// 重置事件的 next 函数（以防事件被重用）
	event.setNextFunc(nil)

	// 构建调用链（从后向前）
	for i := len(handlers) - 1; i >= 0; i-- {
		i := i
		old := event.nextFunc()
		event.setNextFunc(func() error {
			event.setNextFunc(old)
			return handlers[i](event)
		})
	}

	// 开始执行钩子链
	return event.Next()
}

// generateHookId 生成一个随机的钩子 ID
func generateHookId() string {
	return pseudorandomString(20)
}

// Resolver 定义了事件必须实现的接口
// 用于支持钩子链的继续执行
type Resolver interface {
	// Next 继续执行钩子链中的下一个处理器
	Next() error

	// nextFunc 获取下一个处理器函数（内部使用）
	nextFunc() func() error

	// setNextFunc 设置下一个处理器函数（内部使用）
	setNextFunc(func() error)
}

// Event 是所有钩子事件的基础结构
// 自定义事件必须嵌入此类型
type Event struct {
	next func() error // 下一个处理器函数
}

// Next 实现 Resolver 接口
// 继续执行钩子链中的下一个处理器
//
// 如果不调用 Next()，钩子链将在当前处理器处停止
func (e *Event) Next() error {
	if e.next != nil {
		return e.next()
	}
	return nil
}

// nextFunc 实现 Resolver 接口
// 返回下一个处理器函数
func (e *Event) nextFunc() func() error {
	return e.next
}

// setNextFunc 实现 Resolver 接口
// 设置下一个处理器函数
func (e *Event) setNextFunc(f func() error) {
	e.next = f
}
