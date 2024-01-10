package mp

import (
	"fmt"
	"notesServer/gates/storage"
	"reflect"
	"strings"
	"sync"
)

type Map struct {
	mp        map[int64]any
	idInitial int64        // идентификатор первого добавляемого элемента
	idCounter int64        // идентификатор следующего добавляемого элемента
	V         reflect.Type // фиксируется при добавлении первого элемента, сбрасывается при удалении последнего элемента
	mu        sync.RWMutex
}

// NewMap возвращает новую таблицу, первый элемент которой будет иметь идентификатор initID
func NewMap(initID int64) (m *Map) {
	return &Map{idInitial: initID, idCounter: initID, mp: make(map[int64]any), V: nil}
}

// Len возвращает количество элементов в таблице
func (m *Map) Len() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.mp))
}

// Add добавляет значение в таблицу и возвращает его идентификатор
func (m *Map) Add(value any) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Согласование типа элементов
	if m.V == nil {
		m.V = reflect.TypeOf(value)
	} else if m.V != reflect.TypeOf(value) {
		return 0, storage.ErrMismatchType
	}

	m.mp[m.idCounter] = value
	m.idCounter++
	return m.idCounter - 1, nil
}

// RemoveByID удаляет элемент из таблицы по идентификатору
func (m *Map) RemoveByID(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.mp, id)

	// Сброс типа элементов
	if len(m.mp) == 0 {
		m.V = nil
	}
}

// RemoveByValue удаляет один элемент из таблицы по значению
func (m *Map) RemoveByValue(value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.mp {
		if v == value {
			delete(m.mp, k)

			// Сброс типа элементов
			if len(m.mp) == 0 {
				m.V = nil
			}

			return
		}
	}

}

// RemoveAllByValue удаляет все элементы из таблицы по значению
func (m *Map) RemoveAllByValue(value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.mp {
		if v == value {
			delete(m.mp, k)
		}
	}

	// Сброс типа элементов
	if len(m.mp) == 0 {
		m.V = nil
	}
}

// GetByID возвращает значение элемента по идентификатору.
// Если элемента с таким идентификатором нет, то возвращается 0 и false.
func (m *Map) GetByID(id int64) (value any, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok = m.mp[id]
	return value, ok
}

// GetByValue возвращает идентификатор первого найденного элемента по значению.
// Если элемента с таким значением нет, то возвращается 0 и false.
func (m *Map) GetByValue(value any) (id int64, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Согласование типа элементов
	if m.V != reflect.TypeOf(value) {
		return 0, false
	}

	for k, v := range m.mp {
		if v == value {
			return k, true
		}
	}
	return 0, false
}

// GetAllByValue возвращает идентификаторы всех найденных элементов с указанным значением.
// Если элементов с таким значением нет, возвращается nil и false.
func (m *Map) GetAllByValue(value any) ([]int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Согласование типа элементов
	if reflect.TypeOf(value) != m.V {
		return nil, false
	}

	var ids []int64
	for k, v := range m.mp {
		if v == value {
			ids = append(ids, k)
		}
	}

	if len(ids) == 0 {
		return nil, false
	}
	return ids, true
}

// UpdateByID обновляет значение элемента по идентификатору
// Если элемента с таким ID нет, функция возвращает false и nil.
// Если тип value отличается от типов уже присутствующих в хранилище элементов, возвращается false и ErrMismatchType.
func (m *Map) UpdateByID(id int64, value interface{}) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Согласование типа элементов
	if m.V != reflect.TypeOf(value) {
		return false, storage.ErrMismatchType
	}

	_, ok := m.mp[id]
	if !ok {
		return false, nil
	}

	m.mp[id] = value
	return true, nil
}

// GetAll возвращает все элементы хранилища в виде map[int64]any.
// Ключи map соответствуют идентификаторам элементов. Значения map соответствуют значению элементов.
// Если хранилище пусто, возвращается nil и false.
func (m *Map) GetAll() (map[int64]any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.mp) == 0 {
		return nil, false
	}

	// Создание копии, которую можно будет безопасно использовать после разблокировки m.mu
	mpCopy := make(map[int64]any, len(m.mp))
	for k, v := range m.mp {
		mpCopy[k] = v
	}

	return mpCopy, true
}

// Clear очищает таблицу
func (m *Map) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mp = make(map[int64]any)
	m.V = nil
	m.idCounter = m.idInitial
}

// Print выводит таблицу в консоль
func (m *Map) Print() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Определяем максимальные длины строковых представлений ключей и значений
	maxKeyLen := len(fmt.Sprint(m.idCounter - 1))
	maxValLen := 0
	for _, val := range m.mp {
		valLen := len(fmt.Sprint(val))
		if valLen > maxValLen {
			maxValLen = valLen
		}
	}

	// Печатаем шапку таблицы
	fmt.Printf("%-*s | %-*s\n", maxKeyLen, "ID", maxValLen, "Value")
	fmt.Println(strings.Repeat("-", maxKeyLen+3+maxValLen))

	// Печатаем тело таблицы
	for key, val := range m.mp {
		fmt.Printf("%-*v | %-*v\n", maxKeyLen, key, maxValLen, val)
	}
}
