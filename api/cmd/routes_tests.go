package main

import "net/http"

func simulatePanic(w http.ResponseWriter, r *http.Request) error {
	panic("falha simulada")
}
