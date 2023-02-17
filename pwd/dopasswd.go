package pwd

import (
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/rotisserie/eris"
	"golang.org/x/crypto/ssh/terminal"
)

type PwsdOption struct {
	Password string
	Encrypt  bool
}

func PerformPasswdAction(opt PwsdOption) error {
	for {
		fmt.Print("Enter New Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return eris.Wrap(err, "failed to read the password from stdin")
		}
		opt.Password = string(bytePassword)
		opt.Password = strings.TrimSpace(opt.Password)
		if opt.Password != "" {
			break
		}
	}
	fmt.Println("")
	for {
		fmt.Print("Confirm New Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return eris.Wrap(err, "failed to read the confirm password from the stdin")
		}
		cfmPwd := string(bytePassword)
		cfmPwd = strings.TrimSpace(cfmPwd)
		if cfmPwd != "" && opt.Password == cfmPwd {
			break
		} else {
			fmt.Println("\nError: password not match")
		}
	}
	fmt.Println("")

	// change password to all users in all hosts
	pStatus := LoadHostStatus()
	for i := 0; i < len(pStatus.Data.Hosts); i++ {
		pState := &pStatus.Data.Hosts[i]

		for j := 0; j < len(pState.Users); j++ {
			user := pState.Users[j]
			// decrypt password
			if strings.HasPrefix(user.Password, "enc:") {
				user.Password = simpleDecrypt(user.Password[4:len(user.Password)])
			}
			// update password
			msg, err := passwdHost(pState.Host, pState.Port, user.UserId, user.Password, opt.Password)
			if err != nil {
				return fmt.Errorf("error: [%s]:%s %s", pState.Host, user.UserId, err.Error())
			}
			fmt.Printf("[%s]:%s %s\n", pState.Host, user.UserId, msg)

			user.Password = opt.Password
			if opt.Encrypt {
				user.Password = "enc:" + simpleEncrypt(user.Password)
			}
			pState.Users[j] = user
		}
		pState.Update = time.Now().Format("2006-01-02 15:04:05")
		pStatus.SaveState(pState)
	}
	return nil
}

func passwdHost(host string, port int, userId, oldPassword, newPassword string) (string, error) {
	remote := NewSSHTerm(userId, host, port, oldPassword)
	err := remote.Connect()
	if err != nil {
		return "", eris.Wrap(err, "failed to connect the host")
	}
	defer remote.Disconnect()

	// change password
	msg, err := remote.ChangePassword(newPassword)
	if err != nil {
		return "", eris.Wrap(err, "failed to change password")
	}
	return msg, nil
}
