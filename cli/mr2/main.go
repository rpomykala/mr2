// Copyright (c) 2019-present Cloud <cloud@txthinking.com>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of version 3 of the GNU General Public
// License as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/txthinking/mr2"
	"github.com/urfave/cli"
)

var debug bool
var debugListen string

func main() {
	app := cli.NewApp()
	app.Name = "Mr.2"
	app.Version = "20200102"
	app.Usage = "Expose local server to external network"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			Usage:       "Enable debug, more logs",
			Destination: &debug,
		},
		&cli.StringFlag{
			Name:        "listen",
			Aliases:     []string{"l"},
			Usage:       "Listen address for debug",
			Value:       "127.0.0.1:6060",
			Destination: &debugListen,
		},
	}
	app.Commands = []*cli.Command{
		&cli.Command{
			Name:  "server",
			Usage: "Run as server mode",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "listen",
					Aliases: []string{"l"},
					Usage:   "Listen address, like: 1.2.3.4:5",
				},
				&cli.StringFlag{
					Name:    "password",
					Aliases: []string{"p"},
					Usage:   "Password",
				},
				&cli.StringSliceFlag{
					Name:    "portPassword",
					Aliases: []string{"P"},
					Usage:   "Only allow this port and password, like '1000 password'. If you specify this parameter, --password will be ignored",
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("listen") == "" {
					cli.ShowCommandHelp(c, "server")
					return nil
				}
				if debug {
					go func() {
						log.Println(http.ListenAndServe(debugListen, nil))
					}()
				}
				s, err := mr2.NewServer(c.String("listen"), c.String("password"), c.StringSlice("portPassword"))
				if err != nil {
					return err
				}
				go func() {
					sigs := make(chan os.Signal, 1)
					signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
					<-sigs
					s.Shutdown()
				}()
				return s.ListenAndServe()
			},
		},
		&cli.Command{
			Name:  "client",
			Usage: "Run as client mode",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "server",
					Aliases: []string{"s"},
					Usage:   "Server address, like: 1.2.3.4:5",
				},
				&cli.StringFlag{
					Name:    "password",
					Aliases: []string{"p"},
					Usage:   "Password",
				},
				&cli.Int64Flag{
					Name:    "serverPort",
					Aliases: []string{"P"},
					Usage:   "Server port you want to use. When server run as port mode",
				},
				&cli.StringFlag{
					Name:    "serverDomain",
					Aliases: []string{"D"},
					Usage:   "Server subdomain you want to use. When server run as domain mode. Only support official server now.",
				},
				&cli.StringFlag{
					Name:    "clientServer, c",
					Aliases: []string{"c"},
					Usage:   "Client server address, like: 1.2.3.4:5",
				},
				&cli.StringFlag{
					Name:  "clientDirectory",
					Usage: "Client directory, like: /path/to/www. If you specify this parameter, --clientServer will be ignored",
				},
				&cli.Int64Flag{
					Name:  "clientPort",
					Usage: "Work with --clientDirectory",
					Value: 54321,
				},
				&cli.Int64Flag{
					Name:  "tcpTimeout",
					Value: 60,
					Usage: "connection tcp keepalive timeout (s), works with --serverPort",
				},
				&cli.Int64Flag{
					Name:  "tcpDeadline",
					Value: 0,
					Usage: "connection deadline time (s), works with --serverPort",
				},
				&cli.Int64Flag{
					Name:  "udpDeadline",
					Value: 60,
					Usage: "connection deadline time (s), works with --serverPort",
				},
			},
			Action: func(c *cli.Context) error {
				cs := c.String("clientServer")
				if c.String("clientDirectory") != "" {
					go func() {
						log.Println(http.ListenAndServe(":"+strconv.FormatInt(c.Int64("clientPort"), 10), http.FileServer(http.Dir(c.String("clientDirectory")))))
					}()
					cs = "localhost:" + strconv.FormatInt(c.Int64("clientPort"), 10)
				}
				s := mr2.NewClient(c.String("server"), c.String("password"), c.Int64("serverPort"), c.String("serverDomain"), cs, c.Int64("tcpTimeout"), c.Int64("tcpDeadline"), c.Int64("udpDeadline"))
				go func() {
					sigs := make(chan os.Signal, 1)
					signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
					<-sigs
					s.Shutdown()
				}()
				return s.ListenAndServe()
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
