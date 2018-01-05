package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/subchen/go-cli"
	"github.com/subchen/go-log"
	"github.com/subchen/go-stack"
	"github.com/subchen/go-xmldom"
	"github.com/ungerik/go-dry"
)

const (
	RUN_ONCE_FILE = "/etc/ovfenv-installer.done"
)

var (
	BuildVersion   string
	BuildGitRev    string
	BuildGitCommit string
	BuildDate      string

	runOnce    bool
	runLogFile string
)

func main() {
	app := cli.NewApp()
	app.Name = "ovfenv-installer"
	app.Usage = "Configure networking from vSphere ovfEnv properties"
	app.UsageText = "[ OPTIONS ]"
	app.Authors = "Guoqiang Chen <subchen@gmail.com>"

	app.Flags = []*cli.Flag{
		{
			Name:  "run-once",
			Usage: "run once only",
			Value: &runOnce,
		},
		{
			Name:        "log-file",
			Usage:       "save log to file",
			Placeholder: "path",
			Value:       &runLogFile,
		},
	}

	app.Examples = `
		You can append following command-line into /etc/rc.d/rc.local (chmod +x)
		>> ovfenv-installer --run-once --log-file=/var/log/ovfenv-installer.log
	`

	// set compiler version
	if BuildVersion != "" {
		app.Version = BuildVersion + "-" + BuildGitRev
	}
	app.BuildGitCommit = BuildGitCommit
	app.BuildDate = BuildDate

	// cli action
	app.Action = func(c *cli.Context) {
		run()
	}

	app.Run(os.Args)
}

func run() {
	// run once check
	if runOnce && dry.FileExists(RUN_ONCE_FILE) {
		fmt.Println("skipped to run again in --run-once mode")
		os.Exit(0)
	}

	// run log file setting
	if runLogFile != "" {
		fw, err := os.Create(runLogFile)
		gstack.PanicIfErr(err)
		defer fw.Close()

		fmt.Printf("see log: %s\n", runLogFile)
		log.SetWriter(fw)
		log.SetFlags(log.F_TIME)
	} else {
		log.SetWriter(os.Stdout)
		log.SetFlags(log.F_TIME | log.F_COLOR)
	}

	defer func() {
		if err := recover(); err != nil {
			log.Error(gstack.AsErrorString(err))
			os.Exit(1)
		}
	}()

	log.Info("generating ovfenv ...")
	xml, err := generateOvfEnvXml()
	gstack.PanicIfErr(err)
	log.Info(strings.TrimSpace(xml))

	log.Info("parsing ovfenv ...")
	props, nics := parseOvfEnvXml(xml)

	configureNetworking(props, nics)
	configureHostname(props)
	configureDNS(props)
	configureNTP(props)

	if runOnce {
		dry.FileSetString(RUN_ONCE_FILE, "done")
	}

	log.Info("completed!")
}

func generateOvfEnvXml() (xml string, err error) {
	if _, err := exec.LookPath("vmtoolsd"); err == nil {
		return gstack.ExecCommandOutput("vmtoolsd", "--cmd", "info-get guestinfo.ovfenv")
	}

	if _, err := exec.LookPath("vmware-guestd"); err == nil {
		return gstack.ExecCommandOutput("vmware-guestd", "--cmd", "info-get guestinfo.ovfenv")
	}

	if _, err := exec.LookPath("vmware-rpctool"); err == nil {
		return gstack.ExecCommandOutput("bash", "vmware-rpctool", "info-get guestinfo.ovfenv")
	}

	panic("open-vm-tools is not installed")
}

func parseOvfEnvXml(xml string) (props map[string]string, nics int) {
	dom, err := xmldom.ParseXML(xml)
	gstack.PanicIfErr(err)

	props = make(map[string]string)
	for _, node := range dom.Root.GetChild("PropertySection").GetChildren("Property") {
		key := node.GetAttributeValue("key")
		value := node.GetAttributeValue("value")
		value = strings.TrimSpace(value)
		if len(value) > 0 {
			props[key] = value
			log.Infof("get prop: %s = %s", key, value)
		}
	}

	nics = len(dom.Root.GetChild("EthernetAdapterSection").Children)
	log.Infof("get nics: %d", nics)

	return props, nics
}

