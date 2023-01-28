/**
Monitoria de serviço HTTP

Esse projeto é uma simulçao de monotoria de serviços
**/

package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	status        int
	ServerName    string
	ServerURL     string
	TempoExecucao float64
	dataFalha     string
}

func checkServer(servidores []Server) []Server {
	var downServers []Server
	for _, servidor := range servidores {
		agora := time.Now()
		get, err := http.Get(servidor.ServerURL)
		if err != nil {
			fmt.Printf("Server %s is down [%s]\n", servidor.ServerName, err.Error())
			servidor.status = 0
			servidor.dataFalha = agora.Format(time.UnixDate)
			downServers = append(downServers, servidor)
			continue
		}
		servidor.status = get.StatusCode
		if servidor.status != 200 {
			servidor.dataFalha = agora.Format(time.UnixDate)
			downServers = append(downServers, servidor)
		}
		servidor.TempoExecucao = time.Since(agora).Seconds()
		fmt.Printf("Status: [%d] Tempo de carga: [%f] URL: [%s]\n", servidor.status, servidor.TempoExecucao, servidor.ServerURL)
	}
	return downServers
}

func CriarListaServidores(serverList *os.File) []Server {

	csvReader := csv.NewReader(serverList)
	data, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println("Ocorreu um erro ao executar o get(url)")
		panic(err)
	}

	var servidores []Server
	for i, line := range data {
		if i > 0 {
			servidor := Server{
				ServerName: line[0],
				ServerURL:  line[1],
			}
			servidores = append(servidores, servidor)
		}
	}
	return servidores
}

func openFiles(serverListFile string, downtimeFile string) (*os.File, *os.File) {
	serverList, err := os.OpenFile(serverListFile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	downtimeList, err := os.OpenFile(downtimeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return serverList, downtimeList
}

func generateDowntime(downtimeList *os.File, downServers []Server) {
	csvWriter := csv.NewWriter(downtimeList)
	for _, servidor := range downServers {
		line := []string{servidor.ServerName, servidor.ServerURL, servidor.dataFalha, fmt.Sprintf("%f", servidor.TempoExecucao), fmt.Sprintf("%d", servidor.status)}
		csvWriter.Write(line)
	}
	csvWriter.Flush()
}

func main() {
	serverList, downtimeList := openFiles(os.Args[1], os.Args[2])
	defer serverList.Close()
	defer downtimeList.Close()
	servidores := CriarListaServidores(serverList)
	downServers := checkServer(servidores)
	generateDowntime(downtimeList, downServers)
}
