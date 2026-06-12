package core

import (
	"strconv"

	"github.com/korfra/veyra/model"
	"github.com/korfra/veyra/store"
)

type Registry struct {
	store *store.Store
}

func New(store *store.Store) *Registry {
	return &Registry{store: store}
}

func (r *Registry) Add(publicPattern, internalTarget string) error {
	return r.store.Save(model.Rule{
		PublicPattern:  publicPattern,
		InternalTarget: internalTarget,
	})
}

func (r *Registry) Get(publicPattern string) (*model.Rule, error) {
	return r.store.Get(publicPattern)
}

func (r *Registry) List() ([]model.Rule, error) {
	return r.store.List()
}

func (r *Registry) Remove(publicPattern string) error {
	return r.store.Delete(publicPattern)
}

func (r *Registry) SetPorts(httpPort, httpsPort int) error {
	err := r.store.SetSetting("http_port", strconv.Itoa(httpPort))
	if err != nil {
		return err
	}
	return r.store.SetSetting("https_port", strconv.Itoa(httpsPort))
}

func (r *Registry) GetPorts() (int, int, error) {
	httpStr, err := r.store.GetSetting("http_port")
	if err != nil {
		return 80, 443, err
	}
	httpsStr, err := r.store.GetSetting("https_port")
	if err != nil {
		return 80, 443, err
	}

	httpPort := 80
	if httpStr != "" {
		if p, err := strconv.Atoi(httpStr); err == nil {
			httpPort = p
		}
	}

	httpsPort := 443
	if httpsStr != "" {
		if p, err := strconv.Atoi(httpsStr); err == nil {
			httpsPort = p
		}
	}

	return httpPort, httpsPort, nil
}
