package list

import (
	"fmt"
	"notesServer/gates/storage"
	"reflect"
	"sync"
)

type List struct {
	length    int64 // Текущая длина списка (количество узлов)
	firstNode *node // Указатель на первый узел
	lastNode  *node // Указатель на последний узел (для ускорения вставки элемента в конец)
	idInitial int64
	idCounter int64
	V         reflect.Type // фиксируется при добавлении первого элемента, сбрасывается при удалении всех элементов
	mu        sync.RWMutex
}

// NewList создает новый пустой односвязный список
func NewList(initID int64) (l *List) {
	return &List{length: 0, firstNode: nil, lastNode: nil, idInitial: initID, idCounter: initID, V: nil}
}

// Len возвращает количество элементов в списке
func (l *List) Len() int64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.length
}

// Add добавляет элемент в конец списка, возвращает идентификатор добавленного элемента
func (l *List) Add(value any) (id int64, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Согласование типа элементов
	if l.V == nil {
		l.V = reflect.TypeOf(value)
	} else if l.V != reflect.TypeOf(value) {
		return 0, storage.ErrMismatchType
	}

	// Создание нового узла и добавление его в конец
	newNode := &node{id: l.idCounter, value: value}
	l.idCounter++
	l.length++
	// Случай вставки первого элемента, когда не определены первый и последний узлы
	if l.firstNode == nil {
		l.firstNode = newNode
		l.lastNode = newNode
		return l.idCounter, nil
	}
	l.lastNode.nextNode = newNode
	l.lastNode = l.lastNode.nextNode
	return l.idCounter - 1, nil
}

// RemoveByID удаляет элемент по уникальному идентификатору
func (l *List) RemoveByID(id int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Случай out-of-range идентификатора
	if id > l.idCounter || id < l.idInitial {
		return
	}

	// Случай попытки удаления из пустого списка
	if l.firstNode == nil {
		return
	}

	// Случай удаления первого элемента
	if l.firstNode.id == id {
		// Случай удаления первого элемента из двух оставшихся
		if l.firstNode.nextNode == l.lastNode {
			l.lastNode = l.firstNode
		}
		l.firstNode = l.firstNode.nextNode
		// Случай удаления единственного элемента
		if l.firstNode == nil {
			l.lastNode = nil
		}
		l.length--
		// Сброс типа элементов
		if l.length == 0 {
			l.V = nil
		}
		return
	}

	// Проходимся по узлам и останавливаемся на нужном
	prevNode := l.firstNode
	for ; prevNode.nextNode != nil && prevNode.nextNode.id != id; prevNode = prevNode.nextNode {
	}
	// Прошли через весь список и не нашли нужный узел
	if prevNode.nextNode == nil {
		return
	}
	// Случай удаления последнего элемента
	if prevNode.nextNode == l.lastNode {
		l.lastNode = prevNode
		l.lastNode.nextNode = nil
	} else {
		prevNode.nextNode = prevNode.nextNode.nextNode
	}
	l.length--
	// Сброс типа элементов
	if l.length == 0 {
		l.V = nil
	}
	return
}

// RemoveByValue удаляет первый встретившийся элемент с данным значением
func (l *List) RemoveByValue(value any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Согласование типа элементов
	if l.V != reflect.TypeOf(value) {
		return
	}

	l.removeByValueUnsafely(value)
}

func (l *List) removeByValueUnsafely(value any) {
	// Случай попытки удаления из пустого списка
	if l.firstNode == nil {
		return
	}

	// Случай удаления первого элемента
	if l.firstNode.value == value {
		// Случай удаления первого элемента из двух оставшихся
		if l.firstNode.nextNode == l.lastNode {
			l.lastNode = l.firstNode
		}
		l.firstNode = l.firstNode.nextNode
		// Случай удаления единственного элемента
		if l.firstNode == nil {
			l.lastNode = nil
		}
		l.length--
		// Сброс типа элементов
		if l.length == 0 {
			l.V = nil
		}
		return
	}

	// Проходимся по узлам и останавливаемся на нужном
	currentNode := l.firstNode
	for ; currentNode.nextNode != nil && currentNode.nextNode.value != value; currentNode = currentNode.nextNode {
	}
	// Прошли через весь список и не нашли нужный узел
	if currentNode.nextNode == nil {
		return
	}
	// Случай удаления предпоследнего элемента
	if currentNode.nextNode == l.lastNode {
		l.lastNode = currentNode
	}
	currentNode.nextNode = currentNode.nextNode.nextNode
	l.length--
	// Сброс типа элементов
	if l.length == 0 {
		l.V = nil
	}
	return
}