func configureNetworking(props map[string]string, nics int) {
	log.Info("configuring networking ...")
	for i := 0; i < nics; i++ {
		idx := strconv.Itoa(i)
		sb := gstack.NewStringBuilder()
		if ip, ok := props["ip"+idx]; ok && ip != "" {
			gateway, ok := props["gateway"+idx]
			if !ok && i == 0 {
				ipv4 := strings.Split(ip, ".")
				gateway = strings.Join(ipv4[0:3], ".") + ".1"
			}
			subnet, ok := props["subnet"+idx]
			if !ok {
				subnet = "255.255.255.0"
			}

			// static
			sb.Writeln("NAME=eth" + idx)
			sb.Writeln("DEVICE=eth" + idx)
			sb.Writeln("TYPE=Ethernet")
			sb.Writeln("ONBOOT=yes")
			sb.Writeln("BOOTPROTO=static")
			sb.Writeln("IPADDR=" + ip)
			sb.Writeln("GATEWAY=" + gateway)
			sb.Writeln("NETMASK=" + subnet)
			sb.Writeln("IPV6INIT=no")
			sb.Writeln("DEFROUTE=" + gstack.IIfString(i == 0, "yes", "no"))
			sb.Writeln("PEERDNS=no")
			sb.Writeln("NM_CONTROLLED=no")
		} else {
			// dhcp
			sb.Writeln("NAME=eth" + idx)
			sb.Writeln("DEVICE=eth" + idx)
			sb.Writeln("TYPE=Ethernet")
			sb.Writeln("ONBOOT=yes")
			sb.Writeln("BOOTPROTO=dhcp")
			sb.Writeln("#BOOTPROTO=static")
			sb.Writeln("#IPADDR=")
			sb.Writeln("#GATEWAY=")
			sb.Writeln("#NETMASK=255.255.255.0")
			sb.Writeln("IPV6INIT=no")
			sb.Writeln("DEFROUTE=" + gstack.IIfString(i == 0, "yes", "no"))
			sb.Writeln("PEERDNS=no")
			sb.Writeln("NM_CONTROLLED=no")
		}

		filename := fmt.Sprintf("/etc/sysconfig/network-scripts/ifcfg-eth%d", i)
		log.Infof("writing %s ...", filename)
		err := dry.FileSetString(filename, sb.String())
		gstack.PanicIfErr(err)
	}

	if nics > 0 {
		log.Info("executing: systemctl stop NetworkManager.service ...")
		gstack.ExecCommandStdpipe("systemctl", "stop", "NetworkManager.service")

		log.Info("executing: systemctl disable NetworkManager.service ...")
		gstack.ExecCommandStdpipe("systemctl", "disable", "NetworkManager.service")

		log.Info("executing: systemctl restart network.service ...")
		gstack.ExecCommandStdpipe("systemctl", "restart", "network.service")
	}
}

func configureHostname(props map[string]string) {
	log.Info("configuring hostname ...")
	if hostname, ok := props["hostname"]; ok {
		// static
		log.Info("writing /etc/hostname ...")
		err := dry.FileSetString("/etc/hostname", hostname)
		gstack.PanicIfErr(err)

		// transient
		log.Infof("executing: hostname %s ...", hostname)
		gstack.ExecCommandStdpipe("hostname", hostname)
	}
}

func configureDNS(props map[string]string) {
	log.Info("configuring dns ...")

	sb := gstack.NewStringBuilder()

	if domain, ok := props["domain"]; ok {
		sb.Writeln("domain " + domain)
	}

	if dnssearch, ok := props["dnssearch"]; ok {
		sb.Writeln("search " + dnssearch)
	}

	for i := 0; i < 5; i++ {
		idx := strconv.Itoa(i)
		if dns, ok := props["dns"+idx]; ok {
			sb.Writeln("nameserver " + dns)
		}
	}

	if sb.Len() > 0 {
		log.Info("writing /etc/resolv.conf ...")
		err := dry.FileSetString("/etc/resolv.conf", sb.String())
		gstack.PanicIfErr(err)
	}
}

func configureNTP(props map[string]string) {
	log.Info("configuring ntp ...")

	if _, err := exec.LookPath("ntpd"); err != nil {
		log.Info("ntpd is not installed, skipped to configure ntp")
		return
	}

	sb := gstack.NewStringBuilder()

	sb.Writeln("driftfile /var/lib/ntp/drift")
	sb.Writeln("restrict default nomodify notrap nopeer noquery")
	sb.Writeln("restrict 127.0.0.1")
	sb.Writeln("restrict ::1")
	sb.Writeln("")

	ntpServers := 0
	for i := 0; i < 5; i++ {
		idx := strconv.Itoa(i)
		if ntp, ok := props["ntp"+idx]; ok {
			sb.Writeln("server " + ntp)
			ntpServers++
		}
	}

	sb.Writeln("")
	sb.Writeln("includefile /etc/ntp/crypto/pw")
	sb.Writeln("keys /etc/ntp/keys")
	sb.Writeln("disable monitor")

	if ntpServers > 0 {
		log.Info("writing /etc/ntp.conf ...")
		err := dry.FileSetString("/etc/ntp.conf", sb.String())
		gstack.PanicIfErr(err)

		log.Info("executing: systemctl restart ntpd.service ...")
		gstack.ExecCommandStdpipe("systemctl", "restart", "ntpd.service")
	}
}
