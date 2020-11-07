package go1p

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"
)

type OpCLI struct {
	loginTime time.Time
	user      string
	pass      string
	session   string
	vault     string
}

func NewCli() *OpCLI {
	newOpCLi := &OpCLI{}
	return newOpCLi
}

func (o *OpCLI) SignInWithPresetPass(u, p string) error {
	o.user = u
	o.pass = p
	err := o.signInExec()
	if err != nil {
		return err
	}
	return nil
}

func (o *OpCLI) SignIn(u string) error {
	o.user = u
	var oppass string
	fmt.Print("SignIn by 1password,Please enter master password:")
	fmt.Scan(&oppass)
	err := o.signInExec()
	if err != nil {
		return err
	}
	return nil
}

func (o *OpCLI) signInExec() error {
	var cmdArg []string
	cmdArg = []string{"signin", o.user, "--raw"}
	cmd := exec.Command("op", cmdArg...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return ferr
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, o.pass)
	}()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return ferr
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return ferr
	}
	if err := cmd.Start(); err != nil {
		ferr := fmt.Errorf("%s", err)
		return ferr
	}
	sErr, _ := ioutil.ReadAll(stderr)
	sOut, _ := ioutil.ReadAll(stdout)
	if len(sErr) != 0 {
		ferr := fmt.Errorf("%s", sErr)
		return ferr
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Println("SignIn successfully!")
	o.loginTime = time.Now()
	o.session = strings.TrimSpace(string(sOut))
	return nil
}

type ItemRes struct {
	UUID         string      `json:"uuid"`
	TemplateUUID string      `json:"templateUuid"`
	FaveIndex    int64       `json:"faveIndex"`
	Trashed      string      `json:"trashed"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	ChangerUUID  string      `json:"changerUuid"`
	ItemVersion  int64       `json:"itemVersion"`
	VaultUUID    string      `json:"vaultUuid"`
	Details      ItemDetails `json:"details"`
	Overview     Overv       `json:"overview"`
}
type Overv struct {
	URLs          []URL    `json:"URLs"`
	Ainfo         string   `json:"ainfo"`
	B5AccountUUID string   `json:"b5AccountUUID"`
	Pbe           float64  `json:"pbe"`
	Pgrng         bool     `json:"pgrng"`
	Ps            int64    `json:"ps"`
	Tags          []string `json:"tags"`
	Title         string   `json:"title"`
	URL           string   `json:"url"`
}

type URL struct {
	L string `json:"l"`
	U string `json:"u"`
}

type ItemDetails struct {
	Fields          []ItemField   `json:"fields"`
	NotesPlain      string        `json:"notesPlain"`
	PasswordHistory []PasswordHis `json:"passwordHistory"`
	Sections        []Section     `json:"sections,omitempty"`
}

type Section struct {
}

type PasswordHis struct {
	Time  int64  `json:"time"`
	Value string `json:"value"`
}

type ItemField struct {
	Designation string `json:"designation"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

func (o *OpCLI) GetItemFully(itemName string) (ItemRes, error) {
	var getres ItemRes
	var cmdArg []string
	cmdArg = append(cmdArg, "get", "item", itemName, "--session", o.session)
	sOut, err := execCommand("op", cmdArg)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	err = json.Unmarshal(sOut, &getres)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	log.Println("Got", itemName, "successfully")
	return getres, nil
}

type ItemLitelyRes struct {
	Website  string `json:"website"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func (o *OpCLI) GetUsernameAndPassword(itemName string) (ItemLitelyRes, error) {
	var getres ItemLitelyRes
	var cmdArg []string
	cmdArg = append(cmdArg, "get", "item", itemName, "--session", o.session, "--fields", "website,username,password")
	sOut, err := execCommand("op", cmdArg)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	err = json.Unmarshal(sOut, &getres)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	log.Println("Got", itemName, "successfully")
	return getres, nil
}

func (o OpCLI) GetItemWithCustomizedField(itemName string, fieldName ...string) ([]byte, error) {
	fieldNameString := strings.Join(fieldName, ",")
	var cmdArg []string
	var sOut []byte
	cmdArg = append(cmdArg, "get", "item", itemName, "--session", o.session, "--fields", fieldNameString)
	sOut, err := execCommand("op", cmdArg)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	log.Println("Got", itemName, "successfully")
	return sOut, nil
}

func (o *OpCLI) GetSessionLastTime() (time.Duration, error) {
	zero := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)
	if o.loginTime == zero {
		ferr := fmt.Errorf("Did not login yet")
		return time.Duration(0), ferr
	}
	lastTime := time.Now().Sub(o.loginTime)
	return lastTime, nil
}

//if I set set ReadAll(stderr) before ReadAll(stdout) when the stdout
//output with a big size byte, this will freeze at ReadAll(stderr) ->netPoll:220
func (o OpCLI) GetListWithFlag(flags []string, flagContents []string) ([]ItemRes, error) {
	var itemList []ItemRes
	var cmdArg []string
	cmdArg = append(cmdArg, "list", "items", "--session", o.session)
	cmdArg, err := o.addFlagsToCmdArg(cmdArg, flags, flagContents)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	cmdArg = append(cmdArg, "|", "op", "get", "item", "-")
	var sOut []byte
	sOut, err = execCommand("op", cmdArg)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	err = json.Unmarshal(sOut, &itemList)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	if len(itemList) == 0 {
		ferr := fmt.Errorf("No applied flags or flags contents")
		return itemList, ferr
	}
	log.Println("Got", "successfully")
	return itemList, nil
}

func (o OpCLI) addFlagsToCmdArg(cmdArg, flags, flagContents []string) ([]string, error) {
	if len(flags) != len(flagContents) {
		ferr := fmt.Errorf("length of flags or flagContens parameta should be equal")
		return cmdArg, ferr
	}
	var dashedFlags []string
	for _, f := range flags {
		dashedF := "--" + f
		dashedFlags = append(dashedFlags, dashedF)
	}
	for i := 0; i < len(flags); i++ {
		cmdArg = append(cmdArg, dashedFlags[i])
		cmdArg = append(cmdArg, flagContents[i])
	}
	return cmdArg, nil
}

func execCommand(name string, arg []string) ([]byte, error) {
	var sOut, sErr []byte
	cmd := exec.Command(name, arg...)
	stderr, err := cmd.StderrPipe()
	defer stderr.Close()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	if err := cmd.Start(); err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	sOut, err = ioutil.ReadAll(stdout)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	sErr, err = ioutil.ReadAll(stderr)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	if len(sErr) != 0 {
		ferr := fmt.Errorf("%s", sErr)
		return sOut, ferr
	}
	if err := cmd.Wait(); err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	return sOut, nil
}
