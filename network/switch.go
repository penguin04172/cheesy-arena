package network

import (
	"fmt"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"golang.org/x/crypto/ssh"
)

const (
	switchSshPort                  = 22
	switchConnectTimeoutSec        = 1
	switchCommandTimeoutSec        = 30
	switchPollPeriodSec            = 3
	switchRequestBufferSize        = 10
	switchConfigRetryIntervalSec   = 30
	switchConfigBackoffDurationSec = 5
	switchConfigPauseDurationSec   = 2
)

const (
	red1Vlan  = 10
	red2Vlan  = 20
	red3Vlan  = 30
	blue1Vlan = 40
	blue2Vlan = 50
	blue3Vlan = 60
)

type Switch struct {
	address               string
	port                  int
	username              string
	password              string
	configBackoffDuration time.Duration
	configPauseDuration   time.Duration
}

var ServerIpAddress = "10.0.100.5" // The DS will try to connect to this address only.

func NewSwitch(address, username, password string) *Switch {
	return &Switch{
		address:               address,
		port:                  switchSshPort,
		username:              username,
		password:              password,
		configBackoffDuration: switchConfigBackoffDurationSec * time.Second,
		configPauseDuration:   switchConfigPauseDurationSec * time.Second,
	}
}

func (sw *Switch) generateTeamSwitchConfig(team *model.Team, vlan int) (string, error) {
	if vlan != 10 && vlan != 20 && vlan != 30 && vlan != 40 && vlan != 50 && vlan != 60 {
		return "", fmt.Errorf("invalid vlan %d", vlan)
	}

	command := ""
	if team == nil {
		command += fmt.Sprintf(
			"set network.vlan%d.proto='none'\n"+
				"del network.vlan%d.ipaddr\n"+
				"del network.vlan%d.netmask\n"+
				"set dhcp.vlan%d.ignore='1'\n",
			vlan, vlan, vlan, vlan,
		)
	} else {
		teamPartialIp := fmt.Sprintf("%d.%d", team.Id/100, team.Id%100)
		command += fmt.Sprintf(
			"set network.vlan%d.proto='static'\n"+
				"set network.vlan%d.ipaddr='10.%s.4'\n"+
				"set network.vlan%d.netmask='255.255.255.0'\n"+
				"set dhcp.vlan%d.ignore='0'\n",
			vlan,
			vlan,
			teamPartialIp,
			vlan,
			vlan,
		)
	}
	return command, nil
}

func (sw *Switch) ConfigureTeamEthernet(teams [6]*model.Team) error {
	// Remove old team VLANs to reset the switch state.
	removeTeamVlansCommand := ""
	for vlan := 10; vlan <= 60; vlan += 10 {
		command, _ := sw.generateTeamSwitchConfig(nil, vlan)
		removeTeamVlansCommand += command
	}
	_, err := sw.runConfigCommand(removeTeamVlansCommand)
	if err != nil {
		return err
	}
	_, _ = sw.runCommand("uci commit network\n")
	_, _ = sw.runCommand("uci commit dhcp\n")
	_, _ = sw.runCommand("service network restart\n")
	_, _ = sw.runCommand("service dnsmasq restart\n")
	_, _ = sw.runCommand("service odhcpd restart\n")
	time.Sleep(sw.configPauseDuration)

	addTeamVlansCommand := ""
	addTeamVlan := func(team *model.Team, vlan int) {
		if team == nil {
			return
		}
		command, _ := sw.generateTeamSwitchConfig(team, vlan)
		addTeamVlansCommand += command
	}

	addTeamVlan(teams[0], red1Vlan)
	addTeamVlan(teams[1], red2Vlan)
	addTeamVlan(teams[2], red3Vlan)
	addTeamVlan(teams[3], blue1Vlan)
	addTeamVlan(teams[4], blue2Vlan)
	addTeamVlan(teams[5], blue3Vlan)
	if len(addTeamVlansCommand) > 0 {
		_, err = sw.runConfigCommand(addTeamVlansCommand)
		if err != nil {
			return err
		}
		_, _ = sw.runCommand("uci commit network\n")
		_, _ = sw.runCommand("uci commit dhcp\n")
		_, _ = sw.runCommand("service network restart\n")
		_, _ = sw.runCommand("service dnsmasq restart\n")
		_, _ = sw.runCommand("service odhcpd restart\n")
	}

	// Give some time for the configuration to take before another one can be attempted.
	time.Sleep(sw.configBackoffDuration)
	return nil
}

// Logs into the access point via SSH and runs the given shell command.
func (sw *Switch) runCommand(command string) (string, error) {
	// Open an SSH connection to the sw.
	config := &ssh.ClientConfig{User: sw.username,
		Auth:            []ssh.AuthMethod{ssh.Password(sw.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         switchConnectTimeoutSec * time.Second}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sw.address, switchSshPort), config)
	if err != nil {
		return "", err
	}
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	defer conn.Close()

	// Run the command with a timeout.
	commandChan := make(chan sshOutput, 1)
	go func() {
		outputBytes, err := session.Output(command)
		commandChan <- sshOutput{string(outputBytes), err}
	}()
	select {
	case output := <-commandChan:
		return output.output, output.err
	case <-time.After(switchCommandTimeoutSec * time.Second):
		return "", fmt.Errorf("WiFi SSH command timed out after %d seconds", switchCommandTimeoutSec)
	}
}

// Logs into the switch via Telnet and runs the given command in global configuration mode. Reads the output
// and returns it as a string.
func (sw *Switch) runConfigCommand(command string) (string, error) {
	return sw.runCommand(addConfigurationHeader(command))
}
