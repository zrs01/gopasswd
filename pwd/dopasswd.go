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
	Host     string // spcific host to be changed
}

func PerformPasswdAction(opt PwsdOption) error {
	hostStatus := LoadHostStatus()
	if opt.Host != "" {
		if err := hostStatus.IsHostExist(opt.Host); err != nil {
			return eris.Wrap(err, "failed to check host existence")
		}
	}

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
	for i := 0; i < len(hostStatus.Data.Hosts); i++ {
		hostLogin := &hostStatus.Data.Hosts[i]

		allowChangePassword := false
		if opt.Host == "" {
			allowChangePassword = true
		} else if opt.Host != "" && opt.Host == hostLogin.Host {
			allowChangePassword = true
		}

		if allowChangePassword {
			for j := 0; j < len(hostLogin.Users); j++ {
				user := hostLogin.Users[j]
				// decrypt password
				if strings.HasPrefix(user.Password, "enc:") {
					user.Password = simpleDecrypt(user.Password[4:len(user.Password)])
				}
				// ignore if old password is same as new password
				if user.Password == opt.Password {
					fmt.Printf("[%s]:%s ignore update since same password encountered\n", hostLogin.Host, user.UserId)
				} else {
					// update password
					msg, err := passwdHost(hostLogin.Host, hostLogin.Port, user.UserId, user.Password, opt.Password)
					if err != nil {
						return eris.Errorf("error: [%s]:%s %s", hostLogin.Host, user.UserId, err.Error())
					}
					fmt.Printf("[%s]:%s %s\n", hostLogin.Host, user.UserId, msg)
				}
				user.Password = opt.Password
				// encrypt password
				if opt.Encrypt {
					user.Password = "enc:" + simpleEncrypt(user.Password)
				}
				hostLogin.Users[j] = user
			}
			hostLogin.Update = time.Now().Format("2006-01-02 15:04:05")
			hostStatus.SaveState(hostLogin)
		}
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
