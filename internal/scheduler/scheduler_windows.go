//go:build windows

package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

// Register creates (or overwrites) two Task Scheduler tasks for the current user:
// one fired on logon, one fired on workstation unlock (Event ID 4801).
func Register(executablePath string) error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("getting current user: %w", err)
	}

	tasks := []struct {
		name string
		xml  string
	}{
		{"back-cal-logon", logonXML(executablePath, u.Username)},
		{"back-cal-unlock", unlockXML(executablePath, u.Username)},
	}

	for _, task := range tasks {
		if err := registerTask(task.name, task.xml); err != nil {
			return err
		}
		fmt.Printf("registered: %s\n", task.name)
	}
	return nil
}

func registerTask(name, xmlContent string) error {
	f, err := os.CreateTemp("", "back-cal-task-*.xml")
	if err != nil {
		return fmt.Errorf("creating temp file for task %q: %w", name, err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(xmlContent); err != nil {
		f.Close()
		return fmt.Errorf("writing task XML for %q: %w", name, err)
	}
	f.Close()

	out, err := exec.Command("schtasks", "/create", "/tn", name, "/xml", f.Name(), "/f").CombinedOutput()
	if err != nil {
		return fmt.Errorf("registering task %q: %w\n%s", name, err, out)
	}
	return nil
}

func logonXML(execPath, username string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Principals>
    <Principal id="Author">
      <UserId>%s</UserId>
      <LogonType>InteractiveToken</LogonType>
      <RunLevel>LeastPrivilege</RunLevel>
    </Principal>
  </Principals>
  <Triggers>
    <LogonTrigger>
      <Enabled>true</Enabled>
      <UserId>%s</UserId>
    </LogonTrigger>
  </Triggers>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>%s</Command>
      <Arguments>run</Arguments>
    </Exec>
  </Actions>
</Task>`, username, username, execPath)
}

func unlockXML(execPath, username string) string {
	// Event ID 4801: "The workstation was unlocked" from Microsoft-Windows-Security-Auditing.
	// Requires "Audit Other Logon/Logoff Events" to be enabled in Local Security Policy.
	subscription := `&lt;QueryList&gt;&lt;Query Id="0"&gt;&lt;Select Path="Security"&gt;` +
		`*[System[Provider[@Name='Microsoft-Windows-Security-Auditing'] and EventID=4801]]` +
		`&lt;/Select&gt;&lt;/Query&gt;&lt;/QueryList&gt;`

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Principals>
    <Principal id="Author">
      <UserId>%s</UserId>
      <LogonType>InteractiveToken</LogonType>
      <RunLevel>LeastPrivilege</RunLevel>
    </Principal>
  </Principals>
  <Triggers>
    <EventTrigger>
      <Enabled>true</Enabled>
      <Subscription>%s</Subscription>
    </EventTrigger>
  </Triggers>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>%s</Command>
      <Arguments>run</Arguments>
    </Exec>
  </Actions>
</Task>`, username, subscription, execPath)
}
