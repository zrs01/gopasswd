package pwd

import (
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
)

type CheckOption struct {
	Host string // spcific host to be changed
}

func PerformCheckAction(opt CheckOption) error {
	pStatus := LoadHostStatus()
	if opt.Host != "" {
		if err := pStatus.IsHostExist(opt.Host); err != nil {
			return eris.Wrap(err, "failed to check host existence")
		}
	}

	for i := 0; i < len(pStatus.Data.Hosts); i++ {
		hostLogin := &pStatus.Data.Hosts[i]

		allowCheckPassword := false
		if opt.Host == "" {
			allowCheckPassword = true
		} else if opt.Host != "" && opt.Host == hostLogin.Host {
			allowCheckPassword = true
		}

		if allowCheckPassword {
			for j := 0; j < len(hostLogin.Users); j++ {
				user := hostLogin.Users[j]
				// decrypt password
				if strings.HasPrefix(user.Password, "enc:") {
					user.Password = simpleDecrypt(user.Password[4:len(user.Password)])
				}
				err := checkHost(hostLogin.Host, hostLogin.Port, user.UserId, user.Password)
				if err != nil {
					return eris.Wrapf(err, "failed to do the checking")
				}
				fmt.Printf("[%s]:%s connect success\n", hostLogin.Host, user.UserId)
			}
		}
	}
	return nil
}

func checkHost(host string, port int, userId, password string) error {
	remote := NewSSHTerm(userId, host, port, password)
	err := remote.Connect()
	if err != nil {
		return eris.Wrapf(err, "failed connect to %s:%d with %s", host, port, userId)
	}
	defer remote.Disconnect()
	return nil
}
