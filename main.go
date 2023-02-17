package main

import (
	"fmt"
	"os"

	"github.com/rotisserie/eris"
	"github.com/urfave/cli/v2"
	"github.com/zrs01/gopasswd/pwd"
)

var version = "development"

func main() {
	var initOpt = pwd.InitOption{Encrypt: false}
	var pswdOpt = pwd.PwsdOption{Encrypt: false}

	cliapp := cli.NewApp()
	cliapp.Name = "gopasswd"
	cliapp.Usage = "Bulk CentOS remote password changing tools"
	cliapp.Version = version

	// cli.VersionPrinter = func(c *cli.Context) {
	// 	fmt.Printf("%s version %s, %s\n", c.App.Name, c.App.Version, version)
	// }

	cliapp.Commands = []*cli.Command{
		{
			Name: "init",
			// Aliases: []string{"i"},
			Usage: "Initialize the host with the login",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "hosts",
					Aliases:     []string{"s"},
					Usage:       "host name (separated by comma)",
					Destination: &initOpt.Hosts,
				},
				&cli.IntFlag{
					Name:        "port",
					Aliases:     []string{"p"},
					Usage:       "SSH port",
					Value:       0,
					Destination: &initOpt.Port,
				},
				&cli.StringFlag{
					Name:        "user",
					Aliases:     []string{"u"},
					Usage:       "login user ID",
					Destination: &initOpt.User,
				},
				&cli.BoolFlag{
					Name:        "encrypt",
					Aliases:     []string{"e"},
					Usage:       "encrypt the password",
					Destination: &initOpt.Encrypt,
				},
			},
			Action: func(c *cli.Context) error {
				if err := pwd.PerformInitAction(initOpt); err != nil {
					return eris.Wrap(err, "failed to initialize the host")
				}
				return nil
			},
		},
		{
			Name:  "passwd",
			Usage: "Change the password",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "encrypt, e",
					Aliases:     []string{"e"},
					Usage:       "encrypt the password",
					Destination: &initOpt.Encrypt,
				},
			},
			Action: func(c *cli.Context) error {
				if err := pwd.PerformPasswdAction(pswdOpt); err != nil {
					return eris.Wrapf(err, "failed to change password")
				}
				return nil
			},
		},
		{
			Name:  "check",
			Usage: "Check the initialed hosts connectivity",
			Action: func(c *cli.Context) error {
				if err := pwd.PerformCheckAction(); err != nil {
					return eris.Wrap(err, "failed to check hosts")
				}
				return nil
			},
		},
	}

	err := cliapp.Run(os.Args)
	if err != nil {
		fmt.Println(eris.ToString(err, true))
	}
}
