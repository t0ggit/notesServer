package main

import (
	"notesServer/controllers/notesService"
	"notesServer/gates/storage"
	"notesServer/gates/storage/mp"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var st storage.Storage
	st = mp.NewMap(1)

	ns := notesService.NewNotesService(":8080", st)

	signalCh := make(chan os.Signal, 1)     // канал для получения сигнала
	signal.Notify(signalCh, syscall.SIGINT) // привязываем его к сигналу SIGINT
	go func() {
		<-signalCh
		_ = ns.Close()
	}()

	ns.Start()
}
