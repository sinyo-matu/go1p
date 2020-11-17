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

//OpCLI 1Password CLI struct
type OpCLI struct {
	expirationTime   time.Time
	user             string
	pass             string
	session          string
	keepSessionAlive bool
}

//NewCli initial a 1Password CLI struct instance
func NewCli() *OpCLI {
	newOpCLi := &OpCLI{}
	return newOpCLi
}

//SignInWithPresetPass if keepSignInSessionAlive = false, the auth session
//will be invalidated 30 minutes later after you signIn
func (o *OpCLI) SignInWithPresetPass(username, password string, keepSignInSessionAlive bool) error {
	o.user = username
	o.pass = password
	o.keepSessionAlive = keepSignInSessionAlive
	err := o.signInExec()
	if err != nil {
		return err
	}
	return nil
}

//SignIn will request password type in from command line
//if keepSignInSessionAlive = false, the auth session will be invalidated 30 minutes later after you signIn
func (o *OpCLI) SignIn(username string, keepSignInSessionAlive bool) error {
	o.user = username
	o.keepSessionAlive = keepSignInSessionAlive
	var oppass string
	fmt.Print("SignIn by 1password,Please enter master password:")
	fmt.Scan(&oppass)
	o.pass = oppass
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
	o.expirationTime = time.Now().Add(time.Second * 1790)
	o.session = strings.TrimSpace(string(sOut))
	return nil
}

//ItemRes 1Password Item with fully fields
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

//Overv ItemRes field
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

//URL ItemRes field
type URL struct {
	L string `json:"l"`
	U string `json:"u"`
}

//ItemDetails ItemRes field
type ItemDetails struct {
	Fields          []ItemField   `json:"fields"`
	NotesPlain      string        `json:"notesPlain"`
	PasswordHistory []PasswordHis `json:"passwordHistory"`
	Sections        []Section     `json:"sections,omitempty"`
}

//Section ItemRes field
type Section struct {
}

//PasswordHis ItemRes field
type PasswordHis struct {
	Time  int64  `json:"time"`
	Value string `json:"value"`
}

//ItemField ItemRes field
type ItemField struct {
	Designation string `json:"designation"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

//GetItemFully get an item with all of fields
func (o *OpCLI) GetItemFully(itemName string) (ItemRes, error) {
	var getres ItemRes
	err := o.checkSessionAliveOrSignIn()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
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

//ItemLitelyRes 1Password item struct with simplely 3 fields
type ItemLitelyRes struct {
	Website  string `json:"website"`
	Password string `json:"password"`
	Username string `json:"username"`
}

//GetUsernameAndPassword To get item with simplely 3 fields usually we need
func (o *OpCLI) GetUsernameAndPassword(itemName string) (ItemLitelyRes, error) {
	var getres ItemLitelyRes
	err := o.checkSessionAliveOrSignIn()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return getres, ferr
	}
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

//GetItemWithCustomizedField To get a item with the fields you want
func (o OpCLI) GetItemWithCustomizedField(itemName string, fieldName ...string) ([]byte, error) {
	fieldNameString := strings.Join(fieldName, ",")
	var cmdArg []string
	var sOut []byte
	err := o.checkSessionAliveOrSignIn()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	cmdArg = append(cmdArg, "get", "item", itemName, "--session", o.session, "--fields", fieldNameString)
	sOut, err = execCommand("op", cmdArg)
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return sOut, ferr
	}
	log.Println("Got", itemName, "successfully")
	return sOut, nil
}

//if I set set ReadAll(stderr) before ReadAll(stdout) when the stdout
//output with a big size byte, this will freeze at ReadAll(stderr) ->netPoll:220

//GetListWithFlag you can use like GetListWithFlag([]string{"categories", "tags"}, []string{"Secure Note,Bank Account", "news,finance"})
func (o OpCLI) GetListWithFlag(flags []string, flagContents []string) ([]ItemRes, error) {
	var itemList []ItemRes
	err := o.checkSessionAliveOrSignIn()
	if err != nil {
		ferr := fmt.Errorf("%s", err)
		return itemList, ferr
	}
	var cmdArg []string
	cmdArg = append(cmdArg, "list", "items", "--session", o.session)
	cmdArg, err = o.addFlagsToCmdArg(cmdArg, flags, flagContents)
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
		ferr := fmt.Errorf("length of flags or flagContens parameta MUST be equal")
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

func (o *OpCLI) checkSessionAliveOrSignIn() error {
	if time.Now().After(o.expirationTime) {
		if o.keepSessionAlive {
			log.Println("Session was invalidated, will automatically signIn again")
			err := o.signInExec()
			if err != nil {
				return err
			}
			return nil
		}
		log.Println("Session was invalidated, need manually signIn")
	}
	return nil
}

//GetItemFromChannel a helper func for GetUsernameAndPassword
func GetItemFromChannel(cli *OpCLI, itemName string) (<-chan ItemLitelyRes, <-chan error) {
	ch := make(chan ItemLitelyRes)
	errChan := make(chan error)
	go func() {
		result, err := cli.GetUsernameAndPassword(itemName)
		errChan <- err
		ch <- result
	}()
	return ch, errChan
}
