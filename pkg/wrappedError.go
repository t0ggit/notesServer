package pkg

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	logFileName = "log.txt"
	errorTag    = "[ERROR]"
	messageTag  = "(msg)"
)

// WrappedError представляет собой структуру для обертывания ошибки и записи ее в логи.
// Реализует интерфейс error.
// Дополнительно реализует функционал вывода сообщений (не ошибок) в логи и консоль.
type WrappedError struct {
	functionName string   // Имя функции (где произошла ошибка?)
	comment      string   // Комментарий к ошибке (что именно вызвало ошибку?)
	err          error    // Ошибка, которая будет обернута
	timestamp    string   // Время последнего обновления ошибки методом Specify()
	logFile      *os.File // Указатель на файл для записи логов
}

// NewWrappedError создает новый экземпляр WrappedError с именем функции, но без комментария.
// То есть уже известно, где ошибка может произойти, но что именно за ошибка еще неизвестно.
func NewWrappedError(funcName string) *WrappedError {
	return &WrappedError{funcName, "", nil, "[]", nil}
}

// NewWrappedErrorWithFile аналогична NewWrappedError, но с указателем на файл записи в логи.
// Если файл не удалось открыть, то экземпляр не создается и возвращается nil, error.
func NewWrappedErrorWithFile(funcName string) (*WrappedError, error) {
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &WrappedError{funcName, "", nil, "[]", file}, nil
}

// Specify обновляет экземпляр, если переданная ошибка не nil. Перезаписываются err и comment.
// То есть уже известно, что это за ошибка. Функция, в которой появляется ошибка, указывается при создании.
func (e *WrappedError) Specify(err error, comment string) *WrappedError {
	if err != nil {
		e.err = err
		e.comment = comment
		e.timestamp = fmt.Sprintf("[%s]", time.Now().Format(time.RFC3339))
	}
	return e
}

// Error возвращает строковое представление ошибки с комментарием и именем функции.
// Имплементируется метод интерфейса error.
func (e *WrappedError) Error() string {
	if e.err == nil {
		return ""
	}
	return fmt.Sprintf("'%s' in function '%s' invoked '%s'", e.comment, e.functionName, e.err.Error())
}

// LogError выводит ошибку в стандартный вывод и записывает ее в файл логов (если файл логов был открыт).
// Если ошибки нет, то ничего не делает.
func (e *WrappedError) LogError() {
	if e.err != nil {
		log.Println(e.timestamp, errorTag, e.Error())
		if e.logFile != nil {
			_, writeError := fmt.Fprintln(e.logFile, e.timestamp, errorTag, e.Error())
			if writeError != nil {
				log.Println("Failed to write log into opened file:", writeError)
			}
		}
	}
}

// LogMsg выводит сообщение в стандартный вывод и записывает ее в файл логов (если файл логов был открыт).
// Это сообщение не является ошибкой, но выводится в консоль и в файл логов.
func (e *WrappedError) LogMsg(msg string) {
	msgTimestamp := fmt.Sprintf("[%s]", time.Now().Format(time.RFC3339))
	log.Println(msgTimestamp, messageTag, fmt.Sprintf("'%s' from function '%s'", msg, e.functionName))
	if e.logFile != nil {
		_, writeError := fmt.Fprintln(e.logFile, msgTimestamp, messageTag, fmt.Sprintf("'%s' from function '%s'", msg, e.functionName))
		if writeError != nil {
			log.Println("Failed to write log into opened file:", writeError)
		}
	}
}

// Close закрывает файл логов.
// Не возвращает ошибку даже если файл уже был закрыт.
func (e *WrappedError) Close() {
	e.logFile.Close()
}
