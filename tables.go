package main

import (
	"path"
	"strings"
)

type CassandraTableMeta struct {
	Folder        string
	KeyspaceName  string
	DataDirectory string
}

func (t *CassandraTableMeta) GetManifestPath() string {
	dataDir := strings.TrimLeft(t.DataDirectory, "/")
	return path.Join(dataDir, t.KeyspaceName, t.Folder)
}

func (t *CassandraTableMeta) GetDataPath() string {
	return path.Join(t.DataDirectory, t.KeyspaceName, t.Folder)
}
