//go:build windows

package scheduler

import (
	"fmt"
	"os"
	"os/exec"
)

func Register(executablePath string) error {
	if err := registerTask("back-cal-logon", logonXML(executablePath)); err != nil {
		return fmt.Errorf("registering logon task: %w", err)
	}
	if err := registerTask("back-cal-unlock", unlockXML(executablePath)); err != nil {
		return fmt.Errorf("registering unlock task: %w", err)
	}
	return nil
}

func registerTask(name, xml string) error {
	f, err := os.CreateTemp("", "back-cal-task-*.xml")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(xml); err != nil {
		f.Close()
		return err
	}
	f.Close()

	cmd := exec.Command("schtasks", "/create", "/tn", name, "/xml", f.Name(), "/f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func logonXML(execPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Triggers>
    <LogonTrigger><Enabled>true</Enabled></LogonTrigger>
  </Triggers>
  <Actions Context="Author">
    <Exec><Command>%s</Command><Arguments>run</Arguments></Exec>
  </Actions>
  <Settings><MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy></Settings>
</Task>`, execPath)
}

func unlockXML(execPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Triggers>
    <EventTrigger>
      <Enabled>true</Enabled>
      <Subscription>&lt;QueryList&gt;&lt;Query Id="0"&gt;&lt;Select Path="Security"&gt;*[System[Provider[@Name='Microsoft-Windows-Security-Auditing'] and EventID=4801]]&lt;/Select&gt;&lt;/Query&gt;&lt;/QueryList&gt;</Subscription>
    </EventTrigger>
  </Triggers>
  <Actions Context="Author">
    <Exec><Command>%s</Command><Arguments>run</Arguments></Exec>
  </Actions>
  <Settings><MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy></Settings>
</Task>`, execPath)
}
