package pwd

import (
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
)

func PerformCheckAction() error {
	pStatus := LoadHostStatus()
	for i := 0; i < len(pStatus.Data.Hosts); i++ {
		pState := &pStatus.Data.Hosts[i]

		for j := 0; j < len(pState.Users); j++ {
			user := pState.Users[j]
			// decrypt password
			if strings.HasPrefix(user.Password, "enc:") {
				user.Password = simpleDecrypt(user.Password[4:len(user.Password)])
			}
			err := checkHost(pState.Host, pState.Port, user.UserId, user.Password)
			if err != nil {
				return eris.Wrapf(err, "failed to do the checking")
			}
			fmt.Printf("[%s]:%s connect success\n", pState.Host, user.UserId)
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
