// Reference: https://forum.golangbridge.org/t/crypto-ssh-how-to-read-the-stdout-pipe-end/5714/26
package pwd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"golang.org/x/crypto/ssh"
)

type SSHTerm struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  int

	Client *ssh.Client
}

func NewSSHTerm(user string, host string, port int, password string) *SSHTerm {
	return &SSHTerm{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  2,
	}
}

func (s *SSHTerm) Connect() error {
	config := &ssh.ClientConfig{
		User: s.User,
		// Auth: []ssh.AuthMethod{
		// 	ssh.PublicKeys(key),
		// },
		//alternatively, you could use a password
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Connect
	client, err := ssh.Dial("tcp", s.Host+":"+strconv.Itoa(s.Port), config)
	if err != nil {
		return eris.Wrapf(err, "failed to connect to %s", s.Host)
	}
	if client == nil {
		return eris.New("can't connect, maybe password incorrect")
	}
	s.Client = client
	return nil
}

func (s *SSHTerm) Disconnect() error {
	if s.Client != nil {
		err := s.Client.Close()
		if err == nil {
			s.Client = nil
		}
		return eris.Wrapf(err, "failed to disconnect %s", s.Host)
	}
	return nil
}

func (s *SSHTerm) ChangePassword(newPassword string) (string, error) {
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	session, err := s.Client.NewSession()
	if err != nil {
		return "", eris.Wrap(err, "failed to create new session")
	}
	defer session.Close()

	sshOut, err := session.StdoutPipe()
	if err != nil {
		return "", eris.Wrap(err, "failed to get stdout pipe")
	}
	sshIn, err := session.StdinPipe()
	if err != nil {
		return "", eris.Wrap(err, "failed to get stdin pipe")
	}
	if err := session.RequestPty("xterm", 0, 200, modes); err != nil {
		return "", eris.Wrap(err, "failed to make remote pty")
	}
	if err := session.Shell(); err != nil {
		return "", eris.Wrap(err, "failed to start login shell")
	}

	step := func(expect []string, cmd string) error {
		lastMsg := expect[len(expect)-1]
		content := s.readBuff(expect, sshOut, s.Timeout)
		if s.isMatch(s.getLastLine(content), lastMsg) {
			if _, err := s.writeBuff(cmd, sshIn); err != nil {
				return eris.Wrap(err, "failed to get last message from the shell")
			}
			return nil
		}
		return eris.New("Last message is not equal to " + lastMsg)
	}

	// CentOS
	prompt := s.User + "@.*\\$"
	content := s.readBuff([]string{prompt, "\\(current\\) UNIX password:"}, sshOut, s.Timeout)
	if s.isMatch(s.getLastLine(content), prompt+".*$") {
		if _, err := s.writeBuff("passwd", sshIn); err != nil {
			return "", eris.Wrap(err, "failed to send 'passwd' to the shell")
		}
		if err := step([]string{prompt, "\\(current\\) UNIX password:"}, s.Password); err != nil {
			return "", eris.Wrap(err, "failed to get the message '\\(current\\) UNIX password:' from the shell")
		}
	} else {
		// change password immedately
		if _, err := s.writeBuff(s.Password, sshIn); err != nil {
			return "", eris.Wrap(err, "failed to send new password to the shell")
		}
	}

	if err := step([]string{prompt, "New password:"}, newPassword); err != nil {
		return "", eris.Wrap(err, "failed to get the message 'New password:' from the shell")
	}
	content = s.readBuff([]string{"Retype new password:", "BAD PASSWORD:"}, sshOut, s.Timeout)
	if strings.Contains(content, "BAD PASSWORD:") {
		if line, err := s.getMatchLine(content, "BAD PASSWORD:"); err != nil {
			return "", eris.Wrap(err, "failed to get the message 'BAD PASSWORD:' from the shell")
		} else {
			return "", eris.New(strings.TrimSpace(line))
		}
	} else {
		// retype new password
		if _, err := s.writeBuff(newPassword, sshIn); err != nil {
			return "", eris.Wrap(err, "failed to send new password to the shell")
		}
	}

	content = s.readBuff([]string{prompt}, sshOut, s.Timeout)
	if strings.Contains(content, "successfully") {
		if line, err := s.getMatchLine(content, "successfully"); err != nil {
			return "", eris.Wrap(err, "failed to get the message 'successfully' from the shell")
		} else {
			return strings.ReplaceAll(line, "\n", "\\n"), nil
		}
	}
	return "", eris.New(content)
}

func (s *SSHTerm) getLastLine(content string) string {
	parts := strings.Split(content, "\n")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (s *SSHTerm) isMatch(content string, pattern string) bool {
	reg := regexp.MustCompile(pattern)
	m := reg.FindStringSubmatch(content)
	return len(m) != 0
}

func (s *SSHTerm) getMatchLine(content string, pattern string) (string, error) {
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		if s.isMatch(lines[i], pattern) {
			return lines[i], nil
		}
	}
	return "", errors.New("No match for " + pattern)
}

func (s *SSHTerm) readBuffForString(whattoexpect []string, sshOut io.Reader, buffRead chan<- string) {
	buf := make([]byte, 1000)
	waitingString := ""
	for {
		n, err := sshOut.Read(buf) //this reads the ssh terminal
		if err != nil && err != io.EOF {
			fmt.Println(err)
			break
		}
		if err == io.EOF || n == 0 {
			break
		}

		current := string(buf[:n])
		waitingString += current

		match := false
		for i := 0; i < len(whattoexpect); i++ {
			// fmt.Println(">>> ", current, " [vs] ", whattoexpect[i], " [=] ", s.isMatch(current, whattoexpect[i]))
			if s.isMatch(current, whattoexpect[i]) {
				match = true
				break
			}
		}
		if IsDebug() {
			fmt.Printf("%v", current)
		}
		if match {
			// fmt.Printf("match = %v, %d, %v", match, len(match), whattoexpect)
			break
		}
		// fmt.Printf("2*** %v\r", current)
	}
	buffRead <- waitingString
}

func (s *SSHTerm) readBuff(whattoexpect []string, sshOut io.Reader, timeoutSeconds int) string {
	ch := make(chan string)
	go func(whattoexpect []string, sshOut io.Reader) {
		buffRead := make(chan string)
		go s.readBuffForString(whattoexpect, sshOut, buffRead)
		select {
		case ret := <-buffRead:
			ch <- ret
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			s.handleError(fmt.Errorf("%d", timeoutSeconds), true, "Timeout: took longer than %s seconds, perhaps you've entered incorrect details?")
		}
	}(whattoexpect, sshOut)
	return <-ch
}

func (s *SSHTerm) writeBuff(command string, sshIn io.WriteCloser) (int, error) {
	if IsDebug() {
		fmt.Print(command + "\r")
	}
	returnCode, err := sshIn.Write([]byte(command + "\r"))
	return returnCode, err
}

func (s *SSHTerm) handleError(e error, fatal bool, customMessage ...string) {
	var errorMessage string
	if e != nil {
		if len(customMessage) > 0 {
			errorMessage = strings.Join(customMessage, " ")
		} else {
			errorMessage = "%s"
		}
		if fatal {
			log.Fatalf(errorMessage, e)
		} else {
			log.Print(errorMessage, e)
		}
	}
}
