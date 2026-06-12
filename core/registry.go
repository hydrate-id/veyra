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
	rules, err := r.List()
	if err != nil {
		return err
	}

	newRule := model.Rule{
		PublicPattern:  publicPattern,
		InternalTarget: internalTarget,
	}

	if err := ValidateRule(newRule, rules); err != nil {
		return err
	}

	return r.store.Save(newRule)
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

func (r *Registry) SetAdminPort(port int) error {
	return r.store.SetSetting("admin_port", strconv.Itoa(port))
}

func (r *Registry) GetAdminPort() (int, error) {
	str, err := r.store.GetSetting("admin_port")
	if err != nil {
		return 2019, err
	}
	if str == "" {
		return 2019, nil
	}
	port, err := strconv.Atoi(str)
	if err != nil {
		return 2019, nil
	}
	return port, nil
}
