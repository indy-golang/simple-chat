# simple-chat
A simple IRC-like chat created for the Apr 7, 2015 Indy Golang Meeting

## General

This is a very simple chat server written in Go, for demonstrating concurrency and the use of channels.

This package is self contained and doesn't require any dependencies. 

The server is basically a webserver which serves the files necessary to run the chat clients in a browser, plus it handles the chat traffic via websockets.

## Usage

- Download / clone the project
- run the server: `go run server.go`
- connect to the server from any computer in your local network: `http://_server_ip_or_name_:8000`

That's it! 
