package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sync"
)

type UriWriter struct {
	files   map[string]*os.File
	sockets map[string]net.Conn
	mu      sync.Mutex
}

func NewUriWriter() *UriWriter {
	return &UriWriter{
		files:   make(map[string]*os.File),
		sockets: make(map[string]net.Conn),
	}
}

func (uw *UriWriter) GetWriter(uriString string) (io.Writer, error) {
	u, err := url.Parse(uriString)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "file":
		return uw.getFileWriter(u.Path)
	case "tcp", "udp", "unix":
		return uw.getSocketWriter(u)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
}

func (uw *UriWriter) getFileWriter(filePath string) (io.Writer, error) {
	uw.mu.Lock()
	defer uw.mu.Unlock()

	if file, ok := uw.files[filePath]; ok {
		return file, nil
	}

	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	uw.files[filePath] = file

	return file, nil
}

func (uw *UriWriter) getSocketWriter(u *url.URL) (io.Writer, error) {
	uw.mu.Lock()
	defer uw.mu.Unlock()

	key := u.Scheme + "://" + u.Host
	if conn, ok := uw.sockets[key]; ok {
		return conn, nil
	}

	conn, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}
	uw.sockets[key] = conn

	return conn, nil
}

func (uw *UriWriter) CloseAll() error {
	uw.mu.Lock()
	defer uw.mu.Unlock()

	for _, file := range uw.files {
		if err := file.Close(); err != nil {
			return err
		}
	}

	for _, conn := range uw.sockets {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
