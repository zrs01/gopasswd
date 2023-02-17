package pwd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	"golang.org/x/crypto/ssh/terminal"
)

type InitOption struct {
	User     string
	Hosts    string
	Port     int
	Password string
	Encrypt  bool
}

func PerformInitAction(opt InitOption) error {
	if opt.Hosts == "" {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter Host(s): ")
			opt.Hosts, _ = reader.ReadString('\n')
			opt.Hosts = strings.TrimSpace(opt.Hosts)
			// cli.ShowCommandHelpAndExit(c, "init", 1)
			if opt.Hosts != "" {
				break
			}
		}
	}

	// get the port from console
	if opt.Port == 0 {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter Port (22): ")
			ps, _ := reader.ReadString('\n')
			ps = strings.TrimSpace(ps)
			if ps == "" {
				ps = "22"
			}
			pi, err := strconv.Atoi(ps)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
			} else {
				opt.Port = pi
				break
			}
		}
	}

	// get the user name from console
	if opt.User == "" {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter Login User: ")
			opt.User, _ = reader.ReadString('\n')
			opt.User = strings.TrimSpace(opt.User)
			if opt.User != "" {
				break
			}
		}
	}

	if opt.Password == "" {
		for {
			fmt.Print("Enter Password: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return eris.Wrap(err, "failed to read the password")
			}
			opt.Password = string(bytePassword)
			opt.Password = strings.TrimSpace(opt.Password)
			if opt.Password != "" {
				break
			}
		}
	}
	fmt.Println("")

	hs := LoadHostStatus()
	hosts := strings.Split(opt.Hosts, ",")
	for i := 0; i < len(hosts); i++ {
		sta := hs.GetState(hosts[i])
		err := initHost(sta, hosts[i], opt.Port, opt.User, opt.Password, opt.Encrypt)
		if err != nil {
			return fmt.Errorf("Error: [" + hosts[i] + "] " + err.Error())
		}
		fmt.Println("[" + hosts[i] + "] updated")
		hs.SaveState(sta)

	}
	return nil
}

func initHost(pState *HostLogin, host string, port int, user string, password string, encrypt bool) error {
	remote := NewSSHTerm(user, host, port, password)
	err := remote.Connect()
	if err != nil {
		return eris.Wrap(err, "failed to connect remote host")
	}
	defer remote.Disconnect()

	t := time.Now()
	pState.Host = host
	pState.Port = port
	_, index, ok := lo.FindIndexOf(pState.Users, func(u HostUser) bool {
		return u.UserId == user
	})

	// create encrypted password if require
	pwd := password
	if encrypt {
		pwd = "enc:" + simpleEncrypt(password)
	}

	if ok {
		pState.Users[index].UserId = user
		pState.Users[index].Password = pwd
	} else {
		pState.Users = append(pState.Users, HostUser{UserId: user, Password: pwd})
	}
	pState.Update = t.Format("2006-01-02 15:04:05")
	return nil
}
