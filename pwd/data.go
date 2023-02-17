package pwd

import (
	"encoding/json"
	"os"
)

type HostUser struct {
	UserId   string `json:"user"`
	Password string `json:"password"`
}

type HostLogin struct {
	Host  string     `json:"host"`
	Port  int        `json:"port"`
	Users []HostUser `json:"users"`
	// User     string `json:"user"`
	// Password string `json:"password"`
	Update string `json:"update"`
}

type HostStateData struct {
	Hosts []HostLogin `json:"hosts"`
}

type HostStatus struct {
	File string
	Data HostStateData
}

func LoadHostStatus() *HostStatus {
	status := &HostStatus{
		File: "data.json",
		Data: HostStateData{},
	}
	status.load()
	return status
}

func (s *HostStatus) load() error {
	if FileExists(s.File) {
		jsonFile, err := os.ReadFile(s.File)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(jsonFile), &s.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HostStatus) save() error {
	jsonFile, err := json.MarshalIndent(s.Data, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(s.File, jsonFile, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *HostStatus) GetState(host string) *HostLogin {
	for i := 0; i < len(s.Data.Hosts); i++ {
		if s.Data.Hosts[i].Host == host {
			return &s.Data.Hosts[i]
		}
	}
	return &HostLogin{}
}

func (s *HostStatus) SaveState(state *HostLogin) error {
	var exist = false
	for i := 0; i < len(s.Data.Hosts); i++ {
		if s.Data.Hosts[i].Host == state.Host {
			s.Data.Hosts[i] = *state
			exist = true
			break
		}
	}
	if !exist {
		s.Data.Hosts = append(s.Data.Hosts, *state)
	}
	return s.save()
}
