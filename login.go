package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"io/ioutil"
	"net"
	"encoding/json"
	"encoding/binary"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

type Query struct {
    Type  string `json:"type"`
    User string `json:"username"`
    Response string `json:"response"`
    Cmd []string `json:"cmd"`
}

const (
    protocol = "unix"
)

var sockAddr = os.Getenv("GREETD_SOCK")

func send_query(query Query) string {

    conn, err := net.Dial(protocol, sockAddr)
    if err != nil {
      log.Fatal(err)
    }

    queryb, _ := json.Marshal(query)
    lenb := make([]byte, 4)
    binary.LittleEndian.PutUint32(lenb, uint32(len(queryb)) )
    _, err = conn.Write( lenb )
    _, err = conn.Write( queryb )
    if err != nil {
      log.Fatal(err)
    }

    err = conn.(*net.UnixConn).CloseWrite()
    if err != nil {
      log.Fatal(err)
    }

    resp, err := ioutil.ReadAll(conn)
    if err != nil {
      log.Fatal(err)
    }

    conn.Close()
    return string(resp)
}

func get_users() []string {

    var users []string
    readFile, err := os.Open("/etc/passwd")
    if err != nil {
      log.Fatal(err)
    }
    fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)

    for fileScanner.Scan() {
        line := fileScanner.Text()
        user := strings.Split(line, ":")[0]
        uid, _  := strconv.Atoi(strings.Split(line, ":")[2])
        if (uid > 999) {
          users = append(users, user)
        }
    }

    readFile.Close()
    return users
}

func get_sessions() []string {
    var sessions []string
    files, err := ioutil.ReadDir("/usr/share/wayland-sessions/")
    if err != nil {
      log.Fatal(err)
    }

    for _, file := range files {
      if strings.Contains(file.Name(),".desktop") {
        ses := strings.Split(file.Name(), ".desktop")[0]
        sessions = append(sessions, ses)
      }
    }
    return sessions
}

func login(username string, password string, cmd string) string {

    send_query(Query{Type: "create_session", User: username} )
    send_query( Query{Type: "post_auth_message_response", Response: password} )
	// change dinit with your desktop wrapper script
    cmd_arr := []string{"dinit","--wm"}
    if (cmd == "shell") {
      cmd_arr = []string{"/bin/bash"}
    } else {
      cmd_arr = []string{"dinit","--wm",cmd}
    }
    login_ret := send_query(Query{Type: "start_session", Cmd: cmd_arr } )

    return string(login_ret)
}

func login_activate(form *tview.Form) bool {
    _, user := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
    pass := form.GetFormItem(1).(*tview.InputField).GetText()
    _, wm := form.GetFormItem(2).(*tview.DropDown).GetCurrentOption()
    retstr := login(user,pass,wm)
    if strings.Contains(retstr, "success") {
      return true
    } else {
      form.GetFormItem(1).(*tview.InputField).SetText("")
      form.SetFocus(0)
      return false
    }
}

func main() {
    users := get_users()
    sessions := get_sessions()
    users = append(users, "root")
    sessions = append(sessions, "shell")

    app := tview.NewApplication()
    form := tview.NewForm()
    form.AddDropDown("User", users, 0, nil).
        AddPasswordField("Password", "", 10, '*', nil).
        AddDropDown("Desktop", sessions, 0, nil).
        AddButton("Login", func() {
          if login_activate(form) {
            app.Stop()
          }
        }).
        AddButton("Reboot", func() {
          app.Stop()
          exec.Command("reboot").Run()
        })
        pf := form.GetFormItem(1).(*tview.InputField)
        pf.SetDoneFunc(func(key tcell.Key) {
          if key == tcell.KeyEnter {
            if login_activate(form) {
              app.Stop()
            }
          }
        })
        form.SetFocus(1)
        form.SetBorder(true).SetTitle("Milis Linux Login Manager").SetTitleAlign(tview.AlignLeft)
        if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
          panic(err)
        }
}
