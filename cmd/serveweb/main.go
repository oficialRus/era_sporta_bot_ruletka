// Сервер для Mini App: раздаёт web/ на порту 5173. Запуск из корня проекта: go run ./cmd/serveweb
package main

import (
	"log"
	"net"
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("web")))
	log.Println("Mini App: http://localhost:5173")
	if ip := localIP(); ip != "" {
		log.Printf("Для телефона в той же Wi-Fi: http://%s:5173", ip)
	}
	if err := http.ListenAndServe(":5173", nil); err != nil {
		log.Fatal(err)
	}
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}
