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
}

func testeste() {
}

func test(t ...int) {
	fmt.Println(len(t))
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
	cmd := exec.Command("op", "signin", o.user, "--raw")
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
	Pbe           int64    `json:"pbe"`
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
	cmd := exec.Command("op", "get", "item", itemName, "--session", o.session)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	log.Println("Getting", itemName)
	if err := cmd.Start(); err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	sErr, _ := ioutil.ReadAll(stderr)
	sOut, _ := ioutil.ReadAll(stdout)
	if len(sErr) != 0 {
		ferr := fmt.Errorf("%s", sErr)
		return getres, ferr
	}
	if err := cmd.Wait(); err != nil {
		return getres, err
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
	cmd := exec.Command("op", "get", "item", itemName, "--session", o.session, "--fields", "website,username,password")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	log.Println("Getting", itemName)
	if err := cmd.Start(); err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	sErr, _ := ioutil.ReadAll(stderr)
	sOut, _ := ioutil.ReadAll(stdout)
	if len(sErr) != 0 {
		ferr := fmt.Errorf("%s", sErr)
		return getres, ferr
	}
	if err := cmd.Wait(); err != nil {
		return getres, err
	}
	err = json.Unmarshal(sOut, &getres)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
	log.Println("Got", itemName, "successfully")
	return getres, nil
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

//this freezed at ioutil.ReadAll(stderr) -> netPoll.go:220
func (o *OpCLI) GetItemListByCategories(categories ...string) ([]ItemRes, error) {
	var itemList []ItemRes
	commandArg := []string{"list", "items", "--session", o.session}
	commandArg = append(commandArg, categories...)
	commandArg = append(commandArg, "|", "op", "get", "item", "-")
	cmd := exec.Command("op", commandArg...)
	stderr, err := cmd.StderrPipe()
	defer stderr.Close()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	log.Println("Listing items")
	if err := cmd.Start(); err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	sErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	sOut, err := ioutil.ReadAll(stdout)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	if len(sErr) != 0 {
		ferr := fmt.Errorf("%s", sErr)
		return itemList, ferr
	}
	if err := cmd.Wait(); err != nil {
		return itemList, err
	}
	fmt.Println(string(sOut))
	err = json.Unmarshal(sOut, &itemList)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	log.Println("Got", "successfully")
	return itemList, nil
}