// RemoveAllByValue удаляет все элементы с данным значением
func (l *List) RemoveAllByValue(value any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for currentNode := l.firstNode; currentNode != nil; currentNode = currentNode.nextNode {
		if currentNode.value == value {
			l.removeByValueUnsafely(currentNode.value)
		}
	}
}

// GetByID возвращает значение элемента с данным идентификатором
func (l *List) GetByID(id int64) (value any, ok bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if id > l.idCounter || id < l.idInitial {
		return 0, false
	}

	currentNode := l.firstNode
	for ; currentNode.id != id; currentNode = currentNode.nextNode {
		if currentNode.nextNode == nil {
			return 0, false
		}
	}
	return currentNode.value, true
}

// GetByValue возвращает идентификатор первого по порядку элемента с данным значением
// Если не находит элемента с данным значением, возвращает 0 и false
func (l *List) GetByValue(value any) (id int64, ok bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for currentNode := l.firstNode; currentNode != nil; currentNode = currentNode.nextNode {
		if currentNode.value == value {
			return currentNode.id, true
		}
	}
	return 0, false
}

// GetAllByValue возвращает индексы всех элементов с данным значением
// Если элементы с данным значением не найдены, возвращает nil и false
func (l *List) GetAllByValue(value any) (ids []int64, ok bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for currentNode := l.firstNode; currentNode != nil; currentNode = currentNode.nextNode {
		if currentNode.value == value {
			ids = append(ids, currentNode.id)
		}
	}
	// Случай пустого списка
	if len(ids) == 0 {
		return nil, false
	}
	return ids, true
}

// UpdateByID обновляет значение элемента по идентификатору
// Если элемента с таким идентификатором нет, возвращает false и nil.
// При несоответствии типа элемента возвращает false и ErrMismatchType
func (l *List) UpdateByID(id int64, value any) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Случай пустого списка
	if l.length == 0 {
		return false, nil
	}

	// Согласование типа элементов
	if l.V != reflect.TypeOf(value) {
		return false, storage.ErrMismatchType
	}

	// Проходимся по узлам
	for currentNode := l.firstNode; currentNode != nil; currentNode = currentNode.nextNode {
		if currentNode.id == id {
			currentNode.value = value
			return true, nil
		}
	}

	// Прошли по всем узлам и не нашли нужный
	return false, nil
}

// GetAll возвращает все элементы списка в виде map[int64]any. Если список пуст, возвращает nil и false.
func (l *List) GetAll() (map[int64]any, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Случай пустого списка
	if l.length == 0 {
		return nil, false
	}

	mp := make(map[int64]any, l.length)

	for currentNode := l.firstNode; currentNode != nil; currentNode = currentNode.nextNode {
		mp[currentNode.id] = currentNode.value
	}
	return mp, true
}

// Clear удаляет все элементы из списка
func (l *List) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.length = 0
	l.firstNode = nil
	l.lastNode = nil
	l.idCounter = l.idInitial
}

// Print выводит список в консоль
func (l *List) Print() {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Случай пустого списка
	if l.length == int64(0) {
		fmt.Printf("[]\n")
		return
	}

	fmt.Printf("[")
	currentNode := l.firstNode
	for ; currentNode.nextNode != nil; currentNode = currentNode.nextNode {
		fmt.Printf("{%d: %v}, ", currentNode.id, currentNode.value)
	}
	fmt.Printf("{%d: %v}]\n", currentNode.id, currentNode.value)
	return
}
