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

	cliapp := cli.NewApp()
	cliapp.Name = "gopasswd"
	cliapp.Usage = "Bulk CentOS remote password changing tools"
	cliapp.Version = version
	cliapp.Commands = []*cli.Command{}

	cliapp.Commands = append(cliapp.Commands, func() *cli.Command {
		return &cli.Command{
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
		}
	}())

	cliapp.Commands = append(cliapp.Commands, func() *cli.Command {
		var pswdOpt = pwd.PwsdOption{Encrypt: false}
		return &cli.Command{
			Name:  "passwd",
			Usage: "Change the password",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "encrypt",
					Aliases:     []string{"e"},
					Usage:       "encrypt the password",
					Destination: &pswdOpt.Encrypt,
				},
				&cli.StringFlag{
					Name:        "host",
					Aliases:     []string{"s"},
					Usage:       "host name to be changed password",
					Destination: &pswdOpt.Host,
				},
			},
			Action: func(c *cli.Context) error {
				return pwd.PerformPasswdAction(pswdOpt)
			},
		}
	}())

	cliapp.Commands = append(cliapp.Commands, func() *cli.Command {
		var checkOpt = pwd.CheckOption{}
		return &cli.Command{
			Name:  "check",
			Usage: "Check the initialed hosts connectivity",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "host",
					Aliases:     []string{"s"},
					Usage:       "host name to be changed password",
					Destination: &checkOpt.Host,
				},
			},
			Action: func(c *cli.Context) error {
				if err := pwd.PerformCheckAction(checkOpt); err != nil {
					return eris.Wrap(err, "failed to check hosts")
				}
				return nil
			},
		}
	}())

	err := cliapp.Run(os.Args)
	if err != nil {
		format := eris.NewDefaultStringFormat(eris.FormatOptions{
			InvertOutput: true,
			WithTrace:    true,
			InvertTrace:  true,
		})
		fmt.Println(eris.ToCustomString(err, format))
	}
}
