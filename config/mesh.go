package config

import (
	"fmt"

	"github.com/zoobz-io/check"
)

// Mesh holds configuration for the aegis mesh node.
type Mesh struct {
	ID      string `env:"APP_MESH_ID" default:"cicero"`
	Name    string `env:"APP_MESH_NAME" default:"Cicero Translation Service"`
	Host    string `env:"APP_MESH_HOST" default:"0.0.0.0"`
	Port    int    `env:"APP_MESH_PORT" default:"5001"`
	CertDir string `env:"APP_MESH_CERT_DIR" default:"./certs"`
}

// Addr returns the mesh listen address.
func (m Mesh) Addr() string {
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}

// Validate checks that the mesh configuration is valid.
func (m Mesh) Validate() error {
	return check.All(
		check.Str(m.ID, "id").Required().V(),
		check.Str(m.Name, "name").Required().V(),
		check.Str(m.CertDir, "cert_dir").Required().V(),
		check.Int(m.Port, "port").Positive().Max(65535).V(),
	).Err()
}
